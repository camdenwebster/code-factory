package core

// ToolClass is an abstract capability, independent of any harness's tool names.
// The CLI classifies a concrete tool (Edit, Bash, apply_patch, shell, ...) into
// one of these before consulting the policy.
type ToolClass string

const (
	ClassRead      ToolClass = "read"
	ClassShell     ToolClass = "shell"
	ClassTest      ToolClass = "test"
	ClassMutate    ToolClass = "mutate"     // editing/writing SOURCE files
	ClassWriteDocs ToolClass = "write_docs" // writing under docs/
	ClassControl   ToolClass = "control"    // writing the engine's .compound/ files
	ClassOther     ToolClass = "other"
)

// Allowed reports whether a tool class is permitted in a phase. This is the
// firewall the PreToolUse hook enforces; it is the single source of truth that
// the bash fallback mirrors.
func Allowed(p Phase, c ToolClass) bool {
	switch p {
	case PhaseStrategy, PhaseIdeate, PhaseBrainstorm, PhasePlan:
		// Plan, don't code: read + write the phase's DOCS + the engine's
		// control files. Source edits and shell are denied (that's the "HOW").
		return c == ClassRead || c == ClassWriteDocs || c == ClassControl
	case PhaseWork, PhaseDebug:
		return true // the HOW phase: everything
	case PhaseCodeReview:
		// Verify, don't edit — but may write the engine's control files.
		return c == ClassRead || c == ClassTest || c == ClassControl
	case PhaseCompound, PhaseCompoundRefresh, PhaseProductPulse:
		return c == ClassRead || c == ClassWriteDocs || c == ClassControl
	default:
		return false // done, failed
	}
}

// AllowedClasses returns the tool classes permitted in a phase (stable order).
func AllowedClasses(p Phase) []ToolClass {
	var out []ToolClass
	for _, c := range []ToolClass{ClassRead, ClassShell, ClassTest, ClassMutate, ClassWriteDocs, ClassControl} {
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
		return "read + writes to docs/ and .compound/ (no source edits or shell)"
	case PhaseWork, PhaseDebug:
		return "read, edit, shell, build/test"
	case PhaseCodeReview:
		return "read + build/test + .compound/ writes (no source edits)"
	case PhaseCompound, PhaseCompoundRefresh, PhaseProductPulse:
		return "read + writes under docs/ and .compound/"
	default:
		return "none"
	}
}
