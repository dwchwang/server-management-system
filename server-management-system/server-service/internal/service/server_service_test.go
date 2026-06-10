package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/vcs-sms/server-service/internal/dto"
	"github.com/vcs-sms/server-service/internal/model"
	"github.com/vcs-sms/shared/kafka"
	"gorm.io/gorm"
)

// ── Mock Repository ──

type mockServerRepo struct {
	servers                map[string]*model.Server // key = server_id
	createShouldFail       bool
	existsByServerID       map[string]bool
	existsByServerName     map[string]bool
	existsByNameExcludeErr error
}

func newMockServerRepo() *mockServerRepo {
	return &mockServerRepo{
		servers:            make(map[string]*model.Server),
		existsByServerID:   make(map[string]bool),
		existsByServerName: make(map[string]bool),
	}
}

func (m *mockServerRepo) addServer(s *model.Server) {
	m.servers[s.ServerID] = s
	m.existsByServerID[s.ServerID] = true
	m.existsByServerName[s.ServerName] = true
}

func (m *mockServerRepo) Create(ctx context.Context, s *model.Server) error {
	if m.createShouldFail {
		return errors.New("db error")
	}
	m.addServer(s)
	return nil
}

func (m *mockServerRepo) FindByServerID(ctx context.Context, serverID string) (*model.Server, error) {
	s, ok := m.servers[serverID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return s, nil
}

func (m *mockServerRepo) FindAll(ctx context.Context, filter *dto.ServerFilter) ([]model.Server, int64, error) {
	var result []model.Server
	for _, s := range m.servers {
		if filter.Status != "" && s.Status != filter.Status {
			continue
		}
		result = append(result, *s)
	}
	return result, int64(len(result)), nil
}

func (m *mockServerRepo) Update(ctx context.Context, s *model.Server) error {
	m.servers[s.ServerID] = s
	return nil
}

func (m *mockServerRepo) Delete(ctx context.Context, serverID string) error {
	_, ok := m.servers[serverID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.servers, serverID)
	delete(m.existsByServerID, serverID)
	return nil
}

func (m *mockServerRepo) ExistsByServerID(ctx context.Context, serverID string) (bool, error) {
	return m.existsByServerID[serverID], nil
}

func (m *mockServerRepo) ExistsByServerName(ctx context.Context, serverName string) (bool, error) {
	return m.existsByServerName[serverName], nil
}

func (m *mockServerRepo) ExistsByServerNameExclude(ctx context.Context, serverName string, excludeID string) (bool, error) {
	if m.existsByNameExcludeErr != nil {
		return false, m.existsByNameExcludeErr
	}
	for id, s := range m.servers {
		if s.ServerName == serverName && id != excludeID {
			return true, nil
		}
	}
	return false, nil
}

// ── Test Helper ──

func newTestServerService() (ServerService, *mockServerRepo) {
	repo := newMockServerRepo()
	producer := kafka.NewDummyProducer(zerolog.Logger{})
	svc := NewServerService(repo, nil, producer)
	return svc, repo
}

func makeServer(id, name, ip string) *model.Server {
	now := time.Now().UTC()
	return &model.Server{
		ServerID:   id,
		ServerName: name,
		Status:     "off",
		IPv4:       ip,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// ── Create Tests ──

func TestCreateServer_Success(t *testing.T) {
	svc, _ := newTestServerService()

	req := &dto.CreateServerRequest{
		ServerID:   "SRV-001",
		ServerName: "web-server-01",
		IPv4:       "192.168.1.100",
		OS:         "Ubuntu 22.04",
	}

	resp, err := svc.CreateServer(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}
	if resp.ServerID != "SRV-001" {
		t.Errorf("expected ServerID 'SRV-001', got '%s'", resp.ServerID)
	}
	if resp.Status != "off" {
		t.Errorf("expected Status 'off', got '%s'", resp.Status)
	}
}

func TestCreateServer_DuplicateServerID(t *testing.T) {
	svc, repo := newTestServerService()
	repo.addServer(makeServer("SRV-001", "existing", "10.0.0.1"))

	req := &dto.CreateServerRequest{
		ServerID:   "SRV-001",
		ServerName: "new-server",
		IPv4:       "10.0.0.2",
	}

	_, err := svc.CreateServer(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for duplicate server_id")
	}
}

func TestCreateServer_DuplicateServerName(t *testing.T) {
	svc, repo := newTestServerService()
	repo.addServer(makeServer("SRV-001", "my-server", "10.0.0.1"))

	req := &dto.CreateServerRequest{
		ServerID:   "SRV-002",
		ServerName: "my-server",
		IPv4:       "10.0.0.2",
	}

	_, err := svc.CreateServer(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for duplicate server_name")
	}
}

// ── Get Tests ──

func TestGetServer_Success(t *testing.T) {
	svc, repo := newTestServerService()
	repo.addServer(makeServer("SRV-001", "test-server", "10.0.0.1"))

	resp, err := svc.GetServer(context.Background(), "SRV-001")
	if err != nil {
		t.Fatalf("GetServer failed: %v", err)
	}
	if resp.ServerName != "test-server" {
		t.Errorf("expected 'test-server', got '%s'", resp.ServerName)
	}
}

func TestGetServer_NotFound(t *testing.T) {
	svc, _ := newTestServerService()

	_, err := svc.GetServer(context.Background(), "NONEXIST")
	if err == nil {
		t.Fatal("expected error for non-existent server")
	}
}

// ── List Tests ──

func TestListServers_NoFilter(t *testing.T) {
	svc, repo := newTestServerService()
	repo.addServer(makeServer("SRV-001", "server-a", "10.0.0.1"))
	repo.addServer(makeServer("SRV-002", "server-b", "10.0.0.2"))

	resp, err := svc.ListServers(context.Background(), &dto.ServerFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected 2 servers, got %d", resp.Total)
	}
}

func TestListServers_Pagination(t *testing.T) {
	svc, repo := newTestServerService()
	for i := 0; i < 15; i++ {
		repo.addServer(makeServer("SRV-"+string(rune('A'+i)), "server-"+string(rune('A'+i)), "10.0.0."+string(rune('1'+i))))
	}

	resp, err := svc.ListServers(context.Background(), &dto.ServerFilter{Page: 1, PageSize: 5})
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}
	if resp.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Page)
	}
	if resp.Total != 15 {
		t.Errorf("expected 15 total, got %d", resp.Total)
	}
	// Note: actual pagination truncation is handled at DB layer (repository),
	// not in the service. Mock repository returns all records.
}

