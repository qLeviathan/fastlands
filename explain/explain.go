// Package explain builds hierarchical natural-deduction proofs
// from inference results and renders them as indented tree strings.
package explain

import (
	"fmt"
	"strings"

	"github.com/qleviathan/fastlands/kinds"
)

// ProofNode is one node in a natural-deduction proof tree.
type ProofNode struct {
	Level    int
	Tag      string // "conclusion", "evidence", "premise", "rule"
	Text     string
	Children []ProofNode
}

// BuildProof creates a hierarchical proof tree from an InferenceResult.
//
// Structure:
//
//	Root (conclusion): Final prediction with Confidence and Verified status.
//	  Child (evidence): Kind name and prediction from the model result.
//	    Grandchildren: classified from ProofTrace entries:
//	      - lines containing "has_feature" or starting with "fact:" are tagged "premise"
//	      - lines containing "rule" are tagged "rule"
//	      - everything else is tagged "evidence"
func BuildProof(ir kinds.InferenceResult) ProofNode {
	// Root: conclusion node.
	root := ProofNode{
		Level: 1,
		Tag:   "conclusion",
		Text: fmt.Sprintf("FINAL: %s (Confidence=%d/%d, Verified=%v)",
			ir.Final, ir.Confidence.Num, ir.Confidence.Den, ir.Verified),
	}

	// Child: evidence node with kind and prediction.
	evidence := ProofNode{
		Level: 2,
		Tag:   "evidence",
		Text: fmt.Sprintf("AR[%s]: prediction=%s",
			ir.Result.Kind.Name, ir.Result.Prediction),
	}

	// Grandchildren: classify each proof trace line.
	for _, line := range ir.Result.ProofTrace {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		child := ProofNode{
			Level: 3,
		}

		switch {
		case strings.Contains(trimmed, "has_feature") || strings.HasPrefix(trimmed, "fact:"):
			child.Tag = "premise"
			child.Text = trimmed
		case strings.Contains(trimmed, "rule"):
			child.Tag = "rule"
			child.Text = formatRuleLine(trimmed)
		default:
			child.Tag = "evidence"
			child.Text = trimmed
		}

		evidence.Children = append(evidence.Children, child)
	}

	root.Children = append(root.Children, evidence)
	return root
}

// FormatProof renders a proof tree as an indented string.
//
// Each line has the form:
//
//	[+] <level>. [<tag>] <text>
//
// Children are indented by two spaces per level.
func FormatProof(node ProofNode) string {
	lines := FormatProofLines(node)
	return strings.Join(lines, "\n")
}

// FormatProofLines returns each line of the formatted proof as a separate string.
func FormatProofLines(node ProofNode) []string {
	var lines []string
	collectLines(&lines, node, 0)
	return lines
}

// collectLines walks the tree and appends formatted lines.
func collectLines(lines *[]string, node ProofNode, indent int) {
	prefix := strings.Repeat("  ", indent)
	line := fmt.Sprintf("%s[+] %d. [%s] %s", prefix, node.Level, node.Tag, node.Text)
	*lines = append(*lines, line)

	for _, child := range node.Children {
		collectLines(lines, child, indent+1)
	}
}

// formatRuleLine normalizes a rule trace line into a display format.
// Input like "rule R1-defend: threat_level=high AND supply_available=low => action(defend)"
// becomes "rule[R1-defend]: threat_level=high ^ supply_available=low => action(defend)".
func formatRuleLine(line string) string {
	// Try to parse "rule <ID>: <body> => <head>" format.
	if !strings.HasPrefix(line, "rule ") {
		return line
	}

	rest := strings.TrimPrefix(line, "rule ")
	colonIdx := strings.Index(rest, ":")
	if colonIdx < 0 {
		return line
	}

	ruleID := strings.TrimSpace(rest[:colonIdx])
	bodyHead := strings.TrimSpace(rest[colonIdx+1:])

	// Replace " AND " with " ^ " for natural-deduction style.
	bodyHead = strings.ReplaceAll(bodyHead, " AND ", " ^ ")

	return fmt.Sprintf("rule[%s]: %s", ruleID, bodyHead)
}
