package runner

import (
	"regexp"
	"strings"

	"github.com/camdenwebster/code-factory/internal/core"
)

// Triager is an optional capability: a runner whose agent can confirm or
// override the heuristic triage category. The heuristic proposes; the agent
// disposes. A runner that doesn't implement this just uses the heuristic.
type Triager interface {
	ConfirmCategory(task string, proposed core.Category) (core.Category, error)
}

var reCategory = regexp.MustCompile(`(?i)\b(feature|bug|chore|strategy|ideate|refresh|pulse)\b`)

// ConfirmCategory asks the headless agent to classify the request, falling back
// to the heuristic proposal on any error (fail safe — a flaky classifier must
// never block the entry point).
func (r CLIRunner) ConfirmCategory(task string, proposed core.Category) (core.Category, error) {
	var name string
	var args []string
	switch r.Kind {
	case "codex":
		name, args = "codex", []string{"exec", "--sandbox", "read-only"}
	default: // claude — no tools needed, it just answers a word
		name, args = "claude", []string{"-p"}
	}
	out, code, err := r.Exec(".", name, append(args, r.Extra...), triagePrompt(task, proposed))
	if err != nil || code != 0 {
		return proposed, nil
	}
	return scanCategory(out, proposed), nil
}

func triagePrompt(task string, proposed core.Category) string {
	return "Classify this engineering request into exactly one category: " +
		"feature, bug, chore, strategy, ideate, refresh, or pulse.\n" +
		"A keyword heuristic proposed \"" + string(proposed) + "\". Confirm it or pick a better fit.\n" +
		"Request: " + task + "\n\nReply with ONLY the single category word."
}

// scanCategory takes the LAST category word in the agent's reply (its final
// answer), falling back if none is present.
func scanCategory(out string, fallback core.Category) core.Category {
	m := reCategory.FindAllString(out, -1)
	if len(m) == 0 {
		return fallback
	}
	return core.Category(strings.ToLower(m[len(m)-1]))
}
