#!/usr/bin/env bash
# PreToolUse — the tool-policy firewall. Denies any tool not permitted in the
# current phase, BEFORE it runs (stronger than the orchestrator's after-the-fact
# check). This is `toolPolicyViolated` enforced up front.
#
# Codex caveat: PreToolUse fires for the `shell` tool but NOT yet for
# `apply_patch` (openai/codex#16732). On Codex, enforce read-only phases at the
# sandbox level too (see examples/codex/config.toml) — don't rely on this hook
# alone to block file edits.
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
read_input

phase="$(current_phase)"
tool="$(jqr '.tool_name')"
cls="$(classify_tool "$tool")"

if have_compound; then
  # Delegate to the single source of truth; exit 10 == deny(+reason on stdout).
  if reason="$(compound policy --phase "$phase" --tool "$tool" --input-json "$(jqr '.tool_input')")"; then
    exit 0
  else
    [[ $? -eq 10 ]] && emit_deny "${reason:-blocked by ToolPolicy}"
    exit 0
  fi
fi

if policy_allows "$phase" "$cls"; then
  exit 0
else
  emit_deny "CompoundEngine: phase '$phase' forbids $tool ($cls). Allowed here: $(allowed_summary "$phase"). Finish this phase or run /ce-<next> to advance — do not skip stages."
fi
