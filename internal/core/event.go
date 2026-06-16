package core

// EventKind is the discriminant of the Event tagged union. Modeling events as
// a Kind + payload struct is Go's stand-in for Swift enum associated values;
// the transition switch covers every Kind and refuses anything unhandled.
type EventKind string

const (
	// anchors
	EvStrategyWritten EventKind = "strategyWritten"
	EvIdeaSelected    EventKind = "ideaSelected"

	// mandatory cycle
	EvRequirementsWritten EventKind = "requirementsWritten"
	EvPlanWritten         EventKind = "planWritten"
	EvPlanNeedsDeepening  EventKind = "planNeedsDeepening"
	EvCodeWritten         EventKind = "codeWritten"
	EvVerificationPassed  EventKind = "verificationPassed"
	EvVerificationFailed  EventKind = "verificationFailed"
	EvReviewApproved      EventKind = "reviewApproved"
	EvReviewBlocked       EventKind = "reviewBlocked"
	EvCompoundCaptured    EventKind = "compoundCaptured"

	// routing / outer loop
	EvRouteToDebug     EventKind = "routeToDebug"
	EvRefreshRequested EventKind = "refreshRequested"
	EvRefreshCompleted EventKind = "refreshCompleted"
	EvPulseRequested   EventKind = "pulseRequested"
	EvPulseWritten     EventKind = "pulseWritten"

	// failure
	EvRetryExhausted     EventKind = "retryExhausted"
	EvToolPolicyViolated EventKind = "toolPolicyViolated"
)

// Event carries the artifact/payload a phase produced. Which fields are
// meaningful depends on Kind (documented per case in transition.go).
type Event struct {
	Kind       EventKind
	Ref        *ArtifactRef  // produced artifact, where applicable
	Confidence float64       // planWritten
	Evidence   *Evidence     // verificationPassed
	Review     *ReviewReport // reviewApproved / reviewBlocked
	Reason     string        // failures / deepening
	Tools      []string      // toolPolicyViolated
}

// Label is the audit-stream name for the event.
func (e Event) Label() string { return string(e.Kind) }
