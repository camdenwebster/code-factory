// Package runner is the harness seam. A Runner turns a per-phase WorkUnit into
// a WorkResult by invoking some agent (a mock, claude -p, or codex exec).
// Interpret maps a structured result back into a PhaseEvent — the same mapping
// used by Model A (run) and Model B (advance --from-result).
package runner

import "github.com/camdenwebster/code-factory/internal/core"

// WorkUnit is a scoped unit of work handed to the agent for one phase.
type WorkUnit struct {
	Phase          core.Phase
	Instructions   string
	AllowedClasses []core.ToolClass
	ContextRefs    []core.ArtifactRef
	WorkingDir     string
}

// StructuredResult is what the agent emits at the end of a phase
// (.compound/result.json).
type StructuredResult struct {
	Phase            core.Phase         `json:"phase"`
	ProducedArtifact *core.ArtifactRef  `json:"producedArtifact,omitempty"`
	Confidence       float64            `json:"confidence,omitempty"`
	DetectedTrack    core.Track         `json:"detectedTrack,omitempty"`
	Review           *core.ReviewReport `json:"review,omitempty"`
}

// Result is the harness output for one WorkUnit.
type Result struct {
	Raw        string
	Structured *StructuredResult
	ToolsUsed  []core.ToolClass
}

// Runner is the only seam that touches a harness.
type Runner interface {
	Identifier() string
	Run(WorkUnit) (Result, error)
}

// Interpret maps a StructuredResult to a PhaseEvent for the given phase. This
// is the Go port of the Swift Orchestrator.interpret and is shared verbatim by
// both integration models.
func Interpret(phase core.Phase, sr StructuredResult) core.Event {
	switch phase {
	case core.PhaseStrategy:
		return core.Event{Kind: core.EvStrategyWritten, Ref: sr.ProducedArtifact}
	case core.PhaseIdeate:
		return core.Event{Kind: core.EvIdeaSelected, Ref: sr.ProducedArtifact}
	case core.PhaseBrainstorm:
		if sr.DetectedTrack == core.TrackBug {
			return core.Event{Kind: core.EvRouteToDebug}
		}
		return core.Event{Kind: core.EvRequirementsWritten, Ref: sr.ProducedArtifact}
	case core.PhasePlan:
		return core.Event{Kind: core.EvPlanWritten, Ref: sr.ProducedArtifact, Confidence: sr.Confidence}
	case core.PhaseWork, core.PhaseDebug:
		return core.Event{Kind: core.EvCodeWritten, Ref: sr.ProducedArtifact}
	case core.PhaseCodeReview:
		rep := sr.Review
		if rep == nil {
			rep = &core.ReviewReport{}
		}
		if rep.HasBlockingP1() {
			return core.Event{Kind: core.EvReviewBlocked, Review: rep}
		}
		return core.Event{Kind: core.EvReviewApproved, Review: rep}
	case core.PhaseCompound:
		return core.Event{Kind: core.EvCompoundCaptured, Ref: sr.ProducedArtifact}
	case core.PhaseCompoundRefresh:
		return core.Event{Kind: core.EvRefreshCompleted}
	case core.PhaseProductPulse:
		return core.Event{Kind: core.EvPulseWritten, Ref: sr.ProducedArtifact}
	default:
		return core.Event{Kind: core.EvRetryExhausted}
	}
}
