---
description: Fix the bug (bug track)
argument-hint: <optional search terms>
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---
You are in the **debug** phase (bug track). Reproduce, then fix.

!`compound init --phase debug --track bug --if-absent >/dev/null; compound state`

Search prior fixes FIRST:
!`compound search --terms "$ARGUMENTS"`

Rules:
- Write a failing test that reproduces the bug, then make it pass.
- The Stop hook gates this turn on green tests.

When fixed and green, write `.compound/result.json`:

```json
{"producedArtifact":{"path":"<diff-or-summary-ref>","kind":"diff"}}
```
