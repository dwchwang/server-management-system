package simulator

import (
	"net"
	"sync"
)

// FakeServer represents a single simulated server with a TCP listener.
type FakeServer struct {
	Index      int
	Port       int
	UptimeRate float64
	listener   net.Listener
	isOnline   bool
	mu         sync.Mutex
}

// NewFakeServer creates a new fake server
func NewFakeServer(index, port int, uptimeRate float64) *FakeServer {
	return &FakeServer{
		Index:      index,
		Port:       port,
		UptimeRate: uptimeRate,
		isOnline:   false,
	}
}

// StartListening opens the TCP port and starts accepting connections.
func (s *FakeServer) StartListening() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isOnline {
		return nil
	}

	ln, err := net.Listen("tcp", formatPort(s.Port))
	if err != nil {
		return err
	}

	s.listener = ln
	s.isOnline = true

	// Accept loop: accept then immediately close (sufficient for TCP health-check)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			conn.Close()
		}
	}()

	return nil
}

// StopListening closes the TCP port making connection attempts fail with "connection refused".
func (s *FakeServer) StopListening() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
	s.isOnline = false
}

// IsOnline returns the current online status
func (s *FakeServer) IsOnline() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isOnline
}

// Close shuts down the listener
func (s *FakeServer) Close() {
	s.StopListening()
}

func formatPort(port int) string {
	return net.JoinHostPort("", intToStr(port))
}

func intToStr(n int) string {
	// Simple itoa for port formatting
	if n == 0 {
		return "0"
	}
	var digits []byte
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
