package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/camdenwebster/code-factory/internal/checkpoint"
	"github.com/camdenwebster/code-factory/internal/core"
	"github.com/camdenwebster/code-factory/internal/orchestrator"
	"github.com/camdenwebster/code-factory/internal/runner"
	"github.com/camdenwebster/code-factory/internal/store"
	"github.com/camdenwebster/code-factory/internal/verify"
)

// Exit-code contract (see spec/cli.md): 0 ok/allow/pass · 1 fail · 2 usage ·
// 10 policy deny · 11 policy ask.
const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2
	exitDeny  = 10
	exitAsk   = 11
)

// common flags shared by every subcommand.
type ctx struct {
	dir   string
	state string
	json  bool
}

func bindCommon(fs *flag.FlagSet) *ctx {
	c := &ctx{}
	fs.StringVar(&c.dir, "dir", ".", "repo root")
	fs.StringVar(&c.state, "state", "", "checkpoint path (default <dir>/.compound/state.json)")
	fs.BoolVar(&c.json, "json", false, "machine-readable JSON output")
	return c
}

func (c *ctx) statePath() string {
	if c.state != "" {
		return c.state
	}
	if env := os.Getenv("COMPOUND_STATE_DIR"); env != "" {
		return filepath.Join(env, "state.json")
	}
	return filepath.Join(c.dir, ".compound", "state.json")
}

// errNoCycle is returned by load() when no checkpoint exists yet — i.e. there
// is no active CompoundEngine cycle. Inspection commands report this instead of
// silently fabricating a brainstorm phase.
var errNoCycle = errors.New("no active cycle")

func (c *ctx) load() (core.MachineState, error) {
	if !checkpoint.Exists(c.statePath()) {
		return core.MachineState{}, errNoCycle
	}
	return checkpoint.Load(c.statePath())
}

func fail(format string, a ...any) int {
	fmt.Fprintf(os.Stderr, "compound: "+format+"\n", a...)
	return exitFail
}

// ---- init ----------------------------------------------------------------
func cmdInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	c := bindCommon(fs)
	phase := fs.String("phase", string(core.PhaseBrainstorm), "initial phase")
	track := fs.String("track", string(core.TrackKnowledge), "initial track (knowledge|bug)")
	scheme := fs.String("scheme", "", "xcodebuild scheme (unset => swift test)")
	force := fs.Bool("force", false, "overwrite existing checkpoint")
	ifAbsent := fs.Bool("if-absent", false, "no-op (exit 0) if a checkpoint already exists")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	if checkpoint.Exists(c.statePath()) {
		if *ifAbsent {
			return exitOK // a cycle is already running; leave it untouched
		}
		if !*force {
			return fail("checkpoint already exists at %s (use --force)", c.statePath())
		}
	}
	s := core.NewState()
	s.Phase = core.Phase(*phase)
	s.Track = core.Track(*track)
	s.Scheme = *scheme
	if err := checkpoint.Save(c.statePath(), s); err != nil {
		return fail("%v", err)
	}
	fmt.Printf("initialized %s at phase=%s\n", c.statePath(), s.Phase)
	return exitOK
}

