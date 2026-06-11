package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/vcs-sms/report-service/internal/service"
)

// DailyReportCron runs the daily report generation on a cron schedule.
type DailyReportCron struct {
	service  service.ReportService
	schedule string // cron expression, e.g. "0 8 * * *" (8:00 AM daily)
	logger   zerolog.Logger
}

// NewDailyReportCron creates a new DailyReportCron.
func NewDailyReportCron(svc service.ReportService, schedule string, logger zerolog.Logger) *DailyReportCron {
	if schedule == "" {
		schedule = "0 8 * * *" // Default: 8:00 AM every day
	}
	return &DailyReportCron{
		service:  svc,
		schedule: schedule,
		logger:   logger,
	}
}

// Start begins the cron scheduler. Blocks until ctx is cancelled.
func (c *DailyReportCron) Start(ctx context.Context) {
	cronRunner := cron.New()

	_, err := cronRunner.AddFunc(c.schedule, func() {
		c.logger.Info().Str("schedule", c.schedule).Msg("Daily report cron triggered")

		if err := c.service.SendDailyReport(ctx); err != nil {
			c.logger.Error().Err(err).Msg("Daily report cron failed")
		}
	})

	if err != nil {
		c.logger.Error().Err(err).Str("schedule", c.schedule).Msg("Failed to register cron job")
		return
	}

	cronRunner.Start()
	c.logger.Info().Str("schedule", c.schedule).Msg("Daily report cron scheduler started")

	<-ctx.Done()

	c.logger.Info().Msg("Daily report cron scheduler stopping")
	cronRunner.Stop()
}
