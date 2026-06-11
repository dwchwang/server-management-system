package email

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vcs-sms/report-service/internal/dto"
)

func TestRenderDailyReport_Success(t *testing.T) {
	data := &ReportData{
		ReportDate:   "01/06/2026",
		TotalServers: 100,
		ServersOn:    95,
		ServersOff:   5,
		AvgUptimePct: 95.5,
		LowUptimeServers: []dto.ServerUptime{
			{ServerID: "srv-1", ServerName: "Server 1", UptimePct: 50.0},
			{ServerID: "srv-2", ServerName: "Server 2", UptimePct: 75.0},
		},
	}

	html, err := RenderDailyReport(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "01/06/2026")
	assert.Contains(t, html, "100")
	assert.Contains(t, html, "95")
	assert.Contains(t, html, "5")
	assert.Contains(t, html, "95.50%")
	assert.Contains(t, html, "srv-1")
	assert.Contains(t, html, "Server 1")
	assert.Contains(t, html, "50.00%")
	assert.Contains(t, html, "srv-2")
	assert.Contains(t, html, "Server 2")
}

func TestRenderDailyReport_EmptyServers(t *testing.T) {
	data := &ReportData{
		ReportDate:       "01/06/2026",
		TotalServers:     50,
		ServersOn:        50,
		ServersOff:       0,
		AvgUptimePct:     100.0,
		LowUptimeServers: nil,
	}

	html, err := RenderDailyReport(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "50")
	assert.Contains(t, html, "100.00%")
	// Should still render even with empty low uptime servers
	assert.Contains(t, html, "Top 0")
}

func TestRenderDailyReport_TemplatePathError(t *testing.T) {
	original := dailyReportTemplatePathOverride
	dailyReportTemplatePathOverride = func() (string, error) {
		return "", fmt.Errorf("path unavailable")
	}
	t.Cleanup(func() {
		dailyReportTemplatePathOverride = original
	})

	html, err := RenderDailyReport(&ReportData{})

	assert.Error(t, err)
	assert.Empty(t, html)
	assert.Contains(t, err.Error(), "path unavailable")
}

func TestRenderDailyReport_TemplateParseError(t *testing.T) {
	original := dailyReportTemplatePathOverride
	dailyReportTemplatePathOverride = func() (string, error) {
		return "missing-template.html", nil
	}
	t.Cleanup(func() {
		dailyReportTemplatePathOverride = original
	})

	html, err := RenderDailyReport(&ReportData{})

	assert.Error(t, err)
	assert.Empty(t, html)
	assert.Contains(t, err.Error(), "failed to parse template")
}
