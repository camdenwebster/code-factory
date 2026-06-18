You are in the **compound** phase (workspace-write, writes limited to `docs/`).

Capture the learning as a SolutionDoc via the store (it validates the two-track
schema and files it by category):

```bash
compound capture --from - <<'JSON'
{"module":"<m>","date":"<YYYY-MM-DD>","problemType":"<type>","component":"<c>",
 "severity":"medium","tags":["..."],
 "symptoms":["..."],"rootCause":"...","resolutionType":"...",
 "body":"## What happened\n..."}
JSON
```

(Bug track: 1–5 symptoms + rootCause + resolutionType. Knowledge track: appliesWhen.)

Then write `.compound/result.json`:
`{"producedArtifact":{"path":"docs/solutions/<category>/<slug>.md","kind":"solution"}}`
The next cycle's plan retrieves this — that retrieval is the compounding.
