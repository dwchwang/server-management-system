package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/vcs-sms/report-service/internal/dto"
)

type mockReportService struct {
	dailyCalled chan struct{}
}

func (m *mockReportService) GetSummary(ctx context.Context, startDate, endDate time.Time) (*dto.ReportSummaryResponse, error) {
	return nil, nil
}

func (m *mockReportService) SendReport(ctx context.Context, req *dto.SendReportRequest) (*dto.SendReportResponse, error) {
	return nil, nil
}

func (m *mockReportService) SendDailyReport(ctx context.Context) error {
	if m.dailyCalled != nil {
		select {
		case m.dailyCalled <- struct{}{}:
		default:
		}
	}
	return nil
}

func TestDailyReportCron_DefaultFiveFieldScheduleIsValid(t *testing.T) {
	dailyCron := NewDailyReportCron(nil, "", zerolog.Nop())
	cronRunner := cron.New()

	_, err := cronRunner.AddFunc(dailyCron.schedule, func() {})

	assert.NoError(t, err)
	assert.Equal(t, "0 8 * * *", dailyCron.schedule)
}

func TestDailyReportCron_StartReturnsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dailyCron := NewDailyReportCron(&mockReportService{}, "0 8 * * *", zerolog.Nop())
	done := make(chan struct{})

	go func() {
		dailyCron.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after context cancellation")
	}
}

func TestDailyReportCron_StartReturnsOnInvalidSchedule(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dailyCron := NewDailyReportCron(&mockReportService{}, "not-a-schedule", zerolog.Nop())
	done := make(chan struct{})

	go func() {
		dailyCron.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not return on invalid schedule")
	}
}

func TestDailyReportCron_TriggersDailyReport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan struct{}, 1)
	dailyCron := NewDailyReportCron(&mockReportService{dailyCalled: called}, "@every 1s", zerolog.Nop())
	done := make(chan struct{})

	go func() {
		dailyCron.Start(ctx)
		close(done)
	}()

	select {
	case <-called:
		cancel()
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("daily report was not triggered")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after cancel")
	}
}
