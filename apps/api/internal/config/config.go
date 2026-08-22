package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTP   HTTPConfig
	Logger LoggerConfig
}

type HTTPConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func (c HTTPConfig) Addr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

type LoggerConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	port, err := envInt("HTTP_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("HTTP_PORT: %w", err)
	}

	readTimeout, err := envDuration("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("HTTP_READ_TIMEOUT: %w", err)
	}
	writeTimeout, err := envDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("HTTP_WRITE_TIMEOUT: %w", err)
	}
	idleTimeout, err := envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("HTTP_IDLE_TIMEOUT: %w", err)
	}

	return &Config{
		HTTP: HTTPConfig{
			Host:         envString("HTTP_HOST", "0.0.0.0"),
			Port:         port,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		Logger: LoggerConfig{
			Level:  envString("LOG_LEVEL", "info"),
			Format: envString("LOG_FORMAT", "text"),
		},
	}, nil
}

func envString(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid int", v)
	}
	return n, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid duration", v)
	}
	return d, nil
}
