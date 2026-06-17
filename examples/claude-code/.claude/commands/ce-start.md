---
description: Triage a request and start the cycle in the right phase
argument-hint: <what you want to do>
allowed-tools: Read, Grep, Glob, Bash(compound:*)
---
Triage the request and seed the machine (features start at brainstorm, bugs at
brainstorm on the bug track, chores at plan, and strategy/ideate/refresh/pulse
at their own phases):

!`compound start --task "$ARGUMENTS" --force`

Phase context for this cycle:
!`compound rehydrate`

Proceed with the phase indicated above, following its rules — the per-phase
firewall enforces them (read-only for brainstorm/plan; build/test-only for
review). When the phase is complete, write `.compound/result.json` so the Stop
hook advances you.

If the triage looks wrong, you don't need to restart: brainstorm's routing will
re-route a misclassified bug, or run `/ce-status` and then the specific
`/ce-<phase>` you want.
