package server

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/metric"
	otelapi "go.opentelemetry.io/otel/metric"
)

const (
	metricsNewBlockLatency   = "new_block_latency"
	metricTimeSinceLastBlock = "time_since_last_block"
)

var (
	metricDescriptions = map[string]string{
		metricsNewBlockLatency:   "Statistics on how late a node receives blocks compared to the earliest observed ones",
		metricTimeSinceLastBlock: "Time passed since last block was received",
	}
)

var (
	ErrSetupMetricsFailed = errors.New("failed to setup metrics")
)

type metrics struct {
	newBlockLatency       otelapi.Float64Histogram
	secondsSinceLastBlock otelapi.Float64Observable
}

func (m *metrics) setup(meter otelapi.Meter, observe func(ctx context.Context, o metric.Observer) error) error {
	secondsSinceLastBlock, err := meter.Float64ObservableGauge(metricTimeSinceLastBlock,
		otelapi.WithDescription(metricDescriptions[metricTimeSinceLastBlock]),
		otelapi.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("%w: %w: %s",
			ErrSetupMetricsFailed, err, metricTimeSinceLastBlock,
		)
	}
	m.secondsSinceLastBlock = secondsSinceLastBlock

	newBlockLatency, err := meter.Float64Histogram(metricsNewBlockLatency,
		metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 1.5, 3, 6, 12),
		otelapi.WithDescription(metricDescriptions[metricsNewBlockLatency]),
		otelapi.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("%w: %w: %s",
			ErrSetupMetricsFailed, err, metricsNewBlockLatency,
		)
	}
	m.newBlockLatency = newBlockLatency

	if _, err := meter.RegisterCallback(observe,
		m.secondsSinceLastBlock,
	); err != nil {
		return err
	}

	return nil
}
