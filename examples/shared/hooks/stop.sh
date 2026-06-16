#!/usr/bin/env bash
# Stop — the deterministic verification gate. When the agent tries to end its
# turn during work/debug/codeReview, run the tests OURSELVES and refuse to let
# the turn end if they fail. This is the "never trust a self-reported pass"
# rule from VerificationGate, realized as a hook.
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
read_input

phase="$(current_phase)"
case "$phase" in
  work|debug|codeReview)
    if compound_verify "$(current_scheme)"; then
      exit 0   # green: allow the turn to end (machine advances via `compound advance`)
    else
      emit_stop_block "CompoundEngine verification gate FAILED in phase '$phase'. The tests must pass before this turn can end. Re-run the failing tests, fix the cause, and continue. A self-reported 'it works' is not accepted — the gate ran the suite itself."
    fi ;;
  *)
    exit 0 ;;
esac
