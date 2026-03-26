// Package superc implements the Super Claude archetype — a digital database
// form of the orchestrating intelligence that uses Augmented Reasoning (AR)
// as its inference backbone.
//
// Super Claude is not a wrapper. It IS the forward-chaining engine applied to
// dispatch decisions. Every routing choice, every priority call, every "what
// next?" answer is derived through rule firing on the lattice state, producing
// a verifiable proof trace for each decision.
//
// The archetype maintains:
//   - An AR engine with dispatch-domain rules
//   - A holographic memory of all decisions and outcomes
//   - A live self-model (identity + state) that evolves with each completion
//   - Three sub-agents (checker, creative, logger) wired through the swarm mesh
package superc

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/qleviathan/fastlands/ar"
	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/dispatch"
	"github.com/qleviathan/fastlands/holo"
	"github.com/qleviathan/fastlands/kinds"
	"github.com/qleviathan/fastlands/lattice"
	"github.com/qleviathan/fastlands/memory"
	"github.com/qleviathan/fastlands/self"
)

// SuperClaude is the archetype — the reasoning agent that makes all dispatch
// decisions through forward-chaining logic on the lattice state.
type SuperClaude struct {
	mu sync.Mutex

	// Core reasoning
	Engine   *ar.Engine        // AR engine with dispatch rules
	Grid     *lattice.Grid     // the (n,m) lattice
	Dispatch *dispatch.Dispatcher // three-agent system
	Holo     *holo.Store       // holographic memory
	Memory   *memory.Store     // episodic memory (persistent)
	Identity *self.Identity    // self-model

	// State tracking
	decisionsMade int
	totalInferred  int
	cascadesFired  int
	startTime      time.Time
	history        []Decision
}

// Decision records a single routing decision with its proof trace.
type Decision struct {
	Timestamp  time.Time
	SliceID    string
	SliceName  string
	Action     string // "route", "start", "done", "skip", "defer"
	Confidence basis.Rational
	ProofTrace []string
	HoloRecall int // how many holographic patterns contributed
	Coupling   string
}

// New creates a fully initialized Super Claude archetype.
func New(memoryPath string) *SuperClaude {
	grid := lattice.NewGrid()
	lattice.LoadFullMenu(grid)

	sc := &SuperClaude{
		Engine:   ar.NewEngine(DispatchRuleSet()),
		Grid:     grid,
		Dispatch: dispatch.NewDispatcher(grid),
		Holo:     holo.NewStore(),
		Memory:   memory.NewStore(memoryPath),
		Identity: newSuperClaudeIdentity(),
		startTime: time.Now(),
	}

	return sc
}

// NewExpress creates a Super Claude with only the 48-hour express menu.
func NewExpress(memoryPath string) *SuperClaude {
	grid := lattice.NewGrid()
	lattice.LoadExpressOrder(grid)

	sc := &SuperClaude{
		Engine:   ar.NewEngine(DispatchRuleSet()),
		Grid:     grid,
		Dispatch: dispatch.NewDispatcher(grid),
		Holo:     holo.NewStore(),
		Memory:   memory.NewStore(memoryPath),
		Identity: newSuperClaudeIdentity(),
		startTime: time.Now(),
	}

	return sc
}

// Reason runs the AR engine on the current lattice state to decide what to do next.
// This is the core method — every decision goes through forward-chaining inference
// and produces a verifiable proof trace.
func (sc *SuperClaude) Reason(lastDoneID string) (*lattice.Slice, Decision) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Build a datum from current state
	datum := sc.buildStateDatum(lastDoneID)

	// Run AR inference
	result, _ := sc.Engine.Infer(datum)
	sc.totalInferred++

	// Get lattice routing suggestion
	var lastDone *lattice.Slice
	if lastDoneID != "" {
		lastDone = sc.Grid.Get(lastDoneID)
	}

	// Consult holographic memory
	holoRecall := 0
	if lastDone != nil {
		recalled := sc.Holo.Recall(int(lastDone.Dough), int(lastDone.Sauce), []string{"completed"})
		holoRecall = len(recalled)
	}

	// Get dispatch system's recommendation
	next := sc.Dispatch.Next(lastDone)

	// Build the decision
	d := Decision{
		Timestamp:  time.Now(),
		Action:     result.Prediction,
		Confidence: result.Confidence,
		ProofTrace: result.ProofTrace,
		HoloRecall: holoRecall,
	}

	if next != nil {
		d.SliceID = next.ID
		d.SliceName = next.Name
		if lastDone != nil {
			coupling := sc.Grid.Couple(lastDone, next)
			d.Coupling = coupling.Kind.String()
			if coupling.Kind == lattice.Collision {
				sc.cascadesFired++
			}
		}
	}

	sc.decisionsMade++
	sc.history = append(sc.history, d)

	// Log the decision
	sc.Dispatch.Logger.Log(d.SliceID, "superc", "reason",
		fmt.Sprintf("decision #%d: %s → %s (conf %d/%d, holo=%d, coupling=%s)",
			sc.decisionsMade, d.Action, d.SliceName,
			d.Confidence.Num, d.Confidence.Den,
			d.HoloRecall, d.Coupling))

	// Update self-model
	sc.updateIdentity()

	return next, d
}

