package kafka

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// DummyProducer is a placeholder Kafka producer.
// Will be replaced with real Sarama or confluent-kafka-go implementation in Phase 1.
type DummyProducer struct {
	mu     sync.Mutex
	closed bool
	logger zerolog.Logger
}

// NewDummyProducer creates a dummy producer for development
func NewDummyProducer(logger zerolog.Logger) *DummyProducer {
	return &DummyProducer{logger: logger}
}

func (p *DummyProducer) Publish(topic string, key string, value interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	p.logger.Info().
		Str("topic", topic).
		Str("key", key).
		Interface("value", value).
		Msg("Kafka event published (dummy)")

	return nil
}

func (p *DummyProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}
