// Package repl provides an interactive command loop for caslo.
// It reads from stdin, writes to stdout, and supports inference,
// memory inspection, sequence validation, and full pipeline runs.
package repl

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/qleviathan/fastlands/ar"
	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/kinds"
	"github.com/qleviathan/fastlands/memory"
	"github.com/qleviathan/fastlands/oeis"
	"github.com/qleviathan/fastlands/pipeline"
	"github.com/qleviathan/fastlands/report"
)

// REPL is the interactive command loop.
type REPL struct {
	Engine  *ar.Engine
	Memory  *memory.Store
	Catalog *oeis.Catalog
}

// New creates a REPL connected to the given engine and memory store.
func New(engine *ar.Engine, mem *memory.Store) *REPL {
	return &REPL{
		Engine:  engine,
		Memory:  mem,
		Catalog: oeis.NewCatalog(),
	}
}

// Run starts the main read-eval-print loop. It blocks until the user
// types "quit" or "exit", or stdin reaches EOF.
func (r *REPL) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("caslo interactive shell. Type 'help' for commands.")

	for {
		fmt.Print("caslo> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line == "quit" || line == "exit" {
			fmt.Println("goodbye.")
			return
		}

		output := r.processCommand(line)
		if output != "" {
			fmt.Println(output)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "repl: read error: %v\n", err)
	}
}

// processCommand dispatches a single input line and returns output text.
func (r *REPL) processCommand(line string) string {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return ""
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "help":
		return helpText()

	case "infer":
		return r.cmdInfer(args)

	case "memory":
		return r.cmdMemory()

	case "patterns":
		return r.cmdPatterns()

	case "rules":
		return r.cmdRules()

	case "oeis":
		return r.cmdOEIS(args)

	case "basis":
		return r.cmdBasis(args)

	case "run":
		return r.cmdRun()

	default:
		return fmt.Sprintf("unknown command: %s (type 'help' for available commands)", cmd)
	}
}

// helpText returns the command listing.
func helpText() string {
	var b strings.Builder
	b.WriteString("Available commands:\n")
	b.WriteString("  help                          show this message\n")
	b.WriteString("  infer <feature>=<value> ...   run inference on features\n")
	b.WriteString("  memory                        show memory summary\n")
	b.WriteString("  patterns                      show learned patterns\n")
	b.WriteString("  rules                         show loaded rules\n")
	b.WriteString("  oeis <n1> <n2> <n3> ...       validate integer sequence against catalog\n")
	b.WriteString("  basis <n>                     show canonical decomposition of integer n\n")
	b.WriteString("  run                           run full pipeline with defaults\n")
	b.WriteString("  quit / exit                   exit the shell\n")
	return b.String()
}

// cmdInfer parses feature=value pairs and runs inference.
func (r *REPL) cmdInfer(args []string) string {
	if len(args) == 0 {
		return "usage: infer <feature>=<value> ..."
	}

	features := make(map[string]int64)
	for _, arg := range args {
		idx := strings.Index(arg, "=")
		if idx < 1 {
			return fmt.Sprintf("invalid feature pair: %q (expected key=value)", arg)
		}
		key := arg[:idx]
		valStr := arg[idx+1:]
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return fmt.Sprintf("invalid integer value for %s: %q", key, valStr)
		}
		features[key] = val
	}

	datum := kinds.Datum{Features: features}
	result, err := r.Engine.Infer(datum)
	if err != nil {
		return fmt.Sprintf("inference error: %v", err)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Prediction: %s\n", result.Prediction))
	b.WriteString(fmt.Sprintf("Confidence: %d/%d\n", result.Confidence.Num, result.Confidence.Den))
	b.WriteString(fmt.Sprintf("Kind:       %s\n", result.Kind.Name))
	if len(result.ProofTrace) > 0 {
		b.WriteString("Proof trace:\n")
		for i, step := range result.ProofTrace {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step))
		}
	}

	// Record the pattern in memory.
	featureKeys := make([]string, 0, len(features))
	for k := range features {
		featureKeys = append(featureKeys, k)
	}
	r.Memory.RecordPattern(featureKeys, result.Prediction, result.Confidence.Num, result.Confidence.Den)

	return b.String()
}

