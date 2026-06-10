package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/vcs-sms/auth-service/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupAuthTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	dialector := postgres.New(postgres.Config{Conn: mockDB, DriverName: "postgres"})
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}
	return db, mock
}

func TestUserRepository_Create(t *testing.T) {
	db, mock := setupAuthTestDB(t)
	repo := NewUserRepository(db)

	userID := uuid.New()
	roleID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "auth_schema"."users"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &model.User{
		ID: userID, Username: "test", Email: "test@test.com",
		PasswordHash: "hash", FullName: "Test", RoleID: roleID, IsActive: true,
	}
	err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	db, mock := setupAuthTestDB(t)
	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "full_name", "role_id", "is_active", "created_at", "updated_at"}).
		AddRow(uuid.New(), "testuser", "test@test.com", "hash", "Test", uuid.New(), true, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "auth_schema"."users" WHERE username = $1 AND "auth_schema"."users"."deleted_at" IS NULL ORDER BY "auth_schema"."users"."id" LIMIT $2`)).
		WithArgs("testuser", 1).
		WillReturnRows(rows)

	user, err := repo.FindByUsername(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("expected 'testuser', got '%s'", user.Username)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	db, mock := setupAuthTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "auth_schema"."users"`)).
		WithArgs("nonexistent", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByUsername(context.Background(), "nonexistent")
	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, mock := setupAuthTestDB(t)
	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "full_name", "role_id", "is_active", "created_at", "updated_at"}).
		AddRow(uuid.New(), "testuser", "test@test.com", "hash", "Test", uuid.New(), true, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "auth_schema"."users" WHERE email = $1 AND "auth_schema"."users"."deleted_at" IS NULL ORDER BY "auth_schema"."users"."id" LIMIT $2`)).
		WithArgs("test@test.com", 1).
		WillReturnRows(rows)

	user, err := repo.FindByEmail(context.Background(), "test@test.com")
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if user.Email != "test@test.com" {
		t.Errorf("expected 'test@test.com', got '%s'", user.Email)
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock := setupAuthTestDB(t)
	repo := NewUserRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "full_name", "role_id", "is_active", "created_at", "updated_at"}).
		AddRow(id, "testuser", "test@test.com", "hash", "Test", uuid.New(), true, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "auth_schema"."users" WHERE id = $1 AND "auth_schema"."users"."deleted_at" IS NULL ORDER BY "auth_schema"."users"."id" LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnRows(rows)

	user, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if user.ID != id {
		t.Errorf("expected ID match")
	}
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	db, mock := setupAuthTestDB(t)
	repo := NewUserRepository(db)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "auth_schema"."users" SET "last_login_at"=$1,"updated_at"=$2 WHERE id = $3 AND "auth_schema"."users"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateLastLogin(context.Background(), id)
	if err != nil {
		t.Fatalf("UpdateLastLogin failed: %v", err)
	}
}

func TestUserRepository_FindRoleByName(t *testing.T) {
	db, mock := setupAuthTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
		AddRow(roleID, "admin", "Admin role", nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "auth_schema"."roles" WHERE name = $1 ORDER BY "auth_schema"."roles"."id" LIMIT $2`)).
		WithArgs("admin", 1).
		WillReturnRows(rows)

	// Permissions preload
	permRows := sqlmock.NewRows([]string{"id", "role_id", "scope", "created_at"}).
		AddRow(uuid.New(), roleID, "server:create", nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "auth_schema"."role_permissions" WHERE "auth_schema"."role_permissions"."role_id" = $1`)).
		WithArgs(roleID).
		WillReturnRows(permRows)

	role, err := repo.FindRoleByName(context.Background(), "admin")
	if err != nil {
		t.Fatalf("FindRoleByName failed: %v", err)
	}
	if role.Name != "admin" {
		t.Errorf("expected 'admin', got '%s'", role.Name)
	}
}
