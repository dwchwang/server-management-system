package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Server represents a managed server in the system.
type Server struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ServerID    string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"server_id"`
	ServerName  string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"server_name"`
	Status      string         `gorm:"type:varchar(20);not null;default:'off'" json:"status"`
	IPv4        string         `gorm:"type:varchar(15);not null" json:"ipv4"`
	OS          string         `gorm:"type:varchar(100)" json:"os,omitempty"`
	CPUCores    *int           `gorm:"type:integer" json:"cpu_cores,omitempty"`
	RAMGB       *float64       `gorm:"type:decimal(10,2)" json:"ram_gb,omitempty"`
	DiskGB      *float64       `gorm:"type:decimal(10,2)" json:"disk_gb,omitempty"`
	Location    string         `gorm:"type:varchar(255)" json:"location,omitempty"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the default table name for GORM.
func (Server) TableName() string { return "server_schema.servers" }
