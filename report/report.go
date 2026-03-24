// Package report generates evaluation reports from pipeline results.
// All numeric values use exact integer or rational arithmetic — no
// floating point.
package report

import (
	"fmt"
	"strings"

	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/explain"
	"github.com/qleviathan/fastlands/kinds"
)

// Report is a structured evaluation report.
type Report struct {
	Title    string
	Sections []Section
}

// Section is one named block inside a Report.
type Section struct {
	Heading string
	Body    string
}

// Generate creates a full report from pipeline outputs.
func Generate(
	title string,
	results []kinds.InferenceResult,
	accuracy basis.Rational,
	memorySummary, swarmSummary, futuresSummary, rewriterSummary string,
	datasets []string,
) Report {
	r := Report{Title: title}

	// 1. Executive Summary
	r.Sections = append(r.Sections, buildExecutiveSummary(results, accuracy, datasets))

	// 2. Dataset Results
	r.Sections = append(r.Sections, buildDatasetResults(results))

	// 3. Proofs
	r.Sections = append(r.Sections, buildProofs(results))

	// 4. Memory Summary
	r.Sections = append(r.Sections, Section{
		Heading: "Memory Summary",
		Body:    memorySummary,
	})

	// 5. Swarm Summary
	r.Sections = append(r.Sections, Section{
		Heading: "Swarm Summary",
		Body:    swarmSummary,
	})

	// 6. Futures Summary
	r.Sections = append(r.Sections, Section{
		Heading: "Futures Summary",
		Body:    futuresSummary,
	})

	// 7. Rewriter Summary
	r.Sections = append(r.Sections, Section{
		Heading: "Rewriter Summary",
		Body:    rewriterSummary,
	})

	return r
}

// Format renders a Report as a human-readable string.
func Format(r Report) string {
	var b strings.Builder

	// Title banner.
	border := strings.Repeat("=", len(r.Title)+4)
	b.WriteString(border + "\n")
	b.WriteString("  " + r.Title + "\n")
	b.WriteString(border + "\n\n")

	for i, sec := range r.Sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("--- " + sec.Heading + " ---\n")
		b.WriteString(sec.Body)
		if !strings.HasSuffix(sec.Body, "\n") {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// buildExecutiveSummary produces the top-level overview section.
func buildExecutiveSummary(results []kinds.InferenceResult, accuracy basis.Rational, datasets []string) Section {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Accuracy:           %d/%d\n", accuracy.Num, accuracy.Den))
	b.WriteString(fmt.Sprintf("Datasets processed: %d\n", len(datasets)))
	for _, ds := range datasets {
		b.WriteString(fmt.Sprintf("  - %s\n", ds))
	}
	b.WriteString(fmt.Sprintf("Total items:        %d\n", len(results)))

	verified := 0
	for _, ir := range results {
		if ir.Verified {
			verified++
		}
	}
	b.WriteString(fmt.Sprintf("Verified:           %d/%d\n", verified, len(results)))

	return Section{
		Heading: "Executive Summary",
		Body:    b.String(),
	}
}

// buildDatasetResults produces per-item prediction detail.
func buildDatasetResults(results []kinds.InferenceResult) Section {
	var b strings.Builder

	if len(results) == 0 {
		b.WriteString("(no results)\n")
		return Section{Heading: "Dataset Results", Body: b.String()}
	}

	for i, ir := range results {
		status := "unverified"
		if ir.Verified {
			status = "verified"
		}
		b.WriteString(fmt.Sprintf("  [%d] prediction=%-12s confidence=%d/%-4d %s\n",
			i+1, ir.Final, ir.Confidence.Num, ir.Confidence.Den, status))
	}

	return Section{
		Heading: "Dataset Results",
		Body:    b.String(),
	}
}

// buildProofs produces the full proof trace section using explain.FormatProof.
func buildProofs(results []kinds.InferenceResult) Section {
	var b strings.Builder

	if len(results) == 0 {
		b.WriteString("(no proofs)\n")
		return Section{Heading: "Proofs", Body: b.String()}
	}

	for i, ir := range results {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("-- Item %d --\n", i+1))
		b.WriteString(explain.FormatProof(explain.BuildProof(ir)))
	}

	return Section{
		Heading: "Proofs",
		Body:    b.String(),
	}
}
