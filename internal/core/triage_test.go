package core

import "testing"

func TestTriageClassification(t *testing.T) {
	cases := []struct {
		task string
		cat  Category
		ph   Phase
		trk  Track
	}{
		{"add offline mode to the sync engine", CatFeature, PhaseBrainstorm, TrackKnowledge},
		{"users get logged out on token refresh, looks like a bug", CatBug, PhaseBrainstorm, TrackBug},
		{"the app crashes on launch", CatBug, PhaseBrainstorm, TrackBug},
		{"refactor the networking layer", CatChore, PhasePlan, TrackKnowledge},
		{"bump dependencies to latest", CatChore, PhasePlan, TrackKnowledge},
		{"set the product strategy for Q3", CatStrategy, PhaseStrategy, TrackKnowledge},
		{"consolidate solution docs and dedupe", CatRefresh, PhaseCompoundRefresh, TrackKnowledge},
		{"write the product pulse report", CatPulse, PhaseProductPulse, TrackKnowledge},
	}
	for _, c := range cases {
		got := Triage(c.task)
		if got.Category != c.cat || got.Phase != c.ph || got.Track != c.trk {
			t.Errorf("Triage(%q) = {%s %s %s}, want {%s %s %s}",
				c.task, got.Category, got.Phase, got.Track, c.cat, c.ph, c.trk)
		}
	}
}

func TestParseCategory(t *testing.T) {
	if _, ok := ParseCategory("bug"); !ok {
		t.Fatal("bug should parse")
	}
	if _, ok := ParseCategory("nonsense"); ok {
		t.Fatal("nonsense should not parse")
	}
}
