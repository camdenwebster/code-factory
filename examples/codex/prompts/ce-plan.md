You are in the **plan** phase (read-only sandbox).

First run `compound search --terms "$ARGUMENTS"` — local solutions beat web
research. Then read the brainstorm and write an implementation plan to
`docs/plans/<slug>.md` with stable unit IDs and test scenarios.

Assess your confidence (0.0–1.0). Below 0.70 the machine loops you back to
deepen; do real research rather than inflating it.

When complete, write `.compound/result.json`:
`{"producedArtifact":{"path":"docs/plans/<slug>.md","kind":"plan"},"confidence":0.85}`
