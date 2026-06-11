package mocks

import (
	"context"
	"time"

	"github.com/vcs-sms/report-service/internal/dto"
)

// UptimeCalculatorMock is a mock implementation of repository.UptimeCalculator.
type UptimeCalculatorMock struct {
	GetUptimeSummaryFunc    func(ctx context.Context, startDate, endDate time.Time) (*dto.ReportSummaryResponse, error)
	GetLowUptimeServersFunc func(ctx context.Context, startDate, endDate time.Time, topN int) ([]dto.ServerUptime, error)
}

func (m *UptimeCalculatorMock) GetUptimeSummary(ctx context.Context, startDate, endDate time.Time) (*dto.ReportSummaryResponse, error) {
	if m.GetUptimeSummaryFunc != nil {
		return m.GetUptimeSummaryFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

func (m *UptimeCalculatorMock) GetLowUptimeServers(ctx context.Context, startDate, endDate time.Time, topN int) ([]dto.ServerUptime, error) {
	if m.GetLowUptimeServersFunc != nil {
		return m.GetLowUptimeServersFunc(ctx, startDate, endDate, topN)
	}
	return nil, nil
}
