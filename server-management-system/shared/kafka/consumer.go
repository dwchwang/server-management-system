package kafka

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// DummyConsumer is a placeholder Kafka consumer.
// Will be replaced with real Sarama or confluent-kafka-go implementation in Phase 1.
type DummyConsumer struct {
	mu       sync.Mutex
	handlers map[string]EventHandler // topic → handler
	closed   bool
	logger   zerolog.Logger
}

// NewDummyConsumer creates a dummy consumer for development
func NewDummyConsumer(logger zerolog.Logger) *DummyConsumer {
	return &DummyConsumer{
		handlers: make(map[string]EventHandler),
		logger:   logger,
	}
}

func (c *DummyConsumer) Subscribe(topic, groupID string, handler EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("consumer is closed")
	}

	c.handlers[topic] = handler
	c.logger.Info().
		Str("topic", topic).
		Str("group_id", groupID).
		Msg("Kafka consumer subscribed (dummy)")

	return nil
}

func (c *DummyConsumer) Start() error {
	c.logger.Info().Msg("Kafka consumer started (dummy mode — no actual consumption)")
	return nil
}

func (c *DummyConsumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}
