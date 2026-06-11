package mocks

import "context"

// ServerCounterMock is a mock implementation of repository.ServerCounter.
type ServerCounterMock struct {
	CountActiveServersFunc func(ctx context.Context) (int, error)
}

func (m *ServerCounterMock) CountActiveServers(ctx context.Context) (int, error) {
	if m.CountActiveServersFunc != nil {
		return m.CountActiveServersFunc(ctx)
	}
	return 0, nil
}
