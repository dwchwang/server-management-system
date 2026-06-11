package email

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"runtime"

	"github.com/vcs-sms/report-service/internal/dto"
)

// ReportData holds the data for rendering the daily report HTML email.
type ReportData struct {
	ReportDate       string
	TotalServers     int
	ServersOn        int
	ServersOff       int
	AvgUptimePct     float64
	LowUptimeServers []dto.ServerUptime
}

var dailyReportTemplatePathOverride func() (string, error)

func resolveDailyReportTemplatePath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	return filepath.Join(filepath.Dir(filename), "templates", "daily_report.html"), nil
}

// RenderDailyReport renders the daily report HTML email template.
func RenderDailyReport(data *ReportData) (string, error) {
	// Get the path to the template file relative to this source file
	templatePath, err := resolveDailyReportTemplatePath()
	if dailyReportTemplatePathOverride != nil {
		templatePath, err = dailyReportTemplatePathOverride()
	}
	if err != nil {
		return "", err
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}
