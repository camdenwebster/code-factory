package runner

import (
	"testing"

	"github.com/camdenwebster/code-factory/internal/core"
)

func TestScanCategory(t *testing.T) {
	cases := []struct {
		out      string
		fallback core.Category
		want     core.Category
	}{
		{"bug", core.CatFeature, core.CatBug},
		{"This is clearly a chore.", core.CatFeature, core.CatChore},
		{"Maybe feature, but actually a chore", core.CatFeature, core.CatChore}, // last wins
		{"no category word here", core.CatRefresh, core.CatRefresh},             // fallback
	}
	for _, c := range cases {
		if got := scanCategory(c.out, c.fallback); got != c.want {
			t.Errorf("scanCategory(%q) = %s, want %s", c.out, got, c.want)
		}
	}
}

func TestConfirmCategoryParsesAgentReply(t *testing.T) {
	r := CLIRunner{Kind: "claude", Exec: func(d, name string, args []string, stdin string) (string, int, error) {
		return "bug\n", 0, nil
	}}
	got, err := r.ConfirmCategory("something feels off", core.CatFeature)
	if err != nil || got != core.CatBug {
		t.Fatalf("got %s (err %v), want bug", got, err)
	}
}

func TestConfirmCategoryFailsSafe(t *testing.T) {
	// Agent errors / non-zero exit must fall back to the heuristic proposal.
	r := CLIRunner{Kind: "claude", Exec: func(d, name string, args []string, stdin string) (string, int, error) {
		return "", 1, nil
	}}
	got, _ := r.ConfirmCategory("x", core.CatChore)
	if got != core.CatChore {
		t.Fatalf("fail-safe broken: got %s, want chore", got)
	}
}
