package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/filanov/netctrl-server/internal/storage"
	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

// AgentService implements the AgentService gRPC service
type AgentService struct {
	v1.UnimplementedAgentServiceServer
	storage storage.Storage
}

// NewAgentService creates a new agent service
func NewAgentService(store storage.Storage) *AgentService {
	return &AgentService{
		storage: store,
	}
}

// RegisterAgent registers or updates an agent to a cluster
func (s *AgentService) RegisterAgent(ctx context.Context, req *v1.RegisterAgentRequest) (*v1.RegisterAgentResponse, error) {
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Verify cluster exists
	exists, err := s.storage.ClusterExists(ctx, req.ClusterId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check cluster existence: %v", err))
	}
	if !exists {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("cluster %s not found", req.ClusterId))
	}

	now := timestamppb.Now()

	// Check if agent already exists
	existingAgent, err := s.storage.GetAgent(ctx, req.Id)
	if err == nil {
		// Agent exists, update it
		existingAgent.ClusterId = req.ClusterId
		existingAgent.Hostname = req.Hostname
		existingAgent.IpAddress = req.IpAddress
		existingAgent.Version = req.Version
		existingAgent.Status = v1.AgentStatus_AGENT_STATUS_ACTIVE
		existingAgent.LastSeen = now
		existingAgent.UpdatedAt = now

		if err := s.storage.UpdateAgent(ctx, existingAgent); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update agent: %v", err))
		}

		log.Printf("Agent re-registered: id=%s, cluster=%s, hostname=%s, ip=%s",
			existingAgent.Id, existingAgent.ClusterId, existingAgent.Hostname, existingAgent.IpAddress)

		return &v1.RegisterAgentResponse{
			Agent: existingAgent,
		}, nil
	}

	// Agent doesn't exist, create new one
	agent := &v1.Agent{
		Id:        req.Id,
		ClusterId: req.ClusterId,
		Hostname:  req.Hostname,
		IpAddress: req.IpAddress,
		Version:   req.Version,
		Status:    v1.AgentStatus_AGENT_STATUS_ACTIVE,
		LastSeen:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.storage.CreateAgent(ctx, agent); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create agent: %v", err))
	}

	log.Printf("Agent registered: id=%s, cluster=%s, hostname=%s, ip=%s",
		agent.Id, agent.ClusterId, agent.Hostname, agent.IpAddress)

	return &v1.RegisterAgentResponse{
		Agent: agent,
	}, nil
}

// GetAgent retrieves an agent by ID
func (s *AgentService) GetAgent(ctx context.Context, req *v1.GetAgentRequest) (*v1.GetAgentResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "agent ID is required")
	}

	agent, err := s.storage.GetAgent(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("agent not found: %s", req.Id))
	}

	return &v1.GetAgentResponse{
		Agent: agent,
	}, nil
}

// ListAgents lists all agents, optionally filtered by cluster
func (s *AgentService) ListAgents(ctx context.Context, req *v1.ListAgentsRequest) (*v1.ListAgentsResponse, error) {
	agents, err := s.storage.ListAgents(ctx, req.ClusterId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list agents: %v", err))
	}

	return &v1.ListAgentsResponse{
		Agents: agents,
	}, nil
}

// UnregisterAgent removes an agent
func (s *AgentService) UnregisterAgent(ctx context.Context, req *v1.UnregisterAgentRequest) (*v1.UnregisterAgentResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "agent ID is required")
	}

	if err := s.storage.DeleteAgent(ctx, req.Id); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("agent not found: %s", req.Id))
	}

	return &v1.UnregisterAgentResponse{
		Success: true,
	}, nil
}

// GetInstructions polls for pending instructions and updates agent heartbeat
func (s *AgentService) GetInstructions(ctx context.Context, req *v1.GetInstructionsRequest) (*v1.GetInstructionsResponse, error) {
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent ID is required")
	}

	// Verify agent exists and update last_seen (implicit heartbeat)
	agent, err := s.storage.GetAgent(ctx, req.AgentId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("agent not found: %s", req.AgentId))
	}

	// Update agent's last_seen timestamp and set status to active
	now := timestamppb.Now()
	agent.LastSeen = now
	agent.UpdatedAt = now
	agent.Status = v1.AgentStatus_AGENT_STATUS_ACTIVE

	if err := s.storage.UpdateAgent(ctx, agent); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update agent heartbeat: %v", err))
	}

	// Generate instructions for the agent
	instructions := s.generateInstructions(agent)

	// Return instructions with default poll interval
	return &v1.GetInstructionsResponse{
		Instructions:        instructions,
		PollIntervalSeconds: 60, // Default: poll every 60 seconds (1 minute)
		ServerTime:          now,
	}, nil
}

