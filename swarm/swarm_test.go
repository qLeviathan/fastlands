package swarm

import (
	"testing"

	"github.com/qleviathan/fastlands/agents"
	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/kinds"
	"github.com/qleviathan/fastlands/memory"
)

func TestMeshNetwork(t *testing.T) {
	mesh := NewMesh(8)
	mesh.Register("a")
	mesh.Register("b")
	mesh.AddRoute("a", "b")

	if mesh.Size() != 2 {
		t.Errorf("mesh size = %d, want 2", mesh.Size())
	}

	peers := mesh.Peers("a")
	if len(peers) != 1 || peers[0] != "b" {
		t.Errorf("peers of a = %v, want [b]", peers)
	}

	msg := agents.Message{Type: agents.MsgRequest, From: "a", To: "b"}
	if err := mesh.Send("b", msg); err != nil {
		t.Fatalf("Send error: %v", err)
	}

	received, ok := mesh.Receive("b")
	if !ok {
		t.Fatal("expected message")
	}
	if received.From != "a" {
		t.Errorf("received.From = %s, want a", received.From)
	}

	mesh.Unregister("b")
	if mesh.Size() != 1 {
		t.Errorf("mesh size after unregister = %d, want 1", mesh.Size())
	}
}

func TestAutoScaler(t *testing.T) {
	as := NewAutoScaler()
	decisions := as.Evaluate(2, 10)
	if len(decisions) == 0 {
		t.Fatal("expected scale-up decision")
	}
	if decisions[0].Direction != ScaleUp {
		t.Errorf("direction = %d, want ScaleUp", decisions[0].Direction)
	}
}

func TestController(t *testing.T) {
	store := memory.NewStore("")
	ctrl := NewController(store)

	inferFn := func(d kinds.Datum) (kinds.ModelResult, error) {
		return kinds.ModelResult{
			Prediction: "test",
			Confidence: basis.NewRational(1, 1),
			Kind:       kinds.KindLogicPrograms,
			ProofTrace: []string{"auto-test"},
		}, nil
	}

	ds := kinds.DataSet{
		Name: "test",
		Items: []kinds.Datum{
			{Features: map[string]int64{"x": 1}, Label: "test"},
		},
	}

	results := ctrl.RunDataset(ds, inferFn)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Final != "test" {
		t.Errorf("result = %s, want test", results[0].Final)
	}
	if ctrl.Peak() < 1 {
		t.Error("expected peak >= 1")
	}
}
