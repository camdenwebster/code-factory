---
description: Start a Compounding Engineering cycle — capture WHAT, not HOW
argument-hint: <what to build or fix>
allowed-tools: Read, Grep, Glob, Bash(compound:*)
---
Begin a new cycle for: **$ARGUMENTS**

Initialize the machine and show state:
!`compound init --force >/dev/null && compound state`

You are in the **brainstorm** phase. Rules:
- Read-only — no code, no plans. Capture requirements (WHAT, not HOW).
- Write the brief to `docs/brainstorms/<slug>.md`.
- If this is really a bug to fix rather than a feature, set `detectedTrack: "bug"`.

When (and only when) the brief is complete, write `.compound/result.json`:

```json
{"producedArtifact":{"path":"docs/brainstorms/<slug>.md","kind":"requirements"},"detectedTrack":"knowledge"}
```

The Stop hook will advance you to **plan** (or **debug** routing for bugs).
