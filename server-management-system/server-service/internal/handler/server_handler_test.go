package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vcs-sms/server-service/internal/dto"
)

// mockServerService implements service.ServerService for testing.
type mockServerService struct {
	createResult *dto.ServerResponse
	createErr    error
	getResult    *dto.ServerResponse
	getErr       error
	listResult   *dto.ListServerResponse
	listErr      error
	updateResult *dto.ServerResponse
	updateErr    error
	deleteErr    error
}

func (m *mockServerService) CreateServer(ctx context.Context, req *dto.CreateServerRequest) (*dto.ServerResponse, error) {
	return m.createResult, m.createErr
}
func (m *mockServerService) GetServer(ctx context.Context, serverID string) (*dto.ServerResponse, error) {
	return m.getResult, m.getErr
}
func (m *mockServerService) ListServers(ctx context.Context, filter *dto.ServerFilter) (*dto.ListServerResponse, error) {
	return m.listResult, m.listErr
}
func (m *mockServerService) UpdateServer(ctx context.Context, serverID string, req *dto.UpdateServerRequest) (*dto.ServerResponse, error) {
	return m.updateResult, m.updateErr
}
func (m *mockServerService) DeleteServer(ctx context.Context, serverID string) error {
	return m.deleteErr
}

func setupServerTestRouter(handler *ServerHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	servers := r.Group("/api/v1/servers")
	{
		servers.POST("", handler.CreateServer)
		servers.GET("", handler.ListServers)
		servers.GET("/:server_id", handler.GetServer)
		servers.PUT("/:server_id", handler.UpdateServer)
		servers.DELETE("/:server_id", handler.DeleteServer)
	}
	return r
}

func TestCreateServerHandler_ValidBody(t *testing.T) {
	mock := &mockServerService{
		createResult: &dto.ServerResponse{
			ServerID:   "SRV-001",
			ServerName: "test-server",
			Status:     "off",
			IPv4:       "10.0.0.1",
		},
	}
	handler := NewServerHandler(mock)
	router := setupServerTestRouter(handler)

	body := `{"server_id":"SRV-001","server_name":"test-server","ipv4":"10.0.0.1"}`
	req, _ := http.NewRequest("POST", "/api/v1/servers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateServerHandler_InvalidBody(t *testing.T) {
	mock := &mockServerService{}
	handler := NewServerHandler(mock)
	router := setupServerTestRouter(handler)

	body := `{"server_id":""}`
	req, _ := http.NewRequest("POST", "/api/v1/servers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestListServersHandler_DefaultPagination(t *testing.T) {
	mock := &mockServerService{
		listResult: &dto.ListServerResponse{
			Servers:    []dto.ServerResponse{},
			Total:      0,
			Page:       1,
			PageSize:   20,
			TotalPages: 0,
		},
	}
	handler := NewServerHandler(mock)
	router := setupServerTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/v1/servers", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["data"] == nil {
		t.Error("expected data in response")
	}
}

func TestUpdateServerHandler_ValidBody(t *testing.T) {
	mock := &mockServerService{
		updateResult: &dto.ServerResponse{
			ServerID:   "SRV-001",
			ServerName: "updated-server",
			Status:     "off",
			IPv4:       "10.0.0.1",
		},
	}
	handler := NewServerHandler(mock)
	router := setupServerTestRouter(handler)

	body := `{"server_name":"updated-server"}`
	req, _ := http.NewRequest("PUT", "/api/v1/servers/SRV-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteServerHandler_Success(t *testing.T) {
	mock := &mockServerService{}
	handler := NewServerHandler(mock)
	router := setupServerTestRouter(handler)

	req, _ := http.NewRequest("DELETE", "/api/v1/servers/SRV-001", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteServerHandler_MissingID(t *testing.T) {
	mock := &mockServerService{}
	handler := NewServerHandler(mock)
	router := setupServerTestRouter(handler)

	req, _ := http.NewRequest("DELETE", "/api/v1/servers/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be 404 because Gin won't match the route without the param
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
