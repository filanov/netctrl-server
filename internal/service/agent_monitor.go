package service

import (
	"context"
	"log"
	"time"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
	"github.com/filanov/netctrl-server/internal/storage"
)

const (
	// PollIntervalSeconds is the expected agent poll interval
	PollIntervalSeconds = 60

	// InactiveThresholdMultiplier determines how many missed polls before marking inactive
	InactiveThresholdMultiplier = 3

	// MonitorCheckInterval is how often the monitor checks agent states
	MonitorCheckInterval = 30 * time.Second
)

// AgentMonitor monitors agent health and updates their status
type AgentMonitor struct {
	storage storage.Storage
	stopCh  chan struct{}
}

// NewAgentMonitor creates a new agent monitor
func NewAgentMonitor(storage storage.Storage) *AgentMonitor {
	return &AgentMonitor{
		storage: storage,
		stopCh:  make(chan struct{}),
	}
}

// Start begins the agent monitoring loop
func (m *AgentMonitor) Start(ctx context.Context) {
	log.Println("Starting agent monitor...")
	ticker := time.NewTicker(MonitorCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Agent monitor stopping due to context cancellation")
			return
		case <-m.stopCh:
			log.Println("Agent monitor stopped")
			return
		case <-ticker.C:
			m.checkAgentStates(ctx)
		}
	}
}

// Stop stops the agent monitor
func (m *AgentMonitor) Stop() {
	close(m.stopCh)
}

// CheckAgentStatesOnce performs a single check of all agent states (exposed for testing)
func (m *AgentMonitor) CheckAgentStatesOnce(ctx context.Context) {
	m.checkAgentStates(ctx)
}

// checkAgentStates checks all agents and updates their status based on last_seen
func (m *AgentMonitor) checkAgentStates(ctx context.Context) {
	// List all agents
	agents, err := m.storage.ListAgents(ctx, "")
	if err != nil {
		log.Printf("Failed to list agents for monitoring: %v", err)
		return
	}

	// Calculate inactivity threshold: 3 poll intervals
	inactiveThreshold := time.Duration(PollIntervalSeconds*InactiveThresholdMultiplier) * time.Second
	now := time.Now()

	for _, agent := range agents {
		if agent.LastSeen == nil {
			continue
		}

		lastSeenTime := agent.LastSeen.AsTime()
		timeSinceLastSeen := now.Sub(lastSeenTime)

		// Check if agent should be marked as inactive
		if timeSinceLastSeen > inactiveThreshold && agent.Status == v1.AgentStatus_AGENT_STATUS_ACTIVE {
			log.Printf("Marking agent %s as inactive (last seen: %v ago)", agent.Id, timeSinceLastSeen)
			agent.Status = v1.AgentStatus_AGENT_STATUS_INACTIVE

			if err := m.storage.UpdateAgent(ctx, agent); err != nil {
				log.Printf("Failed to update agent %s status: %v", agent.Id, err)
			}
		}
	}
}
