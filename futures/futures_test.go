package futures

import (
	"testing"

	"github.com/qleviathan/fastlands/ar"
	"github.com/qleviathan/fastlands/kinds"
	"github.com/qleviathan/fastlands/memory"
)

func TestPredictorWarmCache(t *testing.T) {
	store := memory.NewStore("")
	store.RecordPattern([]string{"threat_level=8", "supply=2"}, "defend", 4, 5)
	store.RecordPattern([]string{"threat_level=2", "supply=8"}, "expand", 3, 5)

	rules := ar.DefaultRuleSet().Rules
	p := NewPredictor(store, rules, map[string]string{})

	loaded := p.WarmCache(10)
	if loaded != 2 {
		t.Errorf("WarmCache loaded %d, want 2", loaded)
	}

	_, _, sz := p.Stats()
	if sz != 2 {
		t.Errorf("cache size = %d, want 2", sz)
	}
}

func TestPredictorChainAnalysis(t *testing.T) {
	store := memory.NewStore("")
	rs := ar.DefaultRuleSet()
	labels := make(map[string]string)
	for _, lm := range rs.Labels {
		labels[lm.Atom] = lm.Label
	}

	p := NewPredictor(store, rs.Rules, labels)

	datum := kinds.Datum{
		Features: map[string]int64{"threat_level": 8, "supply_available": 2},
	}

	pred, found := p.Predict(datum)
	if !found {
		t.Fatal("expected chain analysis to produce prediction")
	}
	if pred.Source != "chain" {
		t.Errorf("source = %s, want chain", pred.Source)
	}
}

func TestPredictorSummary(t *testing.T) {
	store := memory.NewStore("")
	p := NewPredictor(store, nil, nil)
	s := p.Summary()
	if s != "futures: no predictions attempted" {
		t.Errorf("unexpected summary: %s", s)
	}
}
