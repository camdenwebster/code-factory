package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Regression: before a cycle starts (no state.json), inspection/policy/audit
// commands must succeed, report "no cycle", and crucially NOT fabricate a
// brainstorm checkpoint as a side effect.
func TestNoCycleDoesNotSeedState(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("COMPOUND_STATE_DIR", dir)

	for name, fn := range map[string]func([]string) int{
		"phase":     cmdPhase,
		"state":     cmdState,
		"rehydrate": cmdRehydrate,
		"log":       cmdLog,
	} {
		if rc := fn(nil); rc != exitOK {
			t.Fatalf("%s returned %d on a fresh repo, want %d", name, rc, exitOK)
		}
	}
	if rc := cmdAudit([]string{"--event", "Edit"}); rc != exitOK {
		t.Fatalf("audit returned %d, want %d", rc, exitOK)
	}
	if rc := cmdPolicy([]string{"--tool", "Edit", "--input-json", `{"file_path":"a.go"}`}); rc != exitOK {
		t.Fatalf("policy must allow with no active cycle, got %d", rc)
	}

	if _, err := os.Stat(filepath.Join(dir, "state.json")); err == nil {
		t.Fatal("no command may create state.json before a cycle is started")
	}
}
