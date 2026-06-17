// Command compound is the CompoundEngine orchestration CLI. It is the single
// source of truth for the state machine: the harness hooks (Model B) shell out
// to the small commands, and `compound run` (Model A) is the headless driver.
// See spec/cli.md for the full contract.
package main

import (
	"fmt"
	"os"
)

const usage = `compound — CompoundEngine orchestration CLI

Usage: compound <command> [flags]

Entry:      start             (triage a request -> seed the right phase)
State:      init  phase  state  rehydrate
Policy:     policy            (PreToolUse firewall)
Verify:     verify            (Stop gate)
Transition: advance  audit  log
Knowledge:  search  capture
Driver:     prompt  run       (Model A headless)

Run "compound <command> -h" for flags. Exit codes: 0 ok · 1 fail · 2 usage · 10 deny · 11 ask.`

func main() { os.Exit(dispatch(os.Args[1:])) }

func dispatch(args []string) int {
	if len(args) == 0 {
		fmt.Println(usage)
		return exitUsage
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "start":
		return cmdStart(rest)
	case "init":
		return cmdInit(rest)
	case "phase":
		return cmdPhase(rest)
	case "state":
		return cmdState(rest)
	case "rehydrate":
		return cmdRehydrate(rest)
	case "policy":
		return cmdPolicy(rest)
	case "verify":
		return cmdVerify(rest)
	case "advance":
		return cmdAdvance(rest)
	case "audit":
		return cmdAudit(rest)
	case "log":
		return cmdLog(rest)
	case "search":
		return cmdSearch(rest)
	case "capture":
		return cmdCapture(rest)
	case "prompt":
		return cmdPrompt(rest)
	case "run":
		return cmdRun(rest)
	case "version":
		fmt.Println("compound 0.1.0")
		return exitOK
	case "-h", "--help", "help":
		fmt.Println(usage)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "compound: unknown command %q\n\n%s\n", cmd, usage)
		return exitUsage
	}
}
