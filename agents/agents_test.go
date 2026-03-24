package agents

import (
	"testing"

	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/kinds"
)

func TestSuperClaudeAgent(t *testing.T) {
	boss := NewSuperClaude("boss-0")
	if boss.ID() != "boss-0" {
		t.Errorf("ID = %s", boss.ID())
	}
	if boss.Role() != RoleBoss {
		t.Errorf("Role = %s", boss.Role())
	}

	ir := kinds.InferenceResult{Final: "defend", Verified: true}
	_, err := boss.Process(Message{Type: MsgResult, Payload: ir})
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}

	results := boss.Results()
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestVerifierAgent(t *testing.T) {
	v := NewVerifier("v-0")

	// Valid result
	ir := kinds.InferenceResult{
		Result: kinds.ModelResult{
			Confidence: basis.NewRational(3, 5),
			ProofTrace: []string{"step1", "step2"},
		},
		Final: "defend",
	}
	resp, err := v.Process(Message{Type: MsgVerify, Payload: ir})
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	verified := resp.Payload.(kinds.InferenceResult)
	if !verified.Verified {
		t.Error("expected verified=true for valid result")
	}

	// Invalid: empty proof
	ir2 := kinds.InferenceResult{
		Result: kinds.ModelResult{
			Confidence: basis.NewRational(1, 2),
			ProofTrace: []string{},
		},
	}
	resp2, _ := v.Process(Message{Type: MsgVerify, Payload: ir2})
	unverified := resp2.Payload.(kinds.InferenceResult)
	if unverified.Verified {
		t.Error("expected verified=false for empty proof")
	}
}

func TestModelAgent(t *testing.T) {
	fn := func(d kinds.Datum) (kinds.ModelResult, error) {
		return kinds.ModelResult{
			Prediction: "test",
			Confidence: basis.NewRational(1, 1),
			ProofTrace: []string{"auto"},
		}, nil
	}

	m := NewModelAgent("m-0", fn)
	datum := kinds.Datum{Features: map[string]int64{"x": 1}}

	resp, err := m.Process(Message{Type: MsgRequest, Payload: datum})
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}

	ir := resp.Payload.(kinds.InferenceResult)
	if ir.Final != "test" {
		t.Errorf("Final = %s, want test", ir.Final)
	}
}

func TestEvaluate(t *testing.T) {
	results := []kinds.InferenceResult{
		{Final: "a"},
		{Final: "b"},
		{Final: "c"},
	}
	expected := []string{"a", "b", "x"}

	correct, total := Evaluate(results, expected)
	if correct != 2 || total != 3 {
		t.Errorf("Evaluate = %d/%d, want 2/3", correct, total)
	}
}
