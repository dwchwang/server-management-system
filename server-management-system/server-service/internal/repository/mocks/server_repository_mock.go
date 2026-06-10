package mocks

import (
	"context"

	"github.com/vcs-sms/server-service/internal/dto"
	"github.com/vcs-sms/server-service/internal/model"
)

// ServerRepositoryMock is a test mock for ServerRepository.
type ServerRepositoryMock struct {
	CreateFunc                    func(ctx context.Context, server *model.Server) error
	FindByServerIDFunc            func(ctx context.Context, serverID string) (*model.Server, error)
	FindAllFunc                   func(ctx context.Context, filter *dto.ServerFilter) ([]model.Server, int64, error)
	UpdateFunc                    func(ctx context.Context, server *model.Server) error
	DeleteFunc                    func(ctx context.Context, serverID string) error
	ExistsByServerIDFunc          func(ctx context.Context, serverID string) (bool, error)
	ExistsByServerNameFunc        func(ctx context.Context, serverName string) (bool, error)
	ExistsByServerNameExcludeFunc func(ctx context.Context, serverName string, excludeID string) (bool, error)
}

func (m *ServerRepositoryMock) Create(ctx context.Context, server *model.Server) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, server)
	}
	return nil
}

func (m *ServerRepositoryMock) FindByServerID(ctx context.Context, serverID string) (*model.Server, error) {
	if m.FindByServerIDFunc != nil {
		return m.FindByServerIDFunc(ctx, serverID)
	}
	return nil, nil
}

func (m *ServerRepositoryMock) FindAll(ctx context.Context, filter *dto.ServerFilter) ([]model.Server, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *ServerRepositoryMock) Update(ctx context.Context, server *model.Server) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, server)
	}
	return nil
}

func (m *ServerRepositoryMock) Delete(ctx context.Context, serverID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, serverID)
	}
	return nil
}

func (m *ServerRepositoryMock) ExistsByServerID(ctx context.Context, serverID string) (bool, error) {
	if m.ExistsByServerIDFunc != nil {
		return m.ExistsByServerIDFunc(ctx, serverID)
	}
	return false, nil
}

func (m *ServerRepositoryMock) ExistsByServerName(ctx context.Context, serverName string) (bool, error) {
	if m.ExistsByServerNameFunc != nil {
		return m.ExistsByServerNameFunc(ctx, serverName)
	}
	return false, nil
}

func (m *ServerRepositoryMock) ExistsByServerNameExclude(ctx context.Context, serverName string, excludeID string) (bool, error) {
	if m.ExistsByServerNameExcludeFunc != nil {
		return m.ExistsByServerNameExcludeFunc(ctx, serverName, excludeID)
	}
	return false, nil
}
