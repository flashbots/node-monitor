package server

import (
	"context"
	"net/http"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/flashbots/node-monitor/logutils"
	"github.com/flashbots/node-monitor/state"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

func (s *Server) handleEventEthNewHeader(
	ctx context.Context, id string, ts time.Time, header *ethtypes.Header,
) {
	l := logutils.LoggerFromContext(ctx)

	s.state.UpdateHighestBlockIfNeeded(header.Number, ts)
	s.state.ExecutionEndpoint(id).UpdateHighestBlockIfNeeded(header.Number, ts)

	latency := ts.Sub(s.state.HighestBlockTime())

	s.metrics.newBlockLatency.Record(ctx,
		latency.Seconds(),
		metric.WithAttributes(attribute.KeyValue{
			Key:   "instance_name",
			Value: attribute.StringValue(id),
		}),
	)

	l.Info("Received new header",
		zap.String("block", header.Number.String()),
		zap.String("id", id),
		zap.Duration("latency", latency),
	)
}

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleEventPrometheusObserve(_ context.Context, o metric.Observer) error {
	o.ObserveFloat64(
		s.metrics.secondsSinceLastBlock,
		time.Since(s.state.HighestBlockTime()).Seconds(),

		metric.WithAttributes(
			attribute.KeyValue{
				Key:   "instance_name",
				Value: attribute.StringValue("__global"),
			},
		),
	)

	s.state.IterateExecutionEndpoints(func(id string, ee *state.ExecutionEndpoint) {
		o.ObserveFloat64(
			s.metrics.secondsSinceLastBlock,
			time.Since(ee.HighestBlockTime()).Seconds(),

			metric.WithAttributes(
				attribute.KeyValue{
					Key:   "instance_name",
					Value: attribute.StringValue(id),
				},
			),
		)
	})

	return nil
}
