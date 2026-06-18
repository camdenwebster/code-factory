You are in the **debug** phase, bug track (workspace-write sandbox).

First run `compound search --terms "$ARGUMENTS"` for prior fixes. Write a failing
test that reproduces the bug, then make it pass. The Stop hook gates the turn on
green tests.

When fixed and green, write `.compound/result.json`:
`{"producedArtifact":{"path":"<diff-or-summary-ref>","kind":"diff"}}`
