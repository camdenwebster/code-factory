#!/usr/bin/env bash
# Stop — the deterministic verification gate AND the phase-advance driver.
#
#  1. Verification gate: a turn in work/debug/codeReview cannot end on red
#     tests. The machine runs the suite itself; it never trusts a self-report.
#  2. Advance: if the phase produced a StructuredResult (.compound/result.json),
#     apply the transition and consume it. "result.json present" == "phase done".
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
read_input

phase="$(current_phase)"

# 1. Gate ------------------------------------------------------------------
case "$phase" in
  work|debug)
    # Block the turn on red tests, but record evidence only at codeReview.
    if have_compound; then
      compound verify >/dev/null 2>&1 \
        || emit_stop_block "CompoundEngine: tests are red in '$phase'. Fix them before ending the turn — the gate ran the suite itself."
    else
      compound_verify "$(current_scheme)" \
        || emit_stop_block "CompoundEngine: tests are red in '$phase'. Fix them before ending the turn."
    fi ;;
  codeReview)
    # Run the gate AND record Evidence (so reviewApproved can reach compound).
    if have_compound; then
      compound verify --advance >/dev/null 2>&1 \
        || emit_stop_block "CompoundEngine verification gate FAILED at codeReview. The fix must make the suite pass; re-run and continue."
    else
      compound_verify "$(current_scheme)" \
        || emit_stop_block "CompoundEngine verification gate FAILED at codeReview."
    fi ;;
esac

# 2. Advance ---------------------------------------------------------------
# Consume result.json ONLY on a successful transition. If advance fails (the
# result is malformed or the transition is illegal), keep the file and block so
# the model still has the artifact it needs to fix the problem.
RESULT="$STATE_DIR/result.json"
if [[ -f "$RESULT" ]]; then
  if have_compound; then
    if compound advance --from-result "$RESULT" >/dev/null 2>&1; then
      rm -f "$RESULT"
    else
      emit_stop_block "CompoundEngine could not apply the result for phase '$phase'. .compound/result.json is malformed or the transition is illegal — the file has been preserved. Fix it, then end the turn again."
    fi
  else
    rm -f "$RESULT"   # no binary present: nothing consumes the result
  fi
fi
exit 0
