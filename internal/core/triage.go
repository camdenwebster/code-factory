package core

import "strings"

// Category is the kind of request the triage router recognizes at the entry
// point. It selects which phase/track a cycle starts in.
type Category string

const (
	CatFeature  Category = "feature"
	CatBug      Category = "bug"
	CatChore    Category = "chore"
	CatStrategy Category = "strategy"
	CatIdeate   Category = "ideate"
	CatRefresh  Category = "refresh"
	CatPulse    Category = "pulse"
)

// TriageResult is the entry-point decision: how a request was classified and
// where the machine should start.
type TriageResult struct {
	Category Category `json:"category"`
	Phase    Phase    `json:"phase"`
	Track    Track    `json:"track"`
}

// Seed maps a category to the phase + track a cycle starts in.
func (c Category) Seed() (Phase, Track) {
	switch c {
	case CatBug:
		return PhaseBrainstorm, TrackBug // understand the defect first; plan routes to debug
	case CatChore:
		return PhasePlan, TrackKnowledge // the WHAT is known; go straight to planning the HOW
	case CatStrategy:
		return PhaseStrategy, TrackKnowledge
	case CatIdeate:
		return PhaseIdeate, TrackKnowledge
	case CatRefresh:
		return PhaseCompoundRefresh, TrackKnowledge
	case CatPulse:
		return PhaseProductPulse, TrackKnowledge
	default: // feature (and CatFeature)
		return PhaseBrainstorm, TrackKnowledge
	}
}

// Ordered so more-specific categories win (bug before chore before feature).
var triageRules = []struct {
	cat      Category
	keywords []string
}{
	{CatBug, []string{"bug", "broken", "crash", "regression", "stack trace", "stacktrace",
		"panic", "exception", "fails", "failing", "doesn't work", "does not work", "not working", "hotfix"}},
	{CatRefresh, []string{"consolidate solution", "prune", "stale doc", "refresh the knowledge",
		"knowledge base", "dedupe", "deduplicate"}},
	{CatPulse, []string{"product pulse", "pulse report", "retro", "retrospective", "usage metrics", "health check"}},
	{CatStrategy, []string{"strategy", "roadmap", "north star", "product vision", "positioning", "okr"}},
	{CatIdeate, []string{"ideate", "explore ideas", "opportunities", "what should we build", "brainstorm ideas"}},
	{CatChore, []string{"refactor", "clean up", "cleanup", "rename", "bump", "upgrade dep",
		"dependencies", "dependency", "migrate", "tidy up", "lint", "chore"}},
}

// Triage classifies a free-text request into a category and the phase/track a
// cycle should start in. It is a deterministic heuristic and only a DEFAULT —
// the caller may override (--as), and brainstorm's routeToDebug remains the
// in-loop safety net for a misclassified bug.
func Triage(task string) TriageResult {
	cat := classify(strings.ToLower(task))
	p, tr := cat.Seed()
	return TriageResult{Category: cat, Phase: p, Track: tr}
}

func classify(lower string) Category {
	for _, r := range triageRules {
		for _, kw := range r.keywords {
			if strings.Contains(lower, kw) {
				return r.cat
			}
		}
	}
	return CatFeature
}

// ParseCategory validates an explicit category override.
func ParseCategory(s string) (Category, bool) {
	c := Category(strings.ToLower(strings.TrimSpace(s)))
	switch c {
	case CatFeature, CatBug, CatChore, CatStrategy, CatIdeate, CatRefresh, CatPulse:
		return c, true
	}
	return "", false
}
