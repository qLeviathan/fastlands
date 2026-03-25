// Package dispatch implements the three-agent pizza dispatch system.
//
// Agent 1 (Checker):   Validates every slice — double-checks status, prerequisites, blockers.
// Agent 2 (Creative):  Thinks laterally about slice combinations, then grounds into actionable steps.
// Agent 3 (Logger):    Records the entire build process and generates learning artifacts.
//
// The Dispatcher orchestrates all three, routing slices through the lattice
// and ensuring each step is verified, optimized, and documented.
package dispatch

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/qleviathan/fastlands/lattice"
)

// ──────────────────────────────────────────────────────────
// Agent interfaces
// ──────────────────────────────────────────────────────────

// CheckResult is the output of the Checker agent.
type CheckResult struct {
	SliceID    string
	Passed     bool
	Blockers   []string
	PreReqs    []string // slice IDs that must finish first
	ReadyScore int      // 0-100, how ready this slice is to enter the oven
}

// CreativeInsight is the output of the Creative agent.
type CreativeInsight struct {
	SliceID      string
	Combinations []string // suggested slice combos that compound
	Optimization string   // grounded, actionable next step
	TimeWindow   string   // when to act
	Grounded     bool     // true if the creative idea survived grounding
}

// LogEntry is a single entry in the build log.
type LogEntry struct {
	Timestamp time.Time
	SliceID   string
	Agent     string // "checker", "creative", "logger", "dispatcher"
	Action    string
	Detail    string
	Duration  time.Duration
}

// ──────────────────────────────────────────────────────────
// Checker Agent — validates, double-checks, finds blockers
// ──────────────────────────────────────────────────────────

type Checker struct {
	grid *lattice.Grid
}

func NewChecker(g *lattice.Grid) *Checker {
	return &Checker{grid: g}
}

// Check validates a slice's readiness to enter the oven.
func (c *Checker) Check(s *lattice.Slice) CheckResult {
	r := CheckResult{SliceID: s.ID, Passed: true, ReadyScore: 100}

	// Check 1: Status must be Raw or Prepping
	if s.Status != lattice.Raw && s.Status != lattice.Prepping {
		r.Passed = false
		r.Blockers = append(r.Blockers, fmt.Sprintf("status is %s, not actionable", s.Status))
		r.ReadyScore = 0
		return r
	}

	// Check 2: Find prerequisite slices (coupled slices that should go first)
	coupled := c.grid.Coupled(s)
	for _, other := range coupled {
		cp := c.grid.Couple(s, other)
		// If a collision-coupled slice on a lower temperature shell is still raw,
		// it should probably go first (cascade dependency)
		if cp.Kind == lattice.Collision && other.Temperature < s.Temperature && other.Status == lattice.Raw {
			r.PreReqs = append(r.PreReqs, other.ID)
			r.ReadyScore -= 20
		}
	}

	// Check 3: High-temperature slices without any completed low-temp slices = risky
	if s.Temperature >= 6 {
		hasFoundation := false
		for _, other := range coupled {
			if other.Temperature <= 3 && (other.Status == lattice.Done || other.Status == lattice.InOven) {
				hasFoundation = true
				break
			}
		}
		if !hasFoundation {
			r.Blockers = append(r.Blockers, "artisan slice with no foundation — consider doing a quick slice on the same axis first")
			r.ReadyScore -= 30
		}
	}

	// Check 4: Pepperoni slices (urgent) get a readiness boost
	for _, t := range s.Toppings {
		if t == lattice.Pepperoni {
			r.ReadyScore += 15
			break
		}
	}

	// Check 5: Jalapeño slices (time-sensitive) — flag the window
	for _, t := range s.Toppings {
		if t == lattice.Jalapeno {
			r.Blockers = append(r.Blockers, "time-sensitive window — prioritize if possible")
			break
		}
	}

	if r.ReadyScore < 0 {
		r.ReadyScore = 0
	}
	if r.ReadyScore > 100 {
		r.ReadyScore = 100
	}
	if len(r.Blockers) > 0 && r.ReadyScore < 50 {
		r.Passed = false
	}

	return r
}

// CheckAll validates every raw slice and returns them sorted by readiness.
func (c *Checker) CheckAll() []CheckResult {
	all := c.grid.AllSlices()
	var results []CheckResult
	for _, s := range all {
		if s.Status == lattice.Raw || s.Status == lattice.Prepping {
			results = append(results, c.Check(s))
		}
	}
	// Sort by ReadyScore descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].ReadyScore > results[i].ReadyScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	return results
}

// ──────────────────────────────────────────────────────────
// Creative Agent — lateral thinking, then grounding
// ──────────────────────────────────────────────────────────

type Creative struct {
	grid *lattice.Grid
}

