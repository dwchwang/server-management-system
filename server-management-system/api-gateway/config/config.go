package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the API gateway.
type Config struct {
	App                AppConfig
	JWTSecret          string
	RateLimit          int
	RateLimitWindow    time.Duration
	CORSAllowedOrigins []string
	AuthServiceURL     string
	ServerServiceURL   string
	MonitorServiceURL  string
	ReportServiceURL   string
	FileIOServiceURL   string
	Redis              RedisConfig
	Log                LogConfig
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name string
	Port string
	Env  string
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
	return c.Host + ":" + c.Port
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
	viper.SetDefault("APP_NAME", "api-gateway")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_ENV", "development")

	viper.SetDefault("JWT_SECRET", "vcs-sms-dev-secret-change-in-production")

	viper.SetDefault("RATE_LIMIT", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW_SECONDS", 60)

	viper.SetDefault("AUTH_SERVICE_URL", "http://auth-service:8081")
	viper.SetDefault("SERVER_SERVICE_URL", "http://server-service:8082")
	viper.SetDefault("MONITOR_SERVICE_URL", "http://monitor-service:8083")
	viper.SetDefault("REPORT_SERVICE_URL", "http://report-service:8084")
	viper.SetDefault("FILEIO_SERVICE_URL", "http://fileio-service:8085")

	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_DIR", "logs/gateway")
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
		JWTSecret:          viper.GetString("JWT_SECRET"),
		RateLimit:          viper.GetInt("RATE_LIMIT"),
		RateLimitWindow:    time.Duration(viper.GetInt("RATE_LIMIT_WINDOW_SECONDS")) * time.Second,
		CORSAllowedOrigins: splitCSV(viper.GetString("CORS_ALLOWED_ORIGINS")),
		AuthServiceURL:     viper.GetString("AUTH_SERVICE_URL"),
		ServerServiceURL:   viper.GetString("SERVER_SERVICE_URL"),
		MonitorServiceURL:  viper.GetString("MONITOR_SERVICE_URL"),
		ReportServiceURL:   viper.GetString("REPORT_SERVICE_URL"),
		FileIOServiceURL:   viper.GetString("FILEIO_SERVICE_URL"),
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
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

// splitCSV splits a comma-separated string into a slice, trimming whitespace.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
