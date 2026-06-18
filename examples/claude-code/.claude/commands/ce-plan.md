---
description: Plan the change — confidence-gated, local-first research
argument-hint: <optional search terms>
allowed-tools: Read, Grep, Glob, Bash(compound:*)
---
You are in the **plan** phase (read-only).

!`compound init --phase plan --if-absent >/dev/null; compound advance --to plan >/dev/null 2>&1; compound state`

Search captured solutions FIRST (local-first retrieval beats web research):
!`compound search --terms "$ARGUMENTS"`

Current state:
!`compound state`

Rules:
- Read the brainstorm, then write an implementation plan to `docs/plans/<slug>.md`
  with stable unit IDs and concrete test scenarios.
- Assess your **confidence** (0.0–1.0) that this plan is correct and complete.
  Below the threshold (0.70) the machine loops you back to deepen — do real
  research, don't inflate the number.

When complete, write `.compound/result.json`:

```json
{"producedArtifact":{"path":"docs/plans/<slug>.md","kind":"plan"},"confidence":0.85}
```
