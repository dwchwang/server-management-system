package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for the server service.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	Log      LogConfig
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name string
	Port string
	Env  string
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Schema   string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string.
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode, c.Schema,
	)
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Addr returns the Redis address string.
func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// KafkaConfig holds Kafka connection configuration.
type KafkaConfig struct {
	Brokers string
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level      string
	Dir        string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// LoadConfig reads configuration from .env file and environment variables.
func LoadConfig() *Config {
	viper.AutomaticEnv()

	// Try to load .env from project root (non-fatal if missing)
	viper.SetConfigFile("../.env")
	_ = viper.ReadInConfig()

	// Set defaults
	viper.SetDefault("APP_NAME", "server-service")
	viper.SetDefault("APP_PORT", "8082")
	viper.SetDefault("APP_ENV", "development")

	viper.SetDefault("SERVER_DB_HOST", "localhost")
	viper.SetDefault("SERVER_DB_PORT", "5432")
	viper.SetDefault("SERVER_DB_USER", "server_user")
	viper.SetDefault("SERVER_DB_PASSWORD", "server_pass_secret")
	viper.SetDefault("SERVER_DB_NAME", "vcs_sms")
	viper.SetDefault("SERVER_DB_SCHEMA", "server_schema")
	viper.SetDefault("SERVER_DB_SSLMODE", "disable")

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 1)

	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")

	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_DIR", "logs/server")
	viper.SetDefault("LOG_MAX_SIZE", 100)
	viper.SetDefault("LOG_MAX_BACKUPS", 10)
	viper.SetDefault("LOG_MAX_AGE", 30)
	viper.SetDefault("LOG_COMPRESS", true)

	return &Config{
		App: AppConfig{
			Name: viper.GetString("APP_NAME"),
			Port: viper.GetString("APP_PORT"),
			Env:  viper.GetString("APP_ENV"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("SERVER_DB_HOST"),
			Port:     viper.GetString("SERVER_DB_PORT"),
			User:     viper.GetString("SERVER_DB_USER"),
			Password: viper.GetString("SERVER_DB_PASSWORD"),
			DBName:   viper.GetString("SERVER_DB_NAME"),
			Schema:   viper.GetString("SERVER_DB_SCHEMA"),
			SSLMode:  viper.GetString("SERVER_DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		Kafka: KafkaConfig{
			Brokers: viper.GetString("KAFKA_BROKERS"),
		},
		Log: LogConfig{
			Level:      viper.GetString("LOG_LEVEL"),
			Dir:        viper.GetString("LOG_DIR"),
			MaxSize:    viper.GetInt("LOG_MAX_SIZE"),
			MaxBackups: viper.GetInt("LOG_MAX_BACKUPS"),
			MaxAge:     viper.GetInt("LOG_MAX_AGE"),
			Compress:   viper.GetBool("LOG_COMPRESS"),
		},
	}
}
