package core

// Priority is the P1/P2/P3 rubric: must / should / nice.
type Priority string

const (
	P1 Priority = "p1"
	P2 Priority = "p2"
	P3 Priority = "p3"
)

// SuppressionThreshold mirrors CE: findings below this confidence are dropped.
const SuppressionThreshold = 0.60

// ReviewFinding is one reviewer observation.
type ReviewFinding struct {
	Priority   Priority `json:"priority"`
	Confidence float64  `json:"confidence"`
	Reviewer   string   `json:"reviewer"`
	Message    string   `json:"message"`
}

// ReviewReport is the set of findings from a code review.
type ReviewReport struct {
	Findings []ReviewFinding `json:"findings"`
}

// Surviving returns findings at or above the confidence suppression threshold.
func (r ReviewReport) Surviving() []ReviewFinding {
	out := make([]ReviewFinding, 0, len(r.Findings))
	for _, f := range r.Findings {
		if f.Confidence >= SuppressionThreshold {
			out = append(out, f)
		}
	}
	return out
}

// HasBlockingP1 reports whether any surviving finding is a P1 (must-fix).
func (r ReviewReport) HasBlockingP1() bool {
	for _, f := range r.Surviving() {
		if f.Priority == P1 {
			return true
		}
	}
	return false
}
