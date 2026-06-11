package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vcs-sms/report-service/internal/dto"
)

// mockReportService implements service.ReportService for testing.
type mockReportService struct {
	getSummaryResult   *dto.ReportSummaryResponse
	getSummaryErr      error
	getSummaryStart    time.Time
	getSummaryEnd      time.Time
	sendReportResult   *dto.SendReportResponse
	sendReportErr      error
	sendDailyReportErr error
}

func (m *mockReportService) GetSummary(ctx context.Context, startDate, endDate time.Time) (*dto.ReportSummaryResponse, error) {
	m.getSummaryStart = startDate
	m.getSummaryEnd = endDate
	return m.getSummaryResult, m.getSummaryErr
}

func (m *mockReportService) SendReport(ctx context.Context, req *dto.SendReportRequest) (*dto.SendReportResponse, error) {
	return m.sendReportResult, m.sendReportErr
}

func (m *mockReportService) SendDailyReport(ctx context.Context) error {
	return m.sendDailyReportErr
}

func setupTestRouter(handler *ReportHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	{
		api.GET("/reports/summary", handler.GetSummary)
		api.POST("/reports", handler.SendReport)
	}
	return router
}

func TestGetSummaryHandler_ValidDates(t *testing.T) {
	mockSvc := &mockReportService{
		getSummaryResult: &dto.ReportSummaryResponse{
			StartDate:    "2026-06-01",
			EndDate:      "2026-06-10",
			TotalServers: 100,
			ServersOn:    95,
			ServersOff:   5,
			AvgUptimePct: 95.0,
		},
	}

	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/api/v1/reports/summary?start_date=2026-06-01&end_date=2026-06-10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), mockSvc.getSummaryStart)
	assert.Equal(t, time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC), mockSvc.getSummaryEnd)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "success", response["status"])
}

func TestGetSummaryHandler_MissingDates(t *testing.T) {
	mockSvc := &mockReportService{}
	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/api/v1/reports/summary", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSummaryHandler_InvalidDateFormat(t *testing.T) {
	mockSvc := &mockReportService{}
	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/api/v1/reports/summary?start_date=not-a-date&end_date=2026-06-10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSummaryHandler_EndDateBeforeStartDate(t *testing.T) {
	mockSvc := &mockReportService{}
	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/api/v1/reports/summary?start_date=2026-06-10&end_date=2026-06-01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSummaryHandler_PreviousDayEndDateIsRejected(t *testing.T) {
	mockSvc := &mockReportService{}
	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/api/v1/reports/summary?start_date=2026-06-10&end_date=2026-06-09", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.True(t, mockSvc.getSummaryStart.IsZero())
}

func TestGetSummaryHandler_RangeExceeds90Days(t *testing.T) {
	mockSvc := &mockReportService{}
	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/api/v1/reports/summary?start_date=2026-01-01&end_date=2026-06-30", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSummaryHandler_ServiceError(t *testing.T) {
	mockSvc := &mockReportService{getSummaryErr: fmt.Errorf("ES unavailable")}
	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/api/v1/reports/summary?start_date=2026-06-01&end_date=2026-06-10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSendReportHandler_ValidRequest(t *testing.T) {
	mockSvc := &mockReportService{
		sendReportResult: &dto.SendReportResponse{
			ReportID: "job-123",
			Status:   "completed",
			Message:  "Report sent successfully",
		},
	}

	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	body := `{"start_date":"2026-06-01","end_date":"2026-06-10","email":"user@test.com"}`
	req := httptest.NewRequest("POST", "/api/v1/reports", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSendReportHandler_InvalidEmail(t *testing.T) {
	mockSvc := &mockReportService{}
	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	body := `{"start_date":"2026-06-01","end_date":"2026-06-10","email":"not-an-email"}`
	req := httptest.NewRequest("POST", "/api/v1/reports", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendReportHandler_MissingBody(t *testing.T) {
	mockSvc := &mockReportService{}
	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("POST", "/api/v1/reports", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendReportHandler_ServiceError(t *testing.T) {
	mockSvc := &mockReportService{sendReportErr: fmt.Errorf("SMTP unavailable")}
	handler := NewReportHandler(mockSvc)
	router := setupTestRouter(handler)

	body := `{"start_date":"2026-06-01","end_date":"2026-06-10","email":"user@test.com"}`
	req := httptest.NewRequest("POST", "/api/v1/reports", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
