package kafka

import "context"

// Event represents a generic Kafka event
type Event struct {
	EventID   string      `json:"event_id"`
	EventType string      `json:"event_type"`
	Timestamp string      `json:"timestamp"`
	Source    string      `json:"source"`
	Data      interface{} `json:"data"`
}

// Producer publishes events to Kafka topics
type Producer interface {
	// Publish sends an event to the specified topic
	Publish(ctx context.Context, topic string, key string, value interface{}) error
	// Close shuts down the producer
	Close() error
}

// Consumer subscribes to Kafka topics and processes events
type Consumer interface {
	// Subscribe registers a handler for a topic with a consumer group
	Subscribe(topic, groupID string, handler EventHandler) error
	// Start begins consuming messages (blocking, run in goroutine)
	Start(ctx context.Context) error
	// Close shuts down the consumer
	Close() error
}

// EventHandler is a callback function that processes a received event
type EventHandler func(ctx context.Context, event *Event) error
