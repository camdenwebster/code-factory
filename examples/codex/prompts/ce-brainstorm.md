Begin a new Compounding Engineering cycle for: $ARGUMENTS

Run `compound init --force` then `compound state` to (re)start the machine.

You are in the **brainstorm** phase — read-only (the read-only sandbox enforces
this). Capture requirements (WHAT, not HOW) in `docs/brainstorms/<slug>.md`. If
this is really a bug, set `detectedTrack: "bug"`.

When complete, write `.compound/result.json`:
`{"producedArtifact":{"path":"docs/brainstorms/<slug>.md","kind":"requirements"},"detectedTrack":"knowledge"}`
The Stop hook advances you to plan (or routes bugs to debug).
