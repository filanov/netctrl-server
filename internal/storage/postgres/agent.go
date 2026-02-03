package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// CreateAgent creates a new agent
func (s *Storage) CreateAgent(ctx context.Context, agent *v1.Agent) error {
	networkInterfaces, err := json.Marshal(agent.NetworkInterfaces)
	if err != nil {
		return fmt.Errorf("failed to marshal network interfaces: %w", err)
	}

	query := `
		INSERT INTO agents (
			id, cluster_id, hostname, ip_address, version, status,
			last_seen, created_at, updated_at, hardware_collected, network_interfaces
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = s.pool.Exec(ctx, query,
		agent.Id,
		agent.ClusterId,
		agent.Hostname,
		agent.IpAddress,
		agent.Version,
		agent.Status.String(),
		agent.LastSeen.AsTime(),
		agent.CreatedAt.AsTime(),
		agent.UpdatedAt.AsTime(),
		agent.HardwareCollected,
		networkInterfaces,
	)

	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	return nil
}

// GetAgent retrieves an agent by ID
func (s *Storage) GetAgent(ctx context.Context, id string) (*v1.Agent, error) {
	query := `
		SELECT id, cluster_id, hostname, ip_address, version, status,
		       last_seen, created_at, updated_at, hardware_collected, network_interfaces
		FROM agents
		WHERE id = $1
	`

	var agent v1.Agent
	var statusStr string
	var lastSeen, createdAt, updatedAt time.Time
	var networkInterfacesJSON []byte

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&agent.Id,
		&agent.ClusterId,
		&agent.Hostname,
		&agent.IpAddress,
		&agent.Version,
		&statusStr,
		&lastSeen,
		&createdAt,
		&updatedAt,
		&agent.HardwareCollected,
		&networkInterfacesJSON,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Parse status
	agent.Status = parseAgentStatus(statusStr)

	// Parse timestamps
	agent.LastSeen = timestamppb.New(lastSeen)
	agent.CreatedAt = timestamppb.New(createdAt)
	agent.UpdatedAt = timestamppb.New(updatedAt)

	// Parse network interfaces
	if len(networkInterfacesJSON) > 0 && string(networkInterfacesJSON) != "[]" {
		if err := json.Unmarshal(networkInterfacesJSON, &agent.NetworkInterfaces); err != nil {
			return nil, fmt.Errorf("failed to unmarshal network interfaces: %w", err)
		}
	}

	return &agent, nil
}

// ListAgents lists agents, optionally filtered by cluster
func (s *Storage) ListAgents(ctx context.Context, clusterID string) ([]*v1.Agent, error) {
	var query string
	var args []interface{}

	if clusterID != "" {
		query = `
			SELECT id, cluster_id, hostname, ip_address, version, status,
			       last_seen, created_at, updated_at, hardware_collected, network_interfaces
			FROM agents
			WHERE cluster_id = $1
			ORDER BY created_at DESC
		`
		args = append(args, clusterID)
	} else {
		query = `
			SELECT id, cluster_id, hostname, ip_address, version, status,
			       last_seen, created_at, updated_at, hardware_collected, network_interfaces
			FROM agents
			ORDER BY created_at DESC
		`
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []*v1.Agent
	for rows.Next() {
		var agent v1.Agent
		var statusStr string
		var lastSeen, createdAt, updatedAt time.Time
		var networkInterfacesJSON []byte

		err := rows.Scan(
			&agent.Id,
			&agent.ClusterId,
			&agent.Hostname,
			&agent.IpAddress,
			&agent.Version,
			&statusStr,
			&lastSeen,
			&createdAt,
			&updatedAt,
			&agent.HardwareCollected,
			&networkInterfacesJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		agent.Status = parseAgentStatus(statusStr)
		agent.LastSeen = timestamppb.New(lastSeen)
		agent.CreatedAt = timestamppb.New(createdAt)
		agent.UpdatedAt = timestamppb.New(updatedAt)

		// Parse network interfaces
		if len(networkInterfacesJSON) > 0 && string(networkInterfacesJSON) != "[]" {
			if err := json.Unmarshal(networkInterfacesJSON, &agent.NetworkInterfaces); err != nil {
				return nil, fmt.Errorf("failed to unmarshal network interfaces: %w", err)
			}
		}

		agents = append(agents, &agent)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating agents: %w", err)
	}

	return agents, nil
}

// UpdateAgent updates an existing agent
func (s *Storage) UpdateAgent(ctx context.Context, agent *v1.Agent) error {
	networkInterfaces, err := json.Marshal(agent.NetworkInterfaces)
	if err != nil {
		return fmt.Errorf("failed to marshal network interfaces: %w", err)
	}

	query := `
		UPDATE agents
		SET cluster_id = $2, hostname = $3, ip_address = $4, version = $5,
		    status = $6, last_seen = $7, updated_at = $8,
		    hardware_collected = $9, network_interfaces = $10
		WHERE id = $1
	`

	result, err := s.pool.Exec(ctx, query,
		agent.Id,
		agent.ClusterId,
		agent.Hostname,
		agent.IpAddress,
		agent.Version,
		agent.Status.String(),
		agent.LastSeen.AsTime(),
		agent.UpdatedAt.AsTime(),
		agent.HardwareCollected,
		networkInterfaces,
	)

	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("agent not found")
	}

	return nil
}

// DeleteAgent deletes an agent by ID
func (s *Storage) DeleteAgent(ctx context.Context, id string) error {
	query := `DELETE FROM agents WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("agent not found")
	}

	return nil
}

// parseAgentStatus converts string status to enum
func parseAgentStatus(status string) v1.AgentStatus {
	switch status {
	case "AGENT_STATUS_ACTIVE":
		return v1.AgentStatus_AGENT_STATUS_ACTIVE
	case "AGENT_STATUS_INACTIVE":
		return v1.AgentStatus_AGENT_STATUS_INACTIVE
	default:
		return v1.AgentStatus_AGENT_STATUS_UNSPECIFIED
	}
}
