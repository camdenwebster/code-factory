package main

import (
	"regexp"
	"strings"

	"github.com/camdenwebster/code-factory/internal/core"
)

var reTest = regexp.MustCompile(`(swift\s+test|swift\s+build|xcodebuild)`)

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
		if strings.Contains(fp, "/docs/") || strings.HasPrefix(fp, "docs/") {
			return core.ClassWriteDocs
		}
		return core.ClassMutate
	case "Bash", "shell":
		if reTest.MatchString(commandString(input)) {
			return core.ClassTest
		}
		return core.ClassShell
	}
	if strings.HasPrefix(name, "mcp__") {
		return core.ClassRead
	}
	return core.ClassOther
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
