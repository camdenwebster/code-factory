# CompoundEngine — Agent Instructions (Codex)

This is the **static** context analog of the Claude Code `SessionStart`
`additionalContext` injection. Codex concatenates every `AGENTS.md` from the
repo root down to the cwd at session start. Put this at the repo root (or in
`~/.codex/AGENTS.md` for a global default). For **dynamic** per-session state
(current phase, plan ref), the `SessionStart` hook in `config.toml` injects it
live — this file is the always-true preamble.

## The loop

This repo follows the Compounding Engineering loop. Every change moves through:

```
brainstorm → plan → work → codeReview → compound
```

(bugs take the `debug` path instead of `work`; both converge on `codeReview`.)

## Rules the harness enforces (do not fight them)

- **Phases gate tools.** `brainstorm`/`plan` are read-only — *what*, not *how*.
  `work`/`debug` may edit and run tests. `codeReview` may run tests but not
  edit. `compound` writes only under `docs/`. The `PreToolUse` hook (and, for
  file edits, the sandbox) will block out-of-phase actions.
- **Tests are the gate, not your word.** A turn in `work`/`debug`/`codeReview`
  cannot end until the suite passes; the `Stop` hook runs it for you.
- **Search before you research.** In `plan`/`debug`, scan `docs/solutions/`
  first — prior solutions are why this compounds.
- **Capture in `compound`.** Write a `SolutionDoc` to `docs/solutions/<category>/`
  with the required frontmatter (bug track: 1–5 symptoms + root_cause +
  resolution_type).

## Confidence

If a plan's confidence is below threshold, deepen it (bounded) rather than
proceeding — the state machine will route you back to `plan`.
