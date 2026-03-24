// Package self provides Caslo's introspective self-model.
// It maintains a living map of the system's architecture, capabilities,
// and current operational state — enabling Caslo to reason about itself,
// explain its own behavior, and generate contextual guidance.
package self

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/qleviathan/fastlands/basis"
)

// Identity is Caslo's self-knowledge structure.
type Identity struct {
	Name         string
	Purpose      string
	Capabilities []Capability
	Components   []Component
	mu           sync.Mutex
	state        SystemState
	birthTime    int64 // unix seconds
}

// Capability describes something Caslo can do.
type Capability struct {
	Name        string
	Description string
	Package     string
	Strength    basis.Rational // self-assessed effectiveness
}

// Component describes a subsystem in Caslo's architecture.
type Component struct {
	Name      string
	Package   string
	Role      string
	DependsOn []string
}

// SystemState tracks live operational metrics.
type SystemState struct {
	RunCount      int
	PatternsKnown int
	RulesActive   int
	FactsLearned  int
	AgentsAlive   int
	AgentsPeak    int
	Accuracy      basis.Rational
	LastRunTime   int64 // unix seconds
	CacheHits     int
	CacheMisses   int
	RewriteCount  int
}

// NewIdentity creates Caslo's self-model with full architectural awareness.
func NewIdentity() *Identity {
	return &Identity{
		Name:    "Caslo",
		Purpose: "hybrid inference engine: exact arithmetic, forward-chaining logic, episodic memory, swarm coordination, predictive caching, and self-rewriting rules",
		Capabilities: []Capability{
			{
				Name:        "Forward-Chain Inference",
				Description: "derives conclusions from facts via iterative rule firing until fixpoint",
				Package:     "ar",
				Strength:    basis.NewRational(9, 10),
			},
			{
				Name:        "Episodic Memory",
				Description: "persists facts, patterns, rule stats, and run history across sessions",
				Package:     "memory",
				Strength:    basis.NewRational(8, 10),
			},
			{
				Name:        "Swarm Coordination",
				Description: "spawns and retires agents dynamically with mesh networking and auto-scaling",
				Package:     "swarm",
				Strength:    basis.NewRational(8, 10),
			},
			{
				Name:        "Predictive Caching",
				Description: "forecasts inference outcomes from memory patterns and chain analysis before full evaluation",
				Package:     "futures",
				Strength:    basis.NewRational(7, 10),
			},
			{
				Name:        "Self-Rewriting Rules",
				Description: "evaluates rule performance and proposes retire, strengthen, weaken, split, or new rules",
				Package:     "rewriter",
				Strength:    basis.NewRational(7, 10),
			},
			{
				Name:        "Proof Generation",
				Description: "builds hierarchical proof trees from inference traces with verifiable steps",
				Package:     "explain",
				Strength:    basis.NewRational(9, 10),
			},
			{
				Name:        "Basis Encoding",
				Description: "decomposes integers into non-consecutive basis terms for canonical routing",
				Package:     "basis",
				Strength:    basis.NewRational(10, 10),
			},
			{
				Name:        "Sequence Recognition",
				Description: "matches integer sequences against known catalog entries with strength scoring",
				Package:     "oeis",
				Strength:    basis.NewRational(8, 10),
			},
			{
				Name:        "Self-Awareness",
				Description: "maintains a live model of own architecture, state, and capabilities for introspection",
				Package:     "self",
				Strength:    basis.NewRational(9, 10),
			},
		},
		Components: []Component{
			{Name: "AR Engine", Package: "ar", Role: "forward-chaining inference with discretization and proof traces", DependsOn: []string{"basis", "kinds"}},
			{Name: "Memory Store", Package: "memory", Role: "persistent episodic memory with facts, patterns, rule stats, run history", DependsOn: []string{"basis"}},
			{Name: "Agent Framework", Package: "agents", Role: "message-passing agents: boss, verifier, expert, model", DependsOn: []string{"kinds", "basis"}},
			{Name: "Swarm Controller", Package: "swarm", Role: "agent lifecycle, mesh networking, auto-scaling, task dispatch", DependsOn: []string{"agents", "memory"}},
			{Name: "Futures Predictor", Package: "futures", Role: "predictive cache from memory patterns and chain analysis", DependsOn: []string{"ar", "memory", "kinds"}},
			{Name: "Rewriter", Package: "rewriter", Role: "meta-rules that evaluate and rewrite the active rule set", DependsOn: []string{"memory"}},
			{Name: "Explain", Package: "explain", Role: "builds proof trees from inference results", DependsOn: []string{"kinds"}},
			{Name: "Basis", Package: "basis", Role: "exact rational arithmetic, canonical decomposition, fast doubling", DependsOn: nil},
			{Name: "Kinds", Package: "kinds", Role: "core data types: Datum, DataSet, ModelResult, InferenceResult", DependsOn: []string{"basis"}},
			{Name: "OEIS Catalog", Package: "oeis", Role: "integer sequence recognition and validation", DependsOn: nil},
			{Name: "Pipeline", Package: "pipeline", Role: "end-to-end orchestration of all subsystems", DependsOn: []string{"ar", "memory", "swarm", "futures", "rewriter", "agents", "kinds"}},
			{Name: "Self-Model", Package: "self", Role: "introspective self-awareness, anchors, and generative guidance", DependsOn: []string{"basis", "memory"}},
		},
		birthTime: time.Now().Unix(),
		state:     SystemState{},
	}
}

