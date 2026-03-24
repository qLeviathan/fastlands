package self

import (
	"strings"
	"testing"

	"github.com/qleviathan/fastlands/basis"
)

func TestNewIdentity(t *testing.T) {
	id := NewIdentity()
	if id.Name != "Caslo" {
		t.Errorf("expected name Caslo, got %s", id.Name)
	}
	if len(id.Capabilities) == 0 {
		t.Error("expected capabilities, got none")
	}
	if len(id.Components) == 0 {
		t.Error("expected components, got none")
	}
}

func TestIdentityDescribe(t *testing.T) {
	id := NewIdentity()
	id.UpdateState(SystemState{
		RunCount:      3,
		PatternsKnown: 10,
		RulesActive:   5,
		FactsLearned:  7,
		AgentsAlive:   2,
		AgentsPeak:    4,
		Accuracy:      basis.NewRational(4, 6),
	})

	desc := id.Describe()
	if !strings.Contains(desc, "Caslo") {
		t.Error("description should contain Caslo")
	}
	if !strings.Contains(desc, "Forward-Chain Inference") {
		t.Error("description should list capabilities")
	}
	if !strings.Contains(desc, "Runs:     3") {
		t.Error("description should show run count")
	}
}

func TestFindCapability(t *testing.T) {
	id := NewIdentity()
	cap, ok := id.FindCapability("ar")
	if !ok {
		t.Fatal("expected to find ar capability")
	}
	if cap.Name != "Forward-Chain Inference" {
		t.Errorf("unexpected capability name: %s", cap.Name)
	}

	_, ok = id.FindCapability("nonexistent")
	if ok {
		t.Error("should not find nonexistent package")
	}
}

func TestDependencyChain(t *testing.T) {
	id := NewIdentity()
	chain := id.DependencyChain("pipeline")
	if len(chain) == 0 {
		t.Fatal("expected non-empty dependency chain for pipeline")
	}
	// Pipeline should be last in its own chain.
	if chain[len(chain)-1] != "pipeline" {
		t.Errorf("pipeline should be last in chain, got %s", chain[len(chain)-1])
	}
	// basis should appear before ar in the chain.
	basisIdx := -1
	arIdx := -1
	for i, pkg := range chain {
		if pkg == "basis" {
			basisIdx = i
		}
		if pkg == "ar" {
			arIdx = i
		}
	}
	if basisIdx >= 0 && arIdx >= 0 && basisIdx > arIdx {
		t.Error("basis should appear before ar in dependency chain")
	}
}

func TestBuildAnchor(t *testing.T) {
	id := NewIdentity()
	id.UpdateState(SystemState{
		RulesActive:   5,
		PatternsKnown: 10,
		AgentsAlive:   3,
		Accuracy:      basis.NewRational(3, 4),
	})

	anchor := BuildAnchor(id, "model-1", "model")
	if anchor.AgentID != "model-1" {
		t.Errorf("expected agent ID model-1, got %s", anchor.AgentID)
	}
	if anchor.SystemName != "Caslo" {
		t.Errorf("expected system name Caslo, got %s", anchor.SystemName)
	}
	if anchor.ActiveRules != 5 {
		t.Errorf("expected 5 active rules, got %d", anchor.ActiveRules)
	}
	if len(anchor.Guidance) == 0 {
		t.Error("expected guidance, got none")
	}

	// Model agents should get inference-specific guidance.
	found := false
	for _, g := range anchor.Guidance {
		if strings.Contains(g, "inference") {
			found = true
			break
		}
	}
	if !found {
		t.Error("model anchor should contain inference guidance")
	}

	// All anchors should have the rational arithmetic note.
	found = false
	for _, g := range anchor.Guidance {
		if strings.Contains(g, "rational") {
			found = true
			break
		}
	}
	if !found {
		t.Error("anchor should contain rational arithmetic guidance")
	}
}

