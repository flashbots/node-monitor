package server

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/metric"
	otelapi "go.opentelemetry.io/otel/metric"
)

const (
	metricHighestBlock       = "highest_block"
	metricNewBlockLatency    = "new_block_latency"
	metricTimeSinceLastBlock = "time_since_last_block"
)

var (
	metricDescriptions = map[string]string{
		metricHighestBlock:       "The highest known block",
		metricNewBlockLatency:    "Statistics on how late a node receives blocks compared to the earliest observed ones",
		metricTimeSinceLastBlock: "Time passed since last block was received",
	}
)

var (
	ErrSetupMetricsFailed = errors.New("failed to setup metrics")
)

type metrics struct {
	highestBlock       otelapi.Int64ObservableGauge
	newBlockLatency    otelapi.Float64Histogram
	timeSinceLastBlock otelapi.Float64Observable
}

func (m *metrics) setup(meter otelapi.Meter, observe func(ctx context.Context, o metric.Observer) error) error {
	// highest block
	highestBlock, err := meter.Int64ObservableGauge(metricHighestBlock,
		otelapi.WithDescription(metricDescriptions[metricHighestBlock]),
	)
	if err != nil {
		return fmt.Errorf("%w: %w: %s",
			ErrSetupMetricsFailed, err, metricHighestBlock,
		)
	}
	m.highestBlock = highestBlock

	// new block latency
	newBlockLatency, err := meter.Float64Histogram(metricNewBlockLatency,
		metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 1.5, 3, 6, 12),
		otelapi.WithDescription(metricDescriptions[metricNewBlockLatency]),
		otelapi.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("%w: %w: %s",
			ErrSetupMetricsFailed, err, metricNewBlockLatency,
		)
	}
	m.newBlockLatency = newBlockLatency

	// time since last block
	timeSinceLastBlock, err := meter.Float64ObservableGauge(metricTimeSinceLastBlock,
		otelapi.WithDescription(metricDescriptions[metricTimeSinceLastBlock]),
		otelapi.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("%w: %w: %s",
			ErrSetupMetricsFailed, err, metricTimeSinceLastBlock,
		)
	}
	m.timeSinceLastBlock = timeSinceLastBlock

	// observables
	if _, err := meter.RegisterCallback(observe,
		m.highestBlock,
		m.timeSinceLastBlock,
	); err != nil {
		return err
	}

	return nil
}
