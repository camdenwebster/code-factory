package core

import "testing"

func ref(p string, k ArtifactKind) *ArtifactRef { return &ArtifactRef{Path: p, Kind: k} }

// drive applies a sequence of events and returns the final state.
func drive(s MachineState, events ...Event) MachineState {
	for _, e := range events {
		s = Next(s, e)
	}
	return s
}

// 1. Can never reach compound without a passing Evidence + approved review.
func TestNoCompoundWithoutEvidence(t *testing.T) {
	s := NewState()
	s = drive(s,
		Event{Kind: EvRequirementsWritten, Ref: ref("docs/brainstorms/x.md", KindRequirements)},
		Event{Kind: EvPlanWritten, Ref: ref("docs/plans/x.md", KindPlan), Confidence: 0.9},
		Event{Kind: EvCodeWritten, Ref: ref("diff", KindDiff)},
	)
	if s.Phase != PhaseCodeReview {
		t.Fatalf("want codeReview, got %s", s.Phase)
	}
	// Approve WITHOUT evidence: must be refused, phase unchanged.
	s = Next(s, Event{Kind: EvReviewApproved, Review: &ReviewReport{}})
	if s.Phase != PhaseCodeReview {
		t.Fatalf("approval without evidence must not advance; got %s", s.Phase)
	}
	// Now verify, then approve: reaches compound.
	s = Next(s, Event{Kind: EvVerificationPassed, Evidence: &Evidence{ExitCode: 0}})
	s = Next(s, Event{Kind: EvReviewApproved, Review: &ReviewReport{}})
	if s.Phase != PhaseCompound {
		t.Fatalf("want compound after evidence+approval, got %s", s.Phase)
	}
}

// 2. The deepen loop is bounded: fails after exactly MaxDeepens.
func TestDeepenBounded(t *testing.T) {
	s := NewState()
	s.Phase = PhasePlan
	low := Event{Kind: EvPlanWritten, Ref: ref("p", KindPlan), Confidence: 0.1}
	s = Next(s, low) // deepen 1
	if s.Phase != PhasePlan || s.DeepenCount != 1 {
		t.Fatalf("deepen 1: phase=%s count=%d", s.Phase, s.DeepenCount)
	}
	s = Next(s, low) // deepen 2
	if s.Phase != PhasePlan || s.DeepenCount != 2 {
		t.Fatalf("deepen 2: phase=%s count=%d", s.Phase, s.DeepenCount)
	}
	s = Next(s, low) // exhausted
	if s.Phase != PhaseFailed {
		t.Fatalf("want failed after deepen exhausted, got %s", s.Phase)
	}
}

// 3. Verification retry is bounded: fails after MaxAttempts.
func TestVerificationRetryBounded(t *testing.T) {
	s := NewState()
	s.Phase = PhaseCodeReview
	fail := Event{Kind: EvVerificationFailed, Reason: "boom"}
	for i := 0; i < s.MaxAttempts; i++ {
		s = Next(s, fail)
		if s.Phase != PhaseWork {
			t.Fatalf("attempt %d: want work loopback, got %s", i, s.Phase)
		}
		s.Phase = PhaseCodeReview // simulate work->review for next failure
	}
	s = Next(s, fail)
	if s.Phase != PhaseFailed {
		t.Fatalf("want failed after attempts exhausted, got %s", s.Phase)
	}
}

// 4. Illegal transitions are refused, recorded, and never advance the phase.
func TestIllegalTransitionRefused(t *testing.T) {
	s := NewState() // brainstorm
	before := s.Phase
	s = Next(s, Event{Kind: EvCodeWritten, Ref: ref("d", KindDiff)})
	if s.Phase != before {
		t.Fatalf("illegal event changed phase to %s", s.Phase)
	}
	if len(s.Rejected) != 1 {
		t.Fatalf("want 1 rejected transition, got %d", len(s.Rejected))
	}
	if len(s.AuditLog) != 1 {
		t.Fatalf("illegal transition must still be audited; got %d entries", len(s.AuditLog))
	}
}

// 5. Bug routing flips the track and lands on debug; both tracks reach review.
func TestBugRouting(t *testing.T) {
	s := NewState()
	s = Next(s, Event{Kind: EvRouteToDebug})
	if s.Track != TrackBug || s.Phase != PhasePlan {
		t.Fatalf("routeToDebug: track=%s phase=%s", s.Track, s.Phase)
	}
	s = Next(s, Event{Kind: EvPlanWritten, Ref: ref("p", KindPlan), Confidence: 0.9})
	if s.Phase != PhaseDebug {
		t.Fatalf("bug track high-confidence plan should go to debug, got %s", s.Phase)
	}
	s = Next(s, Event{Kind: EvCodeWritten, Ref: ref("d", KindDiff)})
	if s.Phase != PhaseCodeReview {
		t.Fatalf("debug should converge on codeReview, got %s", s.Phase)
	}
}

// 6. Every transition appends exactly one audit entry.
func TestAuditCompleteness(t *testing.T) {
	s := NewState()
	events := []Event{
		{Kind: EvRequirementsWritten, Ref: ref("b", KindRequirements)},
		{Kind: EvPlanWritten, Ref: ref("p", KindPlan), Confidence: 0.9},
		{Kind: EvCodeWritten, Ref: ref("d", KindDiff)},
		{Kind: EvVerificationPassed, Evidence: &Evidence{}},
		{Kind: EvReviewApproved, Review: &ReviewReport{}},
		{Kind: EvCompoundCaptured, Ref: ref("docs/solutions/x.md", KindSolution)},
	}
	s = drive(s, events...)
	if len(s.AuditLog) != len(events) {
		t.Fatalf("want %d audit entries, got %d", len(events), len(s.AuditLog))
	}
	if s.Phase != PhaseDone {
		t.Fatalf("happy path should end done, got %s", s.Phase)
	}
}

// 7. Review confidence suppression: a low-confidence P1 does not block.
func TestReviewSuppression(t *testing.T) {
	r := ReviewReport{Findings: []ReviewFinding{
		{Priority: P1, Confidence: 0.50}, // below threshold -> suppressed
		{Priority: P3, Confidence: 0.99},
	}}
	if r.HasBlockingP1() {
		t.Fatal("a sub-threshold P1 must be suppressed, not blocking")
	}
	r.Findings[0].Confidence = 0.80
	if !r.HasBlockingP1() {
		t.Fatal("a confident P1 must block")
	}
}

// 8. The outer maintenance loop is reachable from a working phase.
func TestRefreshReachable(t *testing.T) {
	s := NewState()
	s.Phase = PhaseWork
	s = Next(s, Event{Kind: EvRefreshRequested})
	if s.Phase != PhaseCompoundRefresh {
		t.Fatalf("refresh should be reachable from work, got %s", s.Phase)
	}
	s = Next(s, Event{Kind: EvRefreshCompleted})
	if s.Phase != PhaseDone {
		t.Fatalf("refreshCompleted should end done, got %s", s.Phase)
	}
}
