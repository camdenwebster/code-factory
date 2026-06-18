// Package verify is the deterministic verification gate. It runs the test suite
// ITSELF and hashes the output into Evidence — it never trusts a self-reported
// pass. swift-crypto is unnecessary: crypto/sha256 is in the Go stdlib.
package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"time"

	"github.com/camdenwebster/code-factory/internal/core"
)

// CommandRunner runs a command in dir and returns combined output + exit code.
// Indirected so tests can verify the hashing/branching without a toolchain.
type CommandRunner func(dir, name string, args ...string) (output string, exitCode int, err error)

// Exec is the production runner.
func Exec(dir, name string, args ...string) (string, int, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
		err = nil // a non-zero exit is data, not a Go error
	}
	return string(out), code, err
}

// command picks the toolchain: xcodebuild when a scheme is set, else swift test.
func command(scheme string) (string, []string) {
	if scheme != "" {
		return "xcodebuild", []string{"test", "-scheme", scheme,
			"-destination", "platform=iOS Simulator,name=iPhone 16"}
	}
	return "swift", []string{"test"}
}

// Verify runs the gate and returns the resulting PhaseEvent
// (verificationPassed with Evidence, or verificationFailed).
func Verify(dir, scheme string, run CommandRunner) (core.Event, error) {
	name, args := command(scheme)
	out, code, err := run(dir, name, args...)
	if err != nil {
		// Fail closed: an unrunnable toolchain is a failure to PROVE green,
		// not a crash. The gate never lets a turn pass without evidence.
		return core.Event{Kind: core.EvVerificationFailed, Reason: err.Error()}, nil
	}
	if code != 0 {
		return core.Event{Kind: core.EvVerificationFailed, Reason: fmt.Sprintf("exit %d", code)}, nil
	}
	sum := sha256.Sum256([]byte(out))
	return core.Event{
		Kind: core.EvVerificationPassed,
		Evidence: &core.Evidence{
			TestOutputHash: hex.EncodeToString(sum[:]),
			ExitCode:       code,
			Toolchain:      name + " " + args[0],
			Timestamp:      time.Now(),
		},
	}, nil
}
