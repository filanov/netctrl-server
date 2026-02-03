package service_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/filanov/netctrl-server/internal/service"
	"github.com/filanov/netctrl-server/internal/storage/mock"
	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

var _ = Describe("AgentService", func() {
	var (
		agentService   *service.AgentService
		clusterService *service.ClusterService
		ctx            context.Context
		testClusterId  string
	)

	BeforeEach(func() {
		storage := mock.New()
		agentService = service.NewAgentService(storage)
		clusterService = service.NewClusterService(storage)
		ctx = context.Background()

		// Create a test cluster
		createReq := &v1.CreateClusterRequest{
			Name:        "test-cluster",
			Description: "Test cluster for agent tests",
		}
		createResp, err := clusterService.CreateCluster(ctx, createReq)
		Expect(err).NotTo(HaveOccurred())
		testClusterId = createResp.Cluster.Id
	})

	Describe("RegisterAgent", func() {
		It("should register a new agent successfully", func() {
			req := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
				IpAddress: "10.0.1.1",
				Version:   "1.0.0",
			}

			resp, err := agentService.RegisterAgent(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Agent).NotTo(BeNil())
			Expect(resp.Agent.Id).To(Equal("agent-1"))
			Expect(resp.Agent.ClusterId).To(Equal(testClusterId))
			Expect(resp.Agent.Hostname).To(Equal("node1"))
			Expect(resp.Agent.Status).To(Equal(v1.AgentStatus_AGENT_STATUS_ACTIVE))
			Expect(resp.Agent.CreatedAt).NotTo(BeNil())
			Expect(resp.Agent.LastSeen).NotTo(BeNil())
		})

		It("should update existing agent on re-registration", func() {
			// First registration
			req1 := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
				IpAddress: "10.0.1.1",
				Version:   "1.0.0",
			}
			resp1, err := agentService.RegisterAgent(ctx, req1)
			Expect(err).NotTo(HaveOccurred())
			createdAt := resp1.Agent.CreatedAt

			// Second registration with updated info
			req2 := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1-updated",
				IpAddress: "10.0.1.2",
				Version:   "1.0.1",
			}
			resp2, err := agentService.RegisterAgent(ctx, req2)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp2.Agent.Hostname).To(Equal("node1-updated"))
			Expect(resp2.Agent.IpAddress).To(Equal("10.0.1.2"))
			Expect(resp2.Agent.Version).To(Equal("1.0.1"))
			Expect(resp2.Agent.CreatedAt).To(Equal(createdAt))
			Expect(resp2.Agent.UpdatedAt.AsTime()).To(BeTemporally(">=", resp2.Agent.CreatedAt.AsTime()))
		})

		It("should return error when agent ID is missing", func() {
			req := &v1.RegisterAgentRequest{
				ClusterId: testClusterId,
				Hostname:  "node1",
			}

			_, err := agentService.RegisterAgent(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
			Expect(st.Message()).To(ContainSubstring("agent ID is required"))
		})

		It("should return error when cluster ID is missing", func() {
			req := &v1.RegisterAgentRequest{
				Id:       "agent-1",
				Hostname: "node1",
			}

			_, err := agentService.RegisterAgent(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
			Expect(st.Message()).To(ContainSubstring("cluster ID is required"))
		})

		It("should return error when cluster does not exist", func() {
			req := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: "non-existent-cluster",
				Hostname:  "node1",
			}

			_, err := agentService.RegisterAgent(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
			Expect(st.Message()).To(ContainSubstring("cluster"))
			Expect(st.Message()).To(ContainSubstring("not found"))
		})
	})

	Describe("GetAgent", func() {
		It("should retrieve existing agent", func() {
			// Register agent first
			registerReq := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
				IpAddress: "10.0.1.1",
				Version:   "1.0.0",
			}
			_, err := agentService.RegisterAgent(ctx, registerReq)
			Expect(err).NotTo(HaveOccurred())

			// Get agent
			getReq := &v1.GetAgentRequest{Id: "agent-1"}
			getResp, err := agentService.GetAgent(ctx, getReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResp.Agent.Id).To(Equal("agent-1"))
			Expect(getResp.Agent.Hostname).To(Equal("node1"))
		})

		It("should return error for non-existent agent", func() {
			req := &v1.GetAgentRequest{Id: "non-existent"}
			_, err := agentService.GetAgent(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})

		It("should return error when ID is empty", func() {
			req := &v1.GetAgentRequest{Id: ""}
			_, err := agentService.GetAgent(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})
	})

	Describe("ListAgents", func() {
		It("should return empty list when no agents exist", func() {
			req := &v1.ListAgentsRequest{}
			resp, err := agentService.ListAgents(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Agents).To(BeEmpty())
		})

		It("should return all agents", func() {
			// Register two agents
			req1 := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
			}
			req2 := &v1.RegisterAgentRequest{
				Id:        "agent-2",
				ClusterId: testClusterId,
				Hostname:  "node2",
			}
			_, err := agentService.RegisterAgent(ctx, req1)
			Expect(err).NotTo(HaveOccurred())
			_, err = agentService.RegisterAgent(ctx, req2)
			Expect(err).NotTo(HaveOccurred())

			// List agents
			listReq := &v1.ListAgentsRequest{}
			listResp, err := agentService.ListAgents(ctx, listReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(listResp.Agents).To(HaveLen(2))
		})

		It("should filter agents by cluster ID", func() {
			// Create second cluster
			createReq := &v1.CreateClusterRequest{
				Name: "cluster-2",
			}
			createResp, err := clusterService.CreateCluster(ctx, createReq)
			Expect(err).NotTo(HaveOccurred())
			cluster2Id := createResp.Cluster.Id

			// Register agents to different clusters
			req1 := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
			}
			req2 := &v1.RegisterAgentRequest{
				Id:        "agent-2",
				ClusterId: cluster2Id,
				Hostname:  "node2",
			}
			_, err = agentService.RegisterAgent(ctx, req1)
			Expect(err).NotTo(HaveOccurred())
			_, err = agentService.RegisterAgent(ctx, req2)
			Expect(err).NotTo(HaveOccurred())

			// List agents for first cluster
			listReq := &v1.ListAgentsRequest{ClusterId: testClusterId}
			listResp, err := agentService.ListAgents(ctx, listReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(listResp.Agents).To(HaveLen(1))
			Expect(listResp.Agents[0].ClusterId).To(Equal(testClusterId))
		})
	})

	Describe("UnregisterAgent", func() {
		It("should unregister existing agent", func() {
			// Register agent first
			registerReq := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
			}
			_, err := agentService.RegisterAgent(ctx, registerReq)
			Expect(err).NotTo(HaveOccurred())

			// Unregister agent
			unregisterReq := &v1.UnregisterAgentRequest{Id: "agent-1"}
			unregisterResp, err := agentService.UnregisterAgent(ctx, unregisterReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(unregisterResp.Success).To(BeTrue())

			// Verify agent is gone
			getReq := &v1.GetAgentRequest{Id: "agent-1"}
			_, err = agentService.GetAgent(ctx, getReq)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent agent", func() {
			req := &v1.UnregisterAgentRequest{Id: "non-existent"}
			_, err := agentService.UnregisterAgent(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})

		It("should return error when ID is empty", func() {
			req := &v1.UnregisterAgentRequest{Id: ""}
			_, err := agentService.UnregisterAgent(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})
	})

	Describe("GetInstructions", func() {
		var agentId string

		BeforeEach(func() {
			// Register an agent for testing
			registerReq := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
				IpAddress: "10.0.1.1",
				Version:   "1.0.0",
			}
			resp, err := agentService.RegisterAgent(ctx, registerReq)
			Expect(err).NotTo(HaveOccurred())
			agentId = resp.Agent.Id
		})

		It("should return instructions with poll interval", func() {
			req := &v1.GetInstructionsRequest{
				AgentId: agentId,
			}

			resp, err := agentService.GetInstructions(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			// Should request hardware collection for new agents without hardware
			Expect(resp.Instructions).To(HaveLen(1))
			Expect(resp.Instructions[0].Type).To(Equal(v1.InstructionType_INSTRUCTION_TYPE_COLLECT_HARDWARE))
			Expect(resp.PollIntervalSeconds).To(Equal(int32(60)))
			Expect(resp.ServerTime).NotTo(BeNil())
		})

		It("should update agent's last_seen timestamp", func() {
			// Get agent's initial last_seen
			getReq := &v1.GetAgentRequest{Id: agentId}
			getResp, err := agentService.GetAgent(ctx, getReq)
			Expect(err).NotTo(HaveOccurred())
			initialLastSeen := getResp.Agent.LastSeen

			// Call GetInstructions
			instructionsReq := &v1.GetInstructionsRequest{
				AgentId: agentId,
			}
			_, err = agentService.GetInstructions(ctx, instructionsReq)
			Expect(err).NotTo(HaveOccurred())

			// Verify last_seen was updated
			getResp2, err := agentService.GetAgent(ctx, getReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResp2.Agent.LastSeen.AsTime()).To(BeTemporally(">=", initialLastSeen.AsTime()))
			Expect(getResp2.Agent.Status).To(Equal(v1.AgentStatus_AGENT_STATUS_ACTIVE))
		})

		It("should return error for non-existent agent", func() {
			req := &v1.GetInstructionsRequest{
				AgentId: "non-existent",
			}

			_, err := agentService.GetInstructions(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})

		It("should return error when agent ID is empty", func() {
			req := &v1.GetInstructionsRequest{
				AgentId: "",
			}

			_, err := agentService.GetInstructions(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})

	})

	Describe("SubmitInstructionResult", func() {
		var agentId string

		BeforeEach(func() {
			// Register an agent first
			registerReq := &v1.RegisterAgentRequest{
				Id:        "agent-submit-test",
				ClusterId: testClusterId,
				Hostname:  "test-node",
				IpAddress: "10.0.1.1",
			}
			resp, err := agentService.RegisterAgent(ctx, registerReq)
			Expect(err).NotTo(HaveOccurred())
			agentId = resp.Agent.Id
		})

		It("should accept and process hardware collection result", func() {
			req := &v1.SubmitInstructionResultRequest{
				AgentId:       agentId,
				InstructionId: "instruction-123",
				Result: &v1.InstructionResult{
					InstructionType: v1.InstructionType_INSTRUCTION_TYPE_COLLECT_HARDWARE,
					Result: &v1.InstructionResult_HardwareCollection{
						HardwareCollection: &v1.HardwareCollectionResult{
							NetworkInterfaces: []*v1.MellanoxNIC{},
						},
					},
				},
			}

			resp, err := agentService.SubmitInstructionResult(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.Success).To(BeTrue())
		})

		It("should accept and process health check result", func() {
			req := &v1.SubmitInstructionResultRequest{
				AgentId:       agentId,
				InstructionId: "instruction-456",
				Result: &v1.InstructionResult{
					InstructionType: v1.InstructionType_INSTRUCTION_TYPE_HEALTH_CHECK,
					Result: &v1.InstructionResult_HealthCheck{
						HealthCheck: &v1.HealthCheckResult{
							Healthy: true,
						},
					},
				},
			}

			resp, err := agentService.SubmitInstructionResult(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.Success).To(BeTrue())
		})

		It("should return error when agent ID is missing", func() {
			req := &v1.SubmitInstructionResultRequest{
				InstructionId: "instruction-123",
				Result: &v1.InstructionResult{
					InstructionType: v1.InstructionType_INSTRUCTION_TYPE_HEALTH_CHECK,
				},
			}

			_, err := agentService.SubmitInstructionResult(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})

		It("should return error when instruction ID is missing", func() {
			req := &v1.SubmitInstructionResultRequest{
				AgentId: agentId,
				Result: &v1.InstructionResult{
					InstructionType: v1.InstructionType_INSTRUCTION_TYPE_HEALTH_CHECK,
				},
			}

			_, err := agentService.SubmitInstructionResult(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})

		It("should return error when result is missing", func() {
			req := &v1.SubmitInstructionResultRequest{
				AgentId:       agentId,
				InstructionId: "instruction-123",
			}

			_, err := agentService.SubmitInstructionResult(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})

		It("should return error for non-existent agent", func() {
			req := &v1.SubmitInstructionResultRequest{
				AgentId:       "non-existent-agent",
				InstructionId: "instruction-123",
				Result: &v1.InstructionResult{
					InstructionType: v1.InstructionType_INSTRUCTION_TYPE_HEALTH_CHECK,
				},
			}

			_, err := agentService.SubmitInstructionResult(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})
	})
})
