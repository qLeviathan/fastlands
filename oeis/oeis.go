// Package oeis provides integer sequence validation against known sequences.
// Sequences are identified by their catalog ID and validated by structural
// properties and term matching.
package oeis

import "fmt"

// Entry is a known integer sequence with metadata.
type Entry struct {
	ID    string   // catalog identifier (e.g. "A000045")
	Name  string   // human-readable name
	Terms []int64  // known terms
	Check func(n int, val int64) bool // optional structural validator
}

// Catalog holds all known sequences for validation.
type Catalog struct {
	entries map[string]Entry
	index   map[int64][]string // first-term → candidate IDs
}

// NewCatalog returns a catalog pre-loaded with foundational sequences.
func NewCatalog() *Catalog {
	c := &Catalog{
		entries: make(map[string]Entry),
		index:   make(map[int64][]string),
	}
	c.loadBuiltins()
	return c
}

// Register adds a custom sequence to the catalog.
func (c *Catalog) Register(e Entry) {
	c.entries[e.ID] = e
	if len(e.Terms) > 0 {
		first := e.Terms[0]
		c.index[first] = append(c.index[first], e.ID)
	}
}

// Lookup returns a sequence entry by ID.
func (c *Catalog) Lookup(id string) (Entry, bool) {
	e, ok := c.entries[id]
	return e, ok
}

// Match attempts to identify an unknown sequence by comparing terms.
// Returns all matching catalog entries sorted by confidence (match count).
func (c *Catalog) Match(terms []int64) []MatchResult {
	if len(terms) == 0 {
		return nil
	}

	var results []MatchResult
	for id, entry := range c.entries {
		matched := matchTerms(entry.Terms, terms)
		if matched >= 3 {
			results = append(results, MatchResult{
				ID:       id,
				Name:     entry.Name,
				Matched:  matched,
				Total:    len(terms),
				Strength: strengthRatio(matched, len(terms)),
			})
		}
	}

	// Sort by matched count descending (insertion sort, small N)
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Matched > results[j-1].Matched; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	return results
}

// Validate checks if a given sequence matches a specific catalog entry.
func (c *Catalog) Validate(id string, terms []int64) ValidationResult {
	entry, ok := c.entries[id]
	if !ok {
		return ValidationResult{Valid: false, Error: fmt.Sprintf("unknown sequence %s", id)}
	}

	matched := matchTerms(entry.Terms, terms)
	structuralOK := true
	if entry.Check != nil {
		for i, v := range terms {
			if !entry.Check(i, v) {
				structuralOK = false
				break
			}
		}
	}

	return ValidationResult{
		Valid:      matched == len(terms) && structuralOK,
		Matched:    matched,
		Total:      len(terms),
		Structural: structuralOK,
	}
}

// MatchResult describes how well an unknown sequence matches a catalog entry.
type MatchResult struct {
	ID       string
	Name     string
	Matched  int
	Total    int
	Strength [2]int64 // rational fraction [num, den]
}

// ValidationResult describes the outcome of validating terms against a known sequence.
type ValidationResult struct {
	Valid      bool
	Matched    int
	Total      int
	Structural bool
	Error      string
}

// matchTerms counts how many terms of candidate appear in reference (positional).
func matchTerms(reference, candidate []int64) int {
	count := 0
	limit := len(reference)
	if len(candidate) < limit {
		limit = len(candidate)
	}
	for i := 0; i < limit; i++ {
		if reference[i] == candidate[i] {
			count++
		}
	}
	return count
}

// strengthRatio returns [matched, total] as a simplified ratio.
func strengthRatio(matched, total int) [2]int64 {
	if total == 0 {
		return [2]int64{0, 1}
	}
	m, t := int64(matched), int64(total)
	g := gcd(m, t)
	return [2]int64{m / g, t / g}
}

func gcd(a, b int64) int64 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}

// loadBuiltins registers foundational integer sequences.
func (c *Catalog) loadBuiltins() {
	// A000045 - basis sequence (each term = sum of two prior)
	c.Register(Entry{
		ID:   "A000045",
		Name: "Basis Sequence",
		Terms: []int64{
			0, 1, 1, 2, 3, 5, 8, 13, 21, 34,
			55, 89, 144, 233, 377, 610, 987, 1597, 2584, 4181,
			6765, 10946, 17711, 28657, 46368, 75025, 121393, 196418, 317811, 514229,
		},
		Check: func(n int, val int64) bool {
			if n <= 1 {
				return val == int64(n)
			}
			return true // structural check would need prior terms
		},
	})

	// A000032 - companion sequence
	c.Register(Entry{
		ID:   "A000032",
		Name: "Companion Sequence",
		Terms: []int64{
			2, 1, 3, 4, 7, 11, 18, 29, 47, 76,
			123, 199, 322, 521, 843, 1364, 2207, 3571, 5778, 9349,
		},
	})

	// A000040 - prime numbers
	c.Register(Entry{
		ID:   "A000040",
		Name: "Prime Numbers",
		Terms: []int64{
			2, 3, 5, 7, 11, 13, 17, 19, 23, 29,
			31, 37, 41, 43, 47, 53, 59, 61, 67, 71,
			73, 79, 83, 89, 97,
		},
	})

	// A000079 - powers of 2
	c.Register(Entry{
		ID:   "A000079",
		Name: "Powers of Two",
		Terms: []int64{
			1, 2, 4, 8, 16, 32, 64, 128, 256, 512,
			1024, 2048, 4096, 8192, 16384, 32768, 65536,
		},
	})

	// A000290 - perfect squares
	c.Register(Entry{
		ID:   "A000290",
		Name: "Perfect Squares",
		Terms: []int64{
			0, 1, 4, 9, 16, 25, 36, 49, 64, 81,
			100, 121, 144, 169, 196, 225, 256, 289, 324, 361,
		},
	})

	// A000217 - triangular numbers
	c.Register(Entry{
		ID:   "A000217",
		Name: "Triangular Numbers",
		Terms: []int64{
			0, 1, 3, 6, 10, 15, 21, 28, 36, 45,
			55, 66, 78, 91, 105, 120, 136, 153, 171, 190,
		},
	})

	// A000142 - factorials
	c.Register(Entry{
		ID:   "A000142",
		Name: "Factorials",
		Terms: []int64{
			1, 1, 2, 6, 24, 120, 720, 5040, 40320, 362880,
			3628800, 39916800, 479001600, 6227020800,
		},
	})

	// A001113 - Catalan numbers
	c.Register(Entry{
		ID:   "A000108",
		Name: "Catalan Numbers",
		Terms: []int64{
			1, 1, 2, 5, 14, 42, 132, 429, 1430, 4862,
			16796, 58786, 208012, 742900,
		},
	})

	// A000110 - Bell numbers
	c.Register(Entry{
		ID:   "A000110",
		Name: "Bell Numbers",
		Terms: []int64{
			1, 1, 2, 5, 15, 52, 203, 877, 4140, 21147,
			115975, 678570, 4213597,
		},
	})

	// A000041 - partition numbers
	c.Register(Entry{
		ID:   "A000041",
		Name: "Partition Numbers",
		Terms: []int64{
			1, 1, 2, 3, 5, 7, 11, 15, 22, 30,
			42, 56, 77, 101, 135, 176, 231,
		},
	})
}
