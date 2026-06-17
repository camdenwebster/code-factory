# `compound` — CLI specification

The `compound` binary is the executable form of `CompoundCore` + `CompoundKit`.
It is the **single source of truth** for the state machine: both integration
models call the same commands.

- **Model B (harness plugin):** the hook scripts shell out to `compound`
  (`rehydrate`, `policy`, `verify`, `advance`, `audit`).
- **Model A (headless):** `compound run` *is* the `Orchestrator.execute` loop;
  it drives `claude -p` / `codex exec` per phase.

The bash hook scripts ship a faithful fallback so the examples run without the
binary, but when `compound` is on `PATH` the scripts delegate to it — there is
exactly one policy/transition implementation.

---

## Conventions

**Global flags** (accepted by every subcommand):

| Flag | Default | Meaning |
|---|---|---|
| `--dir <path>` | `$PWD` | Repo root |
| `--state <path>` | `<dir>/.compound/state.json` | Checkpoint location |
| `--json` | off | Emit machine-readable JSON instead of human text |
| `--quiet` | off | Suppress non-essential stderr |

**Exit codes** (stable contract the hooks depend on):

| Code | Meaning |
|---|---|
| `0` | success / allow / tests passed |
| `1` | generic failure / tests failed |
| `2` | invalid usage or corrupt state |
| `10` | policy **deny** (reason on stdout) |
| `11` | policy **ask** (defer to user) |

---

## 0. Entry — triage front door

### `compound start`
Classify a request and seed the cycle in the right phase, instead of always
starting at `brainstorm`.
```
compound start --task "<request>" [--as CATEGORY] [--scheme NAME] [--force]
```
- Triage maps the task to a `CATEGORY` and a seed `(phase, track)`:
  `feature`→brainstorm · `bug`→brainstorm (bug track) · `chore`→plan ·
  `strategy`→strategy · `ideate`→ideate · `refresh`→compoundRefresh ·
  `pulse`→productPulse.
- `--as` overrides the heuristic. Refuses if a cycle is already in progress
  unless `--force`. `--json` prints the `TriageResult`.
- The classifier is a deterministic default; brainstorm's `routeToDebug` remains
  the in-loop safety net for a misclassified bug.

## 1. State & inspection

### `compound init`
Create a fresh checkpoint (lower-level than `start`; no triage).
```
compound init [--phase brainstorm] [--track knowledge] [--scheme NAME]
              [--force] [--if-absent]
```
- `--scheme` set ⇒ verification uses `xcodebuild`; unset ⇒ `swift test` (SwiftPM).
- `--force` overwrites an existing checkpoint; `--if-absent` makes it a no-op
  (exit 0) when one already exists — used by the `/ce-<phase>` commands to seed
  state on a fresh repo without clobbering a running cycle.

### `compound phase`
Print the current phase string (e.g. `plan`). Exit `0`.

### `compound state`
Dump the full `MachineState`. `--json` for the Codable encoding; otherwise a
human summary (phase, track, attempts, refs, last review).

### `compound rehydrate`
Print the `SessionStart` `additionalContext` string — "you are in phase X,
plan=Y, honor the tool policy." Used by `session-start.sh`. Exit `0`.

---

## 2. Policy — the PreToolUse firewall

### `compound policy`
Decide whether a tool is allowed in the current (or given) phase. Mirror of
`ToolPolicy.allowed(for:)`.
```
compound policy --tool <NAME> [--phase <P>] [--input-json '<tool_input>']
```
- `--tool` is the **harness** tool name (`Edit`, `Bash`, `apply_patch`,
  `shell`, `mcp__server__tool`, …); `compound` classifies it internally.
- `--input-json` lets the classifier inspect arguments (e.g. a `Bash` command
  that is really `swift test` ⇒ class `test`; a `Write` under `docs/` ⇒
  `write_docs`).
- **Output:** exit `0` allow · exit `10` deny (human reason on stdout) · exit
  `11` ask. With `--json`, prints `{"decision":"allow|deny|ask","class":"…","reason":"…"}`.

---

## 3. Verification — the Stop gate

### `compound verify`
Run the test suite **itself**, hash the output into an `Evidence` record, and
write it to state. Never trusts a self-reported pass.
```
compound verify [--scheme NAME] [--advance]
```
- Picks `xcodebuild test -scheme … -destination 'platform=iOS Simulator,…'`
  when a scheme is set, else `swift test`.
- Exit `0` pass / `1` fail. `--json` prints the `Evidence`
  (`testOutputHash`, `exitCode`, `toolchain`, `timestamp`).
- `--advance` also applies the resulting `verificationPassed` /
  `verificationFailed` event via the transition (convenience for the Stop hook).

---

## 4. Transitions — advancing the machine

### `compound advance`
Apply one `Transition.next(state, event)` and persist. Two forms:

