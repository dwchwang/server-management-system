package repository

import (
	"context"

	"gorm.io/gorm"
)

// ServerCounter reads server inventory data needed by reports.
type ServerCounter interface {
	CountActiveServers(ctx context.Context) (int, error)
}

type serverCounterRepo struct {
	db *gorm.DB
}

// NewServerCounterRepo creates a ServerCounter backed by PostgreSQL.
func NewServerCounterRepo(db *gorm.DB) ServerCounter {
	return &serverCounterRepo{db: db}
}

func (r *serverCounterRepo) CountActiveServers(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("server_schema.servers").
		Where("deleted_at IS NULL").
		Count(&count).Error
	return int(count), err
}
