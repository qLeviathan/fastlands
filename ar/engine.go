package ar

import (
	"fmt"
	"strings"

	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/kinds"
)

// AtomSpec represents a logical atom.
type AtomSpec struct {
	Predicate string   `json:"predicate"`
	Args      []string `json:"args,omitempty"`
}

// RuleSpec represents a single inference rule.
type RuleSpec struct {
	ID       string     `json:"id"`
	Head     AtomSpec   `json:"head"`
	Body     []AtomSpec `json:"body"`
	Priority int        `json:"priority,omitempty"`
}

// RuleSet is a named collection of rules with metadata.
type RuleSet struct {
	Name       string           `json:"name"`
	Domain     string           `json:"domain"`
	Thresholds map[string]int64 `json:"thresholds"`
	Rules      []RuleSpec       `json:"rules"`
	Labels     []LabelMapping   `json:"labels"`
}

// LabelMapping maps a derived atom to a human-readable label.
type LabelMapping struct {
	Atom  string `json:"atom"`
	Label string `json:"label"`
}

// Engine is the forward-chaining inference engine.
// Each Infer() call creates isolated state — no bleed between calls.
type Engine struct {
	Rules      []RuleSpec
	Labels     map[string]string
	Thresholds map[string]int64
	Domain     string
}

// NewEngine creates an engine from a RuleSet.
func NewEngine(rs RuleSet) *Engine {
	labels := make(map[string]string, len(rs.Labels))
	for _, lm := range rs.Labels {
		labels[lm.Atom] = lm.Label
	}
	thresholds := rs.Thresholds
	if thresholds == nil {
		thresholds = make(map[string]int64)
	}
	return &Engine{
		Rules:      rs.Rules,
		Labels:     labels,
		Thresholds: thresholds,
		Domain:     rs.Domain,
	}
}

// Infer runs forward-chaining on a single datum.
//
//  1. Discretize features into atoms using thresholds (low/mid/high).
//  2. Seed the working set with those atoms.
//  3. Loop: check each rule; if all body atoms are satisfied, derive head.
//  4. Continue until fixpoint (no new atoms produced).
//  5. Map derived atoms to labels.
//  6. Return prediction, confidence, and proof trace.
func (e *Engine) Infer(datum kinds.Datum) (kinds.ModelResult, error) {
	// Phase 1: discretize all features into initial atoms.
	known := make(map[string]bool)
	var trace []string

	for feature, value := range datum.Features {
		level := e.discretize(feature, value)
		key := feature + "=" + level
		known[key] = true
		trace = append(trace, fmt.Sprintf("fact: %s is %s (value=%d)", feature, level, value))
	}

	// Phase 2: forward-chaining worklist loop until fixpoint.
	totalRules := int64(len(e.Rules))
	if totalRules == 0 {
		return kinds.ModelResult{
			Prediction: "unknown",
			Confidence: basis.Rational{Num: 0, Den: 1},
			Kind:       kinds.KindLogicPrograms,
			ProofTrace: trace,
		}, nil
	}

	var firedRules []string
	changed := true
	for changed {
		changed = false
		for idx := range e.Rules {
			rule := &e.Rules[idx]
			headKey := atomKey(rule.Head)

			// Skip if we already derived this head.
			if known[headKey] {
				continue
			}

			// Check if all body atoms are satisfied.
			allSatisfied := true
			for _, bodyAtom := range rule.Body {
				if !known[atomKey(bodyAtom)] {
					allSatisfied = false
					break
				}
			}
			if !allSatisfied {
				continue
			}

			// Derive the head atom.
			known[headKey] = true
			changed = true
			firedRules = append(firedRules, rule.ID)

			// Build a proof step.
			var bodyParts []string
			for _, ba := range rule.Body {
				bodyParts = append(bodyParts, atomKey(ba))
			}
			trace = append(trace, fmt.Sprintf("rule %s: %s => %s",
				rule.ID, strings.Join(bodyParts, " AND "), headKey))
		}
	}

	// Phase 3: determine prediction from derived atoms via label mappings.
	prediction := "unknown"
	for atom, label := range e.Labels {
		if known[atom] {
			prediction = label
			break
		}
	}

	// If no label matched, look for any derived head atoms as prediction.
	if prediction == "unknown" {
		for _, rule := range e.Rules {
			hk := atomKey(rule.Head)
			if known[hk] {
				prediction = hk
				break
			}
		}
	}

	// Phase 4: compute confidence as fired/total, minimum 1/total.
	fired := int64(len(firedRules))
	if fired < 1 {
		fired = 1
	}
	confidence := basis.RatSimplify(basis.Rational{Num: fired, Den: totalRules})

	return kinds.ModelResult{
		Prediction: prediction,
		Confidence: confidence,
		Kind:       kinds.KindLogicPrograms,
		ProofTrace: trace,
	}, nil
}

// InferAll runs inference on all items in a dataset.
func (e *Engine) InferAll(ds kinds.DataSet) []kinds.InferenceResult {
	results := make([]kinds.InferenceResult, 0, len(ds.Items))
	for _, datum := range ds.Items {
		mr, err := e.Infer(datum)
		if err != nil {
			mr = kinds.ModelResult{
				Prediction: "error",
				Confidence: basis.RatZero(),
				Kind:       kinds.KindLogicPrograms,
				ProofTrace: []string{err.Error()},
			}
		}
		results = append(results, kinds.InferenceResult{
			Result:     mr,
			Final:      mr.Prediction,
			Verified:   false,
			Confidence: mr.Confidence,
			Explained:  len(mr.ProofTrace) > 0,
		})
	}
	return results
}

// discretize converts an int64 feature value to "low", "mid", or "high"
// using the engine's thresholds map. The thresholds map is expected to
// contain "low" and "mid" keys; values at or below "low" are "low",
// values above "low" but at or below "mid" are "mid", and the rest
// are "high".
func (e *Engine) discretize(feature string, value int64) string {
	lowThresh, hasLow := e.Thresholds["low"]
	midThresh, hasMid := e.Thresholds["mid"]

	if !hasLow || !hasMid {
		// Fallback: no thresholds defined, everything is "mid".
		return "mid"
	}

	switch {
	case value <= lowThresh:
		return "low"
	case value <= midThresh:
		return "mid"
	default:
		return "high"
	}
}

// atomKey returns a canonical string key for matching atoms.
// Format: predicate(arg1,arg2,...) or just predicate if no args.
func atomKey(a AtomSpec) string {
	if len(a.Args) == 0 {
		return a.Predicate
	}
	return a.Predicate + "(" + strings.Join(a.Args, ",") + ")"
}
