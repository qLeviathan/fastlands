// Package lattice implements the (n, m) routing grid for the pizza dispatch system.
// Each slice sits at a position on the grid. Slices sharing a dough type or sauce
// compound — the output of one IS the input of the other along the shared dimension.
//
// Dough axis (n): recipe complexity / preparation style
// Sauce axis (m): oven location / kitchen station
// Temperature: diagonal d = n + m (total heat shell)
// Flavor intensity at (n, m) = F(n) · F(m) where F is the basis sequence.
package lattice

import (
	"fmt"
	"sort"
	"sync"

	"github.com/qleviathan/fastlands/basis"
)

// Dough encodes the n-axis (preparation style).
type Dough int

const (
	ThinCrust    Dough = iota // 0 — lowest prep, fastest out the oven
	HandTossed                // 1 — moderate craft
	DeepDish                  // 2 — analytical depth
	StuffedCrust              // 3 — engineering heavy
	Neapolitan                // 4 — artisan, expert-level
)

var doughNames = [...]string{
	"Thin Crust",
	"Hand Tossed",
	"Deep Dish",
	"Stuffed Crust",
	"Neapolitan",
}

func (d Dough) String() string {
	if int(d) < len(doughNames) {
		return doughNames[d]
	}
	return fmt.Sprintf("Dough(%d)", d)
}

// Sauce encodes the m-axis (kitchen station / oven slot).
type Sauce int

const (
	Marinara    Sauce = iota // 0
	Alfredo                  // 1
	Pesto                    // 2
	BBQ                      // 3
	Buffalo                  // 4
	GarlicButter             // 5
	Vodka                    // 6
	Ranch                    // 7
	HotHoney                 // 8
	Truffle                  // 9
)

var sauceNames = [...]string{
	"Marinara",
	"Alfredo",
	"Pesto",
	"BBQ",
	"Buffalo",
	"Garlic Butter",
	"Vodka",
	"Ranch",
	"Hot Honey",
	"Truffle",
}

func (s Sauce) String() string {
	if int(s) < len(sauceNames) {
		return sauceNames[s]
	}
	return fmt.Sprintf("Sauce(%d)", s)
}

// Topping marks urgency/character of a slice.
type Topping int

const (
	Pepperoni Topping = iota // high urgency
	Mushroom                 // deep focus required
	Olive                    // passive / async
	Jalapeno                 // time-sensitive window
	Basil                    // creative component
)

var toppingNames = [...]string{"Pepperoni", "Mushroom", "Olive", "Jalapeño", "Basil"}

func (t Topping) String() string {
	if int(t) < len(toppingNames) {
		return toppingNames[t]
	}
	return fmt.Sprintf("Topping(%d)", t)
}

// Slice is a single task positioned at (n, m) on the lattice.
type Slice struct {
	ID       string
	Name     string
	Dough    Dough
	Sauce    Sauce
	Toppings []Topping
	Status   SliceStatus
	Notes    string

	// Computed fields
	Temperature int            // d = n + m
	Flavor      basis.Rational // F(n) · F(m)
	Polarity    int            // (-1)^(n+m)
}

type SliceStatus int

const (
	Raw      SliceStatus = iota // not started
	Prepping                    // in progress
	InOven                      // submitted / waiting
	Done                        // delivered
	Burnt                       // failed / abandoned
)

var statusNames = [...]string{"Raw", "Prepping", "In Oven", "Done", "Burnt"}

func (s SliceStatus) String() string {
	if int(s) < len(statusNames) {
		return statusNames[s]
	}
	return "Unknown"
}

// Grid is the (n, m) lattice holding all slices.
type Grid struct {
	mu     sync.RWMutex
	slices map[string]*Slice          // id → slice
	byPos  map[[2]int][]*Slice        // (n,m) → slices at that position
	byDough map[Dough][]*Slice        // shared n-axis
	bySauce map[Sauce][]*Slice        // shared m-axis
	byDiag  map[int][]*Slice          // same temperature shell
	seq    []uint64                    // basis sequence cache
}

// NewGrid creates an empty lattice with precomputed basis terms.
func NewGrid() *Grid {
	seq := basis.Sequence(20) // first 20 basis terms
	return &Grid{
		slices:  make(map[string]*Slice),
		byPos:   make(map[[2]int][]*Slice),
		byDough: make(map[Dough][]*Slice),
		bySauce: make(map[Sauce][]*Slice),
		byDiag:  make(map[int][]*Slice),
		seq:     seq,
	}
}

