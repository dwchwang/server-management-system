package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for the report service.
type Config struct {
	App      AppConfig
	ReportDB DatabaseConfig
	Redis    RedisConfig
	ES       ESConfig
	SMTP     SMTPConfig
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

// ESConfig holds Elasticsearch connection configuration.
type ESConfig struct {
	Addresses string
	Username  string
	Password  string
	IndexName string
}

// SMTPConfig holds SMTP email configuration.
type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	AdminEmail string
}

// KafkaConfig holds Kafka connection configuration.
type KafkaConfig struct {
	Brokers string
}

// LogConfig holds logger configuration.
type LogConfig struct {
	Level      string
	Dir        string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// LoadConfig reads configuration from environment variables via Viper.
func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	viper.SetDefault("REPORT_PORT", "8084")
	viper.SetDefault("REPORT_DB_SSLMODE", "disable")
	viper.SetDefault("REDIS_DB", "2")
	viper.SetDefault("ES_INDEX_NAME", "server-status-logs")
	viper.SetDefault("SMTP_HOST", "smtp.gmail.com")
	viper.SetDefault("SMTP_PORT", "587")
	viper.SetDefault("SMTP_FROM", "VCS-SMS <noreply@vcs-sms.com>")
	viper.SetDefault("SMTP_ADMIN_EMAIL", "admin@company.com")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_DIR", "/var/log/vcs-sms/report")
	viper.SetDefault("LOG_MAX_SIZE", "100")
	viper.SetDefault("LOG_MAX_BACKUPS", "5")
	viper.SetDefault("LOG_MAX_AGE", "30")
	viper.SetDefault("LOG_COMPRESS", "true")

	cfg := &Config{
		App: AppConfig{
			Name: "report-service",
			Port: viper.GetString("REPORT_PORT"),
			Env:  viper.GetString("APP_ENV"),
		},
		ReportDB: DatabaseConfig{
			Host:     viper.GetString("REPORT_DB_HOST"),
			Port:     viper.GetString("REPORT_DB_PORT"),
			User:     viper.GetString("REPORT_DB_USER"),
			Password: viper.GetString("REPORT_DB_PASSWORD"),
			DBName:   viper.GetString("REPORT_DB_NAME"),
			Schema:   viper.GetString("REPORT_DB_SCHEMA"),
			SSLMode:  viper.GetString("REPORT_DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		ES: ESConfig{
			Addresses: viper.GetString("ES_HOST"),
			Username:  viper.GetString("ES_USERNAME"),
			Password:  viper.GetString("ES_PASSWORD"),
			IndexName: viper.GetString("ES_INDEX_NAME"),
		},
		SMTP: SMTPConfig{
			Host:       viper.GetString("SMTP_HOST"),
			Port:       viper.GetInt("SMTP_PORT"),
			Username:   viper.GetString("SMTP_USERNAME"),
			Password:   viper.GetString("SMTP_PASSWORD"),
			From:       viper.GetString("SMTP_FROM"),
			AdminEmail: viper.GetString("SMTP_ADMIN_EMAIL"),
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

	return cfg
}
