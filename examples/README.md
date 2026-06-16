# CompoundEngine — harness integration examples

Example lifecycle-hook configurations that wire the CompoundEngine state machine
into two coding agents: **Claude Code** and **Codex CLI**. Verified against the
current hook contracts of both tools (June 2026).

The key fact that makes this small: **Codex's hooks engine reuses Claude Code's
hook wire format** (same event names, same stdin-JSON in, same
stdout-JSON-or-exit-2 out, same `permissionDecision` / `additionalContext`
outputs). So the hook *scripts* are shared; only the config files differ.

```
examples/
├── shared/hooks/            # harness-agnostic scripts (BOTH tools call these)
│   ├── lib.sh               # policy table + tool classifier + emitters
│   ├── session-start.sh     # bootstrap / rehydrate MachineState, inject context
│   ├── pre-tool-use.sh      # tool-policy firewall (deny out-of-phase tools)
│   ├── stop.sh              # verification gate (refuse to stop on red tests)
│   └── post-tool-use.sh     # append-only audit stream
├── claude-code/
│   └── .claude/settings.json
└── codex/
    ├── config.toml          # inline [hooks] form + approval/sandbox
    ├── hooks.json           # equivalent standalone form (use one, not both)
    └── AGENTS.md            # static session context
```

## How the hooks map to the state machine

| Hook event | CompoundEngine role | Source-of-truth symbol |
|---|---|---|
| `SessionStart` | **Bootstrap/resume.** Load `MachineState` (or seed at `brainstorm`); inject "you are in phase X". Never the driver. | `CheckpointStore`, `MachineState` |
| `PreToolUse` | **Tool firewall.** Deny any tool not in the phase's allow-list, *before* it runs. | `ToolPolicy.allowed(for:)` |
| `Stop` | **Verification gate.** Run the tests ourselves; block the turn from ending if red. | `VerificationGate.verify`, `Evidence` |
| `PostToolUse` | **Audit.** Append `(phase, tool, ts)` — compliance trail + eval dataset. | `MachineState.auditLog` |

The per-phase tool policy the firewall enforces (mirror of the Swift enum):

| Phase | Allowed |
|---|---|
| `strategy` `ideate` `brainstorm` `plan` | read / search only (WHAT, not HOW) |
| `work` `debug` | read, edit, shell, build/test |
| `codeReview` | read + build/test only (no edits) |
| `compound` `compoundRefresh` `productPulse` | read + writes under `docs/` |
| `done` `failed` | nothing |

## The `compound` CLI contract

The scripts are self-contained (pure bash + `jq`) so they run **without** the
Swift core. In production each delegates to the `compound` binary — the
`CompoundCore` build from the implementation plan — when it is on `PATH`. That
binary is the single source of truth; the bash is a faithful fallback, not a
second policy implementation. Expected subcommands:

```
compound rehydrate                         # → additionalContext string (SessionStart)
compound phase                             # → current phase
compound policy --phase P --tool T --input-json '{…}'
                                           #   exit 0 allow / exit 10 deny(+reason on stdout)
compound verify [--scheme S]               # runs swift test / xcodebuild, hashes → Evidence
                                           #   exit 0 pass / non-zero fail
compound advance --event E [--ref R] [--confidence X]   # applies Transition.next, saves checkpoint
compound audit --event E --tool T          # append AuditEntry
```

State lives at `.compound/state.json` (override with `COMPOUND_STATE_DIR`).

## Two ways to run it

**Model B — interactive (the daily driver).** Developer is in a normal `claude`
or `codex` session; the hooks above enforce the invariants while they drive with
`/ce-*` phase intents. `SessionStart` resumes after a crash.

**Model A — headless (CI / unattended).** Phase = process boundary. One agent
invocation per phase, with the sandbox/tools matching the phase. This is also
the cleanest way to get a hard edit-firewall on Codex (see the gap below):

```bash
# Claude Code, plan phase: research only
claude -p "$(compound prompt --phase plan)" \
  --allowedTools "Read,Grep,Glob" --output-format stream-json

# Codex, plan phase: sandbox enforces read-only regardless of tool coverage
codex exec --sandbox read-only --json "$(compound prompt --phase plan)"

# Codex, work phase: writes allowed
codex exec --sandbox workspace-write --json "$(compound prompt --phase work)"
```

## Per-harness notes & the one real gap

**Claude Code**
- `${CLAUDE_PROJECT_DIR}` is expanded to the repo root — paths work as written.
- `PreToolUse` covers every tool (Edit/Write/Bash/MCP), so the firewall is complete in-session.
- Settings precedence: managed > local > project > user; commit the project file.

**Codex CLI**
- Hooks are **stable and on by default**; the engine is literally Claude-compatible.
- Replace `/ABS/PATH/TO/code-factory` — Codex does **not** expand `${CLAUDE_PROJECT_DIR}`.
- Use **either** the `[hooks]` table in `config.toml` **or** `hooks.json`, not both in one layer.
- **The gap:** `PreToolUse` fires for the `shell` tool but **not for `apply_patch`**
  (file edits) yet — `openai/codex#16732`. So the hook alone can't block edits during
  read-only phases. Two mitigations, both used above:
  1. `sandbox_mode = "read-only"` (or per-phase `codex exec --sandbox read-only`)
     blocks writes at the OS level, independent of tool coverage.
  2. `approval_policy` as a backstop.
- First-run **hook trust**: newly discovered hooks may prompt for trust unless
  set via managed config; account for it in automation.
- `notify` is **not** a gate — it fires once (turn complete), detached, and
  cannot block. Use the `Stop` hook for verification, not `notify`.

## Trying it locally

```bash
chmod +x examples/shared/hooks/*.sh
# Simulate a PreToolUse event in the plan phase trying to edit a file:
echo '{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"src/x.swift"}}' \
  | COMPOUND_STATE_DIR=/tmp/ce examples/shared/hooks/pre-tool-use.sh
# → permissionDecision: deny  (plan is read-only)
```

Requires `jq`. The `compound` binary is optional for the demo and required in
production.
