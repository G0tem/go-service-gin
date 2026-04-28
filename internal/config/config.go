package config

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// OTelTraceConfig — только специфичные для трейсинга настройки
type OTelTraceConfig struct {
	Enabled     bool          `env:"OTEL_TRACE_ENABLED, default=true"`
	Endpoint    string        `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	Insecure    bool          `env:"OTEL_EXPORTER_OTLP_INSECURE, default=true"`
	SampleRatio float64       `env:"OTEL_TRACE_SAMPLE_RATIO, default=1.0"`
	Timeout     time.Duration `env:"OTEL_TRACE_TIMEOUT, default=10s"`
}

// OTelMetricConfig — только специфичные для метрик настройки
type OTelMetricConfig struct {
	Enabled   bool          `env:"OTEL_METRIC_ENABLED, default=true"`
	Namespace string        `env:"OTEL_METRIC_NAMESPACE, default=order_service"`
	Interval  time.Duration `env:"OTEL_METRIC_INTERVAL, default=15s"`
}

type Config struct {
	// HTTP
	HTTPAddr     string        `env:"HTTP_ADDR, default=:8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT, default=5s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT, default=10s"`

	// Infrastructure
	PostgresDSN string `env:"POSTGRES_DSN, required"`
	RedisAddr   string `env:"REDIS_ADDR, default=redis:6379"`
	RabbitMQURL string `env:"RABBITMQ_URL, required"`

	// Auth
	JWTSecret      string        `env:"JWT_SECRET, required"`
	AccessTokenTTL time.Duration `env:"JWT_ACCESS_TTL, default=15m"`

	// Service metadata
	ServiceName    string `env:"OTEL_SERVICE_NAME, default=order-service"`
	ServiceVersion string `env:"SERVICE_VERSION, default=dev"`
	Environment    string `env:"ENVIRONMENT, default=development"`

	// OpenTelemetry
	OTelTrace  OTelTraceConfig
	OTelMetric OTelMetricConfig
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := envconfig.Process(context.Background(), cfg); err != nil {
		return nil, fmt.Errorf("parse environment variables: %w", err)
	}

	// Валидация
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate проверяет обязательные поля
func (c *Config) validate() error {
	if c.PostgresDSN == "" {
		return fmt.Errorf("POSTGRES_DSN is required")
	}
	if c.RabbitMQURL == "" {
		return fmt.Errorf("RABBITMQ_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.OTelTrace.SampleRatio < 0 || c.OTelTrace.SampleRatio > 1 {
		return fmt.Errorf("OTEL_TRACE_SAMPLE_RATIO must be between 0 and 1")
	}
	return nil
}

// Prefix возвращает имя метрики с префиксом (если задан namespace)
func (c OTelMetricConfig) Prefix(name string) string {
	if c.Namespace == "" {
		return name
	}
	return c.Namespace + "_" + name
}
