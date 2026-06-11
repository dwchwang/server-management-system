package checker

import (
	"context"
	"fmt"
	"net"
	"time"
)

// TCPChecker implements HealthChecker using real TCP connect.
// It dials the server's IPv4:TCPPort and measures response time.
type TCPChecker struct {
	Timeout time.Duration
}

// NewTCPChecker creates a new TCP health checker.
func NewTCPChecker(timeout time.Duration) *TCPChecker {
	return &TCPChecker{Timeout: timeout}
}

// Check performs a TCP health-check on the given server.
// Uses context-aware dialing so cancellation properly aborts in-flight connections.
func (c *TCPChecker) Check(ctx context.Context, server *ServerInfo) *HealthResult {
	start := time.Now()
	addr := fmt.Sprintf("%s:%d", server.IPv4, server.TCPPort)

	result := &HealthResult{
		ServerID:    server.ServerID,
		ServerName:  server.ServerName,
		CheckMethod: "tcp",
		CheckedAt:   time.Now().UTC(),
	}

	// Use DialContext so context cancellation aborts the dial
	dialer := &net.Dialer{Timeout: c.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		result.Status = "off"
		result.ResponseTimeMs = 0
		result.Error = err.Error()
	} else {
		conn.Close()
		result.Status = "on"
		result.ResponseTimeMs = int(elapsed)
	}

	return result
}

// Name returns the checker name.
func (c *TCPChecker) Name() string { return "tcp" }
