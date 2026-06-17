Triage this request and start the cycle in the right phase: $ARGUMENTS

Run `compound start --task "$ARGUMENTS" --force`, then `compound rehydrate` to
load the phase context. Features start at brainstorm, bugs at brainstorm on the
bug track, chores at plan, and strategy/ideate/refresh/pulse at their own phases.

Proceed with the indicated phase (the sandbox/hooks enforce its tool policy).
When complete, write `.compound/result.json` so the Stop hook advances you. If
the triage looks wrong, brainstorm's routing re-routes a misclassified bug, or
run `compound start --as <category> --force` to override.