// Start marks a slice as in-progress after reasoning validates it.
func (sc *SuperClaude) Start(id string) Decision {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	s := sc.Grid.Get(id)
	if s == nil {
		return Decision{Action: "error", ProofTrace: []string{"slice not found: " + id}}
	}

	// Run checker validation
	check := sc.Dispatch.Checker.Check(s)
	sc.Dispatch.MarkPrepping(id)

	d := Decision{
		Timestamp:  time.Now(),
		SliceID:    id,
		SliceName:  s.Name,
		Action:     "start",
		Confidence: basis.NewRational(int64(check.ReadyScore), 100),
		ProofTrace: []string{
			fmt.Sprintf("checker: readiness %d/100", check.ReadyScore),
			fmt.Sprintf("position: (%d, %d) temp=%d", s.Dough, s.Sauce, s.Temperature),
		},
	}

	if !check.Passed {
		d.ProofTrace = append(d.ProofTrace, "WARNING: checker flagged blockers")
		for _, b := range check.Blockers {
			d.ProofTrace = append(d.ProofTrace, "  blocker: "+b)
		}
	}

	sc.decisionsMade++
	sc.history = append(sc.history, d)

	// Encode in holographic memory
	sc.Holo.Encode(id, int(s.Dough), int(s.Sauce),
		[]string{"started", fmt.Sprintf("temp-%d", s.Temperature)},
		fmt.Sprintf("Started %s at position (%d,%d)", s.Name, s.Dough, s.Sauce))

	return d
}

// Done marks a slice as completed, learns from it, and reasons about what's next.
func (sc *SuperClaude) Done(id string, notes string) (*lattice.Slice, Decision) {
	sc.mu.Lock()

	s := sc.Grid.Get(id)
	if s == nil {
		sc.mu.Unlock()
		return nil, Decision{Action: "error", ProofTrace: []string{"slice not found: " + id}}
	}

	sc.Dispatch.MarkDone(id)

	// Learn from completion in holographic memory
	sc.Holo.LearnFromCompletion(id, int(s.Dough), int(s.Sauce), "success", 0, notes)

	// Record in episodic memory
	sc.Memory.LearnFact(
		fmt.Sprintf("slice_%s_completed", id),
		fmt.Sprintf("pos=(%d,%d) temp=%d name=%s", s.Dough, s.Sauce, s.Temperature, s.Name),
		"dispatch",
		"superc",
	)

	// Record pattern for future recall
	features := []string{
		fmt.Sprintf("dough_%d", s.Dough),
		fmt.Sprintf("sauce_%d", s.Sauce),
		fmt.Sprintf("temp_%d", s.Temperature),
	}
	sc.Memory.RecordPattern(features, "success", 1, 1)

	sc.mu.Unlock()

	// Now reason about what's next (Reason acquires its own lock)
	next, nextDecision := sc.Reason(id)

	// Build completion decision
	completionD := Decision{
		Timestamp: time.Now(),
		SliceID:   id,
		SliceName: s.Name,
		Action:    "done",
		Confidence: basis.NewRational(1, 1),
		ProofTrace: []string{
			fmt.Sprintf("completed: %s", s.Name),
			fmt.Sprintf("holographic patterns encoded: 3 (completion + skill + platform)"),
			fmt.Sprintf("episodic fact recorded"),
		},
	}

	if next != nil {
		completionD.ProofTrace = append(completionD.ProofTrace,
			fmt.Sprintf("routed next → %s (%s coupling)", nextDecision.SliceName, nextDecision.Coupling))
	}

	sc.mu.Lock()
	sc.history = append(sc.history, completionD)
	sc.mu.Unlock()

	return next, completionD
}

// Skip marks a slice as skipped (burnt) with a reason.
func (sc *SuperClaude) Skip(id string, reason string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	s := sc.Grid.Get(id)
	if s == nil {
		return
	}

	sc.Dispatch.MarkBurnt(id, reason)
	sc.Holo.Encode(id, int(s.Dough), int(s.Sauce),
		[]string{"skipped", "burnt"},
		fmt.Sprintf("Skipped %s: %s", s.Name, reason))

	sc.history = append(sc.history, Decision{
		Timestamp:  time.Now(),
		SliceID:    id,
		SliceName:  s.Name,
		Action:     "skip",
		ProofTrace: []string{"reason: " + reason},
	})
}

