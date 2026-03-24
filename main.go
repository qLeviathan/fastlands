// Command caslo is the entry point for the caslo inference system.
// It supports three modes: interactive REPL, automatic pipeline run
// with defaults, and config-driven pipeline run from a JSON file.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/qleviathan/fastlands/ar"
	"github.com/qleviathan/fastlands/memory"
	"github.com/qleviathan/fastlands/pipeline"
	"github.com/qleviathan/fastlands/repl"
	"github.com/qleviathan/fastlands/report"
)

func main() {
	auto := flag.Bool("auto", false, "run pipeline with all defaults")
	configPath := flag.String("config", "", "path to config JSON")
	interactive := flag.Bool("repl", false, "start interactive REPL")
	flag.Parse()

	switch {
	case *interactive:
		startREPL()

	case *auto:
		runPipeline(pipeline.DefaultConfig())

	case *configPath != "":
		cfg, err := loadConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		runPipeline(cfg)

	default:
		fmt.Println("caslo - hybrid inference engine")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  caslo --repl           start interactive shell")
		fmt.Println("  caslo --auto           run pipeline with defaults")
		fmt.Println("  caslo --config FILE    run pipeline from config JSON")
		fmt.Println()
		fmt.Println("Starting interactive shell...")
		fmt.Println()
		startREPL()
	}
}

// startREPL initialises the engine and memory, then launches the
// interactive command loop.
func startREPL() {
	rs := ar.DefaultRuleSet()
	engine := ar.NewEngine(rs)
	mem := memory.NewStore("data/memory.json")

	r := repl.New(engine, mem)
	r.Run()
}

// runPipeline executes the full pipeline and prints a formatted report.
func runPipeline(cfg pipeline.Config) {
	result := pipeline.Run(cfg)

	rpt := report.Generate(
		"caslo Pipeline Report",
		result.Results,
		result.Accuracy,
		result.MemorySummary,
		result.SwarmSummary,
		result.FuturesSummary,
		result.RewriterSummary,
		result.Datasets,
	)

	fmt.Print(report.Format(rpt))
	fmt.Printf("\nCompleted in %s\n", result.Duration)
}

// configFile is the JSON structure for --config.
type configFile struct {
	MemoryPath  string `json:"memory_path"`
	RulesPath   string `json:"rules_path"`
	DataDir     string `json:"data_dir"`
	MaxAgents   int    `json:"max_agents"`
	FutureDepth int    `json:"future_depth"`
}

// loadConfig reads a JSON config file and builds a pipeline.Config.
func loadConfig(path string) (pipeline.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return pipeline.Config{}, fmt.Errorf("failed to read config %s: %w", path, err)
	}

	var cf configFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return pipeline.Config{}, fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	cfg := pipeline.DefaultConfig()
	if cf.MemoryPath != "" {
		cfg.MemoryPath = cf.MemoryPath
	}
	if cf.RulesPath != "" {
		cfg.RulesPath = cf.RulesPath
	}
	if cf.DataDir != "" {
		cfg.DataDir = cf.DataDir
	}
	if cf.MaxAgents > 0 {
		cfg.MaxAgents = cf.MaxAgents
	}
	if cf.FutureDepth > 0 {
		cfg.FutureDepth = cf.FutureDepth
	}

	return cfg, nil
}
