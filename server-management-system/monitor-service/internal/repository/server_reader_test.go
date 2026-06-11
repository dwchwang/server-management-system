package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupServerMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	dialector := postgres.New(postgres.Config{Conn: mockDB})
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("Failed to open GORM: %v", err)
	}
	return db, mock
}

func TestServerReader_GetAllActiveServers_Empty(t *testing.T) {
	db, mock := setupServerMockDB(t)
	reader := NewServerReader(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "server_schema"."servers" WHERE deleted_at IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"server_id", "server_name", "ipv4", "status", "created_at"}))

	servers, err := reader.GetAllActiveServers(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if servers == nil {
		t.Error("Expected empty slice, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestServerReader_GetAllActiveServers_WithData(t *testing.T) {
	db, mock := setupServerMockDB(t)
	reader := NewServerReader(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"server_id", "server_name", "ipv4", "status", "created_at"}).
		AddRow("SRV-001", "Server 1", "10.0.0.1", "on", now).
		AddRow("SRV-002", "Server 2", "10.0.0.2", "off", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "server_schema"."servers" WHERE deleted_at IS NULL`)).
		WillReturnRows(rows)

	servers, err := reader.GetAllActiveServers(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}
	if servers[0].ServerID != "SRV-001" {
		t.Errorf("Expected SRV-001, got %s", servers[0].ServerID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestServerReader_GetAllActiveServers_DBError(t *testing.T) {
	db, mock := setupServerMockDB(t)
	reader := NewServerReader(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "server_schema"."servers" WHERE deleted_at IS NULL`)).
		WillReturnError(gorm.ErrInvalidDB)

	_, err := reader.GetAllActiveServers(context.Background())
	if err == nil {
		t.Error("Expected error from DB")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestServerReader_BatchUpdateStatus_Success(t *testing.T) {
	db, mock := setupServerMockDB(t)
	reader := NewServerReader(db)

	changes := []StatusChangeEvent{
		{ServerID: "SRV-001", OldStatus: "on", NewStatus: "off"},
		{ServerID: "SRV-002", OldStatus: "off", NewStatus: "on"},
	}

	// GORM wraps in transaction for batch update
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "server_schema"."servers" SET "status"=$1,"updated_at"=$2 WHERE server_id = $3`)).
		WithArgs("off", sqlmock.AnyArg(), "SRV-001").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "server_schema"."servers" SET "status"=$1,"updated_at"=$2 WHERE server_id = $3`)).
		WithArgs("on", sqlmock.AnyArg(), "SRV-002").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := reader.BatchUpdateStatus(context.Background(), changes)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestServerReader_BatchUpdateStatus_Empty(t *testing.T) {
	db, _ := setupServerMockDB(t)
	reader := NewServerReader(db)

	// No expectations should be set for empty changes
	err := reader.BatchUpdateStatus(context.Background(), nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestServerReader_BatchUpdateStatus_Error(t *testing.T) {
	db, mock := setupServerMockDB(t)
	reader := NewServerReader(db)

	changes := []StatusChangeEvent{
		{ServerID: "SRV-001", OldStatus: "on", NewStatus: "off"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "server_schema"."servers" SET "status"=$1,"updated_at"=$2 WHERE server_id = $3`)).
		WithArgs("off", sqlmock.AnyArg(), "SRV-001").
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	err := reader.BatchUpdateStatus(context.Background(), changes)
	if err == nil {
		t.Error("Expected error from DB")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}
