// Package checkpoint persists MachineState as JSON, giving crash-recovery for
// free (Codable-equivalent round-trip).
package checkpoint

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/camdenwebster/code-factory/internal/core"
)

// Save atomically writes state to path (creating parent dirs).
func Save(path string, s core.MachineState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load reads state from path.
func Load(path string) (core.MachineState, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return core.MachineState{}, err
	}
	var s core.MachineState
	err = json.Unmarshal(b, &s)
	return s, err
}

// Exists reports whether a checkpoint file is present.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
