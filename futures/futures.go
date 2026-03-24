// Package futures implements the working memory layer — a predictive cache
// that forecasts inference outcomes before full engine evaluation.
// Predictions are sourced from episodic memory patterns, cached results,
// and chain analysis of partially-satisfied rule bodies.
package futures

import (
	"fmt"
	"sort"
	"sync"

	"github.com/qleviathan/fastlands/ar"
	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/kinds"
	"github.com/qleviathan/fastlands/memory"
)

// Prediction is a cached forecast.
type Prediction struct {
	Label      string
	ConfNum    int64
	ConfDen    int64
	Source     string // "cache", "memory", "chain"
	ChainDepth int
}

// Predictor forecasts inference outcomes using the working memory layer.
type Predictor struct {
	Memory    *memory.Store
	Rules     []ar.RuleSpec
	Labels    map[string]string
	Cache     map[string]Prediction
	MaxDepth  int
	mu        sync.Mutex
	hits      int
	misses    int
}

// NewPredictor creates a predictor linked to memory and rules.
func NewPredictor(mem *memory.Store, rules []ar.RuleSpec, labels map[string]string) *Predictor {
	return &Predictor{
		Memory:   mem,
		Rules:    rules,
		Labels:   labels,
		Cache:    make(map[string]Prediction),
		MaxDepth: 3,
	}
}

// Predict attempts to forecast the outcome for a datum without full inference.
// Checks in order: cache hit, memory pattern match, chain analysis.
func (p *Predictor) Predict(datum kinds.Datum) (Prediction, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := datumKey(datum)

	// Source 1: cache hit
	if pred, ok := p.Cache[key]; ok {
		p.hits++
		return pred, true
	}

	// Source 2: memory pattern match
	features := sortedFeatures(datum)
	if label, num, den, found := p.Memory.PredictFromHistory(features); found {
		pred := Prediction{
			Label:   label,
			ConfNum: num,
			ConfDen: den,
			Source:  "memory",
		}
		p.Cache[key] = pred
		p.hits++
		return pred, true
	}

	// Source 3: chain analysis
	if pred, ok := p.chainAnalyze(datum); ok {
		p.Cache[key] = pred
		p.hits++
		return pred, true
	}

	p.misses++
	return Prediction{}, false
}

// WarmCache pre-loads predictions from memory patterns.
func (p *Predictor) WarmCache(n int) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	patterns := p.Memory.GetPatterns()
	loaded := 0
	for _, pat := range patterns {
		if loaded >= n {
			break
		}
		key := patternKey(pat.Features)
		if _, exists := p.Cache[key]; exists {
			continue
		}
		p.Cache[key] = Prediction{
			Label:   pat.Prediction,
			ConfNum: pat.ConfNum,
			ConfDen: pat.ConfDen,
			Source:  "memory",
		}
		loaded++
	}
	return loaded
}

// Stats returns cache performance counters.
func (p *Predictor) Stats() (hits, misses, cacheSize int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.hits, p.misses, len(p.Cache)
}

// Summary returns a human-readable summary.
func (p *Predictor) Summary() string {
	h, m, sz := p.Stats()
	total := h + m
	if total == 0 {
		return "futures: no predictions attempted"
	}
	hitRate := basis.NewRational(int64(h), int64(total))
	return fmt.Sprintf("futures: %d/%d hits (%d/%d), cache size %d",
		h, total, hitRate.Num, hitRate.Den, sz)
}

// chainAnalyze performs forward chain analysis to predict outcomes.
// Scans rules for partially-satisfied bodies and estimates completion
// probability from memory pattern frequencies.
func (p *Predictor) chainAnalyze(datum kinds.Datum) (Prediction, bool) {
	features := make(map[string]bool)
	for k := range datum.Features {
		features[k] = true
	}

	type candidate struct {
		ruleID    string
		head      string
		satisfied int
		total     int
		depth     int
	}

	var candidates []candidate

	// Check each rule for partial body satisfaction
	for _, rule := range p.Rules {
		satisfied := 0
		for _, bodyAtom := range rule.Body {
			matched := false
			// Check by args (has_feature style)
			if len(bodyAtom.Args) > 0 && features[bodyAtom.Args[0]] {
				matched = true
			}
			// Check by predicate containing feature name (compact style)
			if !matched {
				for fname := range features {
					if len(bodyAtom.Predicate) > 0 && contains(bodyAtom.Predicate, fname) {
						matched = true
						break
					}
				}
			}
			if matched {
				satisfied++
			}
		}
		total := len(rule.Body)
		if total == 0 {
			continue
		}
		// At least half satisfied
		if satisfied*2 >= total {
			candidates = append(candidates, candidate{
				ruleID:    rule.ID,
				head:      rule.Head.Predicate,
				satisfied: satisfied,
				total:     total,
				depth:     1,
			})
		}
	}

	if len(candidates) == 0 {
		return Prediction{}, false
	}

	// Pick best candidate (highest satisfaction ratio, integer comparison)
	best := candidates[0]
	for _, c := range candidates[1:] {
		// c.satisfied/c.total > best.satisfied/best.total
		// => c.satisfied * best.total > best.satisfied * c.total
		if c.satisfied*best.total > best.satisfied*c.total {
			best = c
		}
	}

	label := best.head
	if l, ok := p.Labels[best.head]; ok {
		label = l
	}

	return Prediction{
		Label:      label,
		ConfNum:    int64(best.satisfied),
		ConfDen:    int64(best.total),
		Source:     "chain",
		ChainDepth: best.depth,
	}, true
}

func datumKey(d kinds.Datum) string {
	features := sortedFeatures(d)
	return patternKey(features)
}

func sortedFeatures(d kinds.Datum) []string {
	keys := make([]string, 0, len(d.Features))
	for k := range d.Features {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		result = append(result, fmt.Sprintf("%s=%d", k, d.Features[k]))
	}
	return result
}

func patternKey(features []string) string {
	s := ""
	for i, f := range features {
		if i > 0 {
			s += "|"
		}
		s += f
	}
	return s
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
