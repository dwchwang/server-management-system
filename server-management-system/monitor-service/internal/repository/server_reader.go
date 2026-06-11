package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ServerInfo represents a server row read from server_schema.servers.
type ServerInfo struct {
	ServerID   string    `gorm:"column:server_id"`
	ServerName string    `gorm:"column:server_name"`
	IPv4       string    `gorm:"column:ipv4"`
	Status     string    `gorm:"column:status"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

// StatusChangeEvent represents a server status change.
type StatusChangeEvent struct {
	ServerID  string    `json:"server_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	ChangedAt time.Time `json:"changed_at"`
}

// ServerReader defines the interface for cross-schema server reads.
type ServerReader interface {
	GetAllActiveServers(ctx context.Context) ([]ServerInfo, error)
	BatchUpdateStatus(ctx context.Context, changes []StatusChangeEvent) error
}

type serverReader struct {
	db *gorm.DB
}

// NewServerReader creates a new ServerReader.
// db must have SELECT permission on server_schema.servers.
func NewServerReader(db *gorm.DB) ServerReader {
	return &serverReader{db: db}
}

// GetAllActiveServers retrieves all non-deleted servers for health-check.
func (r *serverReader) GetAllActiveServers(ctx context.Context) ([]ServerInfo, error) {
	var servers []ServerInfo
	err := r.db.WithContext(ctx).
		Table("server_schema.servers").
		Where("deleted_at IS NULL").
		Find(&servers).Error
	return servers, err
}

// BatchUpdateStatus updates status for servers that changed.
// Uses a single transaction with individual UPDATE statements.
func (r *serverReader) BatchUpdateStatus(ctx context.Context, changes []StatusChangeEvent) error {
	if len(changes) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, change := range changes {
			result := tx.Table("server_schema.servers").
				Where("server_id = ?", change.ServerID).
				Updates(map[string]interface{}{
					"status":     change.NewStatus,
					"updated_at": time.Now().UTC(),
				})
			if result.Error != nil {
				return fmt.Errorf("failed to update server %s: %w", change.ServerID, result.Error)
			}
		}
		return nil
	})
}
