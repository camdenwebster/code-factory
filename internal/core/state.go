package core

import "time"

// ArtifactKind tags a repo-relative artifact reference.
type ArtifactKind string

const (
	KindStrategy     ArtifactKind = "strategy"
	KindIdeation     ArtifactKind = "ideation"
	KindRequirements ArtifactKind = "requirements"
	KindPlan         ArtifactKind = "plan"
	KindDiff         ArtifactKind = "diff"
	KindSolution     ArtifactKind = "solution"
	KindPulse        ArtifactKind = "pulse"
	KindReview       ArtifactKind = "review"
)

// ArtifactRef is a repo-relative pointer (never an absolute path).
type ArtifactRef struct {
	Path string       `json:"path"`
	Kind ArtifactKind `json:"kind"`
}

// AuditEntry is one append-only record: (phase, event-label, timestamp).
type AuditEntry struct {
	Phase Phase     `json:"phase"`
	Event string    `json:"event"`
	At    time.Time `json:"at"`
}

// MachineState is the full, serializable state of one task run.
type MachineState struct {
	Phase    Phase `json:"phase"`
	Track    Track `json:"track"`
	Attempts int   `json:"attempts"`

	MaxAttempts         int     `json:"maxAttempts"`
	DeepenCount         int     `json:"deepenCount"`
	MaxDeepens          int     `json:"maxDeepens"`
	ConfidenceThreshold float64 `json:"confidenceThreshold"`

	StrategyRef     *ArtifactRef  `json:"strategyRef,omitempty"`
	IdeaRef         *ArtifactRef  `json:"ideaRef,omitempty"`
	RequirementsRef *ArtifactRef  `json:"requirementsRef,omitempty"`
	PlanRef         *ArtifactRef  `json:"planRef,omitempty"`
	LastDiffRef     *ArtifactRef  `json:"lastDiffRef,omitempty"`
	Evidence        *Evidence     `json:"evidence,omitempty"`
	LastReview      *ReviewReport `json:"lastReview,omitempty"`
	SolutionRef     *ArtifactRef  `json:"solutionRef,omitempty"`

	Scheme        string       `json:"scheme,omitempty"` // set => xcodebuild; unset => swift test
	FailureReason string       `json:"failureReason,omitempty"`
	Rejected      []string     `json:"rejectedTransitions,omitempty"`
	AuditLog      []AuditEntry `json:"auditLog,omitempty"`
}

// NewState returns a fresh machine seeded at brainstorm with default bounds.
func NewState() MachineState {
	return MachineState{
		Phase:               PhaseBrainstorm,
		Track:               TrackKnowledge,
		MaxAttempts:         3,
		MaxDeepens:          2,
		ConfidenceThreshold: 0.70,
	}
}
