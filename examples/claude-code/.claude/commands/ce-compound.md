---
description: Capture the learning — the defining step
allowed-tools: Read, Grep, Glob, Write, Bash(compound:*)
---
You are in the **compound** phase. Writes are limited to `docs/`.

!`compound state`

Capture what was learned as a SolutionDoc. Compose the JSON and write it via the
store (which validates the two-track schema and files it by category):

```bash
compound capture --from - <<'JSON'
{"module":"<m>","date":"<YYYY-MM-DD>","problemType":"<type>","component":"<c>",
 "severity":"medium","tags":["..."],
 "symptoms":["..."],"rootCause":"...","resolutionType":"...",
 "body":"## What happened\n..."}
JSON
```

(Bug track requires 1–5 symptoms + rootCause + resolutionType; knowledge track
uses `appliesWhen` instead.)

Then write `.compound/result.json` pointing at the captured file:

```json
{"producedArtifact":{"path":"docs/solutions/<category>/<slug>.md","kind":"solution"}}
```

The Stop hook advances the machine to **done**. The next cycle's `/ce-plan` will
retrieve this solution — that retrieval is the compounding.
