# code-factory · CompoundEngine

A harness-agnostic state machine for **Compounding Engineering**:
`brainstorm → plan → work → codeReview → compound`. One pure core, two ways to
run it — an unattended headless orchestrator and an interactive harness plugin
(Claude Code / Codex CLI).

The orchestrator is the **`compound`** CLI, written in Go (see
[`docs/solutions/tooling-decisions/cli-language-go.md`](docs/solutions/tooling-decisions/cli-language-go.md)
for why Go over Swift). It is the single source of truth for the state machine;
the harness hooks shell out to it, and `compound run` is the headless driver.

## Layout

```
cmd/compound/            # the CLI (dispatch, commands, harness-tool classifier)
internal/
  core/                  # PURE state machine — Phase·Event·Transition·ToolPolicy (+ tests)
  store/                 # KnowledgeStore + SolutionDoc two-track schema (+ tests)
  verify/                # deterministic verification gate -> Evidence (+ tests)
  checkpoint/            # JSON crash-recovery for MachineState
  runner/                # AgentRunner seam + mock + interpret()
  orchestrator/          # Model A headless loop (+ end-to-end tests)
spec/                    # cli.md (CLI contract) · system-spec.html (visual spec)
examples/                # Claude Code + Codex hook configs and shared hook scripts
docs/solutions/          # the compounding substrate (captured learnings)
```

## Build & test

```bash
go build ./...                 # builds everything
go test ./...                  # runs the invariant + schema + gate + e2e suites
go build -o compound ./cmd/compound
```

No external dependencies — pure stdlib (`crypto/sha256`, `encoding/json`,
`os/exec`). Cross-compile with `GOOS=darwin GOARCH=arm64 go build ./cmd/compound`.

## How it runs

- **Model A — headless:** `compound run --task "…" --runner mock` is the
  `Orchestrator.execute` loop: scoped agent call per phase, verification gate at
  `codeReview`, transition, checkpoint, until terminal. (`--runner claude|codex`
  shell out per `spec/cli.md`.)
- **Model B — plugin:** the hooks in `examples/shared/hooks/` call
  `compound rehydrate` (SessionStart), `compound policy` (PreToolUse firewall),
  `compound verify` (Stop gate), and `compound advance` (transition). The same
  scripts drive both Claude Code and Codex, because Codex reuses Claude Code's
  hook wire format.

## What's proven by the tests

`go test ./...` asserts the determinism guarantees: can't reach `compound`
without passing `Evidence`; `deepen` and verification retries are bounded;
illegal transitions are refused and audited; bug/knowledge routing converges on
review; the `SolutionDoc` schema enforces the two-track constraints; and a full
mock run reaches `done` (or, with an unprovable gate, `failed` — never
`compound`).

## Status

Core + persistence + gate + headless driver are implemented and tested. The
live `claude`/`codex` runners and the `/ce-*` slash commands are the remaining
wiring; the contracts for both are in `spec/cli.md`.