func NewCreative(g *lattice.Grid) *Creative {
	return &Creative{grid: g}
}

// Brainstorm generates creative insights for a slice, then grounds them.
func (c *Creative) Brainstorm(s *lattice.Slice) CreativeInsight {
	insight := CreativeInsight{SliceID: s.ID}

	coupled := c.grid.Coupled(s)

	// Phase 1: Think creatively about combinations
	for _, other := range coupled {
		cp := c.grid.Couple(s, other)
		if cp.Kind == lattice.Collision {
			// Cascade combo: completing one feeds directly into the next
			insight.Combinations = append(insight.Combinations,
				fmt.Sprintf("CASCADE: %s → %s (same %s, adjacent — output feeds input)", s.Name, other.Name, cp.Axis))
		} else if cp.Kind == lattice.Resonance {
			// Harmonic combo: skip-one pattern
			insight.Combinations = append(insight.Combinations,
				fmt.Sprintf("HARMONIC: %s ↔ %s (stride-2 on %s — skills reinforce at distance)", s.Name, other.Name, cp.Axis))
		}
	}

	// Phase 2: Ground into actionable optimization
	// The grounding step filters creative ideas through reality
	if len(insight.Combinations) > 0 {
		// Find the strongest cascade
		bestCascade := ""
		for _, combo := range insight.Combinations {
			if strings.HasPrefix(combo, "CASCADE:") {
				bestCascade = combo
				break
			}
		}
		if bestCascade != "" {
			insight.Optimization = fmt.Sprintf("Do %s NOW, then immediately pivot to the cascade partner — the work product transfers directly.", s.Name)
			insight.Grounded = true
		} else {
			insight.Optimization = fmt.Sprintf("Complete %s, then leverage the harmonic resonance on your next session.", s.Name)
			insight.Grounded = true
		}
	} else {
		insight.Optimization = fmt.Sprintf("%s is independent — good for parallel background work while waiting on oven timers.", s.Name)
		insight.Grounded = true
	}

	// Phase 3: Time window assessment
	hasPepperoni := false
	hasJalapeno := false
	for _, t := range s.Toppings {
		if t == lattice.Pepperoni {
			hasPepperoni = true
		}
		if t == lattice.Jalapeno {
			hasJalapeno = true
		}
	}
	if hasPepperoni && hasJalapeno {
		insight.TimeWindow = "URGENT + TIME-SENSITIVE: act within hours"
	} else if hasPepperoni {
		insight.TimeWindow = "HIGH PRIORITY: act today"
	} else if hasJalapeno {
		insight.TimeWindow = "TIME-SENSITIVE: window may close"
	} else {
		insight.TimeWindow = "FLEXIBLE: schedule at convenience"
	}

	return insight
}

// BrainstormAll generates insights for all active slices.
func (c *Creative) BrainstormAll() []CreativeInsight {
	all := c.grid.AllSlices()
	var insights []CreativeInsight
	for _, s := range all {
		if s.Status == lattice.Raw || s.Status == lattice.Prepping {
			insights = append(insights, c.Brainstorm(s))
		}
	}
	return insights
}

// ──────────────────────────────────────────────────────────
// Logger Agent — records everything, generates learnings
// ──────────────────────────────────────────────────────────

type Logger struct {
	mu      sync.Mutex
	entries []LogEntry
}

func NewLogger() *Logger {
	return &Logger{}
}

// Log records an event.
func (l *Logger) Log(sliceID, agent, action, detail string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, LogEntry{
		Timestamp: time.Now(),
		SliceID:   sliceID,
		Agent:     agent,
		Action:    action,
		Detail:    detail,
	})
}

// LogTimed records an event with duration.
func (l *Logger) LogTimed(sliceID, agent, action, detail string, dur time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, LogEntry{
		Timestamp: time.Now(),
		SliceID:   sliceID,
		Agent:     agent,
		Action:    action,
		Detail:    detail,
		Duration:  dur,
	})
}

// Entries returns all log entries.
func (l *Logger) Entries() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	cp := make([]LogEntry, len(l.entries))
	copy(cp, l.entries)
	return cp
}