// ---- start (triage front door) -------------------------------------------
func cmdStart(args []string) int {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	c := bindCommon(fs)
	task := fs.String("task", "", "the request to triage")
	as := fs.String("as", "", "override the category (feature|bug|chore|strategy|ideate|refresh|pulse)")
	scheme := fs.String("scheme", "", "xcodebuild scheme")
	force := fs.Bool("force", false, "restart even if a cycle is in progress")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	if *task == "" && *as == "" {
		return fail("--task or --as is required")
	}

	tr, err := triage(*task, *as)
	if err != nil {
		return fail("%v", err)
	}
	if checkpoint.Exists(c.statePath()) && !*force {
		return fail("a cycle is already in progress at %s (use --force to restart)", c.statePath())
	}
	s := core.NewState()
	s.Phase, s.Track, s.Scheme = tr.Phase, tr.Track, *scheme
	if err := checkpoint.Save(c.statePath(), s); err != nil {
		return fail("%v", err)
	}
	if c.json {
		b, _ := json.MarshalIndent(tr, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Printf("triaged as %s → starting in phase '%s' (track %s)\n", tr.Category, tr.Phase, tr.Track)
	}
	return exitOK
}

// ---- triage (heuristic proposal, no seed) --------------------------------
func cmdTriage(args []string) int {
	fs := flag.NewFlagSet("triage", flag.ContinueOnError)
	c := bindCommon(fs)
	task := fs.String("task", "", "the request to classify")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	if *task == "" {
		return fail("--task is required")
	}
	tr := core.Triage(*task)
	if c.json {
		b, _ := json.MarshalIndent(tr, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Printf("proposal: %s → phase '%s' (track %s). "+
			"Categories: feature|bug|chore|strategy|ideate|refresh|pulse.\n",
			tr.Category, tr.Phase, tr.Track)
	}
	return exitOK
}

// triage resolves a category from an explicit override or the heuristic.
func triage(task, as string) (core.TriageResult, error) {
	if as != "" {
		cat, ok := core.ParseCategory(as)
		if !ok {
			return core.TriageResult{}, fmt.Errorf("unknown category %q", as)
		}
		p, trk := cat.Seed()
		return core.TriageResult{Category: cat, Phase: p, Track: trk}, nil
	}
	return core.Triage(task), nil
}

// ---- phase ---------------------------------------------------------------
func cmdPhase(args []string) int {
	fs := flag.NewFlagSet("phase", flag.ContinueOnError)
	c := bindCommon(fs)
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	s, err := c.load()
	if errors.Is(err, errNoCycle) {
		fmt.Println("none")
		return exitOK
	}
	if err != nil {
		return fail("%v", err)
	}
	fmt.Println(s.Phase)
	return exitOK
}

// ---- state ---------------------------------------------------------------
func cmdState(args []string) int {
	fs := flag.NewFlagSet("state", flag.ContinueOnError)
	c := bindCommon(fs)
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	s, err := c.load()
	if errors.Is(err, errNoCycle) {
		if c.json {
			fmt.Println(`{"phase":"none"}`)
		} else {
			fmt.Println("no active cycle — run `/ce-start <request>` (or `compound start`) to begin")
		}
		return exitOK
	}
	if err != nil {
		return fail("%v", err)
	}
	if c.json {
		b, _ := json.MarshalIndent(s, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Printf("phase=%s track=%s attempts=%d/%d deepen=%d/%d\n",
			s.Phase, s.Track, s.Attempts, s.MaxAttempts, s.DeepenCount, s.MaxDeepens)
	}
	return exitOK
}

// ---- rehydrate (SessionStart) --------------------------------------------
func cmdRehydrate(args []string) int {
	fs := flag.NewFlagSet("rehydrate", flag.ContinueOnError)
	c := bindCommon(fs)
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	s, err := c.load()
	if errors.Is(err, errNoCycle) {
		fmt.Print("No CompoundEngine cycle in progress. Start one with `/ce-start <request>` " +
			"(triage + seed). Do not assume a phase or run phase commands until then.")
		return exitOK
	}
	if err != nil {
		return fail("%v", err)
	}
	plan := "none"
	if s.PlanRef != nil {
		plan = s.PlanRef.Path
	}
	fmt.Printf("CompoundEngine resumed. phase=%s track=%s plan=%s. "+
		"Honor the per-phase tool policy (%s); advance only via /ce-<phase>.",
		s.Phase, s.Track, plan, core.AllowedSummary(s.Phase))
	return exitOK
}

// ---- policy (PreToolUse) -------------------------------------------------
func cmdPolicy(args []string) int {
	fs := flag.NewFlagSet("policy", flag.ContinueOnError)
	c := bindCommon(fs)
	tool := fs.String("tool", "", "harness tool name (Edit, Bash, apply_patch, shell, ...)")
	phaseOverride := fs.String("phase", "", "phase (default: current)")
	inputJSON := fs.String("input-json", "{}", "tool_input JSON")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	if *tool == "" {
		return fail("--tool is required")
	}
	phase := core.Phase(*phaseOverride)
	if phase == "" {
		s, err := c.load()
		if errors.Is(err, errNoCycle) {
			// No active cycle => no firewall to enforce.
			if c.json {
				fmt.Println(`{"decision":"allow","reason":"no active cycle"}`)
			}
			return exitOK
		}
		if err != nil {
			return fail("%v", err)
		}
		phase = s.Phase
	}
	var input map[string]any
	_ = json.Unmarshal([]byte(*inputJSON), &input)
	class := classifyTool(*tool, input)

	if core.Allowed(phase, class) {
		if c.json {
			fmt.Printf(`{"decision":"allow","class":%q}`+"\n", class)
		}
		return exitOK
	}
	reason := fmt.Sprintf("phase '%s' forbids %s (%s); allowed: %s",
		phase, *tool, class, core.AllowedSummary(phase))
	if c.json {
		fmt.Printf(`{"decision":"deny","class":%q,"reason":%q}`+"\n", class, reason)
	} else {
		fmt.Println(reason)
	}
	return exitDeny
}

// ---- verify (Stop gate) --------------------------------------------------
func cmdVerify(args []string) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	c := bindCommon(fs)
	scheme := fs.String("scheme", "", "xcodebuild scheme")
	advance := fs.Bool("advance", false, "also apply the resulting event")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	s, err := c.load()
	if err != nil {
		return fail("%v", err)
	}
	sch := *scheme
	if sch == "" {
		sch = s.Scheme
	}
	ev, err := verify.Verify(c.dir, sch, verify.Exec)
	if err != nil {
		return fail("%v", err)
	}
	if *advance {
		s = core.Next(s, ev)
		if err := checkpoint.Save(c.statePath(), s); err != nil {
			return fail("%v", err)
		}
	}
	if c.json && ev.Evidence != nil {
		b, _ := json.MarshalIndent(ev.Evidence, "", "  ")
		fmt.Println(string(b))
	}
	if ev.Kind == core.EvVerificationFailed {
		fmt.Fprintf(os.Stderr, "verification failed: %s\n", ev.Reason)
		return exitFail
	}
	fmt.Println("verification passed")
	return exitOK
}

// ---- advance -------------------------------------------------------------
func cmdAdvance(args []string) int {
	fs := flag.NewFlagSet("advance", flag.ContinueOnError)
	c := bindCommon(fs)
	event := fs.String("event", "", "PhaseEvent kind")
	to := fs.String("to", "", "advance one legal forward edge toward this phase (idempotent)")
	fromResult := fs.String("from-result", "", "StructuredResult JSON file (runs interpret)")
	refPath := fs.String("ref", "", "produced artifact path")
	refKind := fs.String("kind", "", "produced artifact kind")
	confidence := fs.Float64("confidence", 0, "plan confidence")
	track := fs.String("track", "", "override track")
	reason := fs.String("reason", "", "reason text")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	s, err := c.load()
	if err != nil {
		return fail("%v", err)
	}

	if *to != "" {
		return advanceTo(c, s, core.Phase(*to))
	}

	var ev core.Event
	switch {
	case *fromResult != "":
		data, err := os.ReadFile(*fromResult)
		if err != nil {
			return fail("%v", err)
		}
		var sr runner.StructuredResult
		if err := json.Unmarshal(data, &sr); err != nil {
			return fail("bad result json: %v", err)
		}
		ev = runner.Interpret(s.Phase, sr)
	case *event != "":
		ev = core.Event{Kind: core.EventKind(*event), Confidence: *confidence, Reason: *reason}
		if *refPath != "" {
			ev.Ref = &core.ArtifactRef{Path: *refPath, Kind: core.ArtifactKind(*refKind)}
		}
	default:
		return fail("--event or --from-result is required")
	}
	if *track != "" {
		s.Track = core.Track(*track)
	}

	s = core.Next(s, ev)
	if err := checkpoint.Save(c.statePath(), s); err != nil {
		return fail("%v", err)
	}
	fmt.Println(s.Phase)
	return exitOK
}

// advanceTo moves the machine one legal forward edge toward target. It is
// idempotent (no-op when already at target) and refuses non-adjacent or gated
// jumps — so a /ce-<phase> command reliably drives the transition even when the
// agent never wrote a result.json, without letting anyone skip a stage.
func advanceTo(c *ctx, s core.MachineState, target core.Phase) int {
	if s.Phase == target {
		fmt.Println(s.Phase) // already there
		return exitOK
	}
	if target == core.PhaseCompound && s.Phase == core.PhaseCodeReview && s.Evidence == nil {
		return fail("cannot compound: the verification gate has not passed — run the tests first")
	}
	ev, ok := forwardEdge(s, target, c.dir)
	if !ok {
		return fail("cannot advance from %s to %s — finish the intervening phase(s) first", s.Phase, target)
	}
	ns := core.Next(s, ev)
	if ns.Phase != target {
		return fail("transition from %s toward %s was rejected (landed on %s)", s.Phase, target, ns.Phase)
	}
	if err := checkpoint.Save(c.statePath(), ns); err != nil {
		return fail("%v", err)
	}
	// Any result.json belonged to the phase we just left; it's now stale.
	_ = os.Remove(filepath.Join(filepath.Dir(c.statePath()), "result.json"))
	fmt.Println(ns.Phase)
	return exitOK
}

// forwardEdge returns the single legal event that moves s.Phase -> target, if
// they are adjacent on the loop. The gated codeReview->compound edge is allowed
// only once Evidence exists (Next enforces this too).
func forwardEdge(s core.MachineState, target core.Phase, dir string) (core.Event, bool) {
	switch {
	case s.Phase == core.PhaseBrainstorm && target == core.PhasePlan:
		return core.Event{Kind: core.EvRequirementsWritten,
			Ref: latestDoc(dir, "docs/brainstorms", core.KindRequirements)}, true
	case s.Phase == core.PhasePlan && target == core.PhaseWork:
		return core.Event{Kind: core.EvPlanWritten, Confidence: s.ConfidenceThreshold,
			Ref: latestDoc(dir, "docs/plans", core.KindPlan)}, true
	case s.Phase == core.PhasePlan && target == core.PhaseDebug:
		return core.Event{Kind: core.EvRouteToDebug}, true
	case (s.Phase == core.PhaseWork || s.Phase == core.PhaseDebug) && target == core.PhaseCodeReview:
		return core.Event{Kind: core.EvCodeWritten, Ref: latestDoc(dir, "docs/plans", core.KindDiff)}, true
	case s.Phase == core.PhaseCodeReview && target == core.PhaseCompound:
		return core.Event{Kind: core.EvReviewApproved, Review: &core.ReviewReport{}}, true
	}
	return core.Event{}, false
}

// latestDoc returns a ref to the most recently modified .md under root/sub, or nil.
func latestDoc(root, sub string, kind core.ArtifactKind) *core.ArtifactRef {
	entries, err := os.ReadDir(filepath.Join(root, sub))
	if err != nil {
		return nil
	}
	var newest string
	var newestAt time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newestAt) {
			newestAt, newest = info.ModTime(), e.Name()
		}
	}
	if newest == "" {
		return nil
	}
	return &core.ArtifactRef{Path: filepath.ToSlash(filepath.Join(sub, newest)), Kind: kind}
}

