package core

import "time"

// Evidence is the cryptographic proof that the machine ran the tests itself and
// observed them pass. The state machine never trusts a self-reported pass.
type Evidence struct {
	TestOutputHash string    `json:"testOutputHash"`
	ExitCode       int       `json:"exitCode"`
	Toolchain      string    `json:"toolchain"` // "swift test" | "xcodebuild test"
	Timestamp      time.Time `json:"timestamp"`
}
