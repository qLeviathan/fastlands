package oeis

import "testing"

func TestCatalogMatch(t *testing.T) {
	c := NewCatalog()

	// Match basis sequence terms
	results := c.Match([]int64{0, 1, 1, 2, 3, 5, 8, 13})
	if len(results) == 0 {
		t.Fatal("expected at least one match for basis sequence")
	}
	if results[0].ID != "A000045" {
		t.Errorf("best match = %s, want A000045", results[0].ID)
	}
}

func TestCatalogValidate(t *testing.T) {
	c := NewCatalog()

	r := c.Validate("A000045", []int64{0, 1, 1, 2, 3, 5})
	if !r.Valid {
		t.Error("expected valid for correct basis terms")
	}

	r = c.Validate("A000045", []int64{0, 1, 1, 2, 3, 99})
	if r.Valid {
		t.Error("expected invalid for wrong term")
	}
}

func TestCatalogPrimes(t *testing.T) {
	c := NewCatalog()
	r := c.Validate("A000040", []int64{2, 3, 5, 7, 11, 13})
	if !r.Valid {
		t.Error("expected valid for prime sequence")
	}
}

func TestCatalogRegister(t *testing.T) {
	c := NewCatalog()
	c.Register(Entry{
		ID:    "A999999",
		Name:  "Test Sequence",
		Terms: []int64{1, 3, 7, 15, 31},
	})
	results := c.Match([]int64{1, 3, 7, 15, 31})
	found := false
	for _, r := range results {
		if r.ID == "A999999" {
			found = true
		}
	}
	if !found {
		t.Error("custom sequence not found in match results")
	}
}

func TestStrengthRatio(t *testing.T) {
	r := strengthRatio(4, 8)
	if r[0] != 1 || r[1] != 2 {
		t.Errorf("strengthRatio(4,8) = %v, want [1,2]", r)
	}
}
