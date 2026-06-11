package mocks

import (
	"context"

	"github.com/vcs-sms/monitor-service/internal/repository"
)

// ServerReaderMock is a test mock for ServerReader.
type ServerReaderMock struct {
	GetAllActiveServersFunc func(ctx context.Context) ([]repository.ServerInfo, error)
	BatchUpdateStatusFunc   func(ctx context.Context, changes []repository.StatusChangeEvent) error
}

func (m *ServerReaderMock) GetAllActiveServers(ctx context.Context) ([]repository.ServerInfo, error) {
	if m.GetAllActiveServersFunc != nil {
		return m.GetAllActiveServersFunc(ctx)
	}
	return nil, nil
}

func (m *ServerReaderMock) BatchUpdateStatus(ctx context.Context, changes []repository.StatusChangeEvent) error {
	if m.BatchUpdateStatusFunc != nil {
		return m.BatchUpdateStatusFunc(ctx, changes)
	}
	return nil
}
