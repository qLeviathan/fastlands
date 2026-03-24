package ar

import (
	"testing"

	"github.com/qleviathan/fastlands/kinds"
)

func TestDefaultRuleSet(t *testing.T) {
	rs := DefaultRuleSet()
	if len(rs.Rules) == 0 {
		t.Fatal("DefaultRuleSet has no rules")
	}
	if rs.Domain == "" {
		t.Error("DefaultRuleSet has no domain")
	}
}

func TestEngineInfer(t *testing.T) {
	engine := NewEngine(DefaultRuleSet())

	cases := []struct {
		name     string
		features map[string]int64
		want     string
	}{
		{
			name:     "high threat low supply",
			features: map[string]int64{"threat_level": 8, "supply_available": 2},
			want:     "defend",
		},
		{
			name:     "low threat high supply",
			features: map[string]int64{"threat_level": 2, "supply_available": 8},
			want:     "expand",
		},
		{
			name:     "mid threat mid supply",
			features: map[string]int64{"threat_level": 5, "supply_available": 5},
			want:     "hold",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			datum := kinds.Datum{Features: c.features, Label: c.want}
			result, err := engine.Infer(datum)
			if err != nil {
				t.Fatalf("Infer error: %v", err)
			}
			if result.Prediction != c.want {
				t.Errorf("prediction = %s, want %s", result.Prediction, c.want)
			}
			if len(result.ProofTrace) == 0 {
				t.Error("empty proof trace")
			}
			if result.Confidence.Den <= 0 {
				t.Error("invalid confidence denominator")
			}
		})
	}
}

func TestEngineInferAll(t *testing.T) {
	engine := NewEngine(DefaultRuleSet())
	ds := kinds.DataSet{
		Name:   "test",
		Domain: "logistics",
		Items: []kinds.Datum{
			{Features: map[string]int64{"threat_level": 9, "supply_available": 1}, Label: "defend"},
			{Features: map[string]int64{"threat_level": 1, "supply_available": 9}, Label: "expand"},
		},
	}

	results := engine.InferAll(ds)
	if len(results) != 2 {
		t.Fatalf("InferAll returned %d results, want 2", len(results))
	}
}
