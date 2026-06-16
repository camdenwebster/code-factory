package orchestrator

import (
	"testing"

	"github.com/camdenwebster/code-factory/internal/core"
	"github.com/camdenwebster/code-factory/internal/runner"
)

// End-to-end: the mock happy path drives brainstorm -> ... -> done, and the
// verification gate runs at codeReview before the review is trusted.
func TestExecuteHappyPath(t *testing.T) {
	verified := false
	verify := func(dir, scheme string) (core.Event, error) {
		verified = true
		return core.Event{Kind: core.EvVerificationPassed, Evidence: &core.Evidence{ExitCode: 0}}, nil
	}
	final, err := Execute(core.NewState(), runner.HappyPath(), ".", verify, nil)
	if err != nil {
		t.Fatal(err)
	}
	if final.Phase != core.PhaseDone {
		t.Fatalf("want done, got %s (%s)", final.Phase, final.FailureReason)
	}
	if !verified {
		t.Fatal("verification gate must run at codeReview")
	}
	if final.SolutionRef == nil {
		t.Fatal("a completed run must capture a solution")
	}
}

// A failing gate loops back and eventually fails — it never reaches compound.
func TestExecuteGateFailureNeverCompounds(t *testing.T) {
	verify := func(dir, scheme string) (core.Event, error) {
		return core.Event{Kind: core.EvVerificationFailed, Reason: "red"}, nil
	}
	final, _ := Execute(core.NewState(), runner.HappyPath(), ".", verify, nil)
	if final.Phase != core.PhaseFailed {
		t.Fatalf("persistent red tests must end failed, got %s", final.Phase)
	}
	if final.SolutionRef != nil {
		t.Fatal("must never compound on red tests")
	}
}