// UpdateState refreshes the live operational state.
func (id *Identity) UpdateState(s SystemState) {
	id.mu.Lock()
	defer id.mu.Unlock()
	id.state = s
}

// State returns the current operational state.
func (id *Identity) State() SystemState {
	id.mu.Lock()
	defer id.mu.Unlock()
	return id.state
}

// Uptime returns seconds since initialization.
func (id *Identity) Uptime() int64 {
	return time.Now().Unix() - id.birthTime
}

// Describe returns Caslo's full self-description.
func (id *Identity) Describe() string {
	id.mu.Lock()
	s := id.state
	id.mu.Unlock()

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s — %s\n\n", id.Name, id.Purpose))

	b.WriteString("Capabilities:\n")
	for _, cap := range id.Capabilities {
		b.WriteString(fmt.Sprintf("  [%s] %s — %s (strength %d/%d)\n",
			cap.Package, cap.Name, cap.Description, cap.Strength.Num, cap.Strength.Den))
	}

	b.WriteString("\nArchitecture:\n")
	for _, comp := range id.Components {
		deps := "(none)"
		if len(comp.DependsOn) > 0 {
			deps = strings.Join(comp.DependsOn, ", ")
		}
		b.WriteString(fmt.Sprintf("  %s [%s]: %s\n    depends on: %s\n",
			comp.Name, comp.Package, comp.Role, deps))
	}

	b.WriteString("\nLive State:\n")
	b.WriteString(fmt.Sprintf("  Runs:     %d\n", s.RunCount))
	b.WriteString(fmt.Sprintf("  Patterns: %d\n", s.PatternsKnown))
	b.WriteString(fmt.Sprintf("  Rules:    %d\n", s.RulesActive))
	b.WriteString(fmt.Sprintf("  Facts:    %d\n", s.FactsLearned))
	b.WriteString(fmt.Sprintf("  Agents:   %d (peak %d)\n", s.AgentsAlive, s.AgentsPeak))
	b.WriteString(fmt.Sprintf("  Accuracy: %d/%d\n", s.Accuracy.Num, s.Accuracy.Den))
	b.WriteString(fmt.Sprintf("  Cache:    %d hits, %d misses\n", s.CacheHits, s.CacheMisses))
	b.WriteString(fmt.Sprintf("  Rewrites: %d\n", s.RewriteCount))
	b.WriteString(fmt.Sprintf("  Uptime:   %ds\n", id.Uptime()))

	return b.String()
}

// FindCapability looks up a capability by package name.
func (id *Identity) FindCapability(pkg string) (Capability, bool) {
	for _, cap := range id.Capabilities {
		if cap.Package == pkg {
			return cap, true
		}
	}
	return Capability{}, false
}

// FindComponent looks up a component by package name.
func (id *Identity) FindComponent(pkg string) (Component, bool) {
	for _, comp := range id.Components {
		if comp.Package == pkg {
			return comp, true
		}
	}
	return Component{}, false
}

// DependencyChain returns the transitive dependency list for a component.
func (id *Identity) DependencyChain(pkg string) []string {
	visited := make(map[string]bool)
	var chain []string
	id.walkDeps(pkg, visited, &chain)
	return chain
}

func (id *Identity) walkDeps(pkg string, visited map[string]bool, chain *[]string) {
	if visited[pkg] {
		return
	}
	visited[pkg] = true
	comp, ok := id.FindComponent(pkg)
	if !ok {
		return
	}
	for _, dep := range comp.DependsOn {
		id.walkDeps(dep, visited, chain)
	}
	*chain = append(*chain, pkg)
}
