You are in the **work** phase (workspace-write sandbox). Implement the plan in
`docs/plans/`. Edit code and run the test suite as you go.

The Stop hook will not let the turn end while tests are red — keep going until
they pass.

When complete and green, write `.compound/result.json`:
`{"producedArtifact":{"path":"<diff-or-summary-ref>","kind":"diff"}}`
The Stop hook advances you to codeReview.
