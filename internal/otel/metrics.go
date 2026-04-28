//go:build !skip_otel

package otel

import (
	"context"
	"fmt"
	"os"

	promclient "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"github.com/G0tem/go-service-gin/internal/config"
)

type BusinessMetrics struct {
	OrdersCreated metric.Int64Counter
	OrderAmount   metric.Float64Histogram
	ActiveOrders  metric.Int64UpDownCounter
}

func InitMetrics(ctx context.Context, serviceName string, metricCfg config.OTelMetricConfig) (
	promclient.Gatherer,
	*BusinessMetrics,
	*sdkmetric.MeterProvider,
	error,
) {
	if !metricCfg.Enabled {
		otel.SetMeterProvider(sdkmetric.NewMeterProvider())
		return nil, &BusinessMetrics{}, nil, nil
	}

	registry := promclient.NewRegistry()

	exporter, err := promexporter.New(
		promexporter.WithRegisterer(registry),
		promexporter.WithoutUnits(),
		promexporter.WithoutScopeInfo(),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create prometheus exporter: %w", err)
	}

	// 🔑 Resource: используем ServiceName из конфига метрик
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName), // ← берём из параметра
			semconv.ServiceVersion(os.Getenv("SERVICE_VERSION")),
		),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create resource: %w", err)
	}

	// Prometheus exporter — Pull-модель, передаём напрямую
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(provider)
	meter := provider.Meter(serviceName)

	// 🔑 Создаём метрики с префиксом через метод из config
	ordersCreated, err := meter.Int64Counter(
		metricCfg.Prefix("orders_created_total"),
		metric.WithDescription("Total number of successfully created orders"),
		metric.WithUnit("{order}"),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create orders_created metric: %w", err)
	}

	orderAmount, err := meter.Float64Histogram(
		metricCfg.Prefix("order_amount"),
		metric.WithDescription("Distribution of order amounts in USD"),
		metric.WithUnit("USD"),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create order_amount metric: %w", err)
	}

	activeOrders, err := meter.Int64UpDownCounter(
		metricCfg.Prefix("active_orders"),
		metric.WithDescription("Current number of orders in pending/processing state"),
		metric.WithUnit("{order}"),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create active_orders metric: %w", err)
	}

	metrics := &BusinessMetrics{
		OrdersCreated: ordersCreated,
		OrderAmount:   orderAmount,
		ActiveOrders:  activeOrders,
	}

	return registry, metrics, provider, nil
}
