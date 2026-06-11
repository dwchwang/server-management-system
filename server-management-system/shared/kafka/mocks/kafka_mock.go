package mocks

import (
	"context"

	"github.com/vcs-sms/shared/kafka"
)

// ProducerMock is a test mock for kafka.Producer.
type ProducerMock struct {
	PublishFunc func(ctx context.Context, topic string, key string, value interface{}) error
	CloseFunc   func() error
}

func (m *ProducerMock) Publish(ctx context.Context, topic string, key string, value interface{}) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, topic, key, value)
	}
	return nil
}

func (m *ProducerMock) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// ConsumerMock is a test mock for kafka.Consumer.
type ConsumerMock struct {
	SubscribeFunc func(topic, groupID string, handler kafka.EventHandler) error
	StartFunc     func(ctx context.Context) error
	CloseFunc     func() error
}

func (m *ConsumerMock) Subscribe(topic, groupID string, handler kafka.EventHandler) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(topic, groupID, handler)
	}
	return nil
}

func (m *ConsumerMock) Start(ctx context.Context) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx)
	}
	<-ctx.Done()
	return nil
}

func (m *ConsumerMock) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
