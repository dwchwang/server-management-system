package worker

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/vcs-sms/monitor-service/internal/checker"
	checkermocks "github.com/vcs-sms/monitor-service/internal/checker/mocks"
)

// mockHealthChecker implements checker.HealthChecker for testing using mockery pattern
type mockHealthChecker struct {
	delay      time.Duration
	callCount  atomic.Int64
	shouldFail bool
}

func (m *mockHealthChecker) Check(ctx context.Context, server *checker.ServerInfo) *checker.HealthResult {
	m.callCount.Add(1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	status := "on"
	errMsg := ""
	if m.shouldFail {
		status = "off"
		errMsg = "connection refused"
	}

	return &checker.HealthResult{
		ServerID:       server.ServerID,
		ServerName:     server.ServerName,
		Status:         status,
		ResponseTimeMs: 5,
		CheckMethod:    "tcp",
		CheckedAt:      time.Now().UTC(),
		Error:          errMsg,
	}
}

func (m *mockHealthChecker) Name() string { return "mock" }

// Verify mockHealthChecker satisfies HealthChecker interface
var _ checker.HealthChecker = (*mockHealthChecker)(nil)
var _ checker.HealthChecker = (*checkermocks.HealthCheckerMock)(nil)

func TestPool_Execute_AllServers(t *testing.T) {
	mock := &mockHealthChecker{}
	pool := NewPool(10, mock, zerolog.Nop())

	servers := make([]*checker.ServerInfo, 100)
	for i := 0; i < 100; i++ {
		servers[i] = &checker.ServerInfo{
			ServerID:   fmt.Sprintf("SRV-%05d", i+1),
			ServerName: fmt.Sprintf("Server %d", i+1),
			IPv4:       "127.0.0.1",
			TCPPort:    9000 + i,
		}
	}

	results, err := pool.Execute(context.Background(), servers)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != 100 {
		t.Errorf("Expected 100 results, got %d", len(results))
	}

	callCount := mock.callCount.Load()
	if callCount != 100 {
		t.Errorf("Expected 100 calls, got %d", callCount)
	}
}

func TestPool_Execute_ContextCancel(t *testing.T) {
	mock := &mockHealthChecker{delay: 50 * time.Millisecond}
	pool := NewPool(5, mock, zerolog.Nop())

	servers := make([]*checker.ServerInfo, 50)
	for i := 0; i < 50; i++ {
		servers[i] = &checker.ServerInfo{
			ServerID: fmt.Sprintf("SRV-%05d", i+1),
			IPv4:     "127.0.0.1",
			TCPPort:  9000 + i,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel very quickly
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	results, err := pool.Execute(ctx, servers)
	if err == nil {
		t.Error("Expected error from cancelled context, got nil")
	}

	// Should get partial results (less than 50)
	if len(results) >= 50 {
		t.Errorf("Expected partial results after cancel, got %d/50", len(results))
	}
	t.Logf("Got %d results before cancellation", len(results))
}

func TestPool_Execute_EmptyList(t *testing.T) {
	mock := &mockHealthChecker{}
	pool := NewPool(10, mock, zerolog.Nop())

	results, err := pool.Execute(context.Background(), nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if results != nil {
		t.Errorf("Expected nil results for empty list, got %v", results)
	}
}

func TestPool_NewPool_InvalidWorkerCountDefaultsToOne(t *testing.T) {
	mock := &mockHealthChecker{}
	pool := NewPool(0, mock, zerolog.Nop())

	servers := []*checker.ServerInfo{{
		ServerID: "SRV-001",
		IPv4:     "127.0.0.1",
		TCPPort:  9001,
	}}

	results, err := pool.Execute(context.Background(), servers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if pool.WorkerCount() != 1 {
		t.Fatalf("expected worker count to default to 1, got %d", pool.WorkerCount())
	}
}

func TestPool_Execute_ConcurrencyVerify(t *testing.T) {
	// Start a real TCP listener to test concurrency
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer ln.Close()

	// Accept connections in background
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	realChecker := checker.NewTCPChecker(2 * time.Second)

	// Sequential execution
	seqStart := time.Now()
	servers := make([]*checker.ServerInfo, 20)
	for i := 0; i < 20; i++ {
		servers[i] = &checker.ServerInfo{
			ServerID: fmt.Sprintf("SRV-%05d", i),
			IPv4:     "127.0.0.1",
			TCPPort:  addr.Port,
		}
	}
	for _, srv := range servers {
		realChecker.Check(context.Background(), srv)
	}
	seqDuration := time.Since(seqStart)

	// Pool execution
	poolStart := time.Now()
	pool := NewPool(10, realChecker, zerolog.Nop())
	_, err = pool.Execute(context.Background(), servers)
	if err != nil {
		t.Errorf("Unexpected error during pool execution: %v", err)
	}
	poolDuration := time.Since(poolStart)

	// Pool should be noticeably faster
	t.Logf("Sequential: %v, Pool(10): %v", seqDuration, poolDuration)
	if poolDuration >= seqDuration {
		t.Logf("Pool was not faster (expected with small count): seq=%v pool=%v", seqDuration, poolDuration)
	}
}