// cmdMemory shows the memory store summary.
func (r *REPL) cmdMemory() string {
	return r.Memory.Summary()
}

// cmdPatterns shows all learned inference patterns.
func (r *REPL) cmdPatterns() string {
	patterns := r.Memory.GetPatterns()
	if len(patterns) == 0 {
		return "no patterns learned yet"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Learned patterns: %d\n", len(patterns)))
	for i, p := range patterns {
		b.WriteString(fmt.Sprintf("  [%d] features=%v prediction=%s confidence=%d/%d seen=%d\n",
			i+1, p.Features, p.Prediction, p.ConfNum, p.ConfDen, p.SeenCount))
	}
	return b.String()
}

// cmdRules shows the loaded engine rules.
func (r *REPL) cmdRules() string {
	if len(r.Engine.Rules) == 0 {
		return "no rules loaded"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Loaded rules: %d\n", len(r.Engine.Rules)))
	for _, rule := range r.Engine.Rules {
		bodyParts := make([]string, 0, len(rule.Body))
		for _, atom := range rule.Body {
			s := atom.Predicate
			if len(atom.Args) > 0 {
				s += "(" + strings.Join(atom.Args, ",") + ")"
			}
			bodyParts = append(bodyParts, s)
		}
		head := rule.Head.Predicate
		if len(rule.Head.Args) > 0 {
			head += "(" + strings.Join(rule.Head.Args, ",") + ")"
		}
		b.WriteString(fmt.Sprintf("  %s: %s => %s (priority=%d)\n",
			rule.ID, strings.Join(bodyParts, " AND "), head, rule.Priority))
	}
	return b.String()
}

// cmdOEIS validates a sequence of integers against the OEIS catalog.
func (r *REPL) cmdOEIS(args []string) string {
	if len(args) == 0 {
		return "usage: oeis <n1> <n2> <n3> ..."
	}

	terms := make([]int64, 0, len(args))
	for _, arg := range args {
		val, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return fmt.Sprintf("invalid integer: %q", arg)
		}
		terms = append(terms, val)
	}

	matches := r.Catalog.Match(terms)
	if len(matches) == 0 {
		return fmt.Sprintf("no matching sequences found for %v", terms)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Matches for sequence %v:\n", terms))
	for _, m := range matches {
		b.WriteString(fmt.Sprintf("  %s (%s): matched %d/%d terms, strength %d/%d\n",
			m.ID, m.Name, m.Matched, m.Total, m.Strength[0], m.Strength[1]))
	}
	return b.String()
}

// cmdBasis shows the canonical decomposition of an integer.
func (r *REPL) cmdBasis(args []string) string {
	if len(args) != 1 {
		return "usage: basis <n>"
	}

	val, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return fmt.Sprintf("invalid unsigned integer: %q", args[0])
	}

	canon := basis.Decompose(val)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Value: %d\n", canon.Value))
	b.WriteString(fmt.Sprintf("Terms: %v\n", canon.Terms))
	b.WriteString(fmt.Sprintf("Bits:  %b\n", canon.Bits))

	w := basis.Encode(val)
	b.WriteString(fmt.Sprintf("Rank:  %d\n", w.Rank))
	b.WriteString(fmt.Sprintf("Spread: %d\n", w.Spread))

	return b.String()
}

// cmdRun executes the full pipeline with default configuration.
func (r *REPL) cmdRun() string {
	cfg := pipeline.DefaultConfig()
	cfg.MemoryPath = r.Memory.Path

	result := pipeline.Run(cfg)

	rpt := report.Generate(
		"REPL Pipeline Run",
		result.Results,
		result.Accuracy,
		result.MemorySummary,
		result.SwarmSummary,
		result.FuturesSummary,
		result.RewriterSummary,
		result.Datasets,
	)

	return report.Format(rpt)
}
