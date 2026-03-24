package kinds

import (
	"testing"

	"github.com/qleviathan/fastlands/basis"
)

func TestDatumCreation(t *testing.T) {
	d := Datum{
		Features: map[string]int64{"threat_level": 8, "supply": 3},
		Label:    "defend",
	}
	if d.Features["threat_level"] != 8 {
		t.Error("unexpected feature value")
	}
	if d.Label != "defend" {
		t.Error("unexpected label")
	}
}

func TestModelResult(t *testing.T) {
	r := ModelResult{
		Prediction: "defend",
		Confidence: basis.NewRational(3, 5),
		Kind:       KindLogicPrograms,
		ProofTrace: []string{"step1", "step2"},
	}
	if r.Confidence.Num != 3 || r.Confidence.Den != 5 {
		t.Errorf("confidence = %d/%d, want 3/5", r.Confidence.Num, r.Confidence.Den)
	}
}

func TestKindValues(t *testing.T) {
	if KindLogicPrograms.ID != "ar-lp" {
		t.Errorf("KindLogicPrograms.ID = %s", KindLogicPrograms.ID)
	}
	if KindChainEval.ID != "ar-ce" {
		t.Errorf("KindChainEval.ID = %s", KindChainEval.ID)
	}
}
