package repository

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/vcs-sms/report-service/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return db, mock
}

func TestReportJobRepo_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	repo := NewReportJobRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "report_schema"."report_jobs"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	job := &model.ReportJob{
		ReportType:     "on_demand",
		Status:         "pending",
		RecipientEmail: "test@test.com",
	}

	err := repo.Create(context.Background(), job)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportJobRepo_FindByDateRange(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	repo := NewReportJobRepo(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "report_type", "status", "start_date", "end_date",
		"recipient_email", "total_servers", "servers_on", "servers_off",
		"avg_uptime_pct", "error_message", "sent_at", "created_at", "updated_at",
	}).AddRow(
		"job-1", "daily", "completed", now, now,
		"admin@test.com", 100, 95, 5, 95.5, nil, now, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "report_schema"."report_jobs"`)).
		WillReturnRows(rows)

	jobs, err := repo.FindByDateRange(context.Background(), "2026-06-01", "2026-06-30")
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "job-1", jobs[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportJobRepo_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	repo := NewReportJobRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "report_schema"."report_jobs"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	now := time.Now()
	job := &model.ReportJob{
		ID:             "job-1",
		ReportType:     "on_demand",
		Status:         "completed",
		StartDate:      now,
		EndDate:        now,
		RecipientEmail: "user@test.com",
	}

	err := repo.Update(context.Background(), job)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDailySnapshotRepo_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	repo := NewDailySnapshotRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "report_schema"."daily_snapshots"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	snapshot := &model.DailySnapshot{
		TotalServers: 100,
		ServersOn:    95,
		ServersOff:   5,
		AvgUptimePct: 95.0,
	}

	err := repo.Create(context.Background(), snapshot)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportJobRepo_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	repo := NewReportJobRepo(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "report_type", "status", "start_date", "end_date",
		"recipient_email", "total_servers", "servers_on", "servers_off",
		"avg_uptime_pct", "error_message", "sent_at", "created_at", "updated_at",
	}).AddRow(
		"job-1", "on_demand", "completed", now, now,
		"user@test.com", 50, 48, 2, 96.0, nil, now, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "report_schema"."report_jobs"`)).
		WillReturnRows(rows)

	job, err := repo.FindByID(context.Background(), "job-1")
	assert.NoError(t, err)
	assert.Equal(t, "job-1", job.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDailySnapshotRepo_FindByDate(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	repo := NewDailySnapshotRepo(db)
	targetDate := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "snapshot_date", "total_servers", "servers_on", "servers_off",
		"avg_uptime_pct", "low_uptime_servers", "created_at",
	}).AddRow(
		"snap-1", targetDate, 100, 95, 5, 95.5, "[]", now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "report_schema"."daily_snapshots"`)).
		WillReturnRows(rows)

	snapshot, err := repo.FindByDate(context.Background(), targetDate)
	assert.NoError(t, err)
	assert.Equal(t, "snap-1", snapshot.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDailySnapshotRepo_FindByDateRange(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	repo := NewDailySnapshotRepo(db)
	targetDate := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "snapshot_date", "total_servers", "servers_on", "servers_off",
		"avg_uptime_pct", "low_uptime_servers", "created_at",
	}).AddRow(
		"snap-1", targetDate, 100, 95, 5, 95.5, "[]", now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "report_schema"."daily_snapshots"`)).
		WillReturnRows(rows)

	snapshots, err := repo.FindByDateRange(context.Background(), targetDate, targetDate)

	assert.NoError(t, err)
	assert.Len(t, snapshots, 1)
	assert.Equal(t, "snap-1", snapshots[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServerCounterRepo_CountActiveServers(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	repo := NewServerCounterRepo(db)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(10000)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "server_schema"."servers" WHERE deleted_at IS NULL`)).
		WillReturnRows(rows)

	count, err := repo.CountActiveServers(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 10000, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestESUptimeRepo_ParseSummaryBodyUsesInclusiveEndDate(t *testing.T) {
	repo := &esUptimeRepo{}
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	exclusiveEnd := time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	body := strings.NewReader(`{
		"aggregations": {
			"per_server": {
				"buckets": [
					{
						"key": "SRV-001",
						"total_checks": {"value": 10},
						"on_checks": {"doc_count": 8},
						"latest_check": {
							"hits": {"hits": [{"_source": {"status": "on", "server_name": "Server 1"}}]}
						},
						"uptime_rate": {"value": 80}
					},
					{
						"key": "SRV-002",
						"total_checks": {"value": 10},
						"on_checks": {"doc_count": 1},
						"latest_check": {
							"hits": {"hits": [{"_source": {"status": "off", "server_name": "Server 2"}}]}
						},
						"uptime_rate": {"value": 10}
					}
				]
			},
			"avg_uptime": {"value": 45}
		}
	}`)

	summary, err := repo.parseSummaryBody(body, start, exclusiveEnd)

	assert.NoError(t, err)
	assert.Equal(t, "2026-06-01", summary.StartDate)
	assert.Equal(t, "2026-06-10", summary.EndDate)
	assert.Equal(t, 2, summary.TotalServers)
	assert.Equal(t, 1, summary.ServersOn)
	assert.Equal(t, 1, summary.ServersOff)
	assert.Equal(t, int64(20), summary.TotalChecks)
	assert.Equal(t, 45.0, summary.AvgUptimePct)
}

func TestESUptimeRepo_ParseLowUptimeBodySortsAndLimits(t *testing.T) {
	repo := &esUptimeRepo{}
	body := strings.NewReader(`{
		"aggregations": {
			"per_server": {
				"buckets": [
					{
						"key": "SRV-001",
						"total_checks": {"value": 10},
						"on_checks": {"doc_count": 9},
						"server_name": {"hits": {"hits": [{"_source": {"server_name": "Server 1"}}]}}
					},
					{
						"key": "SRV-002",
						"total_checks": {"value": 10},
						"on_checks": {"doc_count": 2},
						"server_name": {"hits": {"hits": [{"_source": {"server_name": "Server 2"}}]}}
					},
					{
						"key": "SRV-003",
						"total_checks": {"value": 10},
						"on_checks": {"doc_count": 5},
						"server_name": {"hits": {"hits": [{"_source": {"server_name": "Server 3"}}]}}
					}
				]
			}
		}
	}`)

	servers, err := repo.parseLowUptimeBody(body, 2)

	assert.NoError(t, err)
	assert.Len(t, servers, 2)
	assert.Equal(t, "SRV-002", servers[0].ServerID)
	assert.Equal(t, 20.0, servers[0].UptimePct)
	assert.Equal(t, "SRV-003", servers[1].ServerID)
	assert.Equal(t, 50.0, servers[1].UptimePct)
}

func TestESUptimeRepo_GetUptimeSummary(t *testing.T) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://example.com"},
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "/server-status-logs/_search", req.URL.Path)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				Body: io.NopCloser(strings.NewReader(`{
					"aggregations": {
						"per_server": {
							"buckets": [{
								"key": "SRV-001",
								"total_checks": {"value": 10},
								"on_checks": {"doc_count": 8},
								"latest_check": {"hits": {"hits": [{"_source": {"status": "on", "server_name": "Server 1"}}]}},
								"uptime_rate": {"value": 80}
							}]
						},
						"avg_uptime": {"value": 80}
					}
				}`)),
			}, nil
		}),
	})
	assert.NoError(t, err)

	repo := NewESUptimeRepo(client, "server-status-logs")
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	summary, err := repo.GetUptimeSummary(context.Background(), start, end)

	assert.NoError(t, err)
	assert.Equal(t, "2026-06-01", summary.EndDate)
	assert.Equal(t, 1, summary.TotalServers)
	assert.Equal(t, 1, summary.ServersOn)
}

func TestESUptimeRepo_GetLowUptimeServers(t *testing.T) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://example.com"},
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				Body: io.NopCloser(strings.NewReader(`{
					"aggregations": {
						"per_server": {
							"buckets": [{
								"key": "SRV-001",
								"total_checks": {"value": 10},
								"on_checks": {"doc_count": 2},
								"server_name": {"hits": {"hits": [{"_source": {"server_name": "Server 1"}}]}}
							}]
						}
					}
				}`)),
			}, nil
		}),
	})
	assert.NoError(t, err)

	repo := NewESUptimeRepo(client, "server-status-logs")
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	servers, err := repo.GetLowUptimeServers(context.Background(), start, end, 10)

	assert.NoError(t, err)
	assert.Len(t, servers, 1)
	assert.Equal(t, 20.0, servers[0].UptimePct)
}

func TestESUptimeRepo_GetUptimeSummaryReturnsESError(t *testing.T) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://example.com"},
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				Body:       io.NopCloser(strings.NewReader(`{"error":"boom"}`)),
			}, nil
		}),
	})
	assert.NoError(t, err)

	repo := NewESUptimeRepo(client, "server-status-logs")
	_, err = repo.GetUptimeSummary(context.Background(), time.Now(), time.Now().Add(time.Hour))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ES search returned error")
}

func TestESUptimeRepo_GetLowUptimeServersReturnsESError(t *testing.T) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://example.com"},
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				Body:       io.NopCloser(strings.NewReader(`{"error":"bad gateway"}`)),
			}, nil
		}),
	})
	assert.NoError(t, err)

	repo := NewESUptimeRepo(client, "server-status-logs")
	_, err = repo.GetLowUptimeServers(context.Background(), time.Now(), time.Now().Add(time.Hour), 10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ES search returned error")
}
