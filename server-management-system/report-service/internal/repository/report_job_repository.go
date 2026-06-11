package repository

import (
	"context"

	"github.com/vcs-sms/report-service/internal/model"
	"gorm.io/gorm"
)

// ReportJobRepo defines the interface for report job persistence.
type ReportJobRepo interface {
	Create(ctx context.Context, job *model.ReportJob) error
	Update(ctx context.Context, job *model.ReportJob) error
	FindByID(ctx context.Context, id string) (*model.ReportJob, error)
	FindByDateRange(ctx context.Context, startDate, endDate string) ([]model.ReportJob, error)
}

type reportJobRepo struct {
	db *gorm.DB
}

// NewReportJobRepo creates a new ReportJobRepo.
func NewReportJobRepo(db *gorm.DB) ReportJobRepo {
	return &reportJobRepo{db: db}
}

func (r *reportJobRepo) Create(ctx context.Context, job *model.ReportJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *reportJobRepo) Update(ctx context.Context, job *model.ReportJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *reportJobRepo) FindByID(ctx context.Context, id string) (*model.ReportJob, error) {
	var job model.ReportJob
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *reportJobRepo) FindByDateRange(ctx context.Context, startDate, endDate string) ([]model.ReportJob, error) {
	var jobs []model.ReportJob
	err := r.db.WithContext(ctx).
		Where("start_date >= ? AND end_date <= ?", startDate, endDate).
		Order("created_at DESC").
		Find(&jobs).Error
	return jobs, err
}
