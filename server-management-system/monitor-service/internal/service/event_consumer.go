package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/vcs-sms/monitor-service/config"
	"github.com/vcs-sms/monitor-service/internal/model"
	"github.com/vcs-sms/monitor-service/internal/repository"
	"github.com/vcs-sms/shared/kafka"
)

// EventConsumer handles Kafka events from other services.
type EventConsumer struct {
	configRepo repository.HealthCheckConfigRepo
	cfg        config.MonitorConfig
	logger     zerolog.Logger
}

// NewEventConsumer creates a new EventConsumer.
func NewEventConsumer(configRepo repository.HealthCheckConfigRepo, cfg config.MonitorConfig, logger zerolog.Logger) *EventConsumer {
	return &EventConsumer{
		configRepo: configRepo,
		cfg:        cfg,
		logger:     logger,
	}
}

// HandleServerCreated auto-creates a health-check config when a new server is created.
func (s *EventConsumer) HandleServerCreated(ctx context.Context, event *kafka.Event) error {
	data, ok := event.Data.(map[string]interface{})
	if !ok || data == nil {
		s.logger.Error().Interface("data", event.Data).Msg("HandleServerCreated: invalid event data type")
		return fmt.Errorf("invalid event data: %T", event.Data)
	}

	serverID, ok := data["server_id"].(string)
	if !ok || serverID == "" {
		s.logger.Error().Interface("data", data).Msg("HandleServerCreated: missing or invalid server_id")
		return fmt.Errorf("missing server_id in event data")
	}

	s.logger.Info().
		Str("server_id", serverID).
		Msg("Auto-creating health-check config for new server")

	config := &model.HealthCheckConfig{
		ID:           uuid.New().String(),
		ServerID:     serverID,
		CheckMethod:  "tcp",
		TCPPort:      s.cfg.DefaultTCPPort,
		TCPTimeoutMs: s.cfg.TCPTimeout,
		UptimeRate:   s.cfg.DefaultUptime,
		IsEnabled:    true,
	}

	if err := s.configRepo.Create(ctx, config); err != nil {
		s.logger.Error().
			Err(err).
			Str("server_id", serverID).
			Msg("Failed to auto-create health-check config")
		return err
	}

	s.logger.Info().
		Str("server_id", serverID).
		Msg("Health-check config auto-created successfully")
	return nil
}

// HandleServerDeleted disables health-check when a server is deleted.
func (s *EventConsumer) HandleServerDeleted(ctx context.Context, event *kafka.Event) error {
	data, ok := event.Data.(map[string]interface{})
	if !ok || data == nil {
		s.logger.Error().Interface("data", event.Data).Msg("HandleServerDeleted: invalid event data type")
		return fmt.Errorf("invalid event data: %T", event.Data)
	}

	serverID, ok := data["server_id"].(string)
	if !ok || serverID == "" {
		s.logger.Error().Interface("data", data).Msg("HandleServerDeleted: missing or invalid server_id")
		return fmt.Errorf("missing server_id in event data")
	}

	s.logger.Info().
		Str("server_id", serverID).
		Msg("Disabling health-check config for deleted server")

	if err := s.configRepo.DisableByServerID(ctx, serverID); err != nil {
		s.logger.Error().
			Err(err).
			Str("server_id", serverID).
			Msg("Failed to disable health-check config")
		return err
	}

	return nil
}
