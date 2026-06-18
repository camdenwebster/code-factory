package runner

import "github.com/camdenwebster/code-factory/internal/core"

// Mock is a scripted runner: it returns a canned StructuredResult per phase.
// Used by the invariant/integration tests and `compound run --runner mock`.
type Mock struct {
	Results  map[core.Phase]StructuredResult
	Category core.Category // optional ConfirmCategory override ("" => echo proposed)
}

func (m Mock) Identifier() string { return "mock" }

// ConfirmCategory lets tests script an agent re-triage; empty echoes the proposal.
func (m Mock) ConfirmCategory(_ string, proposed core.Category) (core.Category, error) {
	if m.Category != "" {
		return m.Category, nil
	}
	return proposed, nil
}

func (m Mock) Run(u WorkUnit) (Result, error) {
	sr, ok := m.Results[u.Phase]
	if !ok {
		sr = StructuredResult{Phase: u.Phase}
	}
	sr.Phase = u.Phase
	return Result{Structured: &sr}, nil
}

// HappyPath is a knowledge-track run that proceeds straight to done.
func HappyPath() Mock {
	return Mock{Results: map[core.Phase]StructuredResult{
		core.PhaseBrainstorm: {ProducedArtifact: &core.ArtifactRef{Path: "docs/brainstorms/x.md", Kind: core.KindRequirements}},
		core.PhasePlan:       {ProducedArtifact: &core.ArtifactRef{Path: "docs/plans/x.md", Kind: core.KindPlan}, Confidence: 0.9},
		core.PhaseWork:       {ProducedArtifact: &core.ArtifactRef{Path: "diff", Kind: core.KindDiff}},
		core.PhaseCodeReview: {Review: &core.ReviewReport{}},
		core.PhaseCompound:   {ProducedArtifact: &core.ArtifactRef{Path: "docs/solutions/conventions/x.md", Kind: core.KindSolution}},
	}}
}
