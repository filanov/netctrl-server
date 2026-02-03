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

var _ = Describe("ClusterService", func() {
	var (
		clusterService *service.ClusterService
		ctx            context.Context
	)

	BeforeEach(func() {
		storage := mock.New()
		clusterService = service.NewClusterService(storage)
		ctx = context.Background()
	})

	Describe("CreateCluster", func() {
		It("should create a cluster with valid name", func() {
			req := &v1.CreateClusterRequest{
				Name:        "test-cluster",
				Description: "Test Description",
			}

			resp, err := clusterService.CreateCluster(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Cluster).NotTo(BeNil())
			Expect(resp.Cluster.Name).To(Equal("test-cluster"))
			Expect(resp.Cluster.Id).NotTo(BeEmpty())
			Expect(resp.Cluster.CreatedAt).NotTo(BeNil())
			Expect(resp.Cluster.UpdatedAt).NotTo(BeNil())
		})

		It("should return error when name is missing", func() {
			req := &v1.CreateClusterRequest{
				Description: "Test Description",
			}

			_, err := clusterService.CreateCluster(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
			Expect(st.Message()).To(ContainSubstring("name is required"))
		})

		It("should create cluster with only name (description optional)", func() {
			req := &v1.CreateClusterRequest{
				Name: "minimal-cluster",
			}

			resp, err := clusterService.CreateCluster(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Cluster).NotTo(BeNil())
			Expect(resp.Cluster.Name).To(Equal("minimal-cluster"))
		})

		It("should return error when name is too long", func() {
			longName := string(make([]byte, 256))
			for i := range longName {
				longName = longName[:i] + "a" + longName[i+1:]
			}
			req := &v1.CreateClusterRequest{
				Name: longName,
			}

			_, err := clusterService.CreateCluster(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
			Expect(st.Message()).To(ContainSubstring("less than 255 characters"))
		})
	})

	Describe("GetCluster", func() {
		It("should retrieve existing cluster", func() {
			createReq := &v1.CreateClusterRequest{
				Name:        "test-cluster",
				Description: "Test Description",
			}
			createResp, err := clusterService.CreateCluster(ctx, createReq)
			Expect(err).NotTo(HaveOccurred())

			getReq := &v1.GetClusterRequest{
				Id: createResp.Cluster.Id,
			}
			getResp, err := clusterService.GetCluster(ctx, getReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResp.Cluster.Id).To(Equal(createResp.Cluster.Id))
			Expect(getResp.Cluster.Name).To(Equal("test-cluster"))
		})

		It("should return error for non-existent cluster", func() {
			req := &v1.GetClusterRequest{
				Id: "non-existent-id",
			}

			_, err := clusterService.GetCluster(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})

		It("should return error when ID is empty", func() {
			req := &v1.GetClusterRequest{
				Id: "",
			}

			_, err := clusterService.GetCluster(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})
	})

	Describe("ListClusters", func() {
		It("should return empty list when no clusters exist", func() {
			req := &v1.ListClustersRequest{}
			resp, err := clusterService.ListClusters(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Clusters).To(BeEmpty())
		})

		It("should return all clusters", func() {
			createReq1 := &v1.CreateClusterRequest{Name: "cluster-1"}
			createReq2 := &v1.CreateClusterRequest{Name: "cluster-2"}

			_, err := clusterService.CreateCluster(ctx, createReq1)
			Expect(err).NotTo(HaveOccurred())
			_, err = clusterService.CreateCluster(ctx, createReq2)
			Expect(err).NotTo(HaveOccurred())

			listReq := &v1.ListClustersRequest{}
			listResp, err := clusterService.ListClusters(ctx, listReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(listResp.Clusters).To(HaveLen(2))
		})
	})

	Describe("UpdateCluster", func() {
		It("should update cluster name and description", func() {
			createReq := &v1.CreateClusterRequest{
				Name:        "original-name",
				Description: "original-description",
			}
			createResp, err := clusterService.CreateCluster(ctx, createReq)
			Expect(err).NotTo(HaveOccurred())

			updateReq := &v1.UpdateClusterRequest{
				Id:          createResp.Cluster.Id,
				Name:        "updated-name",
				Description: "updated-description",
			}
			updateResp, err := clusterService.UpdateCluster(ctx, updateReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(updateResp.Cluster.Name).To(Equal("updated-name"))
			Expect(updateResp.Cluster.Description).To(Equal("updated-description"))
			Expect(updateResp.Cluster.UpdatedAt.AsTime()).To(BeTemporally(">=", createResp.Cluster.UpdatedAt.AsTime()))
		})

		It("should return error for non-existent cluster", func() {
			req := &v1.UpdateClusterRequest{
				Id:   "non-existent-id",
				Name: "new-name",
			}

			_, err := clusterService.UpdateCluster(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})

		It("should return error when ID is empty", func() {
			req := &v1.UpdateClusterRequest{
				Id:   "",
				Name: "new-name",
			}

			_, err := clusterService.UpdateCluster(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})
	})

	Describe("DeleteCluster", func() {
		It("should delete existing cluster", func() {
			createReq := &v1.CreateClusterRequest{Name: "test-cluster"}
			createResp, err := clusterService.CreateCluster(ctx, createReq)
			Expect(err).NotTo(HaveOccurred())

			deleteReq := &v1.DeleteClusterRequest{Id: createResp.Cluster.Id}
			deleteResp, err := clusterService.DeleteCluster(ctx, deleteReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(deleteResp.Success).To(BeTrue())

			getReq := &v1.GetClusterRequest{Id: createResp.Cluster.Id}
			_, err = clusterService.GetCluster(ctx, getReq)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent cluster", func() {
			req := &v1.DeleteClusterRequest{Id: "non-existent-id"}

			_, err := clusterService.DeleteCluster(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.NotFound))
		})

		It("should return error when ID is empty", func() {
			req := &v1.DeleteClusterRequest{Id: ""}

			_, err := clusterService.DeleteCluster(ctx, req)
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(st.Code()).To(Equal(codes.InvalidArgument))
		})
	})
})
