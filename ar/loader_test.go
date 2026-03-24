package ar

import (
	"encoding/json"
	"testing"
)

func TestLoadRulesFromJSON(t *testing.T) {
	rs := DefaultRuleSet()
	data, err := json.Marshal(rs)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	engine, err := LoadRulesFromJSON(data)
	if err != nil {
		t.Fatalf("LoadRulesFromJSON error: %v", err)
	}

	if len(engine.Rules) != len(rs.Rules) {
		t.Errorf("loaded %d rules, want %d", len(engine.Rules), len(rs.Rules))
	}
}

func TestDefaultRuleSetValid(t *testing.T) {
	rs := DefaultRuleSet()
	for _, rule := range rs.Rules {
		if rule.ID == "" {
			t.Error("rule has empty ID")
		}
		if rule.Head.Predicate == "" {
			t.Errorf("rule %s has empty head predicate", rule.ID)
		}
		if len(rule.Body) == 0 {
			t.Errorf("rule %s has empty body", rule.ID)
		}
	}
}
