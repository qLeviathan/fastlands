package holo

import (
	"testing"
)

func TestEncodeAndRecall(t *testing.T) {
	s := NewStore()
	s.Encode("slice-1", 2, 3, []string{"coding", "fast"}, "built a RAG agent")
	s.Encode("slice-2", 2, 4, []string{"coding", "medium"}, "wrote API docs")
	s.Encode("slice-3", 0, 3, []string{"eval", "fast"}, "completed assessment")

	// Recall from (2, 3) — should find all three, strongest first
	results := s.Recall(2, 3, []string{"coding"})
	if len(results) == 0 {
		t.Fatal("expected results from recall")
	}
	if results[0].SliceID != "slice-1" {
		t.Errorf("expected slice-1 as strongest match, got %s", results[0].SliceID)
	}
}

func TestSharedAxisRecall(t *testing.T) {
	s := NewStore()
	// Two slices sharing dough axis (n=3)
	s.Encode("a", 3, 0, []string{"skill"}, "coding on marinara")
	s.Encode("b", 3, 2, []string{"skill"}, "coding on pesto")

	// Recall from (3, 1) — between them on sauce axis
	results := s.Recall(3, 1, []string{"skill"})
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestProjectWave(t *testing.T) {
	s := NewStore()
	w := s.Project(3, 3, 3)

	// Origin should have amplitude 1.0
	if amp, ok := w.Amplitudes[[2]int{3, 3}]; !ok || amp != 1.0 {
		t.Errorf("origin amplitude: got %v, want 1.0", amp)
	}

	// Adjacent axis-aligned should be amplified
	adj := w.Amplitudes[[2]int{3, 4}]
	diag := w.Amplitudes[[2]int{4, 4}]
	if adj <= diag {
		t.Errorf("axis-aligned (%f) should be stronger than diagonal (%f)", adj, diag)
	}
}

func TestInterfere(t *testing.T) {
	s := NewStore()
	w1 := s.Project(2, 2, 2)
	w2 := s.Project(2, 4, 2)

	combined := Interfere(w1, w2)

	// The shared point (2, 3) should have constructive interference
	a1 := w1.Amplitudes[[2]int{2, 3}]
	a2 := w2.Amplitudes[[2]int{2, 3}]
	aCombined := combined.Amplitudes[[2]int{2, 3}]

	if aCombined < a1 && aCombined < a2 {
		t.Error("expected constructive interference at shared point")
	}
}

func TestLearnFromCompletion(t *testing.T) {
	s := NewStore()
	s.LearnFromCompletion("tc-mar-001", 0, 0, "success", 45, "passed assessment fast")

	// Should have created 3 patterns: completion + skill + platform
	stats := s.Stats()
	if stats.TotalPatterns != 3 {
		t.Errorf("expected 3 patterns after completion, got %d", stats.TotalPatterns)
	}
}

func TestSuggestNext(t *testing.T) {
	s := NewStore()
	// No memory yet
	suggestion := s.SuggestNext(0, 0)
	if suggestion == "" {
		t.Error("expected non-empty suggestion")
	}

	// Add some completions
	s.LearnFromCompletion("a", 0, 0, "success", 30, "quick win")
	s.LearnFromCompletion("b", 0, 2, "success", 60, "medium effort")

	suggestion = s.SuggestNext(0, 0)
	if suggestion == "" {
		t.Error("expected non-empty suggestion after learning")
	}
}

func TestHoloStats(t *testing.T) {
	s := NewStore()
	s.Encode("s1", 0, 0, []string{"a"}, "test1")
	s.Encode("s2", 1, 1, []string{"b"}, "test2")

	stats := s.Stats()
	if stats.TotalPatterns != 2 {
		t.Errorf("total: got %d, want 2", stats.TotalPatterns)
	}
	if stats.UniqueSlices != 2 {
		t.Errorf("slices: got %d, want 2", stats.UniqueSlices)
	}
}