// F returns the basis term at index i (F_0=0, F_1=1, F_2=1, F_3=2, ...).
func (g *Grid) F(i int) uint64 {
	if i < 0 {
		return 0
	}
	if i < len(g.seq) {
		return g.seq[i]
	}
	return basis.Term(uint64(i))
}

// Add places a slice on the lattice. Computes temperature, flavor, polarity.
func (g *Grid) Add(s *Slice) {
	g.mu.Lock()
	defer g.mu.Unlock()

	n := int(s.Dough)
	m := int(s.Sauce)
	s.Temperature = n + m
	fn := g.F(n)
	fm := g.F(m)
	s.Flavor = basis.RatMul(
		basis.Rational{Num: int64(fn), Den: 1},
		basis.Rational{Num: int64(fm), Den: 1},
	)
	if (n+m)%2 == 0 {
		s.Polarity = 1
	} else {
		s.Polarity = -1
	}

	g.slices[s.ID] = s
	pos := [2]int{n, m}
	g.byPos[pos] = append(g.byPos[pos], s)
	g.byDough[s.Dough] = append(g.byDough[s.Dough], s)
	g.bySauce[s.Sauce] = append(g.bySauce[s.Sauce], s)
	g.byDiag[s.Temperature] = append(g.byDiag[s.Temperature], s)
}

// Get retrieves a slice by ID.
func (g *Grid) Get(id string) *Slice {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.slices[id]
}

// CouplingKind describes how two slices interact.
type CouplingKind int

const (
	Parallel  CouplingKind = iota // same diagonal, no shared axis → independent
	DoughCoupled                  // shared n (same dough type) → skill transfer
	SauceCoupled                  // shared m (same sauce) → platform transfer
	Collision                     // adjacent indices on shared axis → cascade
	Resonance                     // stride-2 gap on shared axis → harmonic
)

var couplingNames = [...]string{"Parallel", "Dough-Coupled", "Sauce-Coupled", "Collision", "Resonance"}

func (c CouplingKind) String() string {
	if int(c) < len(couplingNames) {
		return couplingNames[c]
	}
	return "Unknown"
}

// Coupling describes the relationship between two slices.
type Coupling struct {
	Kind     CouplingKind
	Strength basis.Rational // how strongly they interact
	Axis     string         // which axis is shared ("dough", "sauce", "diagonal", "none")
	Gap      int            // distance on non-shared axis
}

// Couple computes the interaction between two slices.
func (g *Grid) Couple(a, b *Slice) Coupling {
	n1, m1 := int(a.Dough), int(a.Sauce)
	n2, m2 := int(b.Dough), int(b.Sauce)
	d1, d2 := a.Temperature, b.Temperature

	// Same diagonal = same effort shell → parallel, no interaction
	if d1 == d2 && n1 != n2 && m1 != m2 {
		return Coupling{Kind: Parallel, Strength: basis.Rational{Num: 0, Den: 1}, Axis: "diagonal"}
	}

	// Shared dough (same n)
	if n1 == n2 && m1 != m2 {
		gap := abs(m1 - m2)
		kind := DoughCoupled
		if gap == 1 {
			kind = Collision // adjacent → cascade fires
		} else if gap == 2 {
			kind = Resonance // stride-2 → harmonic
		}
		// Strength: F(n_shared) · |F(m1) - F(m2)| for transfer potential
		fn := int64(g.F(n1))
		fm1 := int64(g.F(m1))
		fm2 := int64(g.F(m2))
		diff := fm1 - fm2
		if diff < 0 {
			diff = -diff
		}
		return Coupling{
			Kind:     kind,
			Strength: basis.Rational{Num: fn * diff, Den: 1},
			Axis:     "dough",
			Gap:      gap,
		}
	}

	// Shared sauce (same m)
	if m1 == m2 && n1 != n2 {
		gap := abs(n1 - n2)
		kind := SauceCoupled
		if gap == 1 {
			kind = Collision
		} else if gap == 2 {
			kind = Resonance
		}
		fm := int64(g.F(m1))
		fn1 := int64(g.F(n1))
		fn2 := int64(g.F(n2))
		diff := fn1 - fn2
		if diff < 0 {
			diff = -diff
		}
		return Coupling{
			Kind:     kind,
			Strength: basis.Rational{Num: fm * diff, Den: 1},
			Axis:     "sauce",
			Gap:      gap,
		}
	}

	// No shared axis
	return Coupling{Kind: Parallel, Strength: basis.Rational{Num: 0, Den: 1}, Axis: "none"}
}

