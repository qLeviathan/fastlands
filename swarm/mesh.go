// Package swarm provides agent orchestration with mesh networking,
// dynamic scaling, and lifecycle management.
package swarm

import (
	"fmt"
	"sync"

	"github.com/qleviathan/fastlands/agents"
)

// MeshNetwork provides peer-to-peer agent communication via channels.
type MeshNetwork struct {
	channels map[string]chan agents.Message
	routes   map[string][]string // who can talk to whom
	mu       sync.RWMutex
	bufSize  int
}

// NewMesh creates a mesh network with the given channel buffer size.
func NewMesh(bufSize int) *MeshNetwork {
	return &MeshNetwork{
		channels: make(map[string]chan agents.Message),
		routes:   make(map[string][]string),
		bufSize:  bufSize,
	}
}

// Register adds an agent to the mesh.
func (m *MeshNetwork) Register(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.channels[id]; !exists {
		m.channels[id] = make(chan agents.Message, m.bufSize)
	}
}

// Unregister removes an agent from the mesh.
func (m *MeshNetwork) Unregister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ch, exists := m.channels[id]; exists {
		close(ch)
		delete(m.channels, id)
	}
	delete(m.routes, id)
	// Remove from other agents' routes
	for k, v := range m.routes {
		filtered := v[:0]
		for _, peer := range v {
			if peer != id {
				filtered = append(filtered, peer)
			}
		}
		m.routes[k] = filtered
	}
}

// AddRoute creates a bidirectional route between two agents.
func (m *MeshNetwork) AddRoute(from, to string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addOneWay(from, to)
	m.addOneWay(to, from)
}

func (m *MeshNetwork) addOneWay(from, to string) {
	for _, peer := range m.routes[from] {
		if peer == to {
			return
		}
	}
	m.routes[from] = append(m.routes[from], to)
}

// Send sends a message to a specific agent. Non-blocking; drops if full.
func (m *MeshNetwork) Send(to string, msg agents.Message) error {
	m.mu.RLock()
	ch, ok := m.channels[to]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("mesh: unknown agent %s", to)
	}
	select {
	case ch <- msg:
		return nil
	default:
		return fmt.Errorf("mesh: channel full for %s", to)
	}
}

// Receive reads one message for an agent. Non-blocking; returns false if empty.
func (m *MeshNetwork) Receive(id string) (agents.Message, bool) {
	m.mu.RLock()
	ch, ok := m.channels[id]
	m.mu.RUnlock()
	if !ok {
		return agents.Message{}, false
	}
	select {
	case msg := <-ch:
		return msg, true
	default:
		return agents.Message{}, false
	}
}

// Broadcast sends a message to all agents in the mesh.
func (m *MeshNetwork) Broadcast(msg agents.Message) int {
	m.mu.RLock()
	ids := make([]string, 0, len(m.channels))
	for id := range m.channels {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	sent := 0
	for _, id := range ids {
		if m.Send(id, msg) == nil {
			sent++
		}
	}
	return sent
}

// Peers returns the list of agents an agent can communicate with.
func (m *MeshNetwork) Peers(id string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, len(m.routes[id]))
	copy(out, m.routes[id])
	return out
}

// Size returns the number of registered agents.
func (m *MeshNetwork) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.channels)
}
