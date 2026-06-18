Triage this request and start the cycle: $ARGUMENTS

1. Run `compound triage --task "$ARGUMENTS"` to see the heuristic proposal.
2. Decide the category yourself — confirm or override it. The keyword heuristic
   is often right but misses nuance (e.g. "fix up the onboarding copy" is a
   feature/chore, not a bug). Pick one of:
   feature · bug · chore · strategy · ideate · refresh · pulse.
3. Run `compound start --task "$ARGUMENTS" --as <your-category> --force`, then
   `compound rehydrate`.

Proceed with the indicated phase (the sandbox/hooks enforce its tool policy).
When complete, write `.compound/result.json` so the Stop hook advances you.
