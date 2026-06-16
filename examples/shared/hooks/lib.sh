#!/usr/bin/env bash
# ============================================================================
# CompoundEngine — shared lifecycle-hook helpers (harness-agnostic)
#
# Works for BOTH Claude Code and Codex CLI: Codex's hooks engine reuses Claude
# Code's wire format (stdin JSON in, stdout JSON or exit-code-2 to block out),
# so the same scripts drive both. The only per-harness differences live in the
# config files (.claude/settings.json vs ~/.codex/config.toml / hooks.json).
#
# These example scripts are SELF-CONTAINED (pure bash + jq) so you can run them
# without the Swift core. In production each one delegates to the `compound`
# CLI (the CompoundCore binary) when it is on PATH — see `compound_*` below.
# The `compound` CLI is the single source of truth for ToolPolicy + Transition;
# the bash fallbacks here are a faithful mirror, not a second implementation.
# ============================================================================
set -euo pipefail

# Repo-relative state. CLAUDE_PROJECT_DIR is set by Claude Code; Codex passes
# `cwd` on stdin, so we fall back to $PWD.
STATE_DIR="${COMPOUND_STATE_DIR:-${CLAUDE_PROJECT_DIR:-$PWD}/.compound}"
STATE_FILE="$STATE_DIR/state.json"
# Let the `compound` binary find the same checkpoint without --state on every call.
export COMPOUND_STATE_DIR="$STATE_DIR"

# Slurp stdin once; every hook event delivers a single JSON object.
read_input() { HOOK_INPUT="$(cat)"; }
jqr() { printf '%s' "${HOOK_INPUT:-}" | jq -r "$1" 2>/dev/null || true; }

current_phase() {
  [[ -f "$STATE_FILE" ]] && jq -r '.phase // "brainstorm"' "$STATE_FILE" || echo brainstorm
}
current_scheme() {
  [[ -f "$STATE_FILE" ]] || return 0
  jq -r '.scheme // empty' "$STATE_FILE" 2>/dev/null || true
}

# ---- Map a harness tool name + its input to an abstract policy class --------
# Classes mirror the Swift `Tool` groupings: read | shell | test | mutate |
# write_docs | other. Handles BOTH naming schemes (Claude: Edit/Write/Bash;
# Codex: apply_patch/shell) and both command shapes (string vs argv array).
classify_tool() {
  local name="$1"
  case "$name" in
    Read|Grep|Glob|LS|WebFetch|WebSearch) echo read ;;
    NotebookEdit|MultiEdit)               echo mutate ;;
    Edit|Write|apply_patch)
      local fp
      fp="$(jqr '.tool_input.file_path // .tool_input.path // .tool_input.input.path // ""')"
      # Reject any traversal; only un-escaped docs/ paths count as write_docs.
      if [[ "$fp" != *..* && ( "$fp" == docs/* || "$fp" == */docs/* ) ]]; then
        echo write_docs
      else echo mutate; fi ;;
    Bash|shell)
      local cmd
      # Claude: .tool_input.command is a string. Codex: ["bash","-lc","..."].
      cmd="$(jqr '.tool_input.command | if type=="array" then join(" ") else (.//"") end')"
      # Chained/substituting commands could smuggle an edit into a test phase;
      # treat them as plain shell (fail closed), only a clean test cmd is test.
      if [[ "$cmd" == *[\;\&\|\<\>\`]* || "$cmd" == *'$('* || "$cmd" == *$'\n'* ]]; then
        echo shell
      elif [[ "$cmd" =~ ^[[:space:]]*(swift[[:space:]]+test|swift[[:space:]]+build|xcodebuild)([[:space:]]|$) ]]; then
        echo test
      else echo shell; fi ;;
    mcp__*) echo read ;;   # tighten per-server in production
    *)      echo other ;;
  esac
}

# ---- The policy table: mirror of Swift `ToolPolicy.allowed(for:)` -----------
# Exit 0 = allow, non-zero = deny, for a given (phase, class).
policy_allows() {
  local phase="$1" cls="$2"
  case "$phase" in
    strategy|ideate|brainstorm|plan)        [[ "$cls" == read ]] ;;
    work|debug)                             [[ "$cls" =~ ^(read|shell|test|mutate|write_docs)$ ]] ;;
    codeReview)                             [[ "$cls" =~ ^(read|test)$ ]] ;;
    compound|compoundRefresh|productPulse)  [[ "$cls" =~ ^(read|write_docs)$ ]] ;;
    *)                                      return 1 ;;
  esac
}

allowed_summary() {
  case "$1" in
    strategy|ideate|brainstorm|plan)        echo "read/search only (WHAT, not HOW)" ;;
    work|debug)                             echo "read, edit, shell, build/test" ;;
    codeReview)                             echo "read + build/test only (no edits)" ;;
    compound|compoundRefresh|productPulse)  echo "read + writes under docs/ only" ;;
    *)                                      echo "none" ;;
  esac
}

# ---- Output emitters (identical wire shape for Claude Code + Codex) ---------
emit_deny() { # PreToolUse hard block
  jq -nc --arg r "$1" \
    '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:$r}}'
  exit 0
}
emit_session_context() { # SessionStart context injection
  jq -nc --arg c "$1" \
    '{hookSpecificOutput:{hookEventName:"SessionStart",additionalContext:$c}}'
  exit 0
}
emit_stop_block() { # Stop: refuse to end the turn, feed reason back to the model
  jq -nc --arg r "$1" \
    '{decision:"block",reason:$r,hookSpecificOutput:{hookEventName:"Stop",additionalContext:$r}}'
  exit 0
}

# ---- Bridges to the Swift `compound` CLI (used when present) ----------------
have_compound() { command -v compound >/dev/null 2>&1; }

# Deterministic verification gate. Returns 0 (pass) / non-zero (fail). Never
# trusts a self-reported pass: it runs the tests itself and (via the binary)
# hashes the output into an Evidence record.
compound_verify() {
  local scheme="$1"
  if have_compound; then
    compound verify ${scheme:+--scheme "$scheme"}; return $?
  fi
  # Fallback so the example runs without the binary:
  if [[ -n "$scheme" ]]; then
    xcodebuild test -scheme "$scheme" \
      -destination 'platform=iOS Simulator,name=iPhone 16' >/dev/null 2>&1
  elif [[ -f Package.swift ]]; then
    swift test >/dev/null 2>&1
  else
    return 0   # nothing to verify in this example checkout
  fi
}
