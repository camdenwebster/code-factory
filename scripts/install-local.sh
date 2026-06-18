#!/usr/bin/env bash
# install-local.sh — build + test the `compound` CLI, then install the
# CompoundEngine slash commands and lifecycle hooks into ~/.claude so you can
# drive the loop from a real Claude Code session.
#
# Usage:
#   scripts/install-local.sh              # test, build, install
#   scripts/install-local.sh --skip-tests # skip `go test` (faster iteration)
#   scripts/install-local.sh --uninstall  # remove the installed bits
#
# Overridable via env: BIN_DIR (default ~/.local/bin), CLAUDE_DIR (~/.claude).
set -euo pipefail

# --- locate the repo root (nearest ancestor with go.mod) ------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$SCRIPT_DIR"
while [[ "$REPO_ROOT" != "/" && ! -f "$REPO_ROOT/go.mod" ]]; do
  REPO_ROOT="$(dirname "$REPO_ROOT")"
done

CLAUDE_DIR="${CLAUDE_DIR:-$HOME/.claude}"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
HOOKS_DST="$CLAUDE_DIR/hooks/compound"
CMDS_DST="$CLAUDE_DIR/commands"
SETTINGS="$CLAUDE_DIR/settings.json"

say()  { printf '\033[1;36m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33mwarning:\033[0m %s\n' "$*" >&2; }
die()  { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

# --- flags ----------------------------------------------------------------
SKIP_TESTS=0
case "${1:-}" in
  --uninstall)
    say "Uninstalling"
    rm -rf "$HOOKS_DST"
    rm -f "$CMDS_DST"/ce-*.md "$BIN_DIR/compound"
    warn "Removed the binary, hooks, and /ce-* commands."
    warn "settings.json was left as-is — restore its hooks from a $SETTINGS.bak.* file if needed."
    exit 0 ;;
  --skip-tests) SKIP_TESTS=1 ;;
  "" ) ;;
  * ) die "unknown argument: $1 (use --skip-tests or --uninstall)" ;;
esac

[[ -f "$REPO_ROOT/go.mod" ]] || die "could not find go.mod above $SCRIPT_DIR"
command -v go >/dev/null || die "the Go toolchain is not on PATH"
command -v jq >/dev/null || die "jq is required (the hooks use it at runtime too)"

cd "$REPO_ROOT"

# --- 1. test --------------------------------------------------------------
if [[ "$SKIP_TESTS" -eq 0 ]]; then
  say "Running tests (go test ./...)"
  go test ./... || die "tests failed — not installing"
else
  warn "skipping tests (--skip-tests)"
fi

# --- 2. build -------------------------------------------------------------
say "Building compound -> $BIN_DIR/compound"
mkdir -p "$BIN_DIR"
go build -o "$BIN_DIR/compound" ./cmd/compound
say "Built $("$BIN_DIR/compound" version)"

# --- 3. install hooks (lib.sh + the four event scripts) -------------------
say "Installing hooks -> $HOOKS_DST"
mkdir -p "$HOOKS_DST"
cp "$REPO_ROOT"/examples/shared/hooks/*.sh "$HOOKS_DST"/
chmod +x "$HOOKS_DST"/*.sh

# --- 4. install /ce-* slash commands --------------------------------------
say "Installing /ce-* commands -> $CMDS_DST"
mkdir -p "$CMDS_DST"
cp "$REPO_ROOT"/examples/claude-code/.claude/commands/ce-*.md "$CMDS_DST"/

# --- 5. wire hooks into settings.json (absolute paths to installed hooks) --
say "Wiring hooks into $SETTINGS"
HOOKS_JSON="$(cat <<JSON
{
  "SessionStart": [{"matcher":"startup|resume|clear|compact","hooks":[{"type":"command","command":"$HOOKS_DST/session-start.sh","timeout":15}]}],
  "PreToolUse":   [{"matcher":"*","hooks":[{"type":"command","command":"$HOOKS_DST/pre-tool-use.sh","timeout":10}]}],
  "PostToolUse":  [{"matcher":"Edit|Write|MultiEdit|NotebookEdit|Bash","hooks":[{"type":"command","command":"$HOOKS_DST/post-tool-use.sh","async":true}]}],
  "Stop":         [{"hooks":[{"type":"command","command":"$HOOKS_DST/stop.sh","timeout":600}]}]
}
JSON
)"

mkdir -p "$CLAUDE_DIR"
if [[ -f "$SETTINGS" ]]; then
  backup="$SETTINGS.bak.$(date +%Y%m%d%H%M%S)"
  cp "$SETTINGS" "$backup"
  say "Backed up existing settings -> $backup"
else
  echo '{}' > "$SETTINGS"
fi
# Replace only the .hooks key; all other settings are preserved.
tmp="$(mktemp)"
jq --argjson h "$HOOKS_JSON" '.hooks = $h' "$SETTINGS" > "$tmp" && mv "$tmp" "$SETTINGS"

# --- 6. summary + PATH check ----------------------------------------------
say "Done."
cat <<EOF

Installed:
  binary    $BIN_DIR/compound
  hooks     $HOOKS_DST/
  commands  $CMDS_DST/   (/ce-start, /ce-brainstorm, /ce-plan, /ce-status, …)
  settings  $SETTINGS    (.hooks replaced; previous version backed up)
EOF

case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *)
    echo
    warn "$BIN_DIR is not on your PATH — Claude's shell won't find 'compound'."
    echo "      Add to your shell rc:  export PATH=\"$BIN_DIR:\$PATH\"" ;;
esac

cat <<EOF

Test it: open 'claude' in any repo and run
  /ce-start add a dark-mode toggle
  /ce-status
Then 'compound state' in that repo shows the cycle (state lives in ./.compound/).
EOF
