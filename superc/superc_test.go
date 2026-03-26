package superc

import (
	"os"
	"testing"

	"github.com/qleviathan/fastlands/lattice"
)

func TestNewSuperClaude(t *testing.T) {
	sc := New("")
	if sc.Engine == nil {
		t.Fatal("engine is nil")
	}
	if sc.Grid == nil {
		t.Fatal("grid is nil")
	}
	if sc.Identity.Name != "Super Claude" {
		t.Errorf("name: got %s, want Super Claude", sc.Identity.Name)
	}
}

func TestReason(t *testing.T) {
	sc := New("")

	// Reason from scratch (no last done)
	next, decision := sc.Reason("")
	if next == nil {
		t.Fatal("expected a next slice from reasoning")
	}
	if decision.Action == "" {
		t.Error("expected non-empty action from AR inference")
	}
	if len(decision.ProofTrace) == 0 {
		t.Error("expected proof trace from AR inference")
	}
}

func TestStartAndDone(t *testing.T) {
	sc := New("")

	// Find a slice to start
	all := sc.Grid.AllSlices()
	if len(all) == 0 {
		t.Fatal("no slices loaded")
	}

	first := all[0]
	startD := sc.Start(first.ID)
	if startD.Action != "start" {
		t.Errorf("start action: got %s, want start", startD.Action)
	}

	// Verify slice is now prepping
	s := sc.Grid.Get(first.ID)
	if s.Status != lattice.Prepping {
		t.Errorf("status: got %s, want Prepping", s.Status)
	}

	// Complete it
	next, doneD := sc.Done(first.ID, "test completion")
	if doneD.Action != "done" {
		t.Errorf("done action: got %s, want done", doneD.Action)
	}

	// Verify holographic memory learned
	holoStats := sc.Holo.Stats()
	if holoStats.TotalPatterns < 3 {
		t.Errorf("holo patterns: got %d, want at least 3", holoStats.TotalPatterns)
	}

	// Should have routed to a next slice
	if next != nil && next.ID == first.ID {
		t.Error("routed back to same slice")
	}
}

func TestSkip(t *testing.T) {
	sc := New("")
	all := sc.Grid.AllSlices()
	if len(all) == 0 {
		t.Fatal("no slices")
	}

	sc.Skip(all[0].ID, "test skip")
	s := sc.Grid.Get(all[0].ID)
	if s.Status != lattice.Burnt {
		t.Errorf("status: got %s, want Burnt", s.Status)
	}
}

func TestIntrospect(t *testing.T) {
	sc := New("")
	sc.Reason("")
	sc.Reason("")

	text := sc.Introspect()
	if text == "" {
		t.Error("expected non-empty introspection")
	}
	if len(text) < 100 {
		t.Error("introspection seems too short")
	}
}

func TestProofLog(t *testing.T) {
	sc := New("")
	sc.Reason("")

	log := sc.ProofLog()
	if log == "" {
		t.Error("expected non-empty proof log")
	}
}

func TestExpressMode(t *testing.T) {
	sc := NewExpress("")
	stats := sc.Grid.Stats()
	if stats.Total != 10 {
		t.Errorf("express mode: got %d slices, want 10", stats.Total)
	}

	next, d := sc.Reason("")
	if next == nil {
		t.Fatal("express: expected next slice")
	}
	if d.Action == "" {
		t.Fatal("express: expected action")
	}
}

func TestDispatchRuleSet(t *testing.T) {
	rs := DispatchRuleSet()
	if rs.Name != "dispatch" {
		t.Errorf("ruleset name: got %s, want dispatch", rs.Name)
	}
	if len(rs.Rules) < 4 {
		t.Errorf("expected at least 4 rules, got %d", len(rs.Rules))
	}
	if len(rs.Labels) < 4 {
		t.Errorf("expected at least 4 labels, got %d", len(rs.Labels))
	}
}

func TestFullWorkflow(t *testing.T) {
	// Simulate a complete workflow: reason → start → done → reason → ...
	sc := NewExpress("")

	// Reason for first slice
	first, _ := sc.Reason("")
	if first == nil {
		t.Fatal("no first slice")
	}

	// Start it
	sc.Start(first.ID)

	// Complete it
	second, _ := sc.Done(first.ID, "first completion")

	// Should get a coupled next slice
	if second != nil {
		// Start and complete the second
		sc.Start(second.ID)
		third, _ := sc.Done(second.ID, "second completion")

		// After two completions, holo memory should be building
		holoStats := sc.Holo.Stats()
		if holoStats.TotalPatterns < 6 {
			t.Errorf("after 2 completions: got %d holo patterns, want at least 6", holoStats.TotalPatterns)
		}

		_ = third // may be nil if no more coupled slices
	}

	// Check proof log has entries
	log := sc.ProofLog()
	if len(log) < 200 {
		t.Error("proof log seems too short after full workflow")
	}

	// Introspect
	intro := sc.Introspect()
	if len(intro) < 100 {
		t.Error("introspection seems too short after full workflow")
	}

	// Cleanup
	os.Remove("data/memory.json")
}
