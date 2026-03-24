package explain

import (
	"strings"
	"testing"

	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/kinds"
)

func TestBuildProof(t *testing.T) {
	ir := kinds.InferenceResult{
		Result: kinds.ModelResult{
			Prediction: "defend",
			Confidence: basis.NewRational(3, 5),
			Kind:       kinds.KindLogicPrograms,
			ProofTrace: []string{
				"has_feature(threat_level, high)",
				"has_feature(supply_available, low)",
				"rule[coa-r1]: threat_high ^ supply_low => defend",
			},
		},
		Final:    "defend",
		Verified: true,
	}

	proof := BuildProof(ir)
	if proof.Tag != "conclusion" {
		t.Errorf("root tag = %s, want conclusion", proof.Tag)
	}
	if len(proof.Children) == 0 {
		t.Error("expected children in proof tree")
	}
}

func TestFormatProof(t *testing.T) {
	ir := kinds.InferenceResult{
		Result: kinds.ModelResult{
			Prediction: "defend",
			Confidence: basis.NewRational(1, 1),
			Kind:       kinds.KindLogicPrograms,
			ProofTrace: []string{"has_feature(x, high)", "rule[r1]: applied"},
		},
		Final:    "defend",
		Verified: true,
	}

	proof := BuildProof(ir)
	out := FormatProof(proof)
	if !strings.Contains(out, "defend") {
		t.Error("formatted proof should contain prediction")
	}
	if !strings.Contains(out, "conclusion") {
		t.Error("formatted proof should contain 'conclusion' tag")
	}
}
