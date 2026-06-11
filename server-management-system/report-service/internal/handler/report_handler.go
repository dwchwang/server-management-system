package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vcs-sms/report-service/internal/dto"
	"github.com/vcs-sms/report-service/internal/service"
	"github.com/vcs-sms/shared/response"
)

// ReportHandler handles HTTP report endpoints.
type ReportHandler struct {
	service service.ReportService
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(svc service.ReportService) *ReportHandler {
	return &ReportHandler{service: svc}
}

// GetSummary handles GET /api/v1/reports/summary
// Query params: start_date, end_date (format: 2006-01-02)
func (h *ReportHandler) GetSummary(c *gin.Context) {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	// Validate required params
	if startStr == "" || endStr == "" {
		response.Error(c, http.StatusBadRequest, "start_date and end_date are required")
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}

	// Validate range
	if endDate.Before(startDate) {
		response.Error(c, http.StatusBadRequest, "end_date must be on or after start_date")
		return
	}

	queryEndDate := endDate.AddDate(0, 0, 1)

	// Max 90 days range
	if queryEndDate.Sub(startDate) > 90*24*time.Hour {
		response.Error(c, http.StatusBadRequest, "Date range must not exceed 90 days")
		return
	}

	// Get summary
	summary, err := h.service.GetSummary(c.Request.Context(), startDate, queryEndDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get report summary: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Report summary retrieved", summary)
}

// SendReport handles POST /api/v1/reports
// Body: { "start_date": "2006-01-02", "end_date": "2006-01-02", "email": "user@example.com" }
func (h *ReportHandler) SendReport(c *gin.Context) {
	var req dto.SendReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	result, err := h.service.SendReport(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send report: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Report sent successfully", result)
}
