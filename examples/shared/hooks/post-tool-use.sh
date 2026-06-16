#!/usr/bin/env bash
# PostToolUse — append to the append-only audit stream. Every mutating tool
# call becomes one AuditEntry: this doubles as the compliance trail and the
# eval dataset (same artifact), mirroring MachineState.auditLog.
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
read_input

phase="$(current_phase)"
tool="$(jqr '.tool_name')"
ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
mkdir -p "$STATE_DIR"
jq -nc --arg p "$phase" --arg t "$tool" --arg at "$ts" \
  '{phase:$p,event:$t,at:$at}' >> "$STATE_DIR/audit.log"
exit 0
