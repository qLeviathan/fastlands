// Package pipeline orchestrates the full caslo inference pipeline.
// It coordinates memory, AR engine, swarm, futures, and rewriter
// into a single Run function that processes datasets end-to-end.
package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/qleviathan/fastlands/agents"
	"github.com/qleviathan/fastlands/ar"
	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/futures"
	"github.com/qleviathan/fastlands/kinds"
	"github.com/qleviathan/fastlands/memory"
	"github.com/qleviathan/fastlands/rewriter"
	"github.com/qleviathan/fastlands/swarm"
)

// Config holds all pipeline parameters.
type Config struct {
	MemoryPath  string
	RulesPath   string
	DataDir     string
	AutoMode    bool
	MaxAgents   int
	FutureDepth int
}

// DefaultConfig returns sensible defaults for the pipeline.
func DefaultConfig() Config {
	return Config{
		MemoryPath:  "data/memory.json",
		RulesPath:   "",
		DataDir:     "",
		AutoMode:    true,
		MaxAgents:   8,
		FutureDepth: 3,
	}
}

// PipelineResult holds the outcome of a full pipeline run.
type PipelineResult struct {
	Results         []kinds.InferenceResult
	Correct         int
	Total           int
	Accuracy        basis.Rational
	RewriteActions  []rewriter.RewriteAction
	MemorySummary   string
	SwarmSummary    string
	FuturesSummary  string
	RewriterSummary string
	Duration        time.Duration
	Datasets        []string
}

// datasetFile is the on-disk JSON format for a dataset.
type datasetFile struct {
	Name   string         `json:"name"`
	Domain string         `json:"domain"`
	Items  []datasetItem  `json:"items"`
}

// datasetItem is one row inside a dataset JSON file.
type datasetItem struct {
	Features map[string]int64 `json:"features"`
	Label    string           `json:"label"`
}

// Run executes the full caslo pipeline according to cfg.
//
//  1. Load memory store from config path (default "data/memory.json").
//  2. Load AR engine (from file if specified, else DefaultRuleSet).
//  3. Create swarm controller.
//  4. Initialize futures predictor, warm cache from memory.
//  5. Discover and process datasets (from DataDir or built-in testdata).
//  6. Run inference on all datasets via swarm.
//  7. Run rewriter evaluation.
//  8. Record run summary in memory.
//  9. Save memory to disk.
//  10. Return PipelineResult.
func Run(cfg Config) PipelineResult {
	start := time.Now()

	// --- Step 1: Load memory store ---
	memPath := cfg.MemoryPath
	if memPath == "" {
		memPath = "data/memory.json"
	}
	ensureDir(memPath)
	mem := memory.NewStore(memPath)

	// --- Step 2: Load AR engine ---
	var engine *ar.Engine
	var ruleSet ar.RuleSet
	if cfg.RulesPath != "" {
		eng, err := ar.LoadRulesFromFile(cfg.RulesPath)
		if err == nil {
			engine = eng
			// Read the ruleset separately for futures access.
			data, readErr := os.ReadFile(cfg.RulesPath)
			if readErr == nil {
				_ = json.Unmarshal(data, &ruleSet)
			}
		}
	}
	if engine == nil {
		ruleSet = ar.DefaultRuleSet()
		engine = ar.NewEngine(ruleSet)
	}

	// --- Step 3: Create swarm controller ---
	ctrl := swarm.NewController(mem)
	if cfg.MaxAgents > 0 {
		ctrl.Scaler.MaxAgents = cfg.MaxAgents
	}

	// --- Step 4: Initialize futures predictor, warm cache from memory ---
	labels := make(map[string]string, len(ruleSet.Labels))
	for _, lm := range ruleSet.Labels {
		labels[lm.Atom] = lm.Label
	}
	pred := futures.NewPredictor(mem, ruleSet.Rules, labels)
	if cfg.FutureDepth > 0 {
		pred.MaxDepth = cfg.FutureDepth
	}
	pred.WarmCache(64)

	// --- Step 5: Discover datasets ---
	datasets := discoverDatasets(cfg.DataDir)
	if len(datasets) == 0 {
		datasets = builtinDatasets()
	}

	// --- Step 6: Run inference on all datasets via swarm ---
	inferFn := func(d kinds.Datum) (kinds.ModelResult, error) {
		// Try futures prediction first.
		if p, ok := pred.Predict(d); ok {
			_ = p // log the hit but still run full inference for accuracy
		}
		return engine.Infer(d)
	}

	var allResults []kinds.InferenceResult
	var allExpected []string
	var datasetNames []string

	for _, ds := range datasets {
		datasetNames = append(datasetNames, ds.Name)
		results := ctrl.RunDataset(ds, inferFn)

		// Collect expected labels for accuracy.
		for _, item := range ds.Items {
			allExpected = append(allExpected, item.Label)
		}
		allResults = append(allResults, results...)

		// Record patterns and rule fires in memory.
		for i, ir := range results {
			featureKeys := sortedFeatureNames(ds.Items[i])
			mem.RecordPattern(featureKeys, ir.Final, ir.Confidence.Num, ir.Confidence.Den)

			// Record rule fires from proof trace.
			for _, line := range ir.Result.ProofTrace {
				if strings.HasPrefix(line, "rule ") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) >= 1 {
						ruleID := strings.TrimPrefix(parts[0], "rule ")
						ruleID = strings.TrimSpace(ruleID)
						correct := false
						if i < len(ds.Items) {
							correct = ir.Final == ds.Items[i].Label
						}
						mem.RecordRuleFire(ruleID, correct, ds.Domain)
					}
				}
			}
		}
	}

	// --- Step 7: Run rewriter evaluation ---
	rw := rewriter.New()
	rwActions := rw.Evaluate(mem)

	// Also check for pattern-based proposals.
	existingRuleIDs := make(map[string]bool, len(ruleSet.Rules))
	for _, r := range ruleSet.Rules {
		existingRuleIDs[r.ID] = true
	}
	proposals := rw.ProposeFromPatterns(mem, existingRuleIDs)
	rwActions = append(rwActions, proposals...)

	// --- Step 8: Compute accuracy and record run summary in memory ---
	correct, total := agents.Evaluate(allResults, allExpected)
	var accuracy basis.Rational
	if total > 0 {
		accuracy = basis.NewRational(int64(correct), int64(total))
	} else {
		accuracy = basis.RatZero()
	}

	mem.RecordRun(datasetNames, total, int64(correct), int64(total), time.Since(start), ctrl.Peak())

	// --- Step 9: Save memory to disk ---
	_ = mem.Save()

	// --- Step 10: Return PipelineResult ---
	return PipelineResult{
		Results:         allResults,
		Correct:         correct,
		Total:           total,
		Accuracy:        accuracy,
		RewriteActions:  rwActions,
		MemorySummary:   mem.Summary(),
		SwarmSummary:    ctrl.Summary(),
		FuturesSummary:  pred.Summary(),
		RewriterSummary: rw.Summary(),
		Duration:        time.Since(start),
		Datasets:        datasetNames,
	}
}

