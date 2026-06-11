package dto

// SendReportRequest represents the request body for sending a report.
type SendReportRequest struct {
	StartDate string `json:"start_date" binding:"required"`  // "2006-01-02"
	EndDate   string `json:"end_date" binding:"required"`    // "2006-01-02"
	Email     string `json:"email" binding:"required,email"` // recipient email
}
