package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MemoryPath == "" {
		t.Error("MemoryPath should have a default")
	}
	if cfg.MaxAgents <= 0 {
		t.Error("MaxAgents should be positive")
	}
}

func TestRunWithDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.MemoryPath = filepath.Join(dir, "test_memory.json")
	cfg.AutoMode = true

	result := Run(cfg)
	if result.Total == 0 {
		t.Error("expected some inference results")
	}
	if result.Duration <= 0 {
		t.Error("expected positive duration")
	}

	// Memory file should exist after run
	if _, err := os.Stat(cfg.MemoryPath); err != nil {
		t.Errorf("memory file not created: %v", err)
	}
}
