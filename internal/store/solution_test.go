package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateBugTrack(t *testing.T) {
	d := SolutionDoc{Module: "m", Date: "2026-06-16", Component: "c", ProblemType: UIBug, Severity: Low}
	if err := d.Validate(); err == nil {
		t.Fatal("bug track with 0 symptoms should fail validation")
	}
	d.Symptoms = []string{"a"}
	d.RootCause = "rc"
	d.ResolutionType = "fix"
	if err := d.Validate(); err != nil {
		t.Fatalf("valid bug doc rejected: %v", err)
	}
	d.Symptoms = []string{"1", "2", "3", "4", "5", "6"}
	if err := d.Validate(); err == nil {
		t.Fatal("more than 5 symptoms should fail")
	}
}

func TestKnowledgeTrackNeedsNoSymptoms(t *testing.T) {
	d := SolutionDoc{Module: "m", Date: "2026-06-16", Component: "c", ProblemType: ToolingDecision, Severity: High}
	if err := d.Validate(); err != nil {
		t.Fatalf("knowledge doc should not require symptoms: %v", err)
	}
}

func TestWriteAndSearch(t *testing.T) {
	root := t.TempDir()
	s := Store{Root: root}
	d := SolutionDoc{
		Module: "code-factory", Date: "2026-06-16", Component: "auth",
		ProblemType: UIBug, Severity: Medium, Tags: []string{"token", "logout"},
		Symptoms: []string{"logged out on refresh"}, RootCause: "expired token",
		ResolutionType: "refresh-retry", Body: "## Fix\n...",
	}
	r, err := s.Write(d, "token-refresh.md")
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if r.Path != "docs/solutions/ui-bugs/token-refresh.md" {
		t.Fatalf("unexpected path %s", r.Path)
	}
	if _, err := os.Stat(filepath.Join(root, r.Path)); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	hits, err := s.Search([]string{"token"}, nil)
	if err != nil || len(hits) != 1 {
		t.Fatalf("search want 1 hit, got %d (err %v)", len(hits), err)
	}
}
