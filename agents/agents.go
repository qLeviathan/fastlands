// Package agents provides the agent framework for caslo.
// Four agent types coordinate via message passing with isolated state.
package agents

import (
	"fmt"
	"sync"

	"github.com/qleviathan/fastlands/basis"
	"github.com/qleviathan/fastlands/kinds"
)

// Role identifies an agent's function in the system.
type Role string

const (
	RoleBoss     Role = "boss"
	RoleVerifier Role = "verifier"
	RoleExpert   Role = "expert"
	RoleModel    Role = "model"
)

// MsgType identifies what a message carries.
type MsgType string

const (
	MsgDirective MsgType = "directive"
	MsgRequest   MsgType = "request"
	MsgResult    MsgType = "result"
	MsgVerify    MsgType = "verify"
	MsgReview    MsgType = "review"
	MsgReport    MsgType = "report"
)

// Message is the unit of inter-agent communication.
type Message struct {
	Type    MsgType
	From    string
	To      string
	Payload interface{}
	Trace   []string
}

// Agent is the interface all agents implement.
type Agent interface {
	ID() string
	Role() Role
	Process(msg Message) (Message, error)
}

// InferFunc is the function signature for running inference on a datum.
type InferFunc func(kinds.Datum) (kinds.ModelResult, error)

// --- SuperClaude (Boss) ---

// SuperClaudeAgent orchestrates the pipeline, dispatches work, collects results.
type SuperClaudeAgent struct {
	id      string
	results []kinds.InferenceResult
	log     []string
	mu      sync.Mutex
}

func NewSuperClaude(id string) *SuperClaudeAgent {
	return &SuperClaudeAgent{id: id}
}

func (a *SuperClaudeAgent) ID() string  { return a.id }
func (a *SuperClaudeAgent) Role() Role  { return RoleBoss }

func (a *SuperClaudeAgent) Process(msg Message) (Message, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch msg.Type {
	case MsgResult:
		if ir, ok := msg.Payload.(kinds.InferenceResult); ok {
			a.results = append(a.results, ir)
			a.log = append(a.log, fmt.Sprintf("collected result: %s (verified=%v)", ir.Final, ir.Verified))
		}
		return Message{Type: MsgReport, From: a.id, Trace: []string{"ack"}}, nil

	case MsgDirective:
		a.log = append(a.log, fmt.Sprintf("directive received: %v", msg.Payload))
		return Message{Type: MsgReport, From: a.id, Trace: []string{"directive-ack"}}, nil

	default:
		return Message{Type: MsgReport, From: a.id, Trace: []string{"noop"}}, nil
	}
}

func (a *SuperClaudeAgent) Results() []kinds.InferenceResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]kinds.InferenceResult, len(a.results))
	copy(out, a.results)
	return out
}

func (a *SuperClaudeAgent) Log() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]string, len(a.log))
	copy(out, a.log)
	return out
}

// --- Verifier ---

// VerifierAgent validates inference results: proof traces exist,
// confidence is well-formed, no empty steps.
type VerifierAgent struct {
	id  string
	mu  sync.Mutex
	log []string
}

func NewVerifier(id string) *VerifierAgent {
	return &VerifierAgent{id: id}
}

func (a *VerifierAgent) ID() string  { return a.id }
func (a *VerifierAgent) Role() Role  { return RoleVerifier }

func (a *VerifierAgent) Process(msg Message) (Message, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if msg.Type != MsgVerify {
		return Message{Type: MsgReport, From: a.id, Trace: []string{"expected verify msg"}}, nil
	}

	ir, ok := msg.Payload.(kinds.InferenceResult)
	if !ok {
		return Message{}, fmt.Errorf("verifier: invalid payload type")
	}

	verified := true
	var trace []string

	// Check proof trace exists
	if len(ir.Result.ProofTrace) == 0 {
		verified = false
		trace = append(trace, "FAIL: empty proof trace")
	}

	// Check confidence is well-formed (denominator > 0)
	if ir.Result.Confidence.Den <= 0 {
		verified = false
		trace = append(trace, "FAIL: invalid confidence denominator")
	}

	// Check numerator is non-negative and <= denominator
	if ir.Result.Confidence.Num < 0 || ir.Result.Confidence.Num > ir.Result.Confidence.Den {
		verified = false
		trace = append(trace, "FAIL: confidence out of range")
	}

	// Check no empty proof steps
	for i, step := range ir.Result.ProofTrace {
		if step == "" {
			verified = false
			trace = append(trace, fmt.Sprintf("FAIL: empty proof step at index %d", i))
		}
	}

	if verified {
		trace = append(trace, "PASS: all checks passed")
	}

	ir.Verified = verified
	a.log = append(a.log, fmt.Sprintf("verified %s: %v", ir.Final, verified))

	return Message{
		Type:    MsgResult,
		From:    a.id,
		Payload: ir,
		Trace:   trace,
	}, nil
}

