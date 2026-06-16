---
description: Show the current CompoundEngine phase and audit trail
allowed-tools: Bash(compound:*)
---
Current machine state:
!`compound state`

Recent audit trail:
!`compound log --tail 15 2>/dev/null || echo "(no audit yet)"`

Allowed tools in this phase, and the next expected step, follow from the phase
above. Use `/ce-<phase>` to continue.