func TestAnchorFormat(t *testing.T) {
	id := NewIdentity()
	anchor := BuildAnchor(id, "verifier-0", "verifier")
	formatted := anchor.Format()
	if !strings.Contains(formatted, "ANCHOR: verifier-0") {
		t.Error("formatted anchor should contain agent ID")
	}
	if !strings.Contains(formatted, "Guidance:") {
		t.Error("formatted anchor should contain guidance section")
	}
}

func TestAnchorDigest(t *testing.T) {
	id := NewIdentity()
	id.UpdateState(SystemState{RulesActive: 5, PatternsKnown: 3})
	anchor := BuildAnchor(id, "model-0", "model")
	digest := anchor.Digest()
	if !strings.Contains(digest, "model-0") {
		t.Error("digest should contain agent ID")
	}
	if !strings.Contains(digest, "rules=5") {
		t.Error("digest should contain rule count")
	}
}

func TestIntrospect(t *testing.T) {
	id := NewIdentity()

	// Fresh system — should suggest running pipeline.
	insights := Introspect(id)
	found := false
	for _, ins := range insights {
		if ins.Category == "suggestion" && strings.Contains(ins.Message, "no runs") {
			found = true
		}
	}
	if !found {
		t.Error("fresh system should suggest running the pipeline")
	}

	// System with low accuracy — should warn.
	id.UpdateState(SystemState{
		RunCount: 1,
		Accuracy: basis.NewRational(1, 5),
	})
	insights = Introspect(id)
	found = false
	for _, ins := range insights {
		if ins.Category == "warning" && strings.Contains(ins.Message, "below 1/2") {
			found = true
		}
	}
	if !found {
		t.Error("low accuracy should trigger warning")
	}
}

func TestExplainSelf(t *testing.T) {
	id := NewIdentity()
	explanation := ExplainSelf(id)
	if !strings.Contains(explanation, "I am Caslo") {
		t.Error("explanation should start with identity")
	}
	if !strings.Contains(explanation, "fixpoint") {
		t.Error("explanation should describe reasoning process")
	}
}

func TestSuggestNextStep(t *testing.T) {
	id := NewIdentity()

	// Fresh system.
	suggestion := SuggestNextStep(id)
	if !strings.Contains(suggestion, "run the pipeline") {
		t.Errorf("fresh system suggestion should say run pipeline, got: %s", suggestion)
	}

	// After some runs.
	id.UpdateState(SystemState{
		RunCount:      5,
		Accuracy:      basis.NewRational(4, 5),
		PatternsKnown: 10,
		CacheHits:     5,
	})
	suggestion = SuggestNextStep(id)
	if suggestion == "" {
		t.Error("should always return a suggestion")
	}
}

func TestFormatInsights(t *testing.T) {
	insights := []Insight{
		{Category: "warning", Message: "test warning", Priority: 1},
		{Category: "suggestion", Message: "test suggestion", Priority: 2},
		{Category: "status", Message: "test status", Priority: 3},
	}
	formatted := FormatInsights(insights)
	if !strings.Contains(formatted, "! [warning]") {
		t.Error("formatted insights should mark warnings with !")
	}
	if !strings.Contains(formatted, "> [suggestion]") {
		t.Error("formatted insights should mark suggestions with >")
	}
}

func TestRoleGuidance(t *testing.T) {
	roles := []string{"boss", "model", "verifier", "expert", "unknown"}
	s := SystemState{RulesActive: 5, PatternsKnown: 3, RunCount: 1, Accuracy: basis.NewRational(1, 2)}

	for _, role := range roles {
		g := roleGuidance(role, s)
		if len(g) < 2 {
			t.Errorf("role %s should have at least 2 guidance lines, got %d", role, len(g))
		}
		// Last two should always be the universal ones.
		last := g[len(g)-1]
		secondLast := g[len(g)-2]
		if !strings.Contains(secondLast, "rational") {
			t.Errorf("role %s missing universal rational guidance", role)
		}
		if !strings.Contains(last, "proof traces") {
			t.Errorf("role %s missing universal proof trace guidance", role)
		}
	}
}
