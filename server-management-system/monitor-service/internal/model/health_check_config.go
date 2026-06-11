package model

import "time"

// HealthCheckConfig represents a row in monitor_schema.health_check_configs.
type HealthCheckConfig struct {
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	ServerID     string    `gorm:"column:server_id;uniqueIndex;not null" json:"server_id"`
	CheckMethod  string    `gorm:"column:check_method;default:tcp" json:"check_method"`
	TCPPort      int       `gorm:"column:tcp_port;default:80" json:"tcp_port"`
	TCPTimeoutMs int       `gorm:"column:tcp_timeout_ms;default:5000" json:"tcp_timeout_ms"`
	UptimeRate   float64   `gorm:"column:uptime_rate;type:decimal(3,2);default:0.95" json:"uptime_rate"`
	IsEnabled    bool      `gorm:"column:is_enabled;default:true" json:"is_enabled"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName overrides the default table name.
func (HealthCheckConfig) TableName() string {
	return "monitor_schema.health_check_configs"
}
