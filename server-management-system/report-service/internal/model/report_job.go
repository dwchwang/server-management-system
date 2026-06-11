package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportJob represents a report generation/sending job.
// Maps to report_schema.report_jobs.
type ReportJob struct {
	ID             string     `gorm:"column:id;primaryKey;type:uuid" json:"id"`
	ReportType     string     `gorm:"column:report_type;not null;check:report_type IN ('daily','on_demand')" json:"report_type"`
	Status         string     `gorm:"column:status;not null;default:pending;check:status IN ('pending','processing','completed','failed')" json:"status"`
	StartDate      time.Time  `gorm:"column:start_date;not null;type:date" json:"start_date"`
	EndDate        time.Time  `gorm:"column:end_date;not null;type:date" json:"end_date"`
	RecipientEmail string     `gorm:"column:recipient_email;not null" json:"recipient_email"`
	TotalServers   *int       `gorm:"column:total_servers" json:"total_servers,omitempty"`
	ServersOn      *int       `gorm:"column:servers_on" json:"servers_on,omitempty"`
	ServersOff     *int       `gorm:"column:servers_off" json:"servers_off,omitempty"`
	AvgUptimePct   *float64   `gorm:"column:avg_uptime_pct;type:decimal(5,2)" json:"avg_uptime_pct,omitempty"`
	ErrorMessage   *string    `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	SentAt         *time.Time `gorm:"column:sent_at" json:"sent_at,omitempty"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
}

// TableName overrides the default table name.
func (ReportJob) TableName() string {
	return "report_schema.report_jobs"
}

// BeforeCreate generates a UUID for new records.
func (r *ReportJob) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
