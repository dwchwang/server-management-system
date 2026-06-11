package simulator

import (
	"net"
	"testing"
	"time"
)

func TestManager_StartStop(t *testing.T) {
	cfg := &Config{
		BasePort:      19001,
		NumServers:    10,
		TickInterval:  5 * time.Second,
		DefaultUptime: 0.95,
	}
	manager := NewManager(cfg)

	if len(manager.servers) != 10 {
		t.Errorf("Expected 10 servers, got %d", len(manager.servers))
	}

	// Start in background
	go manager.RunControlLoop()

	// Give it time for initial evaluation
	time.Sleep(100 * time.Millisecond)

	// Some servers should be online
	onlineCount := 0
	for _, s := range manager.servers {
		if s.IsOnline() {
			onlineCount++
		}
	}
	t.Logf("Initial online count: %d/10", onlineCount)

	// Shutdown
	manager.Shutdown()
	time.Sleep(200 * time.Millisecond)

	// All should be closed
	for _, s := range manager.servers {
		if s.IsOnline() {
			t.Errorf("Server %d still online after shutdown", s.Index)
		}
	}
}

func TestManager_PortReachable(t *testing.T) {
	port := 19101
	server := NewFakeServer(1, port, 1.0)

	err := server.StartListening()
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer server.StopListening()

	if !server.IsOnline() {
		t.Error("Server should be online after StartListening")
	}

	// Try TCP connect
	conn, err := net.DialTimeout("tcp", formatPort(port), 1*time.Second)
	if err != nil {
		t.Errorf("TCP connect failed (port should be reachable): %v", err)
	} else {
		conn.Close()
	}
}

func TestManager_PortClosed(t *testing.T) {
	port := 19102
	_ = NewFakeServer(1, port, 1.0) // created but not started

	// Don't start - port should be closed
	_, err := net.DialTimeout("tcp", formatPort(port), 500*time.Millisecond)
	if err == nil {
		t.Error("TCP connect should fail when port is not listening")
	}
}

func TestManager_EvaluateToggle(t *testing.T) {
	port := 19103
	server := NewFakeServer(1, port, 1.0)

	// Initially offline
	if server.IsOnline() {
		t.Error("Server should start offline")
	}

	// Start
	if err := server.StartListening(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	if !server.IsOnline() {
		t.Error("Server should be online")
	}

	// Stop
	server.StopListening()
	if server.IsOnline() {
		t.Error("Server should be offline after StopListening")
	}

	// Start again
	if err := server.StartListening(); err != nil {
		t.Fatalf("Failed to start again: %v", err)
	}
	if !server.IsOnline() {
		t.Error("Server should be online after restart")
	}

	server.StopListening()
}
