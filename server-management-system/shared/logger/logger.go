package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogConfig holds logger configuration
type LogConfig struct {
	Level      string // "debug", "info", "warn", "error"
	Dir        string // "/var/log/vcs-sms"
	MaxSize    int    // MB before rotation
	MaxBackups int    // number of backup files
	MaxAge     int    // days to keep
	Compress   bool   // compress rotated files
}

// DefaultLogConfig returns a sane default config
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Dir:        "/var/log/vcs-sms",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
}

// NewLogger creates a zerolog.Logger with file rotation and stdout output
func NewLogger(serviceName string, cfg *LogConfig) zerolog.Logger {
	if cfg == nil {
		cfg = DefaultLogConfig()
	}

	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Ensure log directory exists
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create log directory %s: %v\n", cfg.Dir, err)
	}

	// File writer with logrotate
	fileWriter := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s.log", cfg.Dir, serviceName),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// Multi-writer: stdout + file
	multi := zerolog.MultiLevelWriter(os.Stdout, fileWriter)

	return zerolog.New(multi).
		Level(level).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}