// Route returns the next best slice to work on, given what was just completed.
// It finds the strongest-coupled available slice along a shared axis,
// preferring collisions (cascades) over resonance over standard coupling.
func (g *Grid) Route(justDone *Slice) *Slice {
	g.mu.RLock()
	defer g.mu.RUnlock()

	type candidate struct {
		slice    *Slice
		coupling Coupling
		score    int64
	}

	var candidates []candidate
	for _, s := range g.slices {
		if s.ID == justDone.ID || s.Status == Done || s.Status == Burnt {
			continue
		}
		if s.Status != Raw && s.Status != Prepping {
			continue
		}
		c := g.Couple(justDone, s)
		if c.Kind == Parallel {
			continue
		}
		// Score: coupling kind priority + flavor strength
		kindBonus := int64(0)
		switch c.Kind {
		case Collision:
			kindBonus = 1000
		case Resonance:
			kindBonus = 500
		case DoughCoupled, SauceCoupled:
			kindBonus = 100
		}
		score := kindBonus + c.Strength.Num
		candidates = append(candidates, candidate{s, c, score})
	}

	if len(candidates) == 0 {
		return g.bestUncoupled()
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return candidates[0].slice
}

// bestUncoupled returns the highest-flavor raw slice (fallback when no coupling exists).
func (g *Grid) bestUncoupled() *Slice {
	var best *Slice
	var bestFlavor int64
	for _, s := range g.slices {
		if s.Status != Raw {
			continue
		}
		if s.Flavor.Num > bestFlavor || best == nil {
			best = s
			bestFlavor = s.Flavor.Num
		}
	}
	return best
}

// Shell returns all slices on a given temperature diagonal.
func (g *Grid) Shell(temp int) []*Slice {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.byDiag[temp]
}

// Coupled returns all slices that share an axis with the given slice.
func (g *Grid) Coupled(s *Slice) []*Slice {
	g.mu.RLock()
	defer g.mu.RUnlock()
	seen := make(map[string]bool)
	seen[s.ID] = true
	var result []*Slice
	for _, other := range g.byDough[s.Dough] {
		if !seen[other.ID] {
			result = append(result, other)
			seen[other.ID] = true
		}
	}
	for _, other := range g.bySauce[s.Sauce] {
		if !seen[other.ID] {
			result = append(result, other)
			seen[other.ID] = true
		}
	}
	return result
}

// AllSlices returns all slices sorted by flavor intensity descending.
func (g *Grid) AllSlices() []*Slice {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]*Slice, 0, len(g.slices))
	for _, s := range g.slices {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Flavor.Num > result[j].Flavor.Num
	})
	return result
}

// Stats returns a summary of the lattice state.
func (g *Grid) Stats() GridStats {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var st GridStats
	st.Total = len(g.slices)
	for _, s := range g.slices {
		switch s.Status {
		case Raw:
			st.Raw++
		case Prepping:
			st.Prepping++
		case InOven:
			st.InOven++
		case Done:
			st.Done++
		case Burnt:
			st.Burnt++
		}
	}
	st.DoughTypes = len(g.byDough)
	st.SauceTypes = len(g.bySauce)
	st.TempShells = len(g.byDiag)
	return st
}

type GridStats struct {
	Total      int
	Raw        int
	Prepping   int
	InOven     int
	Done       int
	Burnt      int
	DoughTypes int
	SauceTypes int
	TempShells int
}

func (gs GridStats) String() string {
	return fmt.Sprintf(
		"Grid: %d slices (%d raw, %d prepping, %d in-oven, %d done, %d burnt) | %d doughs × %d sauces across %d temp shells",
		gs.Total, gs.Raw, gs.Prepping, gs.InOven, gs.Done, gs.Burnt,
		gs.DoughTypes, gs.SauceTypes, gs.TempShells,
	)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
