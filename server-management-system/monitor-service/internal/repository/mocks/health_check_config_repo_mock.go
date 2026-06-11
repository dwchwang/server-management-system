package mocks

import (
	"context"

	"github.com/vcs-sms/monitor-service/internal/model"
)

// HealthCheckConfigRepoMock is a test mock for HealthCheckConfigRepo.
type HealthCheckConfigRepoMock struct {
	CreateFunc            func(ctx context.Context, config *model.HealthCheckConfig) error
	GetByServerIDFunc     func(ctx context.Context, serverID string) (*model.HealthCheckConfig, error)
	GetAllEnabledFunc     func(ctx context.Context) ([]model.HealthCheckConfig, error)
	UpdateFunc            func(ctx context.Context, config *model.HealthCheckConfig) error
	DisableByServerIDFunc func(ctx context.Context, serverID string) error
}

func (m *HealthCheckConfigRepoMock) Create(ctx context.Context, config *model.HealthCheckConfig) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, config)
	}
	return nil
}

func (m *HealthCheckConfigRepoMock) GetByServerID(ctx context.Context, serverID string) (*model.HealthCheckConfig, error) {
	if m.GetByServerIDFunc != nil {
		return m.GetByServerIDFunc(ctx, serverID)
	}
	return nil, nil
}

func (m *HealthCheckConfigRepoMock) GetAllEnabled(ctx context.Context) ([]model.HealthCheckConfig, error) {
	if m.GetAllEnabledFunc != nil {
		return m.GetAllEnabledFunc(ctx)
	}
	return nil, nil
}

func (m *HealthCheckConfigRepoMock) Update(ctx context.Context, config *model.HealthCheckConfig) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, config)
	}
	return nil
}

func (m *HealthCheckConfigRepoMock) DisableByServerID(ctx context.Context, serverID string) error {
	if m.DisableByServerIDFunc != nil {
		return m.DisableByServerIDFunc(ctx, serverID)
	}
	return nil
}
