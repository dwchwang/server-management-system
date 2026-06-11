package mocks

import (
	"context"

	"github.com/vcs-sms/report-service/internal/model"
)

// ReportJobRepoMock is a mock implementation of repository.ReportJobRepo.
type ReportJobRepoMock struct {
	CreateFunc          func(ctx context.Context, job *model.ReportJob) error
	UpdateFunc          func(ctx context.Context, job *model.ReportJob) error
	FindByIDFunc        func(ctx context.Context, id string) (*model.ReportJob, error)
	FindByDateRangeFunc func(ctx context.Context, startDate, endDate string) ([]model.ReportJob, error)
}

func (m *ReportJobRepoMock) Create(ctx context.Context, job *model.ReportJob) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, job)
	}
	return nil
}

func (m *ReportJobRepoMock) Update(ctx context.Context, job *model.ReportJob) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, job)
	}
	return nil
}

func (m *ReportJobRepoMock) FindByID(ctx context.Context, id string) (*model.ReportJob, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *ReportJobRepoMock) FindByDateRange(ctx context.Context, startDate, endDate string) ([]model.ReportJob, error) {
	if m.FindByDateRangeFunc != nil {
		return m.FindByDateRangeFunc(ctx, startDate, endDate)
	}
	return nil, nil
}
