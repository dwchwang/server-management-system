package mocks

import (
	"context"

	"github.com/vcs-sms/monitor-service/internal/checker"
)

// ESStatusLogRepoMock is a test mock for ESStatusLogRepo.
type ESStatusLogRepoMock struct {
	BulkIndexFunc func(ctx context.Context, results []*checker.HealthResult) error
}

func (m *ESStatusLogRepoMock) BulkIndex(ctx context.Context, results []*checker.HealthResult) error {
	if m.BulkIndexFunc != nil {
		return m.BulkIndexFunc(ctx, results)
	}
	return nil
}
