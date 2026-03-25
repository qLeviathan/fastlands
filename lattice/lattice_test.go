package lattice

import (
	"testing"
)

func TestGridAddAndGet(t *testing.T) {
	g := NewGrid()
	s := &Slice{
		ID: "test-001", Name: "Test Slice",
		Dough: DeepDish, Sauce: Pesto,
		Toppings: []Topping{Pepperoni},
		Status: Raw,
	}
	g.Add(s)

	got := g.Get("test-001")
	if got == nil {
		t.Fatal("expected slice, got nil")
	}
	if got.Temperature != 4 { // DeepDish(2) + Pesto(2)
		t.Errorf("temperature: got %d, want 4", got.Temperature)
	}
	// Sequence is 1,1,2,3,5... so F(2)=2, F(2)=2 → flavor = 4
	if got.Flavor.Num != 4 || got.Flavor.Den != 1 {
		t.Errorf("flavor: got %d/%d, want 4/1", got.Flavor.Num, got.Flavor.Den)
	}
	if got.Polarity != 1 { // (2+2)%2 == 0 → +1
		t.Errorf("polarity: got %d, want 1", got.Polarity)
	}
}

func TestCouplingSharedDough(t *testing.T) {
	g := NewGrid()
	a := &Slice{ID: "a", Dough: StuffedCrust, Sauce: Pesto, Status: Raw}
	b := &Slice{ID: "b", Dough: StuffedCrust, Sauce: BBQ, Status: Raw}
	g.Add(a)
	g.Add(b)

	c := g.Couple(a, b)
	if c.Axis != "dough" {
		t.Errorf("axis: got %s, want dough", c.Axis)
	}
	if c.Gap != 1 { // BBQ(3) - Pesto(2) = 1
		t.Errorf("gap: got %d, want 1", c.Gap)
	}
	if c.Kind != Collision { // gap=1 → cascade
		t.Errorf("kind: got %v, want Collision", c.Kind)
	}
}

func TestCouplingSharedSauce(t *testing.T) {
	g := NewGrid()
	a := &Slice{ID: "a", Dough: ThinCrust, Sauce: Marinara, Status: Raw}
	b := &Slice{ID: "b", Dough: HandTossed, Sauce: Marinara, Status: Raw}
	g.Add(a)
	g.Add(b)

	c := g.Couple(a, b)
	if c.Axis != "sauce" {
		t.Errorf("axis: got %s, want sauce", c.Axis)
	}
	if c.Kind != Collision { // gap=1
		t.Errorf("kind: got %v, want Collision", c.Kind)
	}
}

func TestCouplingParallel(t *testing.T) {
	g := NewGrid()
	// Same diagonal (d=3), no shared axis
	a := &Slice{ID: "a", Dough: ThinCrust, Sauce: BBQ, Status: Raw}   // (0,3) d=3
	b := &Slice{ID: "b", Dough: HandTossed, Sauce: Pesto, Status: Raw} // (1,2) d=3
	g.Add(a)
	g.Add(b)

	c := g.Couple(a, b)
	if c.Kind != Parallel {
		t.Errorf("kind: got %v, want Parallel", c.Kind)
	}
}

func TestRouting(t *testing.T) {
	g := NewGrid()
	done := &Slice{ID: "done", Dough: StuffedCrust, Sauce: Pesto, Status: Done}
	// Collision candidate: same dough, adjacent sauce
	next := &Slice{ID: "next", Dough: StuffedCrust, Sauce: BBQ, Status: Raw}
	// Weaker candidate: same dough, far sauce
	far := &Slice{ID: "far", Dough: StuffedCrust, Sauce: HotHoney, Status: Raw}
	g.Add(done)
	g.Add(next)
	g.Add(far)

	routed := g.Route(done)
	if routed == nil {
		t.Fatal("expected routed slice, got nil")
	}
	if routed.ID != "next" {
		t.Errorf("routed to %s, want 'next' (collision priority)", routed.ID)
	}
}

func TestShellGrouping(t *testing.T) {
	g := NewGrid()
	g.Add(&Slice{ID: "a", Dough: ThinCrust, Sauce: BBQ, Status: Raw})       // d=3
	g.Add(&Slice{ID: "b", Dough: HandTossed, Sauce: Pesto, Status: Raw})    // d=3
	g.Add(&Slice{ID: "c", Dough: DeepDish, Sauce: Marinara, Status: Raw})   // d=2

	shell3 := g.Shell(3)
	if len(shell3) != 2 {
		t.Errorf("shell(3) count: got %d, want 2", len(shell3))
	}
	shell2 := g.Shell(2)
	if len(shell2) != 1 {
		t.Errorf("shell(2) count: got %d, want 1", len(shell2))
	}
}

func TestFullMenuLoad(t *testing.T) {
	g := NewGrid()
	LoadFullMenu(g)
	stats := g.Stats()
	if stats.Total < 10 {
		t.Errorf("full menu has %d slices, expected at least 10", stats.Total)
	}
	if stats.DoughTypes < 3 {
		t.Errorf("full menu has %d dough types, expected at least 3", stats.DoughTypes)
	}
}

func TestExpressOrder(t *testing.T) {
	g := NewGrid()
	LoadExpressOrder(g)
	stats := g.Stats()
	if stats.Total != 10 {
		t.Errorf("express order has %d slices, expected 10", stats.Total)
	}
	if stats.Raw != 10 {
		t.Errorf("express order has %d raw, expected all 10 raw", stats.Raw)
	}
}
