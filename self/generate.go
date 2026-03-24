package self

import (
	"fmt"
	"strings"

	"github.com/qleviathan/fastlands/basis"
)

// Generate produces contextual guidance based on Caslo's current state
// and self-knowledge. It synthesizes what it knows about itself into
// actionable, helpful output.

// Insight is a generated piece of guidance or observation.
type Insight struct {
	Category string // "status", "suggestion", "warning", "explanation"
	Message  string
	Priority int // 1=critical, 2=important, 3=informational
}

// Introspect examines current state and generates relevant insights.
func Introspect(id *Identity) []Insight {
	s := id.State()
	var insights []Insight

	// Status insights.
	insights = append(insights, Insight{
		Category: "status",
		Message:  fmt.Sprintf("%s is operational with %d rules, %d patterns, %d facts", id.Name, s.RulesActive, s.PatternsKnown, s.FactsLearned),
		Priority: 3,
	})

	if s.RunCount == 0 {
		insights = append(insights, Insight{
			Category: "suggestion",
			Message:  "no runs recorded yet — run the pipeline to begin learning patterns and building memory",
			Priority: 2,
		})
	}

	// Accuracy insights.
	if s.RunCount > 0 {
		if s.Accuracy.Den > 0 {
			// Check if accuracy is below 1/2.
			half := basis.NewRational(1, 2)
			if basis.RatCompare(s.Accuracy, half) < 0 {
				insights = append(insights, Insight{
					Category: "warning",
					Message:  fmt.Sprintf("accuracy %d/%d is below 1/2 — consider adding domain rules or reviewing rewriter suggestions", s.Accuracy.Num, s.Accuracy.Den),
					Priority: 1,
				})
			} else {
				insights = append(insights, Insight{
					Category: "status",
					Message:  fmt.Sprintf("accuracy %d/%d across %d runs", s.Accuracy.Num, s.Accuracy.Den, s.RunCount),
					Priority: 3,
				})
			}
		}
	}

	// Pattern insights.
	if s.PatternsKnown > 0 && s.CacheHits == 0 && s.CacheMisses > 0 {
		insights = append(insights, Insight{
			Category: "suggestion",
			Message:  fmt.Sprintf("%d patterns known but 0 cache hits — warm the futures cache to leverage learned patterns", s.PatternsKnown),
			Priority: 2,
		})
	}

	if s.CacheHits > 0 {
		total := s.CacheHits + s.CacheMisses
		hitRate := basis.NewRational(int64(s.CacheHits), int64(total))
		insights = append(insights, Insight{
			Category: "status",
			Message:  fmt.Sprintf("futures cache hit rate: %d/%d (%d hits, %d misses)", hitRate.Num, hitRate.Den, s.CacheHits, s.CacheMisses),
			Priority: 3,
		})
	}

	// Rewrite insights.
	if s.RewriteCount > 0 {
		insights = append(insights, Insight{
			Category: "status",
			Message:  fmt.Sprintf("%d rewrite actions proposed — review and apply to evolve the rule set", s.RewriteCount),
			Priority: 2,
		})
	}

	// Agent insights.
	if s.AgentsPeak > 4 {
		insights = append(insights, Insight{
			Category: "status",
			Message:  fmt.Sprintf("peak agent count reached %d — swarm scaling is active", s.AgentsPeak),
			Priority: 3,
		})
	}

	// Memory growth insights.
	if s.FactsLearned > 100 {
		insights = append(insights, Insight{
			Category: "suggestion",
			Message:  fmt.Sprintf("%d facts accumulated — consider pruning low-use facts to keep memory focused", s.FactsLearned),
			Priority: 3,
		})
	}

	return insights
}

// ExplainSelf generates a natural-language explanation of what Caslo is
// and what it can do, tailored to the current state.
func ExplainSelf(id *Identity) string {
	s := id.State()
	var b strings.Builder

	b.WriteString(fmt.Sprintf("I am %s, a %s.\n\n", id.Name, id.Purpose))

	b.WriteString("What I do:\n")
	b.WriteString("  I take structured data (integer features), apply forward-chaining logic rules\n")
	b.WriteString("  to derive conclusions, verify those conclusions through proof traces, and\n")
	b.WriteString("  remember what I learn across sessions. I coordinate multiple agents in a\n")
	b.WriteString("  swarm to parallelize inference, predict outcomes from cached patterns, and\n")
	b.WriteString("  rewrite my own rules when they underperform.\n\n")

	b.WriteString("How I reason:\n")
	b.WriteString("  1. Features are discretized into atoms (low/mid/high)\n")
	b.WriteString("  2. Rules fire when all body atoms are satisfied\n")
	b.WriteString("  3. Firing continues until no new atoms are derived (fixpoint)\n")
	b.WriteString("  4. Derived atoms map to predictions via label mappings\n")
	b.WriteString("  5. Every step is recorded in a proof trace for verification\n\n")

	b.WriteString("What I know right now:\n")
	if s.RunCount == 0 {
		b.WriteString("  I have not run yet. My rule set and memory are initialized but untested.\n")
	} else {
		b.WriteString(fmt.Sprintf("  I have completed %d runs with accuracy %d/%d.\n", s.RunCount, s.Accuracy.Num, s.Accuracy.Den))
		b.WriteString(fmt.Sprintf("  I know %d patterns, %d facts, and track %d rules.\n", s.PatternsKnown, s.FactsLearned, s.RulesActive))
	}

	insights := Introspect(id)
	hasActions := false
	for _, ins := range insights {
		if ins.Category == "suggestion" || ins.Category == "warning" {
			if !hasActions {
				b.WriteString("\nWhat I suggest:\n")
				hasActions = true
			}
			prefix := "  > "
			if ins.Category == "warning" {
				prefix = "  ! "
			}
			b.WriteString(prefix + ins.Message + "\n")
		}
	}

	return b.String()
}

// SuggestNextStep generates a single actionable suggestion based on current state.
func SuggestNextStep(id *Identity) string {
	s := id.State()

	if s.RunCount == 0 {
		return "run the pipeline to begin: this will process datasets, fire rules, build memory, and establish baseline accuracy"
	}

	if s.Accuracy.Den > 0 {
		half := basis.NewRational(1, 2)
		if basis.RatCompare(s.Accuracy, half) < 0 {
			return "accuracy is low — review the rewriter's proposed actions and consider adding domain-specific rules"
		}
	}

	if s.PatternsKnown > 0 && s.CacheHits == 0 {
		return "patterns exist but cache is cold — run inference to warm the futures predictor"
	}

	if s.RewriteCount > 0 {
		return fmt.Sprintf("the rewriter has %d proposals — review and apply them to evolve the rule set", s.RewriteCount)
	}

	if s.FactsLearned > 50 {
		return "memory is growing — inspect learned facts and patterns to understand what Caslo has internalized"
	}

	return "system is healthy — run more datasets to deepen pattern coverage and improve accuracy"
}

// FormatInsights renders a list of insights as a readable summary.
func FormatInsights(insights []Insight) string {
	if len(insights) == 0 {
		return "(no insights)"
	}

	var b strings.Builder
	for _, ins := range insights {
		prefix := "  "
		switch ins.Category {
		case "warning":
			prefix = "! "
		case "suggestion":
			prefix = "> "
		case "status":
			prefix = "  "
		case "explanation":
			prefix = "? "
		}
		b.WriteString(fmt.Sprintf("%s[%s] %s\n", prefix, ins.Category, ins.Message))
	}
	return b.String()
}
