package core

import "testing"

func TestToolPolicyMatrix(t *testing.T) {
	cases := []struct {
		phase Phase
		class ToolClass
		want  bool
	}{
		{PhasePlan, ClassRead, true},
		{PhasePlan, ClassMutate, false},
		{PhaseBrainstorm, ClassShell, false},
		{PhaseWork, ClassMutate, true},
		{PhaseWork, ClassTest, true},
		{PhaseDebug, ClassWriteDocs, true},
		{PhaseCodeReview, ClassTest, true},
		{PhaseCodeReview, ClassShell, false}, // verify, don't run arbitrary shell
		{PhaseCodeReview, ClassMutate, false},
		{PhaseCompound, ClassWriteDocs, true},
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