// Summary generates a learning summary from the build log.
func (l *Logger) Summary() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) == 0 {
		return "No entries yet."
	}

	var sb strings.Builder
	sb.WriteString("═══ BUILD LOG SUMMARY ═══\n\n")

	// Count by agent
	agentCounts := make(map[string]int)
	sliceCounts := make(map[string]int)
	for _, e := range l.entries {
		agentCounts[e.Agent]++
		if e.SliceID != "" {
			sliceCounts[e.SliceID]++
		}
	}

	sb.WriteString(fmt.Sprintf("Total events: %d\n", len(l.entries)))
	sb.WriteString("By agent:\n")
	for agent, count := range agentCounts {
		sb.WriteString(fmt.Sprintf("  %s: %d events\n", agent, count))
	}

	sb.WriteString("\nMost active slices:\n")
	for id, count := range sliceCounts {
		sb.WriteString(fmt.Sprintf("  %s: %d events\n", id, count))
	}

	// Learning: find patterns
	sb.WriteString("\n═══ LEARNINGS ═══\n")
	checkerBlocks := 0
	creativeGrounded := 0
	for _, e := range l.entries {
		if e.Agent == "checker" && strings.Contains(e.Action, "blocked") {
			checkerBlocks++
		}
		if e.Agent == "creative" && strings.Contains(e.Action, "grounded") {
			creativeGrounded++
		}
	}
	if checkerBlocks > 0 {
		sb.WriteString(fmt.Sprintf("- Checker blocked %d slices — review prerequisites\n", checkerBlocks))
	}
	if creativeGrounded > 0 {
		sb.WriteString(fmt.Sprintf("- Creative grounded %d insights into actionable steps\n", creativeGrounded))
	}

	return sb.String()
}

// FormatLog returns the full log as readable text.
func (l *Logger) FormatLog() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	var sb strings.Builder
	for _, e := range l.entries {
		ts := e.Timestamp.Format("15:04:05")
		dur := ""
		if e.Duration > 0 {
			dur = fmt.Sprintf(" [%s]", e.Duration)
		}
		sb.WriteString(fmt.Sprintf("[%s] %-12s %-20s %s%s\n", ts, e.Agent, e.Action, e.Detail, dur))
	}
	return sb.String()
}

// ──────────────────────────────────────────────────────────
// Dispatcher — orchestrates all three agents
// ──────────────────────────────────────────────────────────

// Dispatcher coordinates checker, creative, and logger agents
// to route slices through the lattice optimally.
type Dispatcher struct {
	Grid     *lattice.Grid
	Checker  *Checker
	Creative *Creative
	Logger   *Logger
}

// NewDispatcher creates the full three-agent system.
func NewDispatcher(g *lattice.Grid) *Dispatcher {
	return &Dispatcher{
		Grid:     g,
		Checker:  NewChecker(g),
		Creative: NewCreative(g),
		Logger:   NewLogger(),
	}
}

// Plan generates a full dispatch plan: what to do, in what order, with creative insights.
type Plan struct {
	Steps    []PlanStep
	Summary  string
	Insights []CreativeInsight
}

type PlanStep struct {
	Order    int
	Slice    *lattice.Slice
	Check    CheckResult
	Insight  CreativeInsight
	Rationale string
}

// GeneratePlan creates a prioritized action plan using all three agents.
func (d *Dispatcher) GeneratePlan() Plan {
	d.Logger.Log("", "dispatcher", "plan_start", "generating full dispatch plan")

	// Step 1: Checker validates everything
	checks := d.Checker.CheckAll()
	d.Logger.Log("", "checker", "check_all", fmt.Sprintf("validated %d slices", len(checks)))

	// Step 2: Creative brainstorms for each passing slice
	var insights []CreativeInsight
	insightMap := make(map[string]CreativeInsight)
	for _, cr := range checks {
		s := d.Grid.Get(cr.SliceID)
		if s == nil {
			continue
		}
		ci := d.Creative.Brainstorm(s)
		insights = append(insights, ci)
		insightMap[cr.SliceID] = ci
		if ci.Grounded {
			d.Logger.Log(cr.SliceID, "creative", "grounded", ci.Optimization)
		}
	}

	// Step 3: Build ordered plan
	var steps []PlanStep
	order := 1
	for _, cr := range checks {
		s := d.Grid.Get(cr.SliceID)
		if s == nil {
			continue
		}

		rationale := ""
		if cr.Passed {
			rationale = fmt.Sprintf("Ready (score %d/100)", cr.ReadyScore)
		} else {
			rationale = fmt.Sprintf("Blocked: %s", strings.Join(cr.Blockers, "; "))
			d.Logger.Log(cr.SliceID, "checker", "blocked", rationale)
		}

		steps = append(steps, PlanStep{
			Order:     order,
			Slice:     s,
			Check:     cr,
			Insight:   insightMap[cr.SliceID],
			Rationale: rationale,
		})
		order++
	}

	// Build summary
	ready := 0
	blocked := 0
	for _, s := range steps {
		if s.Check.Passed {
			ready++
		} else {
			blocked++
		}
	}
	summary := fmt.Sprintf("%d slices ready, %d blocked, %d total in plan", ready, blocked, len(steps))

	d.Logger.Log("", "dispatcher", "plan_complete", summary)

	return Plan{
		Steps:    steps,
		Summary:  summary,
		Insights: insights,
	}
}