// SubmitInstructionResult processes the result of a completed instruction
func (s *AgentService) SubmitInstructionResult(ctx context.Context, req *v1.SubmitInstructionResultRequest) (*v1.SubmitInstructionResultResponse, error) {
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent ID is required")
	}
	if req.InstructionId == "" {
		return nil, status.Error(codes.InvalidArgument, "instruction ID is required")
	}
	if req.Result == nil {
		return nil, status.Error(codes.InvalidArgument, "result is required")
	}

	// Verify agent exists
	agent, err := s.storage.GetAgent(ctx, req.AgentId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("agent not found: %s", req.AgentId))
	}

	// Process the instruction result
	if err := s.processInstructionResult(agent, req.Result); err != nil {
		log.Printf("Failed to process instruction result for agent %s: %v", agent.Id, err)
		return &v1.SubmitInstructionResultResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to process result: %v", err),
		}, nil
	}

	// Update agent in storage with the processed result
	agent.UpdatedAt = timestamppb.Now()
	if err := s.storage.UpdateAgent(ctx, agent); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update agent: %v", err))
	}

	return &v1.SubmitInstructionResultResponse{
		Success: true,
		Message: "Result processed successfully",
	}, nil
}

// processInstructionResult processes the result from an instruction execution
func (s *AgentService) processInstructionResult(agent *v1.Agent, result *v1.InstructionResult) error {
	// Process based on instruction type
	switch result.InstructionType {
	case v1.InstructionType_INSTRUCTION_TYPE_COLLECT_HARDWARE:
		hwResult := result.GetHardwareCollection()
		if hwResult == nil {
			return fmt.Errorf("hardware collection result is missing")
		}

		// Update agent with hardware information (even if empty)
		agent.NetworkInterfaces = hwResult.NetworkInterfaces
		agent.HardwareCollected = true

		if len(hwResult.NetworkInterfaces) > 0 {
			log.Printf("Hardware collected for agent %s: %d NICs", agent.Id, len(hwResult.NetworkInterfaces))
		} else {
			log.Printf("Hardware collected for agent %s: no Mellanox NICs found", agent.Id)
		}

	case v1.InstructionType_INSTRUCTION_TYPE_HEALTH_CHECK:
		healthResult := result.GetHealthCheck()
		if healthResult == nil {
			return fmt.Errorf("health check result is missing")
		}
		log.Printf("Health check from agent %s: healthy=%v", agent.Id, healthResult.Healthy)

	default:
		log.Printf("Unknown instruction type: %v", result.InstructionType)
	}

	return nil
}

// generateInstructions creates instructions for an agent based on its state
func (s *AgentService) generateInstructions(agent *v1.Agent) []*v1.Instruction {
	var instructions []*v1.Instruction

	// Request hardware collection if not yet completed
	if !agent.HardwareCollected {
		instruction := &v1.Instruction{
			Id:        uuid.New().String(),
			Type:      v1.InstructionType_INSTRUCTION_TYPE_COLLECT_HARDWARE,
			Payload:   `{}`,
			CreatedAt: timestamppb.Now(),
		}
		instructions = append(instructions, instruction)
		log.Printf("Requesting hardware collection from agent %s", agent.Id)
	}

	// Future: Add other instruction types here
	// - Health checks
	// - Command execution
	// - Configuration updates

	return instructions
}

// validateRegisterRequest validates the agent registration request
func (s *AgentService) validateRegisterRequest(req *v1.RegisterAgentRequest) error {
	if req.Id == "" {
		return fmt.Errorf("agent ID is required")
	}
	if req.ClusterId == "" {
		return fmt.Errorf("cluster ID is required")
	}
	return nil
}
