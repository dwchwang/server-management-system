package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vcs-sms/monitor-service/internal/checker"
)

// mockESRepo is a simple test double implementing ESStatusLogRepo
type mockESRepo struct {
	bulkErr   error
	callCount int
}

func (m *mockESRepo) BulkIndex(ctx context.Context, results []*checker.HealthResult) error {
	m.callCount++
	return m.bulkErr
}

// Verify mockESRepo satisfies ESStatusLogRepo
var _ ESStatusLogRepo = (*mockESRepo)(nil)

func TestESRepo_BulkIndex_Empty(t *testing.T) {
	mock := &mockESRepo{}
	err := mock.BulkIndex(context.Background(), nil)
	if err != nil {
		t.Errorf("Expected no error for empty results, got %v", err)
	}
	if mock.callCount != 1 {
		t.Errorf("Expected 1 call, got %d", mock.callCount)
	}
}

func TestESRepo_BulkIndex_Success(t *testing.T) {
	mock := &mockESRepo{}
	results := []*checker.HealthResult{
		{
			ServerID:       "SRV-001",
			ServerName:     "Test",
			Status:         "on",
			ResponseTimeMs: 5,
			CheckMethod:    "tcp",
			CheckedAt:      time.Now().UTC(),
		},
	}
	err := mock.BulkIndex(context.Background(), results)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if mock.callCount != 1 {
		t.Errorf("Expected 1 call, got %d", mock.callCount)
	}
}

func TestESRepo_BulkIndex_Error(t *testing.T) {
	mock := &mockESRepo{bulkErr: fmt.Errorf("ES unavailable")}
	results := []*checker.HealthResult{
		{ServerID: "SRV-001", Status: "on"},
	}
	err := mock.BulkIndex(context.Background(), results)
	if err == nil {
		t.Error("Expected error from mock ES")
	}
}
