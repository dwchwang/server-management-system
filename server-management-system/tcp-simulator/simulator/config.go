package simulator

import (
	"os"
	"strconv"
	"time"
)

// Config holds TCP Simulator configuration
type Config struct {
	BasePort      int           // default: 9001
	NumServers    int           // default: 10000
	TickInterval  time.Duration // default: 30s
	DefaultUptime float64       // default: 0.95
}

// LoadConfig reads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		BasePort:      getEnvInt("SIMULATOR_BASE_PORT", 9001),
		NumServers:    getEnvInt("SIMULATOR_NUM_SERVERS", 10000),
		TickInterval:  getEnvDuration("SIMULATOR_TICK_INTERVAL", 30*time.Second),
		DefaultUptime: getEnvFloat("SIMULATOR_DEFAULT_UPTIME_RATE", 0.95),
	}
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
