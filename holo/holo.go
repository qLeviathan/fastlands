// Package holo implements holographic memory for the pizza dispatch system.
//
// Holographic memory stores information as interference patterns — each memory
// is distributed across the entire store, so any fragment can reconstruct
// related memories. This is the key to AR-based learning: patterns compound
// across slices, and completing one task reveals hidden structure in others.
//
// The implementation uses the (n, m) lattice coordinates as the interference
// basis. When you store a memory at position (n, m), it creates ripples along
// both the dough axis (n) and sauce axis (m). Recall works by presenting a
// partial pattern and letting the interference reconstruct the full memory.
package holo

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/qleviathan/fastlands/basis"
)

// Pattern is a single holographic memory trace.
type Pattern struct {
	ID        string
	SliceID   string    // which pizza slice this came from
	N         int       // dough coordinate
	M         int       // sauce coordinate
	Tags      []string  // semantic tags for content-addressable recall
	Content   string    // the actual memory content
	Strength  float64   // interference amplitude (decays over time, refreshes on recall)
	Created   time.Time
	LastRecall time.Time
	RecallCount int
}

// Wave represents the interference pattern from a single memory
// projected onto the lattice. Each (n, m) position gets an amplitude.
type Wave struct {
	Origin    [2]int             // source position
	Amplitudes map[[2]int]float64 // position → amplitude
}

// Store is the holographic memory store.
type Store struct {
	mu       sync.RWMutex
	patterns []*Pattern
	bySlice  map[string][]*Pattern
	byTag    map[string][]*Pattern
	byPos    map[[2]int][]*Pattern
	idMap    map[string]*Pattern
	nextID   int
}

// NewStore creates an empty holographic memory.
func NewStore() *Store {
	return &Store{
		bySlice: make(map[string][]*Pattern),
		byTag:   make(map[string][]*Pattern),
		byPos:   make(map[[2]int][]*Pattern),
		idMap:   make(map[string]*Pattern),
	}
}

// Encode stores a new memory pattern with interference across the lattice.
func (s *Store) Encode(sliceID string, n, m int, tags []string, content string) *Pattern {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	p := &Pattern{
		ID:        fmt.Sprintf("holo-%04d", s.nextID),
		SliceID:   sliceID,
		N:         n,
		M:         m,
		Tags:      tags,
		Content:   content,
		Strength:  1.0,
		Created:   time.Now(),
		LastRecall: time.Now(),
	}

	s.patterns = append(s.patterns, p)
	s.idMap[p.ID] = p
	s.bySlice[sliceID] = append(s.bySlice[sliceID], p)
	pos := [2]int{n, m}
	s.byPos[pos] = append(s.byPos[pos], p)
	for _, tag := range tags {
		s.byTag[tag] = append(s.byTag[tag], p)
	}

	return p
}

