package ar

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadRulesFromFile loads a RuleSet from a JSON file and returns an Engine.
func LoadRulesFromFile(path string) (*Engine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ar: failed to read rules file %s: %w", path, err)
	}
	return LoadRulesFromJSON(data)
}

// LoadRulesFromJSON loads a RuleSet from raw JSON bytes and returns an Engine.
func LoadRulesFromJSON(data []byte) (*Engine, error) {
	var rs RuleSet
	if err := json.Unmarshal(data, &rs); err != nil {
		return nil, fmt.Errorf("ar: failed to parse rules JSON: %w", err)
	}
	if len(rs.Rules) == 0 {
		return nil, fmt.Errorf("ar: rule set %q contains no rules", rs.Name)
	}
	return NewEngine(rs), nil
}

// DefaultRuleSet returns built-in rules for testing in the supply chain domain.
// Thresholds: low <= 3, mid <= 5, high > 5 (using values low:3, mid:5, high:7).
//
// Rules encode strategic decisions based on threat level and supply availability:
//   - high threat + low supply  => defend
//   - low threat  + high supply => expand
//   - mid threat  + mid supply  => hold
//   - high threat + mid supply  => fortify
//   - low threat  + mid supply  => scout
//   - mid threat  + low supply  => conserve
func DefaultRuleSet() RuleSet {
	return RuleSet{
		Name:   "supply-chain-strategy",
		Domain: "supply-chain",
		Thresholds: map[string]int64{
			"low":  3,
			"mid":  5,
			"high": 7,
		},
		Rules: []RuleSpec{
			{
				ID: "R1-defend",
				Head: AtomSpec{Predicate: "action(defend)"},
				Body: []AtomSpec{
					{Predicate: "threat_level=high"},
					{Predicate: "supply_available=low"},
				},
				Priority: 1,
			},
			{
				ID: "R2-expand",
				Head: AtomSpec{Predicate: "action(expand)"},
				Body: []AtomSpec{
					{Predicate: "threat_level=low"},
					{Predicate: "supply_available=high"},
				},
				Priority: 1,
			},
			{
				ID: "R3-hold",
				Head: AtomSpec{Predicate: "action(hold)"},
				Body: []AtomSpec{
					{Predicate: "threat_level=mid"},
					{Predicate: "supply_available=mid"},
				},
				Priority: 1,
			},
			{
				ID: "R4-fortify",
				Head: AtomSpec{Predicate: "action(fortify)"},
				Body: []AtomSpec{
					{Predicate: "threat_level=high"},
					{Predicate: "supply_available=mid"},
				},
				Priority: 2,
			},
			{
				ID: "R5-scout",
				Head: AtomSpec{Predicate: "action(scout)"},
				Body: []AtomSpec{
					{Predicate: "threat_level=low"},
					{Predicate: "supply_available=mid"},
				},
				Priority: 2,
			},
			{
				ID: "R6-conserve",
				Head: AtomSpec{Predicate: "action(conserve)"},
				Body: []AtomSpec{
					{Predicate: "threat_level=mid"},
					{Predicate: "supply_available=low"},
				},
				Priority: 2,
			},
		},
		Labels: []LabelMapping{
			{Atom: "action(defend)", Label: "defend"},
			{Atom: "action(expand)", Label: "expand"},
			{Atom: "action(hold)", Label: "hold"},
			{Atom: "action(fortify)", Label: "fortify"},
			{Atom: "action(scout)", Label: "scout"},
			{Atom: "action(conserve)", Label: "conserve"},
		},
	}
}
