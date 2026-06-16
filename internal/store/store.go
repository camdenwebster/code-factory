package store

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/camdenwebster/code-factory/internal/core"
)

// Store is rooted at a repo and reads/writes docs/solutions/.
type Store struct{ Root string }

func (s Store) solutionsDir() string { return filepath.Join(s.Root, "docs", "solutions") }

// Write validates and renders a SolutionDoc into docs/solutions/<category>/.
// Returns the repo-relative ArtifactRef.
func (s Store) Write(d SolutionDoc, filename string) (core.ArtifactRef, error) {
	if err := d.Validate(); err != nil {
		return core.ArtifactRef{}, err
	}
	rel := filepath.Join("docs", "solutions", d.ProblemType.CategoryDir(), filename)
	abs := filepath.Join(s.Root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return core.ArtifactRef{}, err
	}
	if err := os.WriteFile(abs, []byte(d.Render()), 0o644); err != nil {
		return core.ArtifactRef{}, err
	}
	return core.ArtifactRef{Path: filepath.ToSlash(rel), Kind: core.KindSolution}, nil
}

// Hit is a search result with a crude overlap score.
type Hit struct {
	Path  string `json:"path"`
	Score int    `json:"score"`
}

// Search scans docs/solutions/ for tag/term overlap (local-first retrieval,
// run before any web research). Case-insensitive substring match.
func (s Store) Search(tags, terms []string) ([]Hit, error) {
	dir := s.solutionsDir()
	if _, err := os.Stat(dir); err != nil {
		return nil, nil // no solutions yet
	}
	needles := make([]string, 0, len(tags)+len(terms))
	for _, t := range append(append([]string{}, tags...), terms...) {
		if t = strings.ToLower(strings.TrimSpace(t)); t != "" {
			needles = append(needles, t)
		}
	}
	var hits []Hit
	err := filepath.WalkDir(dir, func(path string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, _ := os.ReadFile(path)
		lower := strings.ToLower(string(data))
		score := 0
		for _, n := range needles {
			if strings.Contains(lower, n) {
				score++
			}
		}
		if score > 0 {
			rel, _ := filepath.Rel(s.Root, path)
			hits = append(hits, Hit{Path: filepath.ToSlash(rel), Score: score})
		}
		return nil
	})
	return hits, err
}
