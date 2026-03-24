package report

import (
	"strings"
	"testing"

	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/kinds"
)

func TestGenerate(t *testing.T) {
	results := []kinds.InferenceResult{
		{
			Result: kinds.ModelResult{
				Prediction: "defend",
				Confidence: basis.NewRational(4, 5),
				Kind:       kinds.KindLogicPrograms,
				ProofTrace: []string{"has_feature(threat, high)"},
			},
			Final:    "defend",
			Verified: true,
		},
	}
	acc := basis.NewRational(4, 5)

	rpt := Generate("Test Report", results, acc,
		"memory: 1 fact", "swarm: 2 agents", "futures: 0 hits", "rewriter: 0 actions",
		[]string{"test-ds"})

	if rpt.Title != "Test Report" {
		t.Errorf("title = %s", rpt.Title)
	}
	if len(rpt.Sections) == 0 {
		t.Error("expected sections in report")
	}
}

func TestFormat(t *testing.T) {
	rpt := Report{
		Title: "Test",
		Sections: []Section{
			{Heading: "Summary", Body: "All good."},
		},
	}
	out := Format(rpt)
	if !strings.Contains(out, "Summary") {
		t.Error("formatted report should contain section heading")
	}
}
