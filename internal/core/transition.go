package core

import (
	"fmt"
	"strings"
	"time"
)

// now is indirected so tests can pin timestamps.
var now = time.Now

// Next applies one (state, event) -> state. It is PURE: no I/O, no subprocess,
// no LLM. The final default arm is the "agent tried to skip a stage" guard —
// illegal transitions are recorded and the phase is never silently advanced.
//
// Soundness fixes over the original Swift skeleton:
//   - refresh/pulse are reachable from any non-terminal phase (outer loop),
//   - reviewApproved cannot reach compound without Evidence (invariant encoded
//     here rather than relying on caller ordering),
//   - no panics: a missing payload simply yields a rejected transition upstream.
func Next(s MachineState, e Event) MachineState {
	s.AuditLog = append(s.AuditLog, AuditEntry{Phase: s.Phase, Event: e.Label(), At: now()})

	switch {
	// ---- Anchors ----
	case s.Phase == PhaseStrategy && e.Kind == EvStrategyWritten:
		s.StrategyRef = e.Ref
		s.Phase = PhaseBrainstorm
	case s.Phase == PhaseIdeate && e.Kind == EvIdeaSelected:
		s.IdeaRef = e.Ref
		s.Phase = PhaseBrainstorm

	// ---- Brainstorm: WHAT, not HOW. May route to the bug track. ----
	case s.Phase == PhaseBrainstorm && e.Kind == EvRequirementsWritten:
		s.RequirementsRef = e.Ref
		s.Phase = PhasePlan
	case s.Phase == PhaseBrainstorm && e.Kind == EvRouteToDebug:
		s.Track = TrackBug
		s.Phase = PhasePlan // still plan first; plan decides debug vs work

	// ---- Plan: confidence-gated, with a bounded deepen loop. ----
	case s.Phase == PhasePlan && e.Kind == EvPlanWritten:
		s.PlanRef = e.Ref
		switch {
		case e.Confidence >= s.ConfidenceThreshold:
			s.Phase = workOrDebug(s.Track)
		case s.DeepenCount < s.MaxDeepens:
			s.DeepenCount++
			s.Phase = PhasePlan // deepen
		default:
			s.Phase = PhaseFailed
			s.FailureReason = fmt.Sprintf("plan confidence %.2f < threshold after %d deepens", e.Confidence, s.DeepenCount)
		}
	case s.Phase == PhasePlan && e.Kind == EvPlanNeedsDeepening:
		if s.DeepenCount < s.MaxDeepens {
			s.DeepenCount++
			s.Phase = PhasePlan
			s.Rejected = append(s.Rejected, "deepen: "+e.Reason)
		} else {
			s.Phase = PhaseFailed
			s.FailureReason = "deepen exhausted: " + e.Reason
		}
	case s.Phase == PhasePlan && e.Kind == EvRouteToDebug:
		s.Track = TrackBug
		s.Phase = PhaseDebug

	// ---- Work / Debug: both produce a diff, both go to review. ----
	case (s.Phase == PhaseWork || s.Phase == PhaseDebug) && e.Kind == EvCodeWritten:
		s.LastDiffRef = e.Ref
		s.Phase = PhaseCodeReview

	// ---- Code review: verification gate + P1/P2/P3 findings. ----
	case s.Phase == PhaseCodeReview && e.Kind == EvVerificationPassed:
		s.Evidence = e.Evidence
		// stay in codeReview; reviewers run next (same phase, next event)
	case s.Phase == PhaseCodeReview && e.Kind == EvVerificationFailed:
		if s.Attempts >= s.MaxAttempts {
			s.Phase = PhaseFailed
			s.FailureReason = "verification: " + e.Reason
		} else {
			s.Attempts++
			s.Phase = workOrDebug(s.Track)
		}
	case s.Phase == PhaseCodeReview && e.Kind == EvReviewApproved:
		s.LastReview = e.Review
		if s.Evidence == nil {
			// INVARIANT: cannot compound without a passing Evidence.
			s.Rejected = append(s.Rejected, "reviewApproved without evidence")
		} else {
			s.Phase = PhaseCompound
		}
	case s.Phase == PhaseCodeReview && e.Kind == EvReviewBlocked:
		s.LastReview = e.Review
		if s.Attempts >= s.MaxAttempts {
			s.Phase = PhaseFailed
			s.FailureReason = "unresolved P1 findings"
		} else {
			s.Attempts++
			s.Evidence = nil // force re-verification after the fix
			s.Phase = workOrDebug(s.Track)
		}

	// ---- Compound: the closing loop. Terminal for a single task. ----
	case s.Phase == PhaseCompound && e.Kind == EvCompoundCaptured:
		s.SolutionRef = e.Ref
		s.Phase = PhaseDone

	// ---- Maintenance & outer loop (enterable from any non-terminal phase) ----
	case e.Kind == EvRefreshRequested && !s.Phase.IsTerminal():
		s.Phase = PhaseCompoundRefresh
	case s.Phase == PhaseCompoundRefresh && e.Kind == EvRefreshCompleted:
		s.Phase = PhaseDone
	case e.Kind == EvPulseRequested && !s.Phase.IsTerminal():
		s.Phase = PhaseProductPulse
	case s.Phase == PhaseProductPulse && e.Kind == EvPulseWritten:
		s.SolutionRef = e.Ref // pulse feeds the NEXT cycle's brainstorm
		s.Phase = PhaseDone

	// ---- Global failure transitions ----
	case e.Kind == EvRetryExhausted:
		s.Phase = PhaseFailed
		s.FailureReason = "retry exhausted"
	case e.Kind == EvToolPolicyViolated:
		s.Phase = PhaseFailed
		s.FailureReason = "tool policy violated: " + strings.Join(e.Tools, ",")

	// ---- Illegal transition: refuse, record, never silently advance. ----
	default:
		s.Rejected = append(s.Rejected, fmt.Sprintf("%s cannot handle %s", s.Phase, e.Kind))
	}
	return s
}

func workOrDebug(t Track) Phase {
	if t == TrackBug {
		return PhaseDebug
	}
	return PhaseWork
}
