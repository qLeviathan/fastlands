// Package memory provides the episodic memory store for caslo,
// a hybrid AI memory system with JSON-backed persistence.
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/qleviathan/fastlands/basis"
)

// Store is the persistent episodic memory, JSON-backed.
type Store struct {
	Path       string             `json:"-"`
	Facts      []MemoryFact       `json:"facts"`
	RuleStats  []RulePerformance  `json:"rule_stats"`
	RunHistory []RunSummary       `json:"run_history"`
	Patterns   []InferencePattern `json:"patterns"`
	mu         sync.Mutex
}

// MemoryFact is a learned piece of knowledge.
type MemoryFact struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Source   string `json:"source"`
	UseCount int    `json:"use_count"`
}

// RulePerformance tracks per-rule accuracy.
type RulePerformance struct {
	RuleID       string   `json:"rule_id"`
	FireCount    int      `json:"fire_count"`
	CorrectCount int      `json:"correct_count"`
	Domains      []string `json:"domains"`
	// Accuracy is CorrectCount/FireCount as integers - computed on read
}

// Accuracy returns the rule's accuracy as a Rational (no floats).
func (rp RulePerformance) Accuracy() basis.Rational {
	if rp.FireCount == 0 {
		return basis.RatZero()
	}
	return basis.RatSimplify(basis.Rational{
		Num: int64(rp.CorrectCount),
		Den: int64(rp.FireCount),
	})
}

// InferencePattern is a recurring feature->prediction pattern.
type InferencePattern struct {
	Features   []string `json:"features"`   // sorted feature names
	Prediction string   `json:"prediction"`
	ConfNum    int64    `json:"conf_num"`  // confidence numerator
	ConfDen    int64    `json:"conf_den"`  // confidence denominator
	SeenCount  int      `json:"seen_count"`
}

// RunSummary records a past execution.
type RunSummary struct {
	Timestamp  string   `json:"timestamp"`
	Datasets   []string `json:"datasets"`
	TotalItems int      `json:"total_items"`
	CorrectNum int64    `json:"correct_num"` // accuracy numerator
	CorrectDen int64    `json:"correct_den"` // accuracy denominator
	Duration   string   `json:"duration"`
	SwarmPeak  int      `json:"swarm_peak"`
}

// NewStore loads from file or creates empty store.
func NewStore(path string) *Store {
	s := &Store{
		Path:       path,
		Facts:      []MemoryFact{},
		RuleStats:  []RulePerformance{},
		RunHistory: []RunSummary{},
		Patterns:   []InferencePattern{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist or unreadable; return empty store.
		return s
	}

	if err := json.Unmarshal(data, s); err != nil {
		// Corrupt file; return empty store.
		return s
	}

	// Restore the path since it's excluded from JSON.
	s.Path = path
	return s
}

// Save persists to JSON file.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("memory: marshal failed: %w", err)
	}
	if err := os.WriteFile(s.Path, data, 0644); err != nil {
		return fmt.Errorf("memory: write failed: %w", err)
	}
	return nil
}

// RecordRuleFire tracks a rule firing event.
func (s *Store) RecordRuleFire(ruleID string, correct bool, domain string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.RuleStats {
		if s.RuleStats[i].RuleID == ruleID {
			s.RuleStats[i].FireCount++
			if correct {
				s.RuleStats[i].CorrectCount++
			}
			// Add domain if not already present.
			if !containsString(s.RuleStats[i].Domains, domain) {
				s.RuleStats[i].Domains = append(s.RuleStats[i].Domains, domain)
			}
			return
		}
	}

	// New rule.
	correctCount := 0
	if correct {
		correctCount = 1
	}
	s.RuleStats = append(s.RuleStats, RulePerformance{
		RuleID:       ruleID,
		FireCount:    1,
		CorrectCount: correctCount,
		Domains:      []string{domain},
	})
}