// Next returns the single best next slice to work on.
func (d *Dispatcher) Next(justDone *lattice.Slice) *lattice.Slice {
	if justDone != nil {
		d.Logger.Log(justDone.ID, "dispatcher", "completed", fmt.Sprintf("%s done, routing next", justDone.Name))
	}

	var next *lattice.Slice
	if justDone != nil {
		next = d.Grid.Route(justDone)
	}

	if next == nil {
		// No coupling-based route — pick highest-readiness from checker
		checks := d.Checker.CheckAll()
		for _, cr := range checks {
			if cr.Passed {
				next = d.Grid.Get(cr.SliceID)
				break
			}
		}
	}

	if next != nil {
		cr := d.Checker.Check(next)
		ci := d.Creative.Brainstorm(next)
		d.Logger.Log(next.ID, "dispatcher", "routed",
			fmt.Sprintf("→ %s (ready=%d, %s)", next.Name, cr.ReadyScore, ci.TimeWindow))
	}

	return next
}

// MarkDone transitions a slice to Done and logs it.
func (d *Dispatcher) MarkDone(id string) {
	s := d.Grid.Get(id)
	if s == nil {
		return
	}
	s.Status = lattice.Done
	d.Logger.Log(id, "dispatcher", "done", fmt.Sprintf("%s completed", s.Name))
}

// MarkPrepping transitions a slice to Prepping.
func (d *Dispatcher) MarkPrepping(id string) {
	s := d.Grid.Get(id)
	if s == nil {
		return
	}
	s.Status = lattice.Prepping
	d.Logger.Log(id, "dispatcher", "prepping", fmt.Sprintf("started %s", s.Name))
}

// MarkInOven transitions a slice to InOven (submitted/waiting).
func (d *Dispatcher) MarkInOven(id string) {
	s := d.Grid.Get(id)
	if s == nil {
		return
	}
	s.Status = lattice.InOven
	d.Logger.Log(id, "dispatcher", "in_oven", fmt.Sprintf("%s submitted, waiting", s.Name))
}

// MarkBurnt transitions a slice to Burnt (failed).
func (d *Dispatcher) MarkBurnt(id string, reason string) {
	s := d.Grid.Get(id)
	if s == nil {
		return
	}
	s.Status = lattice.Burnt
	d.Logger.Log(id, "dispatcher", "burnt", fmt.Sprintf("%s failed: %s", s.Name, reason))
}

// FormatPlan renders a plan as human-readable text.
func FormatPlan(p Plan) string {
	var sb strings.Builder
	sb.WriteString("══════════════════════════════════════\n")
	sb.WriteString("        PIZZA DISPATCH PLAN           \n")
	sb.WriteString("══════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf("  %s\n\n", p.Summary))

	for _, step := range p.Steps {
		status := "READY"
		if !step.Check.Passed {
			status = "BLOCKED"
		}

		sb.WriteString(fmt.Sprintf("─── #%d [%s] %s ───\n", step.Order, status, step.Slice.Name))
		sb.WriteString(fmt.Sprintf("  Position: %s × %s (temp %d, flavor %d/%d)\n",
			step.Slice.Dough, step.Slice.Sauce,
			step.Slice.Temperature, step.Slice.Flavor.Num, step.Slice.Flavor.Den))

		toppings := make([]string, len(step.Slice.Toppings))
		for i, t := range step.Slice.Toppings {
			toppings[i] = t.String()
		}
		if len(toppings) > 0 {
			sb.WriteString(fmt.Sprintf("  Toppings: %s\n", strings.Join(toppings, ", ")))
		}

		sb.WriteString(fmt.Sprintf("  Readiness: %d/100 — %s\n", step.Check.ReadyScore, step.Rationale))

		if step.Slice.Notes != "" {
			sb.WriteString(fmt.Sprintf("  Recipe: %s\n", step.Slice.Notes))
		}

		if step.Insight.Optimization != "" {
			sb.WriteString(fmt.Sprintf("  Strategy: %s\n", step.Insight.Optimization))
		}
		if step.Insight.TimeWindow != "" {
			sb.WriteString(fmt.Sprintf("  Window: %s\n", step.Insight.TimeWindow))
		}
		if len(step.Insight.Combinations) > 0 {
			sb.WriteString("  Combos:\n")
			for _, combo := range step.Insight.Combinations {
				sb.WriteString(fmt.Sprintf("    • %s\n", combo))
			}
		}

		if len(step.Check.PreReqs) > 0 {
			sb.WriteString(fmt.Sprintf("  Prerequisites: %s\n", strings.Join(step.Check.PreReqs, ", ")))
		}
		if len(step.Check.Blockers) > 0 {
			sb.WriteString("  Blockers:\n")
			for _, b := range step.Check.Blockers {
				sb.WriteString(fmt.Sprintf("    ⚠ %s\n", b))
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