// ── Update Tests ──

func TestUpdateServer_Success(t *testing.T) {
	svc, repo := newTestServerService()
	repo.addServer(makeServer("SRV-001", "old-name", "10.0.0.1"))

	newOS := "Ubuntu 24.04"
	req := &dto.UpdateServerRequest{OS: &newOS}

	resp, err := svc.UpdateServer(context.Background(), "SRV-001", req)
	if err != nil {
		t.Fatalf("UpdateServer failed: %v", err)
	}
	if resp.OS != "Ubuntu 24.04" {
		t.Errorf("expected OS 'Ubuntu 24.04', got '%s'", resp.OS)
	}
}

func TestUpdateServer_NotFound(t *testing.T) {
	svc, _ := newTestServerService()
	req := &dto.UpdateServerRequest{}

	_, err := svc.UpdateServer(context.Background(), "NONEXIST", req)
	if err == nil {
		t.Fatal("expected error for non-existent server")
	}
}

func TestUpdateServer_DuplicateServerName(t *testing.T) {
	svc, repo := newTestServerService()
	repo.addServer(makeServer("SRV-001", "server-a", "10.0.0.1"))
	repo.addServer(makeServer("SRV-002", "server-b", "10.0.0.2"))

	newName := "server-b"
	req := &dto.UpdateServerRequest{ServerName: &newName}

	_, err := svc.UpdateServer(context.Background(), "SRV-001", req)
	if err == nil {
		t.Fatal("expected error for duplicate server_name")
	}
}

// ── Delete Tests ──

func TestDeleteServer_Success(t *testing.T) {
	svc, repo := newTestServerService()
	repo.addServer(makeServer("SRV-001", "to-delete", "10.0.0.1"))

	err := svc.DeleteServer(context.Background(), "SRV-001")
	if err != nil {
		t.Fatalf("DeleteServer failed: %v", err)
	}

	// Verify it's gone
	_, err = svc.GetServer(context.Background(), "SRV-001")
	if err == nil {
		t.Fatal("expected server to be deleted")
	}
}

func TestDeleteServer_NotFound(t *testing.T) {
	svc, _ := newTestServerService()

	err := svc.DeleteServer(context.Background(), "NONEXIST")
	if err == nil {
		t.Fatal("expected error for non-existent server")
	}
}
