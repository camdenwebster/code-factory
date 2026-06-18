package verify

import (
	"testing"

	"github.com/camdenwebster/code-factory/internal/core"
)

func TestVerifyPass(t *testing.T) {
	run := func(dir, name string, args ...string) (string, int, error) {
		return "Test Suite 'All tests' passed", 0, nil
	}
	ev, err := Verify(".", "", run)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Kind != core.EvVerificationPassed {
		t.Fatalf("want passed, got %s", ev.Kind)
	}
	if ev.Evidence == nil || ev.Evidence.TestOutputHash == "" {
		t.Fatal("passing verify must produce a hashed Evidence")
	}
}

func TestVerifyFail(t *testing.T) {
	run := func(dir, name string, args ...string) (string, int, error) {
		return "** TEST FAILED **", 65, nil
	}
	ev, _ := Verify(".", "", run)
	if ev.Kind != core.EvVerificationFailed {
		t.Fatalf("want failed, got %s", ev.Kind)
	}
}

func TestVerifyPicksToolchain(t *testing.T) {
	var gotName string
	run := func(dir, name string, args ...string) (string, int, error) {
		gotName = name
		return "", 0, nil
	}
	Verify(".", "MyApp", run) //nolint:errcheck
	if gotName != "xcodebuild" {
		t.Fatalf("scheme set should select xcodebuild, got %s", gotName)
	}
	Verify(".", "", run) //nolint:errcheck
	if gotName != "swift" {
		t.Fatalf("no scheme should select swift, got %s", gotName)
	}
}
