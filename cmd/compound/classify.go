package main

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/camdenwebster/code-factory/internal/core"
)

// reTest matches a clean build/test invocation anchored at the start of the
// command (not merely containing the words somewhere).
var reTest = regexp.MustCompile(`^\s*(swift\s+test|swift\s+build|xcodebuild)(\s|$)`)

// classifyTool maps a harness tool name + its input JSON to an abstract
// ToolClass. It is the single source of truth the bash hook fallback mirrors,
// and it handles both Claude (Edit/Write/Bash) and Codex (apply_patch/shell)
// naming, plus both command shapes (string vs argv array).
func classifyTool(name string, input map[string]any) core.ToolClass {
	switch name {
	case "Read", "Grep", "Glob", "LS", "WebFetch", "WebSearch":
		return core.ClassRead
	case "NotebookEdit", "MultiEdit":
		return core.ClassMutate
	case "Edit", "Write", "apply_patch":
		fp := firstString(input, "file_path", "path")
		switch {
		case isControlPath(fp):
			return core.ClassControl // .compound/ — the engine's own files
		case isDocsPath(fp):
			return core.ClassWriteDocs
		default:
			return core.ClassMutate
		}
	case "Bash", "shell":
		if looksLikeTest(commandString(input)) {
			return core.ClassTest
		}
		return core.ClassShell
	}
	// Unknown tools (Skill, AskUserQuestion, Task, TodoWrite, mcp__*, …) are not
	// workspace mutations — the phase firewall governs edits/shell, not these.
	return core.ClassRead
}

// isControlPath reports whether a write targets the engine's .compound/ dir.
func isControlPath(fp string) bool {
	if fp == "" {
		return false
	}
	c := filepath.ToSlash(filepath.Clean(fp))
	return c == ".compound" || strings.HasPrefix(c, ".compound/") || strings.Contains(c, "/.compound/")
}

// isDocsPath reports whether a write target resolves under docs/ WITHOUT
// escaping it. The path is cleaned first, so "docs/../cmd/main.go" collapses to
// "cmd/main.go" and is correctly classified as a mutation, not a docs write.
func isDocsPath(fp string) bool {
	if fp == "" {
		return false
	}
	c := filepath.ToSlash(filepath.Clean(fp))
	if c == ".." || strings.HasPrefix(c, "../") {
		return false // escapes upward
	}
	return strings.HasPrefix(c, "docs/") || strings.Contains(c, "/docs/")
}

// looksLikeTest reports whether a shell command is a clean, single build/test
// invocation. A command that chains or substitutes (";", "&&", "|", "$(", ...)
// could smuggle an edit past a test-only phase, so it is NOT a test — it falls
// through to the more-restricted shell class. Fail closed.
func looksLikeTest(cmd string) bool {
	if strings.ContainsAny(cmd, ";&|<>\n\x60") || strings.Contains(cmd, "$(") {
		return false
	}
	return reTest.MatchString(cmd)
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// commandString extracts a shell command whether it's a string (Claude) or an
// argv array (Codex: ["bash","-lc","..."]).
func commandString(m map[string]any) string {
	switch c := m["command"].(type) {
	case string:
		return c
	case []any:
		parts := make([]string, 0, len(c))
		for _, p := range c {
			if s, ok := p.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, " ")
	}
	return ""
}
