#!/usr/bin/env bash
# SessionStart — bootstrap / rehydrate the CompoundEngine machine.
# This is the RESUME seam, not the driver: it loads MachineState (or seeds it
# at brainstorm) and injects "you are in phase X" context for the model.
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
read_input

if have_compound; then
  emit_session_context "$(compound rehydrate)"
fi

if [[ -f "$STATE_FILE" ]]; then
  ctx="$(jq -r '
    "CompoundEngine resumed. phase=\(.phase) track=\(.track // "knowledge") " +
    "plan=\(.planRef // "none") attempts=\(.attempts // 0). " +
    "Honor the per-phase tool policy; advance only via /ce-<phase>."' "$STATE_FILE")"
else
  # No checkpoint => no active cycle. Do NOT seed a phase here; the triage front
  # door (/ce-start) is the only thing that begins a cycle.
  ctx="No CompoundEngine cycle in progress. This repo follows the Compounding Engineering loop (brainstorm → plan → work → codeReview → compound). Run /ce-start <request> to triage and begin — do not assume a phase until then."
fi
emit_session_context "$ctx"
