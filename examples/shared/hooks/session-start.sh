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
  mkdir -p "$STATE_DIR"
  printf '%s\n' '{"phase":"brainstorm","track":"knowledge","attempts":0,"deepenCount":0,"scheme":null}' > "$STATE_FILE"
  ctx="CompoundEngine initialized at phase=brainstorm. This repo follows the Compounding Engineering loop: brainstorm → plan → work → codeReview → compound. Read-only research until /ce-work; tests must pass before any turn in work/codeReview can end."
fi
emit_session_context "$ctx"
