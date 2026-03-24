package swarm

import "time"

// ScaleDirection indicates whether to add or remove agents.
type ScaleDirection int

const (
	ScaleNone ScaleDirection = iota
	ScaleUp
	ScaleDown
)

// ScaleDecision is a proposed scaling action.
type ScaleDecision struct {
	Direction ScaleDirection
	Count     int
	Reason    string
}

// AutoScaler evaluates agent load and proposes scaling decisions.
type AutoScaler struct {
	MinAgents     int
	MaxAgents     int
	QueueThreshold int
	IdleTimeout   time.Duration
	lastActivity  map[string]time.Time
}

// NewAutoScaler creates an autoscaler with default thresholds.
func NewAutoScaler() *AutoScaler {
	return &AutoScaler{
		MinAgents:      1,
		MaxAgents:      8,
		QueueThreshold: 5,
		IdleTimeout:    30 * time.Second,
		lastActivity:   make(map[string]time.Time),
	}
}

// RecordActivity marks an agent as active now.
func (as *AutoScaler) RecordActivity(agentID string) {
	as.lastActivity[agentID] = time.Now()
}

// Evaluate proposes scaling decisions based on current state.
func (as *AutoScaler) Evaluate(currentAgents int, queueDepth int) []ScaleDecision {
	var decisions []ScaleDecision

	// Scale UP: queue depth exceeds threshold and room to grow
	if queueDepth > as.QueueThreshold && currentAgents < as.MaxAgents {
		add := queueDepth / as.QueueThreshold
		if currentAgents+add > as.MaxAgents {
			add = as.MaxAgents - currentAgents
		}
		if add > 0 {
			decisions = append(decisions, ScaleDecision{
				Direction: ScaleUp,
				Count:     add,
				Reason:    "queue depth exceeds threshold",
			})
		}
	}

	// Scale DOWN: idle agents beyond minimum
	now := time.Now()
	idle := 0
	for _, lastActive := range as.lastActivity {
		if now.Sub(lastActive) > as.IdleTimeout {
			idle++
		}
	}
	if idle > 0 && currentAgents > as.MinAgents {
		remove := idle
		if currentAgents-remove < as.MinAgents {
			remove = currentAgents - as.MinAgents
		}
		if remove > 0 {
			decisions = append(decisions, ScaleDecision{
				Direction: ScaleDown,
				Count:     remove,
				Reason:    "agents idle beyond timeout",
			})
		}
	}

	return decisions
}

// IdleAgents returns agents that have been idle longer than the timeout.
func (as *AutoScaler) IdleAgents() []string {
	now := time.Now()
	var idle []string
	for id, lastActive := range as.lastActivity {
		if now.Sub(lastActive) > as.IdleTimeout {
			idle = append(idle, id)
		}
	}
	return idle
}
