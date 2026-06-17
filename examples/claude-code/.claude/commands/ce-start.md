---
description: Triage a request (with your confirmation) and start the cycle
argument-hint: <what you want to do>
allowed-tools: Read, Grep, Glob, Bash(compound:*)
---
Request: **$ARGUMENTS**

Heuristic proposal (a starting point, not a verdict):
!`compound triage --task "$ARGUMENTS"`

Now **you** decide the category — confirm the proposal or override it. Read the
request critically; the keyword heuristic is often right but misses nuance
(e.g. "fix up the onboarding copy" is a feature/chore, not a bug). Choose one of:
`feature · bug · chore · strategy · ideate · refresh · pulse`.

Then run these as Bash commands:
1. `compound start --task "$ARGUMENTS" --as <your-category> --force`
2. `compound rehydrate`

Proceed with the phase it reports, following that phase's rules (the per-phase
firewall enforces them). When the phase is complete, write `.compound/result.json`
so the Stop hook advances you.
