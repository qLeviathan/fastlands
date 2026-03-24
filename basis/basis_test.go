package basis

import "testing"

func TestSequence(t *testing.T) {
	seq := Sequence(10)
	want := []uint64{1, 1, 2, 3, 5, 8, 13, 21, 34, 55}
	if len(seq) != len(want) {
		t.Fatalf("Sequence(10): got %d terms, want %d", len(seq), len(want))
	}
	for i, v := range want {
		if seq[i] != v {
			t.Errorf("Sequence(10)[%d] = %d, want %d", i, seq[i], v)
		}
	}
}

func TestTerm(t *testing.T) {
	cases := []struct {
		n    uint64
		want uint64
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 5},
		{10, 55},
		{20, 6765},
	}
	for _, c := range cases {
		got := Term(c.n)
		if got != c.want {
			t.Errorf("Term(%d) = %d, want %d", c.n, got, c.want)
		}
	}
}

func TestCompanionTerm(t *testing.T) {
	// Companion: 2, 1, 3, 4, 7, 11, 18, 29, 47, 76
	cases := []struct {
		n    uint64
		want uint64
	}{
		{0, 2},
		{1, 1},
		{2, 3},
		{3, 4},
		{4, 7},
		{5, 11},
	}
	for _, c := range cases {
		got := CompanionTerm(c.n)
		if got != c.want {
			t.Errorf("CompanionTerm(%d) = %d, want %d", c.n, got, c.want)
		}
	}
}

func TestDecompose(t *testing.T) {
	cases := []struct {
		n     uint64
		terms []uint64
	}{
		{0, nil},
		{1, []uint64{1}},
		{5, []uint64{5}},
		{10, []uint64{8, 2}},
		{20, []uint64{13, 5, 2}},
	}
	for _, c := range cases {
		got := Decompose(c.n)
		if got.Value != c.n {
			t.Errorf("Decompose(%d).Value = %d", c.n, got.Value)
		}
		if len(got.Terms) != len(c.terms) {
			t.Errorf("Decompose(%d).Terms = %v, want %v", c.n, got.Terms, c.terms)
			continue
		}
		for i, v := range c.terms {
			if got.Terms[i] != v {
				t.Errorf("Decompose(%d).Terms[%d] = %d, want %d", c.n, i, got.Terms[i], v)
			}
		}
	}
}

func TestRational(t *testing.T) {
	a := NewRational(3, 6)
	if a.Num != 1 || a.Den != 2 {
		t.Errorf("NewRational(3,6) = %d/%d, want 1/2", a.Num, a.Den)
	}

	b := NewRational(2, 3)
	sum := RatAdd(a, b)
	sum = RatSimplify(sum)
	if sum.Num != 7 || sum.Den != 6 {
		t.Errorf("1/2 + 2/3 = %d/%d, want 7/6", sum.Num, sum.Den)
	}

	prod := RatMul(a, b)
	prod = RatSimplify(prod)
	if prod.Num != 1 || prod.Den != 3 {
		t.Errorf("1/2 * 2/3 = %d/%d, want 1/3", prod.Num, prod.Den)
	}

	cmp := RatCompare(a, b)
	if cmp >= 0 {
		t.Errorf("RatCompare(1/2, 2/3) = %d, want < 0", cmp)
	}
}

func TestEncode(t *testing.T) {
	w := Encode(10)
	if w.Canon.Value != 10 {
		t.Errorf("Encode(10).Canon.Value = %d", w.Canon.Value)
	}
	if w.Rank < 0 {
		t.Errorf("Encode(10).Rank = %d, want >= 0", w.Rank)
	}
	if w.Spread != len(w.Canon.Terms) {
		t.Errorf("Encode(10).Spread = %d, want %d", w.Spread, len(w.Canon.Terms))
	}
}