**Explicit event** (low-level):
```
compound advance --event <PhaseEvent>
       [--ref <path> --kind <plan|diff|solution|…>]
       [--confidence <0..1>] [--track bug|knowledge] [--reason <text>]
```
`--event` is a `PhaseEvent` case: `requirementsWritten`, `planWritten`,
`planNeedsDeepening`, `codeWritten`, `reviewApproved`, `reviewBlocked`,
`compoundCaptured`, `routeToDebug`, `refreshRequested`, `pulseRequested`, …

**From a structured result** (preferred — runs the orchestrator's `interpret`):
```
compound advance --from-result <FILE>
```
Reads a `StructuredResult` the agent emitted at end of phase and maps it to the
correct event for the current phase (table below), then transitions. This is
the same `interpret()` used by Model A, so both models share the mapping.

Prints the new phase (or `--json` the new `MachineState`). Illegal transitions
are refused and recorded in `rejectedTransitions` (exit `0`, phase unchanged) —
the machine never silently advances.

**`StructuredResult` schema** (`.compound/result.json`):
```json
{
  "phase": "plan",
  "producedArtifact": { "path": "docs/plans/retry.md", "kind": "plan" },
  "confidence": 0.82,
  "detectedTrack": "bug",
  "review": { "findings": [
    { "priority": "p1", "confidence": 0.71, "reviewer": "ce-swift-ios", "message": "…" }
  ]}
}
```

**`interpret` mapping (phase → event):**

| Current phase | Result signal | Emitted event |
|---|---|---|
| `brainstorm` | `detectedTrack == bug` | `routeToDebug` |
| `brainstorm` | else | `requirementsWritten` |
| `plan` | `confidence` | `planWritten(confidence)` |
| `work` / `debug` | artifact | `codeWritten` |
| `codeReview` | review has blocking P1 | `reviewBlocked` |
| `codeReview` | else | `reviewApproved` |
| `compound` | artifact | `compoundCaptured` |
| `compoundRefresh` | — | `refreshCompleted` |
| `productPulse` | artifact | `pulseWritten` |

---

## 5. Knowledge store — capture & retrieval

### `compound search` — retrieval (called in `plan` / `debug`, local-first)
```
compound search --tags a,b --terms "token refresh","logout" [--limit 10]
```
Scans `docs/solutions/**` for tag/term overlap **before** any web research.
Prints matching repo-relative paths (`--json` for `[{path, score}]`).

### `compound capture` — the `compound` step
```
compound capture --from <doc.json|->  [--filename NAME]
```
Validates a `SolutionDoc` against the two-track schema (bug track requires 1–5
symptoms + root_cause + resolution_type), renders YAML-frontmatter Markdown, and
writes it to `docs/solutions/<category>/`. Prints the resulting `ArtifactRef`.
Schema violations exit `1` with the failing rule.

---

## 6. Audit

### `compound audit`
```
compound audit --event <label> [--tool <name>]
```
Append one `AuditEntry` `(phase, event, at)` to the audit stream. Called by
`post-tool-use.sh`.

### `compound log`
```
compound log [--tail N] [--phase P]
```
Read back the audit stream — the compliance trail and eval dataset.

---

## 7. Headless driver (Model A)

### `compound run` — the orchestrator loop
```
compound run --task "<description>"
             [--runner claude|codex|mock]
             [--scheme NAME] [--max-attempts 3] [--from-phase brainstorm]
```
Runs `Orchestrator.execute`: for each phase it builds a `WorkUnit` (scoped
tools), invokes the runner, interprets the result, runs the verification gate at
`codeReview`, applies the transition, and checkpoints — until `done`/`failed`.

- `--runner claude` ⇒ `AgentRunner` shells `claude -p … --allowedTools …`.
- `--runner codex` ⇒ shells `codex exec --sandbox <per-phase> --json …`.
- `--runner mock` ⇒ scripted results (for the invariant tests / demos).

Exit `0` if it reaches `done`, `1` if `failed`. `--json` prints the final state.

### `compound prompt` — render a phase prompt
```
compound prompt --phase <P> [--task "<text>"]
```
Prints the phase's instruction template (SKILL.md-equivalent) plus its allowed
tools. Used both inside `run` and for manual headless calls:
```bash
codex exec --sandbox read-only --json "$(compound prompt --phase plan --task "$T")"
```

---

## Who calls what

| Caller | Commands |
|---|---|
| `session-start.sh` | `rehydrate` (fallback: read state) |
| `pre-tool-use.sh` | `policy` |
| `stop.sh` | `verify`, then `advance --from-result` |
| `post-tool-use.sh` | `audit` |
| `/ce-*` slash commands | `prompt`, `search`, `capture` |
| Model A driver | `run` (which internally uses all of the above) |
