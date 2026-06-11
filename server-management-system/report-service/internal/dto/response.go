package dto

// ServerUptime represents a single server's uptime data.
type ServerUptime struct {
	ServerID    string  `json:"server_id"`
	ServerName  string  `json:"server_name"`
	UptimePct   float64 `json:"uptime_pct"`
	TotalChecks int64   `json:"total_checks"`
	OnChecks    int64   `json:"on_checks"`
}

// ReportSummaryResponse represents the summary report data.
type ReportSummaryResponse struct {
	StartDate        string         `json:"start_date"`
	EndDate          string         `json:"end_date"`
	TotalServers     int            `json:"total_servers"`
	ServersOn        int            `json:"servers_on"`
	ServersOff       int            `json:"servers_off"`
	AvgUptimePct     float64        `json:"avg_uptime_pct"`
	TotalChecks      int64          `json:"total_checks"`
	LowUptimeServers []ServerUptime `json:"low_uptime_servers"`
}

// SendReportResponse represents the response after sending a report.
type SendReportResponse struct {
	ReportID string                 `json:"report_id"`
	Status   string                 `json:"status"`
	Message  string                 `json:"message"`
	Summary  *ReportSummaryResponse `json:"summary,omitempty"`
}
