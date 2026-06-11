package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/vcs-sms/monitor-service/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})
	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("Failed to open GORM with mock: %v", err)
	}
	return db, mock
}

func TestConfigRepo_GetAllEnabled(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewConfigRepo(db)

	// Mock: no rows as simple case
	mock.ExpectQuery(`SELECT \* FROM "monitor_schema"\."health_check_configs" WHERE is_enabled = \$1`).
		WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "server_id", "check_method", "tcp_port", "is_enabled"}))

	configs, err := repo.GetAllEnabled(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if configs == nil {
		t.Error("Expected empty slice, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestConfigRepo_GetByServerID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewConfigRepo(db)

	rows := sqlmock.NewRows([]string{"id", "server_id", "check_method", "tcp_port", "is_enabled"}).
		AddRow("uuid-1", "SRV-00001", "tcp", 9001, true)

	mock.ExpectQuery(`SELECT \* FROM "monitor_schema"\."health_check_configs" WHERE server_id = \$1`).
		WithArgs("SRV-00001", 1).
		WillReturnRows(rows)

	cfg, err := repo.GetByServerID(context.Background(), "SRV-00001")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}
	if cfg.ServerID != "SRV-00001" {
		t.Errorf("Expected server_id 'SRV-00001', got '%s'", cfg.ServerID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestConfigRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewConfigRepo(db)

	mock.ExpectExec(`INSERT INTO "monitor_schema"\."health_check_configs"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	cfg := &model.HealthCheckConfig{
		ServerID:     "SRV-00002",
		CheckMethod:  "tcp",
		TCPPort:      8080,
		TCPTimeoutMs: 5000,
		UptimeRate:   0.95,
		IsEnabled:    true,
	}

	err := repo.Create(context.Background(), cfg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestConfigRepo_DisableByServerID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewConfigRepo(db)

	mock.ExpectExec(`UPDATE "monitor_schema"\."health_check_configs" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DisableByServerID(context.Background(), "SRV-00003")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestConfigRepo_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewConfigRepo(db)

	mock.ExpectExec(`UPDATE "monitor_schema"\."health_check_configs" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	cfg := &model.HealthCheckConfig{
		ID: "uuid-1", ServerID: "SRV-00001", CheckMethod: "tcp",
		TCPPort: 8080, TCPTimeoutMs: 5000, UptimeRate: 0.95, IsEnabled: true,
	}

	err := repo.Update(context.Background(), cfg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}
