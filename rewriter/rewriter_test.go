package rewriter

import (
	"testing"

	"github.com/qleviathan/fastlands/memory"
)

func TestRewriterEvaluate(t *testing.T) {
	rw := New()
	store := memory.NewStore("")

	// Add a high-performing rule
	for i := 0; i < 10; i++ {
		store.RecordRuleFire("good-r1", true, "test")
	}

	// Add a poorly-performing rule
	for i := 0; i < 12; i++ {
		store.RecordRuleFire("bad-r1", false, "test")
	}

	actions := rw.Evaluate(store)
	if len(actions) == 0 {
		t.Fatal("expected rewrite actions")
	}

	// Check that good rule is strengthened
	found := false
	for _, a := range actions {
		if a.RuleID == "good-r1" && a.Kind == ActionStrengthen {
			found = true
		}
	}
	if !found {
		t.Error("expected strengthen action for good-r1")
	}

	// Check that bad rule gets retire or weaken
	foundBad := false
	for _, a := range actions {
		if a.RuleID == "bad-r1" && (a.Kind == ActionRetire || a.Kind == ActionWeaken) {
			foundBad = true
		}
	}
	if !foundBad {
		t.Error("expected retire/weaken action for bad-r1")
	}
}

func TestProposeFromPatterns(t *testing.T) {
	rw := New()
	store := memory.NewStore("")

	// Record a strong pattern seen many times
	for i := 0; i < 6; i++ {
		store.RecordPattern([]string{"temp", "humidity"}, "storm", 9, 10)
	}

	existing := map[string]bool{}
	proposals := rw.ProposeFromPatterns(store, existing)
	if len(proposals) == 0 {
		t.Fatal("expected proposed rules from strong pattern")
	}
	if proposals[0].Kind != ActionPropose {
		t.Errorf("action kind = %s, want propose", proposals[0].Kind)
	}
}

func TestRewriterSummary(t *testing.T) {
	rw := New()
	s := rw.Summary()
	if s != "rewriter: no actions taken" {
		t.Errorf("unexpected summary: %s", s)
	}
}
