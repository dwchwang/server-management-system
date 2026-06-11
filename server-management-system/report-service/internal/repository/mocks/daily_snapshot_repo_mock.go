package mocks

import (
	"context"
	"time"

	"github.com/vcs-sms/report-service/internal/model"
)

// DailySnapshotRepoMock is a mock implementation of repository.DailySnapshotRepo.
type DailySnapshotRepoMock struct {
	CreateFunc          func(ctx context.Context, snapshot *model.DailySnapshot) error
	FindByDateFunc      func(ctx context.Context, date time.Time) (*model.DailySnapshot, error)
	FindByDateRangeFunc func(ctx context.Context, startDate, endDate time.Time) ([]model.DailySnapshot, error)
}

func (m *DailySnapshotRepoMock) Create(ctx context.Context, snapshot *model.DailySnapshot) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, snapshot)
	}
	return nil
}

func (m *DailySnapshotRepoMock) FindByDate(ctx context.Context, date time.Time) (*model.DailySnapshot, error) {
	if m.FindByDateFunc != nil {
		return m.FindByDateFunc(ctx, date)
	}
	return nil, nil
}

func (m *DailySnapshotRepoMock) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]model.DailySnapshot, error) {
	if m.FindByDateRangeFunc != nil {
		return m.FindByDateRangeFunc(ctx, startDate, endDate)
	}
	return nil, nil
}
