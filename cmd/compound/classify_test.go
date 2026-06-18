package main

import (
	"testing"

	"github.com/camdenwebster/code-factory/internal/core"
)

func bashInput(cmd string) map[string]any  { return map[string]any{"command": cmd} }
func editInput(path string) map[string]any { return map[string]any{"file_path": path} }

func TestClassifyDocsPathTraversal(t *testing.T) {
	cases := []struct {
		path string
		want core.ToolClass
	}{
		{"docs/solutions/x.md", core.ClassWriteDocs},
		{"/repo/docs/solutions/x.md", core.ClassWriteDocs},
		{".compound/result.json", core.ClassControl}, // engine control file
		{"/repo/.compound/state.json", core.ClassControl},
		{"docs/../cmd/compound/main.go", core.ClassMutate}, // traversal must NOT be docs
		{"/repo/docs/../secret.go", core.ClassMutate},
		{"../docs/x.md", core.ClassMutate}, // escapes upward
		{"src/main.go", core.ClassMutate},
	}
	for _, c := range cases {
		if got := classifyTool("Edit", editInput(c.path)); got != c.want {
			t.Errorf("classify Edit %q = %s, want %s", c.path, got, c.want)
		}
	}
}

func TestClassifyCompoundCommandNotTest(t *testing.T) {
	cases := []struct {
		cmd  string
		want core.ToolClass
	}{
		{"swift test", core.ClassTest},
		{"swift test --filter Foo", core.ClassTest},
		{"xcodebuild test -scheme App", core.ClassTest},
		{"echo swift test; touch cmd/compound/main.go", core.ClassShell}, // chained -> not test
		{"swift build && rm -rf /", core.ClassShell},
		{"swift test | tee out", core.ClassShell},
		{"echo $(swift test)", core.ClassShell},
		{"sed -i s/a/b/ x.go", core.ClassShell},
	}
	for _, c := range cases {
		if got := classifyTool("Bash", bashInput(c.cmd)); got != c.want {
			t.Errorf("classify Bash %q = %s, want %s", c.cmd, got, c.want)
		}
	}
}

// Meta-tools must not be governed by the phase firewall (the deadlock bug:
// Skill/AskUserQuestion were classified as deny-everywhere).
func TestClassifyMetaToolsAreRead(t *testing.T) {
	for _, name := range []string{"AskUserQuestion", "Skill", "Task", "TodoWrite", "ExitPlanMode"} {
		if got := classifyTool(name, nil); got != core.ClassRead {
			t.Errorf("%s should classify as read (allowed), got %s", name, got)
		}
	}
}

// Codex argv-array shape must classify identically.
func TestClassifyCodexArrayCommand(t *testing.T) {
	in := map[string]any{"command": []any{"bash", "-lc", "echo swift test; touch x"}}
	if got := classifyTool("shell", in); got != core.ClassShell {
		t.Errorf("chained codex command = %s, want shell", got)
	}
}
