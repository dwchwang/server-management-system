package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vcs-sms/auth-service/internal/model"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByIDWithRole(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	FindRoleByName(ctx context.Context, name string) (*model.Role, error)
}

// userRepository implements UserRepository using GORM.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user into the database.
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByUsername retrieves an active user by username.
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail retrieves an active user by email.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID retrieves an active user by UUID.
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByIDWithRole retrieves a user by UUID with their role and permissions preloaded.
func (r *userRepository) FindByIDWithRole(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateLastLogin sets the last_login_at timestamp for a user.
func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("last_login_at", &now).Error
}

// FindRoleByName retrieves a role by its name with permissions preloaded.
func (r *userRepository) FindRoleByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("name = ?", name).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}
