package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServerUptimeJSON represents a low-uptime server stored as JSONB.
type ServerUptimeJSON struct {
	ServerID    string  `json:"server_id"`
	ServerName  string  `json:"server_name"`
	UptimePct   float64 `json:"uptime_pct"`
	TotalChecks int64   `json:"total_checks"`
	OnChecks    int64   `json:"on_checks"`
}

// DailySnapshot represents a daily uptime snapshot.
// Maps to report_schema.daily_snapshots.
type DailySnapshot struct {
	ID               string    `gorm:"column:id;primaryKey;type:uuid" json:"id"`
	SnapshotDate     time.Time `gorm:"column:snapshot_date;not null;uniqueIndex;type:date" json:"snapshot_date"`
	TotalServers     int       `gorm:"column:total_servers;not null" json:"total_servers"`
	ServersOn        int       `gorm:"column:servers_on;not null" json:"servers_on"`
	ServersOff       int       `gorm:"column:servers_off;not null" json:"servers_off"`
	AvgUptimePct     float64   `gorm:"column:avg_uptime_pct;not null;type:decimal(5,2)" json:"avg_uptime_pct"`
	LowUptimeServers string    `gorm:"column:low_uptime_servers;type:jsonb" json:"low_uptime_servers,omitempty"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// TableName overrides the default table name.
func (DailySnapshot) TableName() string {
	return "report_schema.daily_snapshots"
}

// BeforeCreate generates a UUID for new records.
func (d *DailySnapshot) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// SetLowUptimeServers marshals the server list to JSONB string.
func (d *DailySnapshot) SetLowUptimeServers(servers []ServerUptimeJSON) error {
	data, err := json.Marshal(servers)
	if err != nil {
		return err
	}
	d.LowUptimeServers = string(data)
	return nil
}

// GetLowUptimeServers unmarshals the JSONB string to server list.
func (d *DailySnapshot) GetLowUptimeServers() ([]ServerUptimeJSON, error) {
	if d.LowUptimeServers == "" || d.LowUptimeServers == "null" {
		return nil, nil
	}
	var servers []ServerUptimeJSON
	if err := json.Unmarshal([]byte(d.LowUptimeServers), &servers); err != nil {
		return nil, err
	}
	return servers, nil
}
