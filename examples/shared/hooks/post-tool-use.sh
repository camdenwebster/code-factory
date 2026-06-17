#!/usr/bin/env bash
# PostToolUse — append to the append-only audit stream. Every mutating tool
# call becomes one AuditEntry: this doubles as the compliance trail and the
# eval dataset (same artifact), mirroring MachineState.auditLog.
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
read_input

[[ -z "$(current_phase)" ]] && exit 0   # no active cycle => nothing to audit
tool="$(jqr '.tool_name')"
# Prefer the binary so the entry lands in MachineState.AuditLog — the same trail
# `compound log` and /ce-status read. Fall back to a side file only without it.
if have_compound; then
  compound audit --event "$tool" >/dev/null 2>&1 || true
else
  phase="$(current_phase)"
  ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  mkdir -p "$STATE_DIR"
  jq -nc --arg p "$phase" --arg t "$tool" --arg at "$ts" \
    '{phase:$p,event:$t,at:$at}' >> "$STATE_DIR/audit.log"
fi
exit 0
