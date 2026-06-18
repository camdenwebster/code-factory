// Package core is the PURE heart of the CompoundEngine state machine.
// It has no I/O, no subprocess, no network — only types and the (State,Event)
// transition function. Everything here is exhaustively unit-testable.
package core

// Phase is a stage in the Compounding Engineering loop. Mandatory cycle:
// brainstorm -> plan -> work -> codeReview -> compound. The others are
// anchors/optional states or the alternate bug path (debug).
type Phase string

const (
	// Anchors / optional (non-linear).
	PhaseStrategy        Phase = "strategy"
	PhaseIdeate          Phase = "ideate"
	PhaseProductPulse    Phase = "productPulse"
	PhaseCompoundRefresh Phase = "compoundRefresh"

	// Mandatory cycle.
	PhaseBrainstorm Phase = "brainstorm"
	PhasePlan       Phase = "plan"
	PhaseWork       Phase = "work"
	PhaseDebug      Phase = "debug" // alternate execution path for bug tracks
	PhaseCodeReview Phase = "codeReview"
	PhaseCompound   Phase = "compound"

	// Terminal.
	PhaseDone   Phase = "done"
	PhaseFailed Phase = "failed"
)

// IsTerminal reports whether the machine has stopped.
func (p Phase) IsTerminal() bool { return p == PhaseDone || p == PhaseFailed }

// Track distinguishes a bug fix from a knowledge/feature change. Decided in
// brainstorm/plan; selects work vs debug.
type Track string

const (
	TrackBug       Track = "bug"
	TrackKnowledge Track = "knowledge"
)
