package dispatch

import (
	"testing"

	"github.com/qleviathan/fastlands/lattice"
)

func TestCheckerPassesReadySlice(t *testing.T) {
	g := lattice.NewGrid()
	s := &lattice.Slice{
		ID: "t1", Name: "Test",
		Dough: lattice.ThinCrust, Sauce: lattice.Marinara,
		Toppings: []lattice.Topping{lattice.Pepperoni},
		Status: lattice.Raw,
	}
	g.Add(s)

	checker := NewChecker(g)
	r := checker.Check(s)
	if !r.Passed {
		t.Errorf("expected ready slice to pass, got blockers: %v", r.Blockers)
	}
	if r.ReadyScore < 100 {
		t.Errorf("pepperoni slice should have readiness >= 100, got %d", r.ReadyScore)
	}
}

func TestCheckerBlocksDoneSlice(t *testing.T) {
	g := lattice.NewGrid()
	s := &lattice.Slice{
		ID: "t1", Name: "Done Slice",
		Dough: lattice.DeepDish, Sauce: lattice.Pesto,
		Status: lattice.Done,
	}
	g.Add(s)

	checker := NewChecker(g)
	r := checker.Check(s)
	if r.Passed {
		t.Error("expected done slice to be blocked")
	}
}

func TestCreativeFindsCascade(t *testing.T) {
	g := lattice.NewGrid()
	a := &lattice.Slice{
		ID: "a", Name: "Slice A",
		Dough: lattice.StuffedCrust, Sauce: lattice.Pesto,
		Status: lattice.Raw,
	}
	b := &lattice.Slice{
		ID: "b", Name: "Slice B",
		Dough: lattice.StuffedCrust, Sauce: lattice.BBQ,
		Status: lattice.Raw,
	}
	g.Add(a)
	g.Add(b)

	creative := NewCreative(g)
	ci := creative.Brainstorm(a)
	if len(ci.Combinations) == 0 {
		t.Error("expected cascade combination for adjacent sauce slices")
	}
	if !ci.Grounded {
		t.Error("expected insight to be grounded")
	}
}

func TestDispatcherPlan(t *testing.T) {
	g := lattice.NewGrid()
	lattice.LoadExpressOrder(g)

	d := NewDispatcher(g)
	plan := d.GeneratePlan()

	if len(plan.Steps) == 0 {
		t.Fatal("expected non-empty plan")
	}
	if plan.Summary == "" {
		t.Error("expected non-empty summary")
	}

	// Verify logger recorded events
	entries := d.Logger.Entries()
	if len(entries) < 3 {
		t.Errorf("expected at least 3 log entries, got %d", len(entries))
	}
}

func TestDispatcherRouting(t *testing.T) {
	g := lattice.NewGrid()
	lattice.LoadExpressOrder(g)

	d := NewDispatcher(g)
	first := g.Get("express-001")
	if first == nil {
		t.Fatal("express-001 not found")
	}

	d.MarkDone("express-001")
	next := d.Next(first)
	if next == nil {
		t.Fatal("expected next slice after completing express-001")
	}
	if next.Status == lattice.Done {
		t.Error("routed to a done slice")
	}
}

func TestLoggerSummary(t *testing.T) {
	logger := NewLogger()
	logger.Log("s1", "checker", "check", "all good")
	logger.Log("s1", "creative", "grounded", "cascade found")
	logger.Log("s2", "dispatcher", "routed", "next up")

	summary := logger.Summary()
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}
