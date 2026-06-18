---
description: Review the change — P1/P2/P3 with confidence gating
allowed-tools: Read, Grep, Glob, Bash(swift test:*), Bash(xcodebuild:*), Bash(compound:*)
---
You are in the **codeReview** phase. You may read and run tests, but not edit.

!`compound init --phase codeReview --if-absent >/dev/null; compound advance --to codeReview >/dev/null 2>&1; compound state`

Review the diff and produce findings, each with a priority and a confidence:
- **p1** = must fix (blocks compounding), **p2** = should, **p3** = nice.
- Findings below 0.60 confidence are suppressed; a surviving p1 sends the change
  back to work/debug (and Evidence is reset so the fix is re-verified).

The Stop hook runs the verification gate itself and records Evidence — you do not
need to assert the tests pass.

When the review is complete, write `.compound/result.json`:

```json
{"review":{"findings":[{"priority":"p1","confidence":0.8,"reviewer":"ce-swift-ios","message":"..."}]}}
```

No surviving p1 → the machine advances to **compound**.
