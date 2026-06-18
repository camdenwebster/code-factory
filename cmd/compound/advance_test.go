package main

import (
	"path/filepath"
	"testing"

	"github.com/camdenwebster/code-factory/internal/checkpoint"
	"github.com/camdenwebster/code-factory/internal/core"
)

func phaseOf(t *testing.T, dir string) core.Phase {
	t.Helper()
	s, err := checkpoint.Load(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return s.Phase
}

// Reproduces the reported bug: after seeding at brainstorm, invoking the next
// phase must actually move the machine — not leave it stuck at brainstorm.
func TestAdvanceToDrivesTransition(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("COMPOUND_STATE_DIR", dir)

	if rc := cmdStart([]string{"--task", "add a settings screen", "--force"}); rc != exitOK {
		t.Fatalf("start rc=%d", rc)
	}
	if got := phaseOf(t, dir); got != core.PhaseBrainstorm {
		t.Fatalf("seeded phase = %s, want brainstorm", got)
	}

	// The fix: /ce-plan drives brainstorm -> plan.
	if rc := cmdAdvance([]string{"--to", "plan"}); rc != exitOK {
		t.Fatalf("advance --to plan rc=%d", rc)
	}
	if got := phaseOf(t, dir); got != core.PhasePlan {
		t.Fatalf("after advance --to plan = %s, want plan", got)
	}

	// Idempotent: running it again is a no-op, not an error.
	if rc := cmdAdvance([]string{"--to", "plan"}); rc != exitOK {
		t.Fatalf("idempotent advance rc=%d", rc)
	}

	// plan -> work (knowledge track).
	if rc := cmdAdvance([]string{"--to", "work"}); rc != exitOK {
		t.Fatalf("advance --to work rc=%d", rc)
	}
	if got := phaseOf(t, dir); got != core.PhaseWork {
		t.Fatalf("after advance --to work = %s, want work", got)
	}

	// Cannot skip a stage: work -> compound is refused.
	if rc := cmdAdvance([]string{"--to", "compound"}); rc == exitOK {
		t.Fatal("work -> compound must be refused (no stage skipping)")
	}
	if got := phaseOf(t, dir); got != core.PhaseWork {
		t.Fatalf("refused advance must not change phase, got %s", got)
	}
}

// The gated edge: codeReview -> compound only after the verification gate set Evidence.
func TestAdvanceToCompoundRequiresEvidence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("COMPOUND_STATE_DIR", dir)
	s := core.NewState()
	s.Phase = core.PhaseCodeReview
	if err := checkpoint.Save(filepath.Join(dir, "state.json"), s); err != nil {
		t.Fatal(err)
	}
	if rc := cmdAdvance([]string{"--to", "compound"}); rc == exitOK {
		t.Fatal("compound without Evidence must be refused")
	}

	s.Evidence = &core.Evidence{ExitCode: 0}
	if err := checkpoint.Save(filepath.Join(dir, "state.json"), s); err != nil {
		t.Fatal(err)
	}
	if rc := cmdAdvance([]string{"--to", "compound"}); rc != exitOK {
		t.Fatalf("compound with Evidence should succeed, rc=%d", rc)
	}
	if got := phaseOf(t, dir); got != core.PhaseCompound {
		t.Fatalf("want compound, got %s", got)
	}
}