// --- PhD (Domain Expert) ---

// PhDAgent analyzes domains, recommends reasoning kinds, reviews results.
type PhDAgent struct {
	id     string
	domain string
	mu     sync.Mutex
	log    []string
}

func NewPhD(id, domain string) *PhDAgent {
	return &PhDAgent{id: id, domain: domain}
}

func (a *PhDAgent) ID() string  { return a.id }
func (a *PhDAgent) Role() Role  { return RoleExpert }

func (a *PhDAgent) Process(msg Message) (Message, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch msg.Type {
	case MsgReview:
		ir, ok := msg.Payload.(kinds.InferenceResult)
		if !ok {
			return Message{}, fmt.Errorf("phd: invalid payload")
		}
		review := fmt.Sprintf("domain=%s prediction=%s kind=%s",
			a.domain, ir.Final, ir.Result.Kind.Name)
		a.log = append(a.log, review)
		return Message{
			Type:    MsgReport,
			From:    a.id,
			Payload: review,
			Trace:   []string{review},
		}, nil

	case MsgRequest:
		// Recommend reasoning kind for domain
		rec := kinds.KindLogicPrograms
		a.log = append(a.log, fmt.Sprintf("recommended %s for domain %s", rec.Name, a.domain))
		return Message{
			Type:    MsgResult,
			From:    a.id,
			Payload: rec,
			Trace:   []string{fmt.Sprintf("recommend: %s", rec.ID)},
		}, nil

	default:
		return Message{Type: MsgReport, From: a.id}, nil
	}
}

// --- Model Agent ---

// ModelAgent wraps an inference function. Runs Infer() on individual datums.
type ModelAgent struct {
	id      string
	inferFn InferFunc
	mu      sync.Mutex
	log     []string
}

func NewModelAgent(id string, fn InferFunc) *ModelAgent {
	return &ModelAgent{id: id, inferFn: fn}
}

func (a *ModelAgent) ID() string  { return a.id }
func (a *ModelAgent) Role() Role  { return RoleModel }

func (a *ModelAgent) Process(msg Message) (Message, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if msg.Type != MsgRequest {
		return Message{Type: MsgReport, From: a.id}, nil
	}

	datum, ok := msg.Payload.(kinds.Datum)
	if !ok {
		return Message{}, fmt.Errorf("model: expected Datum payload")
	}

	result, err := a.inferFn(datum)
	if err != nil {
		a.log = append(a.log, fmt.Sprintf("inference error: %v", err))
		return Message{}, err
	}

	ir := kinds.InferenceResult{
		Result:     result,
		Final:      result.Prediction,
		Verified:   false,
		Confidence: result.Confidence,
		Explained:  len(result.ProofTrace) > 0,
	}

	label := ""
	if datum.Label != "" {
		label = fmt.Sprintf(" (expected=%s)", datum.Label)
	}
	a.log = append(a.log, fmt.Sprintf("inferred %s%s conf=%d/%d",
		result.Prediction, label, result.Confidence.Num, result.Confidence.Den))

	return Message{
		Type:    MsgResult,
		From:    a.id,
		Payload: ir,
		Trace:   result.ProofTrace,
	}, nil
}

// Evaluate computes accuracy over a set of results against expected labels.
// Returns (correct, total) as integers — no floats.
func Evaluate(results []kinds.InferenceResult, expected []string) (correct, total int) {
	total = len(results)
	if len(expected) < total {
		total = len(expected)
	}
	for i := 0; i < total; i++ {
		if results[i].Final == expected[i] {
			correct++
		}
	}
	return correct, total
}

// EvaluateRational returns accuracy as a basis.Rational.
func EvaluateRational(results []kinds.InferenceResult, expected []string) basis.Rational {
	c, t := Evaluate(results, expected)
	if t == 0 {
		return basis.RatZero()
	}
	return basis.NewRational(int64(c), int64(t))
}
