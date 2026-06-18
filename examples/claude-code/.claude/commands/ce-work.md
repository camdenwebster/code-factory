---
description: Implement the plan (feature/knowledge track)
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---
You are in the **work** phase. Implement the plan in `docs/plans/`.

!`compound init --phase work --if-absent >/dev/null; compound advance --to work >/dev/null 2>&1; compound state`

Rules:
- Edit/write code and run the test suite as you go.
- The Stop hook will not let this turn end while tests are red — keep going
  until they pass.

When the implementation is complete and green, write `.compound/result.json`:

```json
{"producedArtifact":{"path":"<diff-or-summary-ref>","kind":"diff"}}
```

The Stop hook advances you to **codeReview**, where the gate records Evidence.