// RecordPattern records or updates a feature->prediction pattern.
func (s *Store) RecordPattern(features []string, prediction string, confNum, confDen int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sorted := make([]string, len(features))
	copy(sorted, features)
	sort.Strings(sorted)

	key := strings.Join(sorted, "|")

	for i := range s.Patterns {
		pKey := strings.Join(s.Patterns[i].Features, "|")
		if pKey == key && s.Patterns[i].Prediction == prediction {
			s.Patterns[i].SeenCount++
			s.Patterns[i].ConfNum = confNum
			s.Patterns[i].ConfDen = confDen
			return
		}
	}

	s.Patterns = append(s.Patterns, InferencePattern{
		Features:   sorted,
		Prediction: prediction,
		ConfNum:    confNum,
		ConfDen:    confDen,
		SeenCount:  1,
	})
}

// PredictFromHistory looks up a matching pattern.
// Returns (prediction, confNum, confDen, found).
func (s *Store) PredictFromHistory(features []string) (string, int64, int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sorted := make([]string, len(features))
	copy(sorted, features)
	sort.Strings(sorted)

	key := strings.Join(sorted, "|")

	var best *InferencePattern
	for i := range s.Patterns {
		pKey := strings.Join(s.Patterns[i].Features, "|")
		if pKey == key {
			if best == nil || s.Patterns[i].SeenCount > best.SeenCount {
				best = &s.Patterns[i]
			}
		}
	}

	if best == nil {
		return "", 0, 0, false
	}
	return best.Prediction, best.ConfNum, best.ConfDen, true
}

// LearnFact stores or updates a fact. Deduplicates by key+domain.
func (s *Store) LearnFact(key, value, domain, source string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Facts {
		if s.Facts[i].Key == key && s.Facts[i].Domain == domain {
			s.Facts[i].Value = value
			s.Facts[i].Source = source
			s.Facts[i].UseCount++
			return
		}
	}

	s.Facts = append(s.Facts, MemoryFact{
		Key:      key,
		Value:    value,
		Domain:   domain,
		Source:    source,
		UseCount: 1,
	})
}

// RecordRun adds a run summary.
func (s *Store) RecordRun(datasets []string, totalItems int, correctNum, correctDen int64, duration time.Duration, swarmPeak int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.RunHistory = append(s.RunHistory, RunSummary{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Datasets:   datasets,
		TotalItems: totalItems,
		CorrectNum: correctNum,
		CorrectDen: correctDen,
		Duration:   duration.String(),
		SwarmPeak:  swarmPeak,
	})
}

// GetRuleStats returns all rule performance data.
func (s *Store) GetRuleStats() []RulePerformance {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]RulePerformance, len(s.RuleStats))
	copy(out, s.RuleStats)
	return out
}

// GetPatterns returns all learned patterns.
func (s *Store) GetPatterns() []InferencePattern {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]InferencePattern, len(s.Patterns))
	copy(out, s.Patterns)
	return out
}

// Summary returns a human-readable summary of memory contents.
func (s *Store) Summary() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var b strings.Builder

	b.WriteString(fmt.Sprintf("Memory Store: %s\n", s.Path))
	b.WriteString(fmt.Sprintf("  Facts:    %d\n", len(s.Facts)))
	b.WriteString(fmt.Sprintf("  Rules:    %d\n", len(s.RuleStats)))
	b.WriteString(fmt.Sprintf("  Patterns: %d\n", len(s.Patterns)))
	b.WriteString(fmt.Sprintf("  Runs:     %d\n", len(s.RunHistory)))

	if len(s.RuleStats) > 0 {
		b.WriteString("\nTop rules by fire count:\n")
		// Sort a copy by fire count descending.
		sorted := make([]RulePerformance, len(s.RuleStats))
		copy(sorted, s.RuleStats)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].FireCount > sorted[j].FireCount
		})
		limit := 5
		if len(sorted) < limit {
			limit = len(sorted)
		}
		for _, rp := range sorted[:limit] {
			acc := rp.Accuracy()
			b.WriteString(fmt.Sprintf("  %s: %d fires, accuracy %d/%d, domains %v\n",
				rp.RuleID, rp.FireCount, acc.Num, acc.Den, rp.Domains))
		}
	}

	if len(s.RunHistory) > 0 {
		last := s.RunHistory[len(s.RunHistory)-1]
		b.WriteString(fmt.Sprintf("\nLast run: %s, %d items, accuracy %d/%d, duration %s\n",
			last.Timestamp, last.TotalItems, last.CorrectNum, last.CorrectDen, last.Duration))
	}

	return b.String()
}

// containsString checks if a slice contains a given string.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
