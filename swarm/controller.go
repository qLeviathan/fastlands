package swarm

import (
	"fmt"
	"sync"

	"github.com/qleviathan/fastlands/agents"
	"github.com/qleviathan/fastlands/kinds"
	"github.com/qleviathan/fastlands/memory"
	"github.com/qleviathan/fastlands/self"
)

// Controller manages agent lifecycle, mesh networking, and task dispatch.
type Controller struct {
	Agents   map[string]agents.Agent
	Mesh     *MeshNetwork
	Scaler   *AutoScaler
	Memory   *memory.Store
	Boss     *agents.SuperClaudeAgent
	Identity *self.Identity
	Anchors  []self.Anchor // anchors issued to spawned agents
	mu       sync.Mutex
	peak     int
	log      []string
}

// NewController creates a swarm controller with boss and mesh.
func NewController(mem *memory.Store) *Controller {
	boss := agents.NewSuperClaude("boss-0")
	mesh := NewMesh(32)
	mesh.Register(boss.ID())

	identity := self.NewIdentity()

	c := &Controller{
		Agents:   map[string]agents.Agent{boss.ID(): boss},
		Mesh:     mesh,
		Scaler:   NewAutoScaler(),
		Memory:   mem,
		Boss:     boss,
		Identity: identity,
		peak:     1,
	}

	// Anchor the boss immediately.
	bossAnchor := self.BuildAnchor(identity, boss.ID(), string(agents.RoleBoss))
	c.Anchors = append(c.Anchors, bossAnchor)
	c.log = append(c.log, fmt.Sprintf("anchored %s: %s", boss.ID(), bossAnchor.Digest()))

	return c
}

// Spawn creates and registers a new agent, issuing it an anchor
// so it immediately understands the system it's joining.
func (c *Controller) Spawn(agent agents.Agent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Agents[agent.ID()] = agent
	c.Mesh.Register(agent.ID())
	c.Mesh.AddRoute(c.Boss.ID(), agent.ID())
	c.Scaler.RecordActivity(agent.ID())

	current := len(c.Agents)
	if current > c.peak {
		c.peak = current
	}

	// Build and issue an anchor for the new agent.
	anchor := self.BuildAnchor(c.Identity, agent.ID(), string(agent.Role()))
	c.Anchors = append(c.Anchors, anchor)

	c.log = append(c.log, fmt.Sprintf("spawned %s (role=%s)", agent.ID(), agent.Role()))
	c.log = append(c.log, fmt.Sprintf("anchored %s: %s", agent.ID(), anchor.Digest()))
}

// Retire removes and unregisters an agent.
func (c *Controller) Retire(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if id == c.Boss.ID() {
		return // never retire the boss
	}
	delete(c.Agents, id)
	c.Mesh.Unregister(id)
	delete(c.Scaler.lastActivity, id)
	c.log = append(c.log, fmt.Sprintf("retired %s", id))
}

// Submit sends a datum to a model agent for inference.
func (c *Controller) Submit(agentID string, datum kinds.Datum) (kinds.InferenceResult, error) {
	c.mu.Lock()
	agent, ok := c.Agents[agentID]
	c.mu.Unlock()

	if !ok {
		return kinds.InferenceResult{}, fmt.Errorf("swarm: unknown agent %s", agentID)
	}

	c.Scaler.RecordActivity(agentID)

	msg := agents.Message{
		Type:    agents.MsgRequest,
		From:    c.Boss.ID(),
		To:      agentID,
		Payload: datum,
	}

	resp, err := agent.Process(msg)
	if err != nil {
		return kinds.InferenceResult{}, err
	}

	ir, ok := resp.Payload.(kinds.InferenceResult)
	if !ok {
		return kinds.InferenceResult{}, fmt.Errorf("swarm: unexpected response type from %s", agentID)
	}

	return ir, nil
}

// RunDataset processes an entire dataset through the swarm.
func (c *Controller) RunDataset(ds kinds.DataSet, inferFn agents.InferFunc) []kinds.InferenceResult {
	// Ensure at least one model agent exists
	modelID := "model-0"
	c.mu.Lock()
	if _, ok := c.Agents[modelID]; !ok {
		c.mu.Unlock()
		c.Spawn(agents.NewModelAgent(modelID, inferFn))
	} else {
		c.mu.Unlock()
	}

	// Ensure verifier exists
	verifierID := "verifier-0"
	c.mu.Lock()
	if _, ok := c.Agents[verifierID]; !ok {
		c.mu.Unlock()
		c.Spawn(agents.NewVerifier(verifierID))
	} else {
		c.mu.Unlock()
	}

	results := make([]kinds.InferenceResult, 0, len(ds.Items))

	for _, datum := range ds.Items {
		ir, err := c.Submit(modelID, datum)
		if err != nil {
			c.log = append(c.log, fmt.Sprintf("error: %v", err))
			continue
		}

		// Verify
		c.mu.Lock()
		verifier, ok := c.Agents[verifierID]
		c.mu.Unlock()
		if ok {
			vmsg := agents.Message{
				Type:    agents.MsgVerify,
				From:    c.Boss.ID(),
				To:      verifierID,
				Payload: ir,
			}
			vresp, err := verifier.Process(vmsg)
			if err == nil {
				if vir, ok := vresp.Payload.(kinds.InferenceResult); ok {
					ir = vir
				}
			}
		}

		// Report to boss
		c.Boss.Process(agents.Message{
			Type:    agents.MsgResult,
			From:    modelID,
			Payload: ir,
		})

		results = append(results, ir)
	}

	return results
}

// Peak returns the peak number of concurrent agents.
func (c *Controller) Peak() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.peak
}

// Log returns the controller's event log.
func (c *Controller) Log() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.log))
	copy(out, c.log)
	return out
}

// Summary returns a human-readable swarm summary.
func (c *Controller) Summary() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return fmt.Sprintf("swarm: %d agents (peak %d), %d events",
		len(c.Agents), c.peak, len(c.log))
}
