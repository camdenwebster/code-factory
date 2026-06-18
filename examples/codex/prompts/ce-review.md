You are in the **codeReview** phase (read-only sandbox): read and run tests, but
do not edit. Review the diff and produce findings, each with a priority and
confidence:

- p1 = must fix (blocks compounding), p2 = should, p3 = nice.
- Findings below 0.60 confidence are suppressed; a surviving p1 sends the change
  back to work/debug with Evidence reset.

The Stop hook runs the verification gate and records Evidence itself.

When complete, write `.compound/result.json`:
`{"review":{"findings":[{"priority":"p1","confidence":0.8,"reviewer":"ce-swift-ios","message":"..."}]}}`
