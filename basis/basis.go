// Package basis provides the core encoding system for caslo.
//
// It implements basis sequence generation via iterative accumulation,
// canonical form decomposition into non-consecutive terms, exact rational
// arithmetic (integer-only, no floating point), and fast doubling for
// O(log n) term computation.
package basis

// Canonical represents a number decomposed into non-consecutive basis terms.
type Canonical struct {
	Bits  uint64   // bitmask: bit i means basis term i is included
	Value uint64   // the original value
	Terms []uint64 // the actual basis terms included
}

// Decompose breaks a positive integer into canonical form using a greedy
// approach, selecting non-consecutive basis terms from largest to smallest.
func Decompose(n uint64) Canonical {
	if n == 0 {
		return Canonical{Value: 0}
	}

	// Build basis terms up to n via accumulation.
	terms := []uint64{1, 1}
	for {
		next := terms[len(terms)-1] + terms[len(terms)-2]
		if next > n {
			break
		}
		terms = append(terms, next)
	}

	remaining := n
	var bits uint64
	var included []uint64
	lastUsed := len(terms) + 2 // track index to enforce non-consecutive constraint

	for i := len(terms) - 1; i >= 0 && remaining > 0; i-- {
		if terms[i] <= remaining && lastUsed-i > 1 {
			remaining -= terms[i]
			bits |= 1 << uint(i)
			included = append(included, terms[i])
			lastUsed = i
		}
	}

	return Canonical{
		Bits:  bits,
		Value: n,
		Terms: included,
	}
}

// ---------------------------------------------------------------------------
// Rational arithmetic — exact fractional math, NO floating point.
// ---------------------------------------------------------------------------

// Rational represents an exact fraction using integer-only arithmetic.
type Rational struct {
	Num int64
	Den int64
}

// GCD computes the greatest common divisor using the iterative Euclidean
// algorithm (fold-style reduction).
func GCD(a, b int64) int64 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// RatSimplify reduces a rational to lowest terms and normalises the sign
// so that the denominator is always positive.
func RatSimplify(r Rational) Rational {
	if r.Den == 0 {
		panic("basis: rational with zero denominator")
	}
	if r.Num == 0 {
		return Rational{Num: 0, Den: 1}
	}
	g := GCD(r.Num, r.Den)
	r.Num /= g
	r.Den /= g
	if r.Den < 0 {
		r.Num = -r.Num
		r.Den = -r.Den
	}
	return r
}

// NewRational creates a simplified rational from numerator and denominator.
func NewRational(num, den int64) Rational {
	if den == 0 {
		den = 1
	}
	return RatSimplify(Rational{Num: num, Den: den})
}

// RatZero returns the rational number 0/1.
func RatZero() Rational {
	return Rational{Num: 0, Den: 1}
}

// RatOne returns the rational number 1/1.
func RatOne() Rational {
	return Rational{Num: 1, Den: 1}
}

// RatFromInt creates a rational from an integer value.
func RatFromInt(n int64) Rational {
	return Rational{Num: n, Den: 1}
}

// RatAdd returns the sum a + b in simplified form.
func RatAdd(a, b Rational) Rational {
	return RatSimplify(Rational{
		Num: a.Num*b.Den + b.Num*a.Den,
		Den: a.Den * b.Den,
	})
}

// RatSub returns the difference a - b in simplified form.
func RatSub(a, b Rational) Rational {
	return RatSimplify(Rational{
		Num: a.Num*b.Den - b.Num*a.Den,
		Den: a.Den * b.Den,
	})
}

// RatMul returns the product a * b in simplified form.
func RatMul(a, b Rational) Rational {
	return RatSimplify(Rational{
		Num: a.Num * b.Num,
		Den: a.Den * b.Den,
	})
}

// RatDiv returns the quotient a / b in simplified form.
func RatDiv(a, b Rational) Rational {
	if b.Num == 0 {
		panic("basis: division by zero rational")
	}
	return RatSimplify(Rational{
		Num: a.Num * b.Den,
		Den: a.Den * b.Num,
	})
}

// RatCompare returns -1 if a < b, 0 if a == b, +1 if a > b.
// Uses only integer cross-multiplication to avoid any floating point.
func RatCompare(a, b Rational) int {
	a = RatSimplify(a)
	b = RatSimplify(b)
	lhs := a.Num * b.Den
	rhs := b.Num * a.Den
	switch {
	case lhs < rhs:
		return -1
	case lhs > rhs:
		return 1
	default:
		return 0
	}
}

// RatNeg returns the negation of r.
func RatNeg(r Rational) Rational {
	return Rational{Num: -r.Num, Den: r.Den}
}

