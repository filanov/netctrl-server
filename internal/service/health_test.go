package service_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/filanov/netctrl-server/internal/service"
	v1 "github.com/filanov/netctrl-server/pkg/api/v1"
)

var _ = Describe("HealthService", func() {
	var (
		healthService *service.HealthService
		ctx           context.Context
	)

	BeforeEach(func() {
		healthService = service.NewHealthService()
		ctx = context.Background()
	})

	Describe("Check", func() {
		It("should return healthy status", func() {
			req := &v1.HealthCheckRequest{}
			resp, err := healthService.Check(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Status).To(Equal(v1.HealthStatus_HEALTH_STATUS_HEALTHY))
			Expect(resp.Message).To(Equal("Service is healthy"))
		})
	})

	Describe("Ready", func() {
		It("should return ready status", func() {
			req := &v1.ReadinessCheckRequest{}
			resp, err := healthService.Ready(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Status).To(Equal(v1.ReadinessStatus_READINESS_STATUS_READY))
			Expect(resp.Message).To(Equal("Service is ready"))
		})
	})
})
