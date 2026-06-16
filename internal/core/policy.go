package core

// ToolClass is an abstract capability, independent of any harness's tool names.
// The CLI classifies a concrete tool (Edit, Bash, apply_patch, shell, ...) into
// one of these before consulting the policy.
type ToolClass string

const (
	ClassRead      ToolClass = "read"
	ClassShell     ToolClass = "shell"
	ClassTest      ToolClass = "test"
	ClassMutate    ToolClass = "mutate"
	ClassWriteDocs ToolClass = "write_docs"
	ClassOther     ToolClass = "other"
)

// Allowed reports whether a tool class is permitted in a phase. This is the
// firewall the PreToolUse hook enforces; it is the single source of truth that
// the bash fallback mirrors.
func Allowed(p Phase, c ToolClass) bool {
	switch p {
	case PhaseStrategy, PhaseIdeate, PhaseBrainstorm, PhasePlan:
		return c == ClassRead // WHAT, not HOW
	case PhaseWork, PhaseDebug:
		switch c {
		case ClassRead, ClassShell, ClassTest, ClassMutate, ClassWriteDocs:
			return true
		}
		return false
	case PhaseCodeReview:
		return c == ClassRead || c == ClassTest // verify, don't edit
	case PhaseCompound, PhaseCompoundRefresh, PhaseProductPulse:
		return c == ClassRead || c == ClassWriteDocs // write to docs/ only
	default:
		return false // done, failed
	}
}

// AllowedClasses returns the tool classes permitted in a phase (stable order).
func AllowedClasses(p Phase) []ToolClass {
	var out []ToolClass
	for _, c := range []ToolClass{ClassRead, ClassShell, ClassTest, ClassMutate, ClassWriteDocs} {
		if Allowed(p, c) {
			out = append(out, c)
		}
	}
	return out
}

// AllowedSummary is a human description of a phase's permissions, for hook
// deny messages.
func AllowedSummary(p Phase) string {
	switch p {
	case PhaseStrategy, PhaseIdeate, PhaseBrainstorm, PhasePlan:
		return "read/search only (WHAT, not HOW)"
	case PhaseWork, PhaseDebug:
		return "read, edit, shell, build/test"
	case PhaseCodeReview:
		return "read + build/test only (no edits)"
	case PhaseCompound, PhaseCompoundRefresh, PhaseProductPulse:
		return "read + writes under docs/ only"
	default:
		return "none"
	}
}
