package runner

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/camdenwebster/code-factory/internal/core"
)

func TestCLIRunnerClaudeScopesReadOnlyPhase(t *testing.T) {
	dir := t.TempDir()
	var gotName string
	var gotArgs []string
	r := CLIRunner{Kind: "claude", StateDir: dir, Task: "t", Exec: func(d, name string, args []string, stdin string) (string, int, error) {
		gotName, gotArgs = name, args
		os.WriteFile(filepath.Join(dir, "result.json"),
			[]byte(`{"producedArtifact":{"path":"docs/plans/x.md","kind":"plan"},"confidence":0.8}`), 0o644)
		return "ok", 0, nil
	}}
	res, err := r.Run(WorkUnit{Phase: core.PhasePlan, WorkingDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if gotName != "claude" {
		t.Fatalf("want claude, got %s", gotName)
	}
	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, "--allowedTools") || strings.Contains(joined, "Edit") {
		t.Fatalf("plan must be read-only (no Edit), got: %s", joined)
	}
	if res.Structured == nil || res.Structured.Confidence != 0.8 {
		t.Fatal("structured result not parsed from result.json")
	}
	if res.Structured.Phase != core.PhasePlan {
		t.Fatalf("phase should be stamped, got %s", res.Structured.Phase)
	}
}

func TestCLIRunnerClaudeUnlocksWork(t *testing.T) {
	dir := t.TempDir()
	var gotArgs []string
	r := CLIRunner{Kind: "claude", StateDir: dir, Exec: func(d, name string, args []string, stdin string) (string, int, error) {
		gotArgs = args
		return "", 0, nil
	}}
	r.Run(WorkUnit{Phase: core.PhaseWork, WorkingDir: dir}) //nolint:errcheck
	if !strings.Contains(strings.Join(gotArgs, " "), "Edit") {
		t.Fatalf("work must permit Edit, got: %v", gotArgs)
	}
}

func TestClaudeToolsCodeReviewIsTestScoped(t *testing.T) {
	tools := claudeTools(core.PhaseCodeReview)
	if slices.Contains(tools, "Bash") {
		t.Fatalf("codeReview must not grant bare Bash (allows arbitrary mutation): %v", tools)
	}
	if !slices.Contains(tools, "Bash(swift test:*)") {
		t.Fatalf("codeReview should allow command-scoped test runs: %v", tools)
	}
}

func TestClaudeToolsWorkHasFullBash(t *testing.T) {
	if !slices.Contains(claudeTools(core.PhaseWork), "Bash") {
		t.Fatal("work should grant full Bash")
	}
}

func TestCLIRunnerCodexSandboxPerPhase(t *testing.T) {
	dir := t.TempDir()
	var gotArgs []string
	r := CLIRunner{Kind: "codex", StateDir: dir, Exec: func(d, name string, args []string, stdin string) (string, int, error) {
		gotArgs = args
		return "", 0, nil
	}}
	r.Run(WorkUnit{Phase: core.PhaseWork, WorkingDir: dir}) //nolint:errcheck
	if !slices.Contains(gotArgs, "workspace-write") {
		t.Fatalf("work should be workspace-write, got %v", gotArgs)
	}
	r.Run(WorkUnit{Phase: core.PhasePlan, WorkingDir: dir}) //nolint:errcheck
	if !slices.Contains(gotArgs, "read-only") {
		t.Fatalf("plan should be read-only, got %v", gotArgs)
	}
}

func TestCLIRunnerNonZeroExitErrors(t *testing.T) {
	dir := t.TempDir()
	r := CLIRunner{Kind: "claude", StateDir: dir, Exec: func(d, name string, args []string, stdin string) (string, int, error) {
		return "boom", 1, nil
	}}
	if _, err := r.Run(WorkUnit{Phase: core.PhaseWork, WorkingDir: dir}); err == nil {
		t.Fatal("a non-zero agent exit must surface as an error")
	}
}