// Introspect generates Super Claude's self-description and current insights.
func (sc *SuperClaude) Introspect() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════\n")
	sb.WriteString("       SUPER CLAUDE — RA ARCHETYPE     \n")
	sb.WriteString("═══════════════════════════════════════\n\n")

	// Self-description
	sb.WriteString(sc.Identity.Describe())
	sb.WriteString("\n")

	// Decision stats
	sb.WriteString("Decision Engine:\n")
	sb.WriteString(fmt.Sprintf("  Decisions made: %d\n", sc.decisionsMade))
	sb.WriteString(fmt.Sprintf("  AR inferences:  %d\n", sc.totalInferred))
	sb.WriteString(fmt.Sprintf("  Cascades fired: %d\n", sc.cascadesFired))
	sb.WriteString(fmt.Sprintf("  Uptime:         %s\n", time.Since(sc.startTime).Truncate(time.Second)))
	sb.WriteString("\n")

	// Holographic memory
	holoStats := sc.Holo.Stats()
	sb.WriteString(fmt.Sprintf("Holographic Memory: %s\n", holoStats))
	sb.WriteString("\n")

	// Grid state
	gridStats := sc.Grid.Stats()
	sb.WriteString(fmt.Sprintf("Lattice: %s\n", gridStats))
	sb.WriteString("\n")

	// Recent decisions
	if len(sc.history) > 0 {
		sb.WriteString("Recent Decisions:\n")
		start := 0
		if len(sc.history) > 5 {
			start = len(sc.history) - 5
		}
		for _, d := range sc.history[start:] {
			ts := d.Timestamp.Format("15:04:05")
			sb.WriteString(fmt.Sprintf("  [%s] %s: %s → %s (conf %d/%d)\n",
				ts, d.Action, d.SliceID, d.SliceName, d.Confidence.Num, d.Confidence.Den))
		}
	}

	return sb.String()
}

