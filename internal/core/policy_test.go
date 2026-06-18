package core

import "testing"

func TestToolPolicyMatrix(t *testing.T) {
	cases := []struct {
		phase Phase
		class ToolClass
		want  bool
	}{
		{PhasePlan, ClassRead, true},
		{PhasePlan, ClassWriteDocs, true}, // must write its plan doc
		{PhasePlan, ClassControl, true},   // must write .compound/result.json
		{PhasePlan, ClassMutate, false},   // but not source
		{PhasePlan, ClassShell, false},
		{PhaseBrainstorm, ClassWriteDocs, true}, // must write its brainstorm doc
		{PhaseBrainstorm, ClassControl, true},
		{PhaseBrainstorm, ClassShell, false},
		{PhaseBrainstorm, ClassMutate, false},
		{PhaseWork, ClassMutate, true},
		{PhaseWork, ClassTest, true},
		{PhaseDebug, ClassWriteDocs, true},
		{PhaseCodeReview, ClassTest, true},
		{PhaseCodeReview, ClassControl, true}, // may write the review result.json
		{PhaseCodeReview, ClassShell, false},  // verify, don't run arbitrary shell
		{PhaseCodeReview, ClassMutate, false},
		{PhaseCompound, ClassWriteDocs, true},
		{PhaseCompound, ClassControl, true},
		{PhaseCompound, ClassMutate, false},
		{PhaseDone, ClassRead, false},
		{PhaseFailed, ClassRead, false},
	}
	for _, c := range cases {
		if got := Allowed(c.phase, c.class); got != c.want {
			t.Errorf("Allowed(%s,%s)=%v want %v", c.phase, c.class, got, c.want)
		}
	}
}
