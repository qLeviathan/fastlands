package kinds

import "github.com/qleviathan/fastlands/basis"

// Datum is a single data point - features are integers, not floats
type Datum struct {
	Features map[string]int64
	Label    string
}

// DataSet is a collection of datums
type DataSet struct {
	Name   string
	Domain string
	Items  []Datum
}

// ModelResult is the output of a single inference
type ModelResult struct {
	Prediction string
	Confidence basis.Rational // exact, no floats
	Kind       Kind
	ProofTrace []string
}

// InferenceResult wraps a full pipeline result
type InferenceResult struct {
	Result     ModelResult
	Final      string
	Verified   bool
	Confidence basis.Rational
	Explained  bool
}
