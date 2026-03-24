package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStoreEmpty(t *testing.T) {
	s := NewStore("")
	if len(s.Facts) != 0 {
		t.Error("new store should have no facts")
	}
	if len(s.RuleStats) != 0 {
		t.Error("new store should have no rule stats")
	}
}

func TestLearnFact(t *testing.T) {
	s := NewStore("")
	s.LearnFact("accuracy", "high", "test", "engine")
	if len(s.Facts) != 1 {
		t.Fatalf("expected 1 fact, got %d", len(s.Facts))
	}
	// Deduplicate
	s.LearnFact("accuracy", "high", "test", "engine")
	if len(s.Facts) != 1 {
		t.Errorf("expected dedup, got %d facts", len(s.Facts))
	}
	if s.Facts[0].UseCount != 2 {
		t.Errorf("use_count = %d, want 2", s.Facts[0].UseCount)
	}
}

func TestRecordRuleFire(t *testing.T) {
	s := NewStore("")
	s.RecordRuleFire("r1", true, "logistics")
	s.RecordRuleFire("r1", false, "logistics")
	s.RecordRuleFire("r1", true, "finance")

	stats := s.GetRuleStats()
	if len(stats) != 1 {
		t.Fatalf("expected 1 rule stat, got %d", len(stats))
	}
	if stats[0].FireCount != 3 {
		t.Errorf("fire_count = %d, want 3", stats[0].FireCount)
	}
	if stats[0].CorrectCount != 2 {
		t.Errorf("correct_count = %d, want 2", stats[0].CorrectCount)
	}
	acc := stats[0].Accuracy()
	if acc.Num != 2 || acc.Den != 3 {
		t.Errorf("accuracy = %d/%d, want 2/3", acc.Num, acc.Den)
	}
}

func TestRecordPattern(t *testing.T) {
	s := NewStore("")
	s.RecordPattern([]string{"threat_level", "supply"}, "defend", 3, 5)
	s.RecordPattern([]string{"threat_level", "supply"}, "defend", 4, 5)

	patterns := s.GetPatterns()
	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}
	if patterns[0].SeenCount != 2 {
		t.Errorf("seen_count = %d, want 2", patterns[0].SeenCount)
	}
}

func TestPredictFromHistory(t *testing.T) {
	s := NewStore("")
	s.RecordPattern([]string{"a", "b"}, "result", 4, 5)

	label, num, den, found := s.PredictFromHistory([]string{"b", "a"}) // should sort
	if !found {
		t.Fatal("expected pattern match")
	}
	if label != "result" {
		t.Errorf("prediction = %s, want result", label)
	}
	if num != 4 || den != 5 {
		t.Errorf("confidence = %d/%d, want 4/5", num, den)
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test_memory.json")

	s := NewStore(path)
	s.LearnFact("key1", "val1", "test", "unit")
	s.RecordRuleFire("r1", true, "test")
	s.RecordRun([]string{"ds1"}, 10, 8, 10, 5*time.Second, 2)
	if err := s.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Reload
	s2 := NewStore(path)
	if len(s2.Facts) != 1 {
		t.Errorf("loaded %d facts, want 1", len(s2.Facts))
	}
	if len(s2.RuleStats) != 1 {
		t.Errorf("loaded %d rule stats, want 1", len(s2.RuleStats))
	}
	if len(s2.RunHistory) != 1 {
		t.Errorf("loaded %d runs, want 1", len(s2.RunHistory))
	}

	// Cleanup
	os.Remove(path)
}