// Recall retrieves memories by interference. Given partial coordinates and/or tags,
// it finds all patterns that resonate and returns them ranked by interference strength.
func (s *Store) Recall(n, m int, tags []string) []*Pattern {
	s.mu.Lock()
	defer s.mu.Unlock()

	scores := make(map[string]float64)

	for _, p := range s.patterns {
		score := 0.0

		// Spatial interference: closer positions resonate more strongly
		// Uses basis-weighted distance on the lattice
		dn := abs(p.N - n)
		dm := abs(p.M - m)

		if dn == 0 && dm == 0 {
			score += 10.0 // exact position match
		} else if dn == 0 || dm == 0 {
			// Shared axis — strong coupling
			gap := dn + dm
			if gap == 1 {
				score += 8.0 // collision coupling
			} else if gap == 2 {
				score += 5.0 // resonance coupling
			} else {
				score += 2.0 / float64(gap) // harmonic decay
			}
		} else {
			// Different axes — interference based on diagonal proximity
			dDiag := abs((p.N + p.M) - (n + m))
			if dDiag == 0 {
				score += 3.0 // same temperature shell
			} else {
				score += 1.0 / float64(dDiag+1)
			}
		}

		// Tag interference: matching tags amplify
		tagMatch := 0
		for _, qt := range tags {
			for _, pt := range p.Tags {
				if qt == pt {
					tagMatch++
				}
			}
		}
		score += float64(tagMatch) * 3.0

		// Strength modulation (recent + frequently recalled = stronger)
		score *= p.Strength

		// Decay: older memories fade unless refreshed
		age := time.Since(p.LastRecall).Hours()
		decay := math.Exp(-age / (24 * 7)) // half-life of ~1 week
		score *= (0.3 + 0.7*decay)          // floor at 30% even for old memories

		if score > 0.1 {
			scores[p.ID] = score
		}
	}

	// Collect and sort by score
	type scored struct {
		pattern *Pattern
		score   float64
	}
	var results []scored
	for id, sc := range scores {
		p := s.idMap[id]
		results = append(results, scored{p, sc})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Refresh recalled patterns (holographic reinforcement)
	out := make([]*Pattern, 0, len(results))
	for _, r := range results {
		r.pattern.LastRecall = time.Now()
		r.pattern.RecallCount++
		r.pattern.Strength = math.Min(2.0, r.pattern.Strength+0.05)
		out = append(out, r.pattern)
	}

	return out
}

// RecallByTag retrieves all patterns with a given tag.
func (s *Store) RecallByTag(tag string) []*Pattern {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byTag[tag]
}

// RecallBySlice retrieves all patterns for a given pizza slice.
func (s *Store) RecallBySlice(sliceID string) []*Pattern {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bySlice[sliceID]
}

// Project creates an interference wave from a source position,
// showing how strongly it would resonate with each lattice point.
// Uses basis sequence for the amplitude envelope.
func (s *Store) Project(n, m int, radius int) Wave {
	seq := basis.Sequence(radius*2 + 1)
	w := Wave{
		Origin:     [2]int{n, m},
		Amplitudes: make(map[[2]int]float64),
	}

	for dn := -radius; dn <= radius; dn++ {
		for dm := -radius; dm <= radius; dm++ {
			pn := n + dn
			pm := m + dm
			if pn < 0 || pm < 0 {
				continue
			}

			dist := abs(dn) + abs(dm)
			if dist == 0 {
				w.Amplitudes[[2]int{pn, pm}] = 1.0
				continue
			}

			// Amplitude falls off as 1/F(dist+1), with axis-aligned bonus
			fDist := seq[min(dist+1, len(seq)-1)]
			if fDist == 0 {
				fDist = 1
			}
			amp := 1.0 / float64(fDist)

			// Axis-aligned positions get amplified (coupling)
			if dn == 0 || dm == 0 {
				amp *= 2.0
			}

			// Same diagonal gets a phase boost
			if (pn + pm) == (n + m) {
				amp *= 1.5
			}

			w.Amplitudes[[2]int{pn, pm}] = amp
		}
	}

	return w
}

// Interfere combines two waves, creating constructive/destructive interference.
func Interfere(a, b Wave) Wave {
	result := Wave{
		Origin:     a.Origin,
		Amplitudes: make(map[[2]int]float64),
	}
	// Copy a
	for pos, amp := range a.Amplitudes {
		result.Amplitudes[pos] = amp
	}
	// Superpose b
	for pos, amp := range b.Amplitudes {
		result.Amplitudes[pos] += amp
	}
	return result
}

// LearnFromCompletion encodes the experience of completing a slice.
// This creates a persistent memory that future routing can leverage.
func (s *Store) LearnFromCompletion(sliceID string, n, m int, outcome string, durationMin int, notes string) {
	tags := []string{"completed", outcome}
	if durationMin < 60 {
		tags = append(tags, "quick")
	} else if durationMin > 240 {
		tags = append(tags, "deep-work")
	}

	content := fmt.Sprintf("Completed %s in %d min: %s. %s", sliceID, durationMin, outcome, notes)
	s.Encode(sliceID, n, m, tags, content)

	// Encode a separate "skill" pattern along the dough axis
	s.Encode(sliceID, n, 0, []string{"skill-growth", fmt.Sprintf("dough-%d", n)},
		fmt.Sprintf("Skill reinforcement on dough axis %d from %s", n, sliceID))

	// Encode a "platform" pattern along the sauce axis
	s.Encode(sliceID, 0, m, []string{"platform-exp", fmt.Sprintf("sauce-%d", m)},
		fmt.Sprintf("Platform experience on sauce axis %d from %s", m, sliceID))
}

// SuggestNext uses holographic recall to suggest what to do next.
// It projects from the most recent completion and finds the strongest resonance.
func (s *Store) SuggestNext(lastN, lastM int) string {
	recalled := s.Recall(lastN, lastM, []string{"completed"})
	if len(recalled) == 0 {
		return "No memory yet — start with the lowest temperature slice available."
	}

	// Find the pattern with highest recall that ISN'T on the same exact position
	for _, p := range recalled {
		if p.N != lastN || p.M != lastM {
			return fmt.Sprintf("Holographic recall suggests position (%d, %d): %s", p.N, p.M, p.Content)
		}
	}

	return "All recalled patterns are at the same position — try a new axis."
}

// Stats returns memory statistics.
type Stats struct {
	TotalPatterns int
	UniqueSlices  int
	UniqueTags    int
	AvgStrength   float64
	OldestAge     time.Duration
	MostRecalled  string
}

func (s *Store) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st := Stats{
		TotalPatterns: len(s.patterns),
		UniqueSlices:  len(s.bySlice),
		UniqueTags:    len(s.byTag),
	}

	if len(s.patterns) > 0 {
		totalStr := 0.0
		maxRecall := 0
		for _, p := range s.patterns {
			totalStr += p.Strength
			age := time.Since(p.Created)
			if age > st.OldestAge {
				st.OldestAge = age
			}
			if p.RecallCount > maxRecall {
				maxRecall = p.RecallCount
				st.MostRecalled = p.ID
			}
		}
		st.AvgStrength = totalStr / float64(len(s.patterns))
	}

	return st
}

func (st Stats) String() string {
	return fmt.Sprintf("Holo: %d patterns, %d slices, %d tags, avg strength %.2f",
		st.TotalPatterns, st.UniqueSlices, st.UniqueTags, st.AvgStrength)
}

// FormatWave renders a wave as a simple grid visualization.
func FormatWave(w Wave, maxN, maxM int) string {
	var sb strings.Builder
	sb.WriteString("Interference pattern:\n")

	for n := 0; n <= maxN; n++ {
		for m := 0; m <= maxM; m++ {
			amp, ok := w.Amplitudes[[2]int{n, m}]
			if !ok || amp < 0.01 {
				sb.WriteString("  ·  ")
			} else if amp >= 0.8 {
				sb.WriteString(" ███ ")
			} else if amp >= 0.5 {
				sb.WriteString(" ▓▓▓ ")
			} else if amp >= 0.3 {
				sb.WriteString(" ▒▒▒ ")
			} else if amp >= 0.1 {
				sb.WriteString(" ░░░ ")
			} else {
				sb.WriteString("  ·  ")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
