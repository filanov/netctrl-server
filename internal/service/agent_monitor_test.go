package service_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/filanov/netctrl-server/internal/service"
	"github.com/filanov/netctrl-server/internal/storage/memory"
	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

var _ = Describe("AgentMonitor", func() {
	var (
		monitor        *service.AgentMonitor
		agentService   *service.AgentService
		clusterService *service.ClusterService
		storage        *memory.Storage
		ctx            context.Context
		testClusterId  string
	)

	BeforeEach(func() {
		storage = memory.New()
		monitor = service.NewAgentMonitor(storage)
		agentService = service.NewAgentService(storage)
		clusterService = service.NewClusterService(storage)
		ctx = context.Background()

		// Create a test cluster
		createReq := &v1.CreateClusterRequest{
			Name:        "test-cluster",
			Description: "Test cluster for monitor tests",
		}
		createResp, err := clusterService.CreateCluster(ctx, createReq)
		Expect(err).NotTo(HaveOccurred())
		testClusterId = createResp.Cluster.Id
	})

	Describe("Agent State Management", func() {
		It("should mark agent as inactive after 3 poll intervals", func() {
			// Register an agent
			registerReq := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
			}
			resp, err := agentService.RegisterAgent(ctx, registerReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Agent.Status).To(Equal(v1.AgentStatus_AGENT_STATUS_ACTIVE))

			// Manually set last_seen to 181 seconds ago (just over 3 intervals)
			agent, err := storage.GetAgent(ctx, "agent-1")
			Expect(err).NotTo(HaveOccurred())

			oldTime := time.Now().Add(-181 * time.Second)
			agent.LastSeen = timestamppb.New(oldTime)
			err = storage.UpdateAgent(ctx, agent)
			Expect(err).NotTo(HaveOccurred())

			// Run monitor check
			monitor.CheckAgentStatesOnce(ctx)

			// Verify agent is now inactive
			updatedAgent, err := storage.GetAgent(ctx, "agent-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedAgent.Status).To(Equal(v1.AgentStatus_AGENT_STATUS_INACTIVE))
		})

		It("should not mark agent as inactive within 3 poll intervals", func() {
			// Register an agent
			registerReq := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
			}
			_, err := agentService.RegisterAgent(ctx, registerReq)
			Expect(err).NotTo(HaveOccurred())

			// Set last_seen to 179 seconds ago (just under 3 intervals)
			agent, err := storage.GetAgent(ctx, "agent-1")
			Expect(err).NotTo(HaveOccurred())

			oldTime := time.Now().Add(-179 * time.Second)
			agent.LastSeen = timestamppb.New(oldTime)
			err = storage.UpdateAgent(ctx, agent)
			Expect(err).NotTo(HaveOccurred())

			// Run monitor check
			monitor.CheckAgentStatesOnce(ctx)

			// Verify agent is still active
			updatedAgent, err := storage.GetAgent(ctx, "agent-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedAgent.Status).To(Equal(v1.AgentStatus_AGENT_STATUS_ACTIVE))
		})

		It("should not change status of already inactive agents", func() {
			// Register an agent
			registerReq := &v1.RegisterAgentRequest{
				Id:        "agent-1",
				ClusterId: testClusterId,
				Hostname:  "node1",
			}
			_, err := agentService.RegisterAgent(ctx, registerReq)
			Expect(err).NotTo(HaveOccurred())

			// Manually set to inactive with old last_seen
			agent, err := storage.GetAgent(ctx, "agent-1")
			Expect(err).NotTo(HaveOccurred())

			oldTime := time.Now().Add(-300 * time.Second)
			agent.LastSeen = timestamppb.New(oldTime)
			agent.Status = v1.AgentStatus_AGENT_STATUS_INACTIVE
			err = storage.UpdateAgent(ctx, agent)
			Expect(err).NotTo(HaveOccurred())

			// Run monitor check
			monitor.CheckAgentStatesOnce(ctx)

			// Verify agent remains inactive
			updatedAgent, err := storage.GetAgent(ctx, "agent-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedAgent.Status).To(Equal(v1.AgentStatus_AGENT_STATUS_INACTIVE))
		})

		It("should handle multiple agents with different states", func() {
			// Register multiple agents
			for i := 1; i <= 3; i++ {
				registerReq := &v1.RegisterAgentRequest{
					Id:        fmt.Sprintf("agent-%d", i),
					ClusterId: testClusterId,
					Hostname:  fmt.Sprintf("node%d", i),
				}
				_, err := agentService.RegisterAgent(ctx, registerReq)
				Expect(err).NotTo(HaveOccurred())
			}

			// Set different last_seen times
			// agent-1: 200s ago (should be inactive)
			agent1, _ := storage.GetAgent(ctx, "agent-1")
			agent1.LastSeen = timestamppb.New(time.Now().Add(-200 * time.Second))
			storage.UpdateAgent(ctx, agent1)

			// agent-2: 100s ago (should stay active)
			agent2, _ := storage.GetAgent(ctx, "agent-2")
			agent2.LastSeen = timestamppb.New(time.Now().Add(-100 * time.Second))
			storage.UpdateAgent(ctx, agent2)

			// agent-3: 10s ago (should stay active)
			agent3, _ := storage.GetAgent(ctx, "agent-3")
			agent3.LastSeen = timestamppb.New(time.Now().Add(-10 * time.Second))
			storage.UpdateAgent(ctx, agent3)

			// Run monitor check
			monitor.CheckAgentStatesOnce(ctx)

			// Verify statuses
			updated1, _ := storage.GetAgent(ctx, "agent-1")
			Expect(updated1.Status).To(Equal(v1.AgentStatus_AGENT_STATUS_INACTIVE))

			updated2, _ := storage.GetAgent(ctx, "agent-2")
			Expect(updated2.Status).To(Equal(v1.AgentStatus_AGENT_STATUS_ACTIVE))

			updated3, _ := storage.GetAgent(ctx, "agent-3")
			Expect(updated3.Status).To(Equal(v1.AgentStatus_AGENT_STATUS_ACTIVE))
		})
	})
})
