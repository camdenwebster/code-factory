package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/camdenwebster/code-factory/internal/core"
)

// ExecFunc runs a headless agent command with a prompt on stdin and returns
// combined output + exit code. Indirected so the runner is testable without a
// real claude/codex binary.
type ExecFunc func(dir, name string, args []string, stdin string) (stdout string, code int, err error)

// CLIRunner drives a headless coding agent (claude -p or codex exec) for one
// phase. It scopes the agent's tools/sandbox to the phase policy and reads back
// the StructuredResult the agent writes to <StateDir>/result.json.
type CLIRunner struct {
	Kind     string // "claude" | "codex"
	StateDir string // directory holding result.json
	Task     string
	Extra    []string // extra CLI args appended to the invocation
	Exec     ExecFunc
}

// NewCLIRunner builds a runner that shells out to the real binary.
func NewCLIRunner(kind, stateDir, task string) CLIRunner {
	return CLIRunner{Kind: kind, StateDir: stateDir, Task: task, Exec: execCmd}
}

func (r CLIRunner) Identifier() string { return r.Kind }

// Run dispatches one phase to the headless agent and parses its result.
func (r CLIRunner) Run(u WorkUnit) (Result, error) {
	if err := os.MkdirAll(r.StateDir, 0o755); err != nil {
		return Result{}, err
	}
	resultPath := filepath.Join(r.StateDir, "result.json")
	_ = os.Remove(resultPath) // clear any stale result before the run

	name, args := r.command(u.Phase)
	args = append(args, r.Extra...)
	out, code, err := r.Exec(u.WorkingDir, name, args, buildPrompt(u, r.Task, resultPath))
	if err != nil {
		return Result{}, err
	}
	res := Result{Raw: out}
	if code != 0 {
		return res, fmt.Errorf("%s exited %d", name, code)
	}
	// A missing/invalid result is not a Go error: the orchestrator interprets a
	// nil Structured as "phase produced nothing" and fails the transition.
	if sr, err := readResult(resultPath, u.Phase); err == nil {
		res.Structured = sr
	}
	return res, nil
}

// command builds the per-phase headless invocation, scoped to the policy.
func (r CLIRunner) command(p core.Phase) (string, []string) {
	switch r.Kind {
	case "codex":
		// Sandbox is the firewall here: Codex's PreToolUse does not cover
		// apply_patch, so read-only phases are enforced at the OS level.
		return "codex", []string{"exec", "--json", "--sandbox", codexSandbox(p)}
	default: // claude
		return "claude", []string{"-p", "--output-format", "stream-json",
			"--allowedTools", strings.Join(claudeTools(p), ",")}
	}
}

// claudeTools maps the phase's allowed classes to concrete Claude Code tools.
func claudeTools(p core.Phase) []string {
	set := map[string]bool{}
	for _, c := range core.AllowedClasses(p) {
		for _, t := range claudeToolsForClass[c] {
			set[t] = true
		}
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

var claudeToolsForClass = map[core.ToolClass][]string{
	core.ClassRead:      {"Read", "Grep", "Glob"},
	core.ClassShell:     {"Bash"},
	core.ClassTest:      {"Bash"},
	core.ClassMutate:    {"Edit", "Write", "MultiEdit"},
	core.ClassWriteDocs: {"Write"},
}

// codexSandbox picks the sandbox mode for a phase: writable only where the
// policy permits mutation or doc writes.
func codexSandbox(p core.Phase) string {
	switch p {
	case core.PhaseWork, core.PhaseDebug, core.PhaseCompound, core.PhaseCompoundRefresh, core.PhaseProductPulse:
		return "workspace-write"
	default:
		return "read-only"
	}
}

func buildPrompt(u WorkUnit, task, resultPath string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are in the %q phase of the Compounding Engineering loop.\n", u.Phase)
	if task != "" {
		fmt.Fprintf(&b, "Task: %s\n", task)
	}
	fmt.Fprintf(&b, "Permitted here: %s.\n", core.AllowedSummary(u.Phase))
	if len(u.ContextRefs) > 0 {
		b.WriteString("Context artifacts:\n")
		for _, ref := range u.ContextRefs {
			fmt.Fprintf(&b, "  - %s\n", ref.Path)
		}
	}
	if u.Instructions != "" {
		b.WriteString(u.Instructions + "\n")
	}
	fmt.Fprintf(&b, "\nWhen (and only when) this phase is complete, write your StructuredResult "+
		"JSON to %s — fields: producedArtifact{path,kind}, confidence (plan), "+
		"detectedTrack (brainstorm), review{findings} (codeReview).\n", resultPath)
	return b.String()
}

func readResult(path string, p core.Phase) (*StructuredResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var sr StructuredResult
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, err
	}
	sr.Phase = p
	return &sr, nil
}

func execCmd(dir, name string, args []string, stdin string) (string, int, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
		err = nil
	}
	return string(out), code, err
}
