package mocks

import (
	"context"

	"github.com/vcs-sms/monitor-service/internal/checker"
)

// HealthCheckerMock is a test mock for HealthChecker.
type HealthCheckerMock struct {
	CheckFunc func(ctx context.Context, server *checker.ServerInfo) *checker.HealthResult
	NameFunc  func() string
}

func (m *HealthCheckerMock) Check(ctx context.Context, server *checker.ServerInfo) *checker.HealthResult {
	if m.CheckFunc != nil {
		return m.CheckFunc(ctx, server)
	}
	return &checker.HealthResult{Status: "on", CheckMethod: "mock"}
}

func (m *HealthCheckerMock) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock"
}
