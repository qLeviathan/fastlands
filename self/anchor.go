package self

import (
	"fmt"
	"strings"

	"github.com/qleviathan/fastlands/basis"
)

// Anchor is a compressed context bundle that agents receive on spawn.
// It gives them immediate understanding of the system they're operating
// within — who they are, what the system does, what state it's in,
// and what they should focus on.
type Anchor struct {
	// Who the agent is.
	AgentID   string
	AgentRole string

	// What the system is.
	SystemName    string
	SystemPurpose string

	// Compressed architecture map.
	Components []ComponentBrief

	// Current state snapshot.
	ActiveRules   int
	KnownPatterns int
	FactsLearned  int
	AgentsAlive   int
	Accuracy      basis.Rational
	RunCount      int

	// Role-specific guidance for this agent.
	Guidance []string

	// Capabilities this agent should know about.
	Capabilities []string
}

// ComponentBrief is a minimal component descriptor for anchor payloads.
type ComponentBrief struct {
	Name string
	Role string
}

// BuildAnchor creates an anchor for a specific agent role from the
// current identity state. The anchor is tailored to what that role
// needs to know immediately.
func BuildAnchor(id *Identity, agentID string, role string) Anchor {
	s := id.State()

	// Build compressed component list.
	briefs := make([]ComponentBrief, 0, len(id.Components))
	for _, comp := range id.Components {
		briefs = append(briefs, ComponentBrief{
			Name: comp.Name,
			Role: comp.Role,
		})
	}

	// Build capability names.
	capNames := make([]string, 0, len(id.Capabilities))
	for _, cap := range id.Capabilities {
		capNames = append(capNames, cap.Name)
	}

	anchor := Anchor{
		AgentID:       agentID,
		AgentRole:     role,
		SystemName:    id.Name,
		SystemPurpose: id.Purpose,
		Components:    briefs,
		ActiveRules:   s.RulesActive,
		KnownPatterns: s.PatternsKnown,
		FactsLearned:  s.FactsLearned,
		AgentsAlive:   s.AgentsAlive,
		Accuracy:      s.Accuracy,
		RunCount:      s.RunCount,
		Capabilities:  capNames,
	}

	// Role-specific guidance — what this agent needs to latch onto.
	anchor.Guidance = roleGuidance(role, s)

	return anchor
}

// roleGuidance generates context-aware guidance for a specific agent role.
func roleGuidance(role string, s SystemState) []string {
	var g []string

	switch role {
	case "boss":
		g = append(g, "you orchestrate the pipeline: dispatch work, collect results, coordinate agents")
		g = append(g, fmt.Sprintf("system has %d active rules and %d learned patterns", s.RulesActive, s.PatternsKnown))
		if s.RunCount > 0 {
			g = append(g, fmt.Sprintf("accuracy across %d runs: %d/%d", s.RunCount, s.Accuracy.Num, s.Accuracy.Den))
		}
		if s.AgentsAlive > 1 {
			g = append(g, fmt.Sprintf("%d agents currently active — monitor for idle retirement", s.AgentsAlive))
		}

	case "model":
		g = append(g, "you run inference on individual datums via the AR engine")
		g = append(g, "each inference is isolated — no state bleed between calls")
		g = append(g, "return ModelResult with prediction, confidence (rational), and proof trace")
		if s.PatternsKnown > 0 {
			g = append(g, fmt.Sprintf("futures cache has %d patterns — check for cache hits before full eval", s.PatternsKnown))
		}

	case "verifier":
		g = append(g, "you validate inference results: proof traces, confidence bounds, step integrity")
		g = append(g, "check: non-empty proof trace, denominator > 0, numerator in [0..denominator], no empty steps")
		g = append(g, "set Verified=true only when all checks pass")

	case "expert":
		g = append(g, "you analyze domains and recommend reasoning kinds for new problems")
		g = append(g, "review inference results for domain-specific correctness")
		if s.RulesActive > 0 {
			g = append(g, fmt.Sprintf("%d rules active — suggest refinements based on domain knowledge", s.RulesActive))
		}

	default:
		g = append(g, fmt.Sprintf("you are role=%s within the Caslo inference system", role))
		g = append(g, "coordinate with boss-0, produce verifiable results, include proof traces")
	}

	// Universal guidance all agents share.
	g = append(g, "all arithmetic is exact rational — never use floating point")
	g = append(g, "all results must include proof traces for verification")

	return g
}

// Format renders an anchor as a human-readable briefing.
func (a Anchor) Format() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("=== ANCHOR: %s (role=%s) ===\n", a.AgentID, a.AgentRole))
	b.WriteString(fmt.Sprintf("System: %s — %s\n\n", a.SystemName, a.SystemPurpose))

	b.WriteString("Architecture:\n")
	for _, comp := range a.Components {
		b.WriteString(fmt.Sprintf("  %s: %s\n", comp.Name, comp.Role))
	}

	b.WriteString(fmt.Sprintf("\nState: %d rules, %d patterns, %d facts, %d agents, accuracy %d/%d\n",
		a.ActiveRules, a.KnownPatterns, a.FactsLearned, a.AgentsAlive,
		a.Accuracy.Num, a.Accuracy.Den))

	b.WriteString(fmt.Sprintf("Run history: %d runs\n", a.RunCount))

	b.WriteString("\nGuidance:\n")
	for i, line := range a.Guidance {
		b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, line))
	}

	b.WriteString("\nCapabilities available:\n")
	for _, cap := range a.Capabilities {
		b.WriteString(fmt.Sprintf("  - %s\n", cap))
	}

	return b.String()
}

// Digest returns a one-line compressed anchor for fast agent ingestion.
func (a Anchor) Digest() string {
	return fmt.Sprintf("%s|%s|rules=%d|patterns=%d|facts=%d|agents=%d|acc=%d/%d|runs=%d",
		a.AgentID, a.AgentRole, a.ActiveRules, a.KnownPatterns, a.FactsLearned,
		a.AgentsAlive, a.Accuracy.Num, a.Accuracy.Den, a.RunCount)
}