// ---- audit ---------------------------------------------------------------
func cmdAudit(args []string) int {
	fs := flag.NewFlagSet("audit", flag.ContinueOnError)
	c := bindCommon(fs)
	event := fs.String("event", "", "event/tool label")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	s, err := c.load()
	if errors.Is(err, errNoCycle) {
		return exitOK // no cycle: nothing to audit, and do NOT create state
	}
	if err != nil {
		return fail("%v", err)
	}
	s.AuditLog = append(s.AuditLog, core.AuditEntry{Phase: s.Phase, Event: *event, At: time.Now()})
	if err := checkpoint.Save(c.statePath(), s); err != nil {
		return fail("%v", err)
	}
	return exitOK
}

// ---- log -----------------------------------------------------------------
func cmdLog(args []string) int {
	fs := flag.NewFlagSet("log", flag.ContinueOnError)
	c := bindCommon(fs)
	tail := fs.Int("tail", 0, "show only the last N entries (0 = all)")
	phase := fs.String("phase", "", "filter by phase")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	s, err := c.load()
	if errors.Is(err, errNoCycle) {
		return exitOK // no cycle: empty trail
	}
	if err != nil {
		return fail("%v", err)
	}
	entries := s.AuditLog
	if *phase != "" {
		filtered := entries[:0:0]
		for _, e := range entries {
			if string(e.Phase) == *phase {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}
	if *tail > 0 && len(entries) > *tail {
		entries = entries[len(entries)-*tail:]
	}
	if c.json {
		b, _ := json.MarshalIndent(entries, "", "  ")
		fmt.Println(string(b))
		return exitOK
	}
	for _, e := range entries {
		fmt.Printf("%s\t%s\t%s\n", e.At.Format(time.RFC3339), e.Phase, e.Event)
	}
	return exitOK
}

// ---- search (retrieval) --------------------------------------------------
func cmdSearch(args []string) int {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	c := bindCommon(fs)
	tags := fs.String("tags", "", "comma-separated tags")
	terms := fs.String("terms", "", "comma-separated terms")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	hits, err := store.Store{Root: c.dir}.Search(splitCSV(*tags), splitCSV(*terms))
	if err != nil {
		return fail("%v", err)
	}
	if c.json {
		b, _ := json.MarshalIndent(hits, "", "  ")
		fmt.Println(string(b))
		return exitOK
	}
	for _, h := range hits {
		fmt.Printf("%d\t%s\n", h.Score, h.Path)
	}
	return exitOK
}

// ---- capture (compound step) ---------------------------------------------
func cmdCapture(args []string) int {
	fs := flag.NewFlagSet("capture", flag.ContinueOnError)
	c := bindCommon(fs)
	from := fs.String("from", "-", "SolutionDoc JSON file ('-' = stdin)")
	filename := fs.String("filename", "", "output filename")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	data, err := readAll(*from)
	if err != nil {
		return fail("%v", err)
	}
	var d store.SolutionDoc
	if err := json.Unmarshal(data, &d); err != nil {
		return fail("bad solution json: %v", err)
	}
	name := *filename
	if name == "" {
		name = slug(d.Component) + ".md"
	}
	ref, err := store.Store{Root: c.dir}.Write(d, name)
	if err != nil {
		return fail("%v", err)
	}
	fmt.Println(ref.Path)
	return exitOK
}

// ---- prompt --------------------------------------------------------------
func cmdPrompt(args []string) int {
	fs := flag.NewFlagSet("prompt", flag.ContinueOnError)
	c := bindCommon(fs)
	phase := fs.String("phase", "", "phase")
	task := fs.String("task", "", "task description")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	p := core.Phase(*phase)
	if p == "" {
		s, _ := c.load()
		p = s.Phase
	}
	fmt.Printf("Phase %s. Task: %s. Allowed: %s.\n", p, *task, core.AllowedSummary(p))
	return exitOK
}

// ---- run (Model A driver) ------------------------------------------------
func cmdRun(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	c := bindCommon(fs)
	task := fs.String("task", "", "task description")
	as := fs.String("as", "", "override triage category (feature|bug|chore|strategy|ideate|refresh|pulse)")
	confirm := fs.Bool("confirm", false, "ask the agent to confirm/override the triage category")
	runnerName := fs.String("runner", "mock", "mock|claude|codex")
	scheme := fs.String("scheme", "", "xcodebuild scheme")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	var r runner.Runner
	switch *runnerName {
	case "mock":
		r = runner.HappyPath()
	case "claude", "codex":
		r = runner.NewCLIRunner(*runnerName, filepath.Dir(c.statePath()), *task)
	default:
		return fail("unknown runner %q (mock|claude|codex)", *runnerName)
	}
	// Triage the task to pick the starting phase instead of always brainstorm.
	tr, err := triage(*task, *as)
	if err != nil {
		return fail("%v", err)
	}
	// Optional agent confirmation: the heuristic proposes, the agent disposes.
	// Skipped when the category was set explicitly via --as.
	if *confirm && *as == "" {
		if tg, ok := r.(runner.Triager); ok {
			if cat, e := tg.ConfirmCategory(*task, tr.Category); e == nil {
				if cat != tr.Category {
					fmt.Printf("agent re-triaged %s → %s\n", tr.Category, cat)
				}
				p, trk := cat.Seed()
				tr = core.TriageResult{Category: cat, Phase: p, Track: trk}
			}
		}
	}
	s := core.NewState()
	s.Phase, s.Track, s.Scheme = tr.Phase, tr.Track, *scheme
	verifyFn := func(dir, sch string) (core.Event, error) { return verify.Verify(dir, sch, verify.Exec) }
	save := func(st core.MachineState) error { return checkpoint.Save(c.statePath(), st) }

	final, ferr := orchestrator.Execute(s, r, c.dir, verifyFn, save)
	if ferr != nil {
		return fail("%v", ferr)
	}
	fmt.Printf("task %q triaged as %s; finished: phase=%s\n", *task, tr.Category, final.Phase)
	if final.Phase == core.PhaseFailed {
		fmt.Fprintln(os.Stderr, "reason: "+final.FailureReason)
		return exitFail
	}
	return exitOK
}

// ---- helpers -------------------------------------------------------------
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func readAll(path string) ([]byte, error) {
	if path == "-" {
		return os.ReadFile("/dev/stdin")
	}
	return os.ReadFile(path)
}

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(s)
	if s == "" {
		return "solution"
	}
	return s
}