// RatAbs returns the absolute value of r.
func RatAbs(r Rational) Rational {
	n := r.Num
	if n < 0 {
		n = -n
	}
	return RatSimplify(Rational{Num: n, Den: r.Den})
}

// ---------------------------------------------------------------------------
// Fast doubling — O(log n) computation of the nth basis term.
// ---------------------------------------------------------------------------

// Term returns the nth basis term using integer-only fast doubling.
//
// The doubling identities used are:
//
//	S(2k)   = S(k) * (2*S(k+1) - S(k))
//	S(2k+1) = S(k)^2 + S(k+1)^2
//
// where basis indexing is: S(0)=0, S(1)=1, S(2)=1, S(3)=2, ...
// We shift by one so that Term(0)=1, Term(1)=1, Term(2)=2, etc.
func Term(n uint64) uint64 {
	a, b := fastDouble(n)
	_ = b
	return a
}

// fastDouble is the core closed-form evaluation engine. It returns
// (S(n), S(n+1)) where S(0)=0, S(1)=1, using iterative bit scanning.
func fastDouble(n uint64) (uint64, uint64) {
	if n == 0 {
		return 0, 1
	}

	// Find the highest set bit position.
	var bits []bool
	for v := n; v > 0; v >>= 1 {
		bits = append(bits, v&1 == 1)
	}

	a, b := uint64(0), uint64(1) // S(0), S(1)

	// Scan from the most significant bit downward (accumulation pass).
	for i := len(bits) - 1; i >= 0; i-- {
		// Doubling step: compute S(2k) and S(2k+1).
		c := a * (2*b - a)  // S(2k) = S(k) * (2*S(k+1) - S(k))
		d := a*a + b*b      // S(2k+1) = S(k)^2 + S(k+1)^2

		if bits[i] {
			// Current bit is 1: advance to (S(2k+1), S(2k+2)).
			a, b = d, c+d
		} else {
			// Current bit is 0: stay at (S(2k), S(2k+1)).
			a, b = c, d
		}
	}

	return a, b
}

// CompanionTerm returns the nth companion sequence term.
// Uses the identity C(n) = 2*S(n+1) - S(n) where S is the basis sequence
// with S(0)=0, S(1)=1.
// Companion values: C(0)=2, C(1)=1, C(2)=3, C(3)=4, C(4)=7, C(5)=11, ...
func CompanionTerm(n uint64) uint64 {
	a, b := fastDouble(n) // S(n), S(n+1)
	return 2*b - a
}

// Sequence returns the first count basis terms via iterative accumulation.
// basis[0]=1, basis[1]=1, basis[2]=2, basis[3]=3, basis[4]=5, basis[5]=8, ...
func Sequence(count int) []uint64 {
	if count <= 0 {
		return nil
	}
	seq := make([]uint64, count)
	seq[0] = 1
	if count == 1 {
		return seq
	}
	seq[1] = 1
	// Chain evaluation: each term is the sum of the two preceding terms.
	for i := 2; i < count; i++ {
		seq[i] = seq[i-1] + seq[i-2]
	}
	return seq
}

// ---------------------------------------------------------------------------
// Weight — canonical basis encoding for routing.
// ---------------------------------------------------------------------------

// Weight encodes a value in canonical basis representation for routing.
type Weight struct {
	Rank   int       // index of highest basis term in the decomposition
	Spread int       // number of terms in the decomposition
	Canon  Canonical // full canonical form
}

// Encode produces a Weight from a positive integer by decomposing it into
// canonical form and extracting routing metadata.
func Encode(n uint64) Weight {
	c := Decompose(n)
	if len(c.Terms) == 0 {
		return Weight{Rank: 0, Spread: 0, Canon: c}
	}

	// Find the highest set bit in the bitmask to determine rank.
	rank := 0
	for b := c.Bits; b > 0; b >>= 1 {
		if b&1 == 1 {
			// This bit position is set; keep tracking the highest.
		}
		rank++
	}
	// rank is now one past the highest bit; subtract 1.
	rank--

	// More precise: find MSB position.
	rank = msb(c.Bits)

	return Weight{
		Rank:   rank,
		Spread: len(c.Terms),
		Canon:  c,
	}
}

// Route returns the routing key from a weight, which is the index of the
// highest basis term in the canonical decomposition.
func (w Weight) Route() int {
	return w.Rank
}

// msb returns the position of the most significant set bit (0-indexed).
// Returns 0 for input 0.
func msb(v uint64) int {
	if v == 0 {
		return 0
	}
	pos := 0
	for v >>= 1; v > 0; v >>= 1 {
		pos++
	}
	return pos
}
