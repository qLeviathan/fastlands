// Package rewriter implements meta-rule evaluation and rule evolution.
// Meta-rules are rules about rules — they read performance data from
// episodic memory and propose modifications to the active rule set.
package rewriter

import (
	"fmt"

	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/memory"
)

// ActionKind describes what a rewrite does.
type ActionKind string

const (
	ActionRetire     ActionKind = "retire"
	ActionStrengthen ActionKind = "strengthen"
	ActionWeaken     ActionKind = "weaken"
	ActionSplit      ActionKind = "split"
	ActionPropose    ActionKind = "propose"
)

// RewriteAction describes a proposed change to a rule.
type RewriteAction struct {
	Kind        ActionKind
	RuleID      string
	Description string
	NewPriority int // for strengthen/weaken
}

// MetaRule evaluates a rule's performance and proposes an action.
type MetaRule struct {
	ID        string
	Name      string
	Condition func(stats memory.RulePerformance) bool
	Action    func(stats memory.RulePerformance) RewriteAction
}

// Rewriter holds meta-rules and evaluates the active rule set.
type Rewriter struct {
	MetaRules []MetaRule
	Applied   []RewriteAction
}

// New creates a rewriter with the default set of meta-rules.
func New() *Rewriter {
	return &Rewriter{
		MetaRules: defaultMetaRules(),
	}
}

// Evaluate runs all meta-rules against all rule stats from memory.
// Returns proposed actions.
func (rw *Rewriter) Evaluate(store *memory.Store) []RewriteAction {
	stats := store.GetRuleStats()
	var actions []RewriteAction

	for _, stat := range stats {
		for _, mr := range rw.MetaRules {
			if mr.Condition(stat) {
				action := mr.Action(stat)
				actions = append(actions, action)
			}
		}
	}

	rw.Applied = append(rw.Applied, actions...)
	return actions
}

// ProposeFromPatterns reads memory patterns and proposes new rules
// for patterns that have been seen enough times but have no matching rule.
func (rw *Rewriter) ProposeFromPatterns(store *memory.Store, existingRuleIDs map[string]bool) []RewriteAction {
	patterns := store.GetPatterns()
	var proposals []RewriteAction

	for _, p := range patterns {
		if p.SeenCount < 5 {
			continue
		}
		// Check confidence threshold: num/den >= 4/5 (i.e. num*5 >= den*4)
		if p.ConfNum*5 < p.ConfDen*4 {
			continue
		}
		// Generate a rule ID from features
		ruleID := fmt.Sprintf("auto-%s", p.Prediction)
		for _, f := range p.Features {
			ruleID += "-" + f
		}
		if existingRuleIDs[ruleID] {
			continue
		}

		proposals = append(proposals, RewriteAction{
			Kind:        ActionPropose,
			RuleID:      ruleID,
			Description: fmt.Sprintf("propose rule from pattern: %v → %s (seen %d, conf %d/%d)", p.Features, p.Prediction, p.SeenCount, p.ConfNum, p.ConfDen),
		})
	}

	rw.Applied = append(rw.Applied, proposals...)
	return proposals
}

// Summary returns a human-readable summary of all applied actions.
func (rw *Rewriter) Summary() string {
	if len(rw.Applied) == 0 {
		return "rewriter: no actions taken"
	}
	s := fmt.Sprintf("rewriter: %d actions\n", len(rw.Applied))
	for _, a := range rw.Applied {
		s += fmt.Sprintf("  [%s] %s: %s\n", a.Kind, a.RuleID, a.Description)
	}
	return s
}

func defaultMetaRules() []MetaRule {
	return []MetaRule{
		{
			ID:   "mr-retire",
			Name: "Retire Underperformers",
			Condition: func(s memory.RulePerformance) bool {
				if s.FireCount < 10 {
					return false
				}
				// accuracy < 40%: correct*10 < fire*4
				return int64(s.CorrectCount)*10 < int64(s.FireCount)*4
			},
			Action: func(s memory.RulePerformance) RewriteAction {
				acc := s.Accuracy()
				return RewriteAction{
					Kind:        ActionRetire,
					RuleID:      s.RuleID,
					Description: fmt.Sprintf("accuracy %d/%d after %d fires", acc.Num, acc.Den, s.FireCount),
				}
			},
		},
		{
			ID:   "mr-strengthen",
			Name: "Strengthen High Performers",
			Condition: func(s memory.RulePerformance) bool {
				if s.FireCount < 5 {
					return false
				}
				// accuracy > 90%: correct*10 > fire*9
				return int64(s.CorrectCount)*10 > int64(s.FireCount)*9
			},
			Action: func(s memory.RulePerformance) RewriteAction {
				return RewriteAction{
					Kind:        ActionStrengthen,
					RuleID:      s.RuleID,
					Description: fmt.Sprintf("high accuracy after %d fires", s.FireCount),
					NewPriority: 10,
				}
			},
		},
		{
			ID:   "mr-split",
			Name: "Split Ambiguous",
			Condition: func(s memory.RulePerformance) bool {
				if s.FireCount < 8 {
					return false
				}
				// accuracy between 40-60%: domain-dependent behavior
				acc := basis.NewRational(int64(s.CorrectCount), int64(s.FireCount))
				return acc.Num*10 >= acc.Den*4 && acc.Num*10 <= acc.Den*6 && len(s.Domains) > 1
			},
			Action: func(s memory.RulePerformance) RewriteAction {
				return RewriteAction{
					Kind:        ActionSplit,
					RuleID:      s.RuleID,
					Description: fmt.Sprintf("ambiguous across %d domains", len(s.Domains)),
				}
			},
		},
		{
			ID:   "mr-weaken",
			Name: "Weaken Early Warning",
			Condition: func(s memory.RulePerformance) bool {
				if s.FireCount < 3 {
					return false
				}
				// accuracy < 60%: correct*10 < fire*6
				return int64(s.CorrectCount)*10 < int64(s.FireCount)*6
			},
			Action: func(s memory.RulePerformance) RewriteAction {
				return RewriteAction{
					Kind:        ActionWeaken,
					RuleID:      s.RuleID,
					Description: fmt.Sprintf("low early accuracy after %d fires", s.FireCount),
					NewPriority: -5,
				}
			},
		},
	}
}
