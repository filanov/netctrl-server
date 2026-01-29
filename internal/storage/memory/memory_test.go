package memory_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/filanov/netctrl-server/internal/storage/memory"
	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

var _ = Describe("Memory Storage", func() {
	var (
		storage *memory.Storage
		ctx     context.Context
	)

	BeforeEach(func() {
		storage = memory.New()
		ctx = context.Background()
	})

	Describe("Cluster Operations", func() {
		var testCluster *v1.Cluster

		BeforeEach(func() {
			testCluster = &v1.Cluster{
				Id:          "test-cluster-id",
				Name:        "Test Cluster",
				Description: "Test Description",
				CreatedAt:   timestamppb.Now(),
				UpdatedAt:   timestamppb.Now(),
			}
		})

		Describe("CreateCluster", func() {
			It("should create a new cluster successfully", func() {
				err := storage.CreateCluster(ctx, testCluster)
				Expect(err).NotTo(HaveOccurred())

				cluster, err := storage.GetCluster(ctx, testCluster.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(cluster.Id).To(Equal(testCluster.Id))
				Expect(cluster.Name).To(Equal(testCluster.Name))
			})

			It("should return error when creating duplicate cluster", func() {
				err := storage.CreateCluster(ctx, testCluster)
				Expect(err).NotTo(HaveOccurred())

				err = storage.CreateCluster(ctx, testCluster)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("already exists"))
			})
		})

		Describe("GetCluster", func() {
			It("should retrieve existing cluster", func() {
				err := storage.CreateCluster(ctx, testCluster)
				Expect(err).NotTo(HaveOccurred())

				cluster, err := storage.GetCluster(ctx, testCluster.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(cluster.Id).To(Equal(testCluster.Id))
			})

			It("should return error for non-existent cluster", func() {
				_, err := storage.GetCluster(ctx, "non-existent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Describe("ListClusters", func() {
			It("should return empty list when no clusters exist", func() {
				clusters, err := storage.ListClusters(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(BeEmpty())
			})

			It("should return all clusters", func() {
				cluster1 := testCluster
				cluster2 := &v1.Cluster{
					Id:          "test-cluster-id-2",
					Name:        "Test Cluster 2",
					Description: "Test Description 2",
					CreatedAt:   timestamppb.Now(),
					UpdatedAt:   timestamppb.Now(),
				}

				err := storage.CreateCluster(ctx, cluster1)
				Expect(err).NotTo(HaveOccurred())
				err = storage.CreateCluster(ctx, cluster2)
				Expect(err).NotTo(HaveOccurred())

				clusters, err := storage.ListClusters(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(2))
			})
		})

		Describe("UpdateCluster", func() {
			It("should update existing cluster", func() {
				err := storage.CreateCluster(ctx, testCluster)
				Expect(err).NotTo(HaveOccurred())

				testCluster.Name = "Updated Name"
				err = storage.UpdateCluster(ctx, testCluster)
				Expect(err).NotTo(HaveOccurred())

				cluster, err := storage.GetCluster(ctx, testCluster.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(cluster.Name).To(Equal("Updated Name"))
			})

			It("should return error when updating non-existent cluster", func() {
				err := storage.UpdateCluster(ctx, testCluster)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Describe("DeleteCluster", func() {
			It("should delete existing cluster", func() {
				err := storage.CreateCluster(ctx, testCluster)
				Expect(err).NotTo(HaveOccurred())

				err = storage.DeleteCluster(ctx, testCluster.Id)
				Expect(err).NotTo(HaveOccurred())

				_, err = storage.GetCluster(ctx, testCluster.Id)
				Expect(err).To(HaveOccurred())
			})

			It("should return error when deleting non-existent cluster", func() {
				err := storage.DeleteCluster(ctx, "non-existent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})

			It("should cascade delete associated agents", func() {
				err := storage.CreateCluster(ctx, testCluster)
				Expect(err).NotTo(HaveOccurred())

				agent1 := &v1.Agent{
					Id:        "agent-1",
					ClusterId: testCluster.Id,
					Hostname:  "node1",
					Status:    v1.AgentStatus_AGENT_STATUS_ACTIVE,
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				}
				agent2 := &v1.Agent{
					Id:        "agent-2",
					ClusterId: testCluster.Id,
					Hostname:  "node2",
					Status:    v1.AgentStatus_AGENT_STATUS_ACTIVE,
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				}

				err = storage.CreateAgent(ctx, agent1)
				Expect(err).NotTo(HaveOccurred())
				err = storage.CreateAgent(ctx, agent2)
				Expect(err).NotTo(HaveOccurred())

				err = storage.DeleteCluster(ctx, testCluster.Id)
				Expect(err).NotTo(HaveOccurred())

				agents, err := storage.ListAgents(ctx, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(agents).To(BeEmpty())
			})
		})

		Describe("ClusterExists", func() {
			It("should return true for existing cluster", func() {
				err := storage.CreateCluster(ctx, testCluster)
				Expect(err).NotTo(HaveOccurred())

				exists, err := storage.ClusterExists(ctx, testCluster.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})

			It("should return false for non-existent cluster", func() {
				exists, err := storage.ClusterExists(ctx, "non-existent")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
		})
	})

	Describe("Agent Operations", func() {
		var (
			testCluster *v1.Cluster
			testAgent   *v1.Agent
		)

		BeforeEach(func() {
			testCluster = &v1.Cluster{
				Id:          "test-cluster-id",
				Name:        "Test Cluster",
				Description: "Test Description",
				CreatedAt:   timestamppb.Now(),
				UpdatedAt:   timestamppb.Now(),
			}
			err := storage.CreateCluster(ctx, testCluster)
			Expect(err).NotTo(HaveOccurred())

			testAgent = &v1.Agent{
				Id:        "test-agent-id",
				ClusterId: testCluster.Id,
				Hostname:  "test-host",
				IpAddress: "10.0.1.1",
				Version:   "1.0.0",
				Status:    v1.AgentStatus_AGENT_STATUS_ACTIVE,
				LastSeen:  timestamppb.Now(),
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			}
		})

		Describe("CreateAgent", func() {
			It("should create a new agent successfully", func() {
				err := storage.CreateAgent(ctx, testAgent)
				Expect(err).NotTo(HaveOccurred())

				agent, err := storage.GetAgent(ctx, testAgent.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(agent.Id).To(Equal(testAgent.Id))
				Expect(agent.ClusterId).To(Equal(testAgent.ClusterId))
			})

			It("should return error when creating agent without ID", func() {
				testAgent.Id = ""
				err := storage.CreateAgent(ctx, testAgent)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("required"))
			})

			It("should return error when creating duplicate agent", func() {
				err := storage.CreateAgent(ctx, testAgent)
				Expect(err).NotTo(HaveOccurred())

				err = storage.CreateAgent(ctx, testAgent)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("already exists"))
			})
		})

		Describe("GetAgent", func() {
			It("should retrieve existing agent", func() {
				err := storage.CreateAgent(ctx, testAgent)
				Expect(err).NotTo(HaveOccurred())

				agent, err := storage.GetAgent(ctx, testAgent.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(agent.Id).To(Equal(testAgent.Id))
			})

			It("should return error for non-existent agent", func() {
				_, err := storage.GetAgent(ctx, "non-existent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Describe("ListAgents", func() {
			It("should return empty list when no agents exist", func() {
				agents, err := storage.ListAgents(ctx, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(agents).To(BeEmpty())
			})

			It("should return all agents when cluster ID not specified", func() {
				agent1 := testAgent
				agent2 := &v1.Agent{
					Id:        "agent-2",
					ClusterId: testCluster.Id,
					Hostname:  "host-2",
					Status:    v1.AgentStatus_AGENT_STATUS_ACTIVE,
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				}

				err := storage.CreateAgent(ctx, agent1)
				Expect(err).NotTo(HaveOccurred())
				err = storage.CreateAgent(ctx, agent2)
				Expect(err).NotTo(HaveOccurred())

				agents, err := storage.ListAgents(ctx, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(agents).To(HaveLen(2))
			})

			It("should filter agents by cluster ID", func() {
				cluster2 := &v1.Cluster{
					Id:          "cluster-2",
					Name:        "Cluster 2",
					Description: "Second cluster",
					CreatedAt:   timestamppb.Now(),
					UpdatedAt:   timestamppb.Now(),
				}
				err := storage.CreateCluster(ctx, cluster2)
				Expect(err).NotTo(HaveOccurred())

				agent1 := testAgent
				agent2 := &v1.Agent{
					Id:        "agent-2",
					ClusterId: cluster2.Id,
					Hostname:  "host-2",
					Status:    v1.AgentStatus_AGENT_STATUS_ACTIVE,
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				}

				err = storage.CreateAgent(ctx, agent1)
				Expect(err).NotTo(HaveOccurred())
				err = storage.CreateAgent(ctx, agent2)
				Expect(err).NotTo(HaveOccurred())

				agents, err := storage.ListAgents(ctx, testCluster.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(agents).To(HaveLen(1))
				Expect(agents[0].ClusterId).To(Equal(testCluster.Id))
			})
		})

		Describe("UpdateAgent", func() {
			It("should update existing agent", func() {
				err := storage.CreateAgent(ctx, testAgent)
				Expect(err).NotTo(HaveOccurred())

				testAgent.Hostname = "updated-host"
				err = storage.UpdateAgent(ctx, testAgent)
				Expect(err).NotTo(HaveOccurred())

				agent, err := storage.GetAgent(ctx, testAgent.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(agent.Hostname).To(Equal("updated-host"))
			})

			It("should return error when updating non-existent agent", func() {
				err := storage.UpdateAgent(ctx, testAgent)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Describe("DeleteAgent", func() {
			It("should delete existing agent", func() {
				err := storage.CreateAgent(ctx, testAgent)
				Expect(err).NotTo(HaveOccurred())

				err = storage.DeleteAgent(ctx, testAgent.Id)
				Expect(err).NotTo(HaveOccurred())

				_, err = storage.GetAgent(ctx, testAgent.Id)
				Expect(err).To(HaveOccurred())
			})

			It("should return error when deleting non-existent agent", func() {
				err := storage.DeleteAgent(ctx, "non-existent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})
	})
})
