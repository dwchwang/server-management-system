package simulator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// SimulatorManager orchestrates all 10,000 fake servers.
// Every tick interval it re-evaluates each server's On/Off state via the MathEngine.
type SimulatorManager struct {
	servers      map[int]*FakeServer
	mathEngine   *MathEngine
	basePort     int
	numServers   int
	tickInterval time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewManager creates a new SimulatorManager from configuration.
func NewManager(cfg *Config) *SimulatorManager {
	engine := NewMathEngine()
	servers := make(map[int]*FakeServer, cfg.NumServers)

	for i := 1; i <= cfg.NumServers; i++ {
		port := cfg.BasePort + i - 1
		servers[i] = NewFakeServer(i, port, cfg.DefaultUptime)
	}

	return &SimulatorManager{
		servers:      servers,
		mathEngine:   engine,
		basePort:     cfg.BasePort,
		numServers:   cfg.NumServers,
		tickInterval: cfg.TickInterval,
	}
}

// RunControlLoop starts the main control loop that periodically re-evaluates all servers.
// This method blocks until Shutdown is called.
func (m *SimulatorManager) RunControlLoop() {
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Initial evaluation
	log.Printf("[tcp-simulator] Starting with %d servers, base port %d, tick %s",
		m.numServers, m.basePort, m.tickInterval)
	m.evaluateAllServers()

	ticker := time.NewTicker(m.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			log.Println("[tcp-simulator] Shutting down...")
			m.shutdownAll()
			m.wg.Wait()
			return
		case <-ticker.C:
			m.evaluateAllServers()
		}
	}
}

// Shutdown gracefully stops the manager
func (m *SimulatorManager) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *SimulatorManager) evaluateAllServers() {
	startTime := time.Now()
	onlineCount := 0
	offlineCount := 0

	// Semaphore to limit concurrent listener open/close operations
	sem := make(chan struct{}, 200)

	var wg sync.WaitGroup
	for _, server := range m.servers {
		wg.Add(1)
		sem <- struct{}{}

		go func(s *FakeServer) {
			defer wg.Done()
			defer func() { <-sem }()

			shouldBeOn := m.mathEngine.ShouldBeOnline(s.UptimeRate, s.Index)

			if shouldBeOn && !s.IsOnline() {
				if err := s.StartListening(); err != nil {
					log.Printf("[tcp-simulator] Failed to start server %d (port %d): %v",
						s.Index, s.Port, err)
				}
			} else if !shouldBeOn && s.IsOnline() {
				s.StopListening()
			}
		}(server)
	}
	wg.Wait()

	// Count final states
	for _, s := range m.servers {
		if s.IsOnline() {
			onlineCount++
		} else {
			offlineCount++
		}
	}

	elapsed := time.Since(startTime)
	log.Printf("[tcp-simulator] Cycle complete: %d ON, %d OFF (took %s)",
		onlineCount, offlineCount, elapsed)
}

func (m *SimulatorManager) shutdownAll() {
	log.Println("[tcp-simulator] Closing all listeners...")
	var wg sync.WaitGroup
	for _, server := range m.servers {
		wg.Add(1)
		go func(s *FakeServer) {
			defer wg.Done()
			s.Close()
		}(server)
	}
	wg.Wait()
	fmt.Println("[tcp-simulator] All listeners closed.")
}
