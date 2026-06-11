package repository

import (
	"context"
	"time"

	"github.com/vcs-sms/report-service/internal/model"
	"gorm.io/gorm"
)

// DailySnapshotRepo defines the interface for daily snapshot persistence.
type DailySnapshotRepo interface {
	Create(ctx context.Context, snapshot *model.DailySnapshot) error
	FindByDate(ctx context.Context, date time.Time) (*model.DailySnapshot, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]model.DailySnapshot, error)
}

type dailySnapshotRepo struct {
	db *gorm.DB
}

// NewDailySnapshotRepo creates a new DailySnapshotRepo.
func NewDailySnapshotRepo(db *gorm.DB) DailySnapshotRepo {
	return &dailySnapshotRepo{db: db}
}

func (r *dailySnapshotRepo) Create(ctx context.Context, snapshot *model.DailySnapshot) error {
	return r.db.WithContext(ctx).Create(snapshot).Error
}

func (r *dailySnapshotRepo) FindByDate(ctx context.Context, date time.Time) (*model.DailySnapshot, error) {
	var snapshot model.DailySnapshot
	dateStr := date.Format("2006-01-02")
	err := r.db.WithContext(ctx).Where("snapshot_date = ?", dateStr).First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *dailySnapshotRepo) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]model.DailySnapshot, error) {
	var snapshots []model.DailySnapshot
	err := r.db.WithContext(ctx).
		Where("snapshot_date >= ? AND snapshot_date <= ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("snapshot_date ASC").
		Find(&snapshots).Error
	return snapshots, err
}
