package prometheus

import (
	"context"

	"go.opentelemetry.io/otel/exporters/prometheus"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func NewMeter(ctx context.Context, name string) (
	otelapi.Meter, error,
) {
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(name)),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := prometheus.New(
		prometheus.WithNamespace("node-monitor"),
	)
	if err != nil {
		return nil, err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)

	meter := provider.Meter(name)

	return meter, nil
}
