// Package store is the file-backed knowledge store: capture (compound) and
// retrieval (plan/debug). It mirrors the CE docs/solutions two-track schema.
package store

import (
	"fmt"
	"strings"

	"github.com/camdenwebster/code-factory/internal/core"
)

// ProblemType is the solution classifier; it determines the track and the
// category directory under docs/solutions/.
type ProblemType string

const (
	// bug track
	BuildError       ProblemType = "build_error"
	TestFailure      ProblemType = "test_failure"
	RuntimeError     ProblemType = "runtime_error"
	PerformanceIssue ProblemType = "performance_issue"
	DatabaseIssue    ProblemType = "database_issue"
	SecurityIssue    ProblemType = "security_issue"
	UIBug            ProblemType = "ui_bug"
	IntegrationIssue ProblemType = "integration_issue"
	LogicError       ProblemType = "logic_error"
	// knowledge track
	BestPractice        ProblemType = "best_practice"
	DocumentationGap    ProblemType = "documentation_gap"
	WorkflowIssue       ProblemType = "workflow_issue"
	DeveloperExperience ProblemType = "developer_experience"
	ArchitecturePattern ProblemType = "architecture_pattern"
	DesignPattern       ProblemType = "design_pattern"
	ToolingDecision     ProblemType = "tooling_decision"
	Convention          ProblemType = "convention"
)

var knowledgeTrack = map[ProblemType]bool{
	BestPractice: true, DocumentationGap: true, WorkflowIssue: true,
	DeveloperExperience: true, ArchitecturePattern: true, DesignPattern: true,
	ToolingDecision: true, Convention: true,
}

// Track reports which track a problem type belongs to.
func (pt ProblemType) Track() core.Track {
	if knowledgeTrack[pt] {
		return core.TrackKnowledge
	}
	return core.TrackBug
}

// CategoryDir maps a problem type to its docs/solutions/ subdirectory.
func (pt ProblemType) CategoryDir() string {
	switch pt {
	case PerformanceIssue:
		return "performance-issues"
	case ArchitecturePattern:
		return "architecture-patterns"
	case UIBug:
		return "ui-bugs"
	case ToolingDecision:
		return "tooling-decisions"
	default:
		return strings.ReplaceAll(string(pt), "_", "-")
	}
}

// Severity is the impact rubric.
type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
	Low      Severity = "low"
)

// SolutionDoc is the core learning artifact written by compound and read by
// plan/debug. Bug and knowledge tracks share required frontmatter and differ in
// their track-specific fields.
type SolutionDoc struct {
	Module      string      `json:"module"`
	Date        string      `json:"date"` // YYYY-MM-DD
	ProblemType ProblemType `json:"problemType"`
	Component   string      `json:"component"`
	Severity    Severity    `json:"severity"`
	Tags        []string    `json:"tags"`

	// bug track
	Symptoms       []string `json:"symptoms,omitempty"` // 1..5 when present
	RootCause      string   `json:"rootCause,omitempty"`
	ResolutionType string   `json:"resolutionType,omitempty"`

	// knowledge track
	AppliesWhen string `json:"appliesWhen,omitempty"`

	Body string `json:"body"`
}

// Validate enforces the two-track schema constraints.
func (d SolutionDoc) Validate() error {
	if d.Module == "" || d.Date == "" || d.Component == "" {
		return fmt.Errorf("schema: module, date and component are required")
	}
	if d.ProblemType.Track() == core.TrackBug {
		if n := len(d.Symptoms); n < 1 || n > 5 {
			return fmt.Errorf("schema: bug track requires 1-5 symptoms (got %d)", n)
		}
		if d.RootCause == "" || d.ResolutionType == "" {
			return fmt.Errorf("schema: bug track requires root_cause + resolution_type")
		}
	}
	return nil
}

// Render produces YAML-frontmatter Markdown.
func (d SolutionDoc) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "---\nmodule: %s\ndate: %s\nproblem_type: %s\ncomponent: %s\nseverity: %s\ntags: [%s]\n",
		d.Module, d.Date, d.ProblemType, d.Component, d.Severity, strings.Join(d.Tags, ", "))
	if len(d.Symptoms) > 0 {
		b.WriteString("symptoms:\n")
		for _, s := range d.Symptoms {
			fmt.Fprintf(&b, "  - %s\n", s)
		}
	}
	if d.RootCause != "" {
		fmt.Fprintf(&b, "root_cause: %s\n", d.RootCause)
	}
	if d.ResolutionType != "" {
		fmt.Fprintf(&b, "resolution_type: %s\n", d.ResolutionType)
	}
	if d.AppliesWhen != "" {
		fmt.Fprintf(&b, "applies_when: %s\n", d.AppliesWhen)
	}
	b.WriteString("---\n\n")
	b.WriteString(d.Body)
	return b.String()
}
