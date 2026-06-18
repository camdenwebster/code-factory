// Package orchestrator is the headless driver (Model A): the loop that drives
// one phase at a time until the machine reaches a terminal state. It is the Go
// equivalent of the Swift Orchestrator.execute.
package orchestrator

import (
	"github.com/camdenwebster/code-factory/internal/core"
	"github.com/camdenwebster/code-factory/internal/runner"
)

// VerifyFn runs the deterministic verification gate for dir/scheme.
type VerifyFn func(dir, scheme string) (core.Event, error)

// SaveFn checkpoints state after each transition (may be nil).
type SaveFn func(core.MachineState) error

// Execute runs the loop. At codeReview it runs the verification gate FIRST
// (when no Evidence yet), then lets reviewers run — mirroring the rule that the
// machine verifies before it trusts a review.
func Execute(s core.MachineState, r runner.Runner, dir string, verify VerifyFn, save SaveFn) (core.MachineState, error) {
	for !s.Phase.IsTerminal() {
		var ev core.Event

		if s.Phase == core.PhaseCodeReview && s.Evidence == nil {
			e, err := verify(dir, s.Scheme)
			if err != nil {
				return s, err
			}
			ev = e
		} else {
			unit := runner.WorkUnit{
				Phase:          s.Phase,
				AllowedClasses: core.AllowedClasses(s.Phase),
				ContextRefs:    contextRefs(s),
				WorkingDir:     dir,
			}
			res, err := r.Run(unit)
			if err != nil {
				return s, err
			}
			// Defense in depth: enforce the tool policy after the fact.
			if illegal := violations(s.Phase, res.ToolsUsed); len(illegal) > 0 {
				ev = core.Event{Kind: core.EvToolPolicyViolated, Tools: illegal}
			} else if res.Structured != nil {
				ev = runner.Interpret(s.Phase, *res.Structured)
			} else {
				ev = core.Event{Kind: core.EvVerificationFailed, Reason: "no structured result"}
			}
		}

		s = core.Next(s, ev)
		if save != nil {
			if err := save(s); err != nil {
				return s, err
			}
		}
	}
	return s, nil
}

func contextRefs(s core.MachineState) []core.ArtifactRef {
	var out []core.ArtifactRef
	for _, r := range []*core.ArtifactRef{s.StrategyRef, s.IdeaRef, s.RequirementsRef, s.PlanRef, s.LastDiffRef, s.SolutionRef} {
		if r != nil {
			out = append(out, *r)
		}
	}
	return out
}

func violations(p core.Phase, used []core.ToolClass) []string {
	var bad []string
	for _, c := range used {
		if !core.Allowed(p, c) {
			bad = append(bad, string(c))
		}
	}
	return bad
}