// discoverDatasets reads all *.json files from dir and parses them as datasets.
func discoverDatasets(dir string) []kinds.DataSet {
	if dir == "" {
		return nil
	}

	pattern := filepath.Join(dir, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil
	}

	sort.Strings(matches)
	var datasets []kinds.DataSet

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var dsf datasetFile
		if err := json.Unmarshal(data, &dsf); err != nil {
			continue
		}
		if len(dsf.Items) == 0 {
			continue
		}

		ds := kinds.DataSet{
			Name:   dsf.Name,
			Domain: dsf.Domain,
		}
		if ds.Name == "" {
			base := filepath.Base(path)
			ds.Name = strings.TrimSuffix(base, filepath.Ext(base))
		}

		for _, item := range dsf.Items {
			ds.Items = append(ds.Items, kinds.Datum{
				Features: item.Features,
				Label:    item.Label,
			})
		}

		datasets = append(datasets, ds)
	}

	return datasets
}

// builtinDatasets returns sample test data when no external files are available.
// Uses the DefaultRuleSet supply-chain domain with integer-only features.
func builtinDatasets() []kinds.DataSet {
	return []kinds.DataSet{
		{
			Name:   "builtin-supply-chain",
			Domain: "supply-chain",
			Items: []kinds.Datum{
				{
					Features: map[string]int64{"threat_level": 8, "supply_available": 2},
					Label:    "defend",
				},
				{
					Features: map[string]int64{"threat_level": 1, "supply_available": 9},
					Label:    "expand",
				},
				{
					Features: map[string]int64{"threat_level": 4, "supply_available": 4},
					Label:    "hold",
				},
				{
					Features: map[string]int64{"threat_level": 7, "supply_available": 5},
					Label:    "fortify",
				},
				{
					Features: map[string]int64{"threat_level": 2, "supply_available": 4},
					Label:    "scout",
				},
				{
					Features: map[string]int64{"threat_level": 4, "supply_available": 2},
					Label:    "conserve",
				},
			},
		},
	}
}

// sortedFeatureNames returns the feature names of a datum in sorted order.
func sortedFeatureNames(d kinds.Datum) []string {
	keys := make([]string, 0, len(d.Features))
	for k := range d.Features {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ensureDir creates the parent directory for a file path if it does not exist.
func ensureDir(filePath string) {
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}
}