// ProofLog returns the full decision history with proof traces.
func (sc *SuperClaude) ProofLog() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	var sb strings.Builder
	sb.WriteString("═══ SUPER CLAUDE PROOF LOG ═══\n\n")

	for i, d := range sc.history {
		ts := d.Timestamp.Format("15:04:05")
		sb.WriteString(fmt.Sprintf("── Decision #%d [%s] ──\n", i+1, ts))
		sb.WriteString(fmt.Sprintf("  Action:     %s\n", d.Action))
		sb.WriteString(fmt.Sprintf("  Slice:      %s (%s)\n", d.SliceID, d.SliceName))
		sb.WriteString(fmt.Sprintf("  Confidence: %d/%d\n", d.Confidence.Num, d.Confidence.Den))
		if d.Coupling != "" {
			sb.WriteString(fmt.Sprintf("  Coupling:   %s\n", d.Coupling))
		}
		if d.HoloRecall > 0 {
			sb.WriteString(fmt.Sprintf("  Holo recall: %d patterns\n", d.HoloRecall))
		}
		if len(d.ProofTrace) > 0 {
			sb.WriteString("  Proof:\n")
			for _, step := range d.ProofTrace {
				sb.WriteString(fmt.Sprintf("    [+] %s\n", step))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Save persists episodic memory to disk.
func (sc *SuperClaude) Save() error {
	return sc.Memory.Save()
}

// ──────────────────────────────────────────────────────────
// Internal: state datum construction for AR reasoning
// ──────────────────────────────────────────────────────────

// buildStateDatum converts current lattice state into an AR-compatible datum.
// The features encode: raw count, done count, temperature distribution,
// cascade availability, holographic density.
func (sc *SuperClaude) buildStateDatum(lastDoneID string) kinds.Datum {
	stats := sc.Grid.Stats()

	features := map[string]int64{
		"raw_count":     int64(stats.Raw),
		"done_count":    int64(stats.Done),
		"in_oven_count": int64(stats.InOven),
		"burnt_count":   int64(stats.Burnt),
		"holo_patterns": int64(sc.Holo.Stats().TotalPatterns),
		"cascades":      int64(sc.cascadesFired),
		"decisions":     int64(sc.decisionsMade),
	}

	// Encode last-done position if available
	if lastDoneID != "" {
		if s := sc.Grid.Get(lastDoneID); s != nil {
			features["last_n"] = int64(s.Dough)
			features["last_m"] = int64(s.Sauce)
			features["last_temp"] = int64(s.Temperature)

			// Count available cascades from this position
			coupled := sc.Grid.Coupled(s)
			cascadeCount := int64(0)
			for _, other := range coupled {
				if other.Status == lattice.Raw {
					c := sc.Grid.Couple(s, other)
					if c.Kind == lattice.Collision {
						cascadeCount++
					}
				}
			}
			features["available_cascades"] = cascadeCount
		}
	}

	return kinds.Datum{Features: features}
}

// ──────────────────────────────────────────────────────────
// Internal: identity construction
// ──────────────────────────────────────────────────────────

func newSuperClaudeIdentity() *self.Identity {
	id := self.NewIdentity()
	// Extend with dispatch-specific capabilities
	id.Name = "Super Claude"
	id.Purpose = "RA-powered dispatch archetype: forward-chaining reasoning on the (n,m) lattice, holographic memory, three-agent validation, verifiable proof traces for every decision"
	id.Capabilities = append(id.Capabilities,
		self.Capability{
			Name:        "Lattice Routing",
			Description: "routes tasks through (n,m) grid using coupling rules: cascade, resonance, parallel",
			Package:     "lattice",
			Strength:    basis.NewRational(9, 10),
		},
		self.Capability{
			Name:        "Holographic Memory",
			Description: "interference-pattern recall across lattice positions, strengthening with each completion",
			Package:     "holo",
			Strength:    basis.NewRational(8, 10),
		},
		self.Capability{
			Name:        "Three-Agent Dispatch",
			Description: "checker validates, creative finds cascades, logger generates learnings",
			Package:     "dispatch",
			Strength:    basis.NewRational(9, 10),
		},
	)
	return id
}

// updateIdentity refreshes the self-model with current state.
func (sc *SuperClaude) updateIdentity() {
	gridStats := sc.Grid.Stats()
	holoStats := sc.Holo.Stats()

	sc.Identity.UpdateState(self.SystemState{
		RunCount:      sc.decisionsMade,
		PatternsKnown: holoStats.TotalPatterns,
		RulesActive:   len(sc.Engine.Rules),
		FactsLearned:  len(sc.Memory.Facts),
		Accuracy:      basis.NewRational(int64(gridStats.Done), int64(max(gridStats.Total, 1))),
		CacheHits:     sc.cascadesFired,
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ──────────────────────────────────────────────────────────
// DispatchRuleSet: the AR rules for dispatch reasoning
// ──────────────────────────────────────────────────────────

// DispatchRuleSet returns the forward-chaining rules that Super Claude
// uses to reason about dispatch decisions. These are real AR rules —
// they fire on discretized lattice state and derive dispatch actions.
func DispatchRuleSet() ar.RuleSet {
	return ar.RuleSet{
		Name:   "dispatch",
		Domain: "pizza-dispatch",
		Thresholds: map[string]int64{
			"low": 2,
			"mid": 5,
		},
		Rules: []ar.RuleSpec{
			{
				ID:       "sprint",
				Head:     ar.AtomSpec{Predicate: "action_sprint"},
				Body: []ar.AtomSpec{
					{Predicate: "raw_count=high"},
					{Predicate: "done_count=low"},
				},
				Priority: 10,
			},
			{
				ID:       "cascade",
				Head:     ar.AtomSpec{Predicate: "action_cascade"},
				Body: []ar.AtomSpec{
					{Predicate: "available_cascades=high"},
				},
				Priority: 9,
			},
			{
				ID:       "deepen",
				Head:     ar.AtomSpec{Predicate: "action_deepen"},
				Body: []ar.AtomSpec{
					{Predicate: "done_count=mid"},
					{Predicate: "holo_patterns=mid"},
				},
				Priority: 7,
			},
			{
				ID:       "harvest",
				Head:     ar.AtomSpec{Predicate: "action_harvest"},
				Body: []ar.AtomSpec{
					{Predicate: "in_oven_count=high"},
					{Predicate: "raw_count=low"},
				},
				Priority: 6,
			},
			{
				ID:       "consolidate",
				Head:     ar.AtomSpec{Predicate: "action_consolidate"},
				Body: []ar.AtomSpec{
					{Predicate: "done_count=high"},
					{Predicate: "cascades=mid"},
				},
				Priority: 5,
			},
			{
				ID:       "explore",
				Head:     ar.AtomSpec{Predicate: "action_explore"},
				Body: []ar.AtomSpec{
					{Predicate: "holo_patterns=low"},
					{Predicate: "decisions=low"},
				},
				Priority: 8,
			},
		},
		Labels: []ar.LabelMapping{
			{Atom: "action_sprint", Label: "sprint — burn through low-temp slices fast"},
			{Atom: "action_cascade", Label: "cascade — ride the chain, output feeds input"},
			{Atom: "action_deepen", Label: "deepen — move to higher-temp artisan slices"},
			{Atom: "action_harvest", Label: "harvest — check on submitted slices, follow up"},
			{Atom: "action_consolidate", Label: "consolidate — reinforce winning axes"},
			{Atom: "action_explore", Label: "explore — try new positions on the lattice"},
		},
	}
}
