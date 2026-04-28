//go:build !skip_otel

package otel

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"github.com/G0tem/go-service-gin/internal/config"
)

// InitTracing инициализирует трейсинг на основе конфига
func InitTracing(ctx context.Context, serviceName string, cfg config.OTelTraceConfig) (func(context.Context) error, error) {
	// Если трейсинг выключен — возвращаем noop-провайдер
	if !cfg.Enabled {
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	var exporter sdktrace.SpanExporter
	var err error

	// Fallback на stdout, если endpoint не задан
	if cfg.Endpoint == "" {
		fmt.Fprintln(os.Stderr, "⚠️  OTEL_EXPORTER_OTLP_ENDPOINT not set, using stdout exporter")
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	} else {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithTimeout(cfg.Timeout),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptracegrpc.New(ctx, opts...)
	}
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName), // ← берём из параметра
			semconv.ServiceVersion(os.Getenv("SERVICE_VERSION")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	// Sampling
	sampler := sdktrace.AlwaysSample()
	if cfg.SampleRatio < 1.0 && cfg.SampleRatio > 0 {
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))
	} else if cfg.SampleRatio == 0 {
		sampler = sdktrace.NeverSample()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}
