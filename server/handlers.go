package server

import (
	"context"
	"math/big"
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

	ee := s.state.ExecutionEndpoint(id)
	name, namespace := ee.Name()

	s.state.UpdateHighestBlockIfNeeded(namespace, header.Number, ts)
	ee.UpdateHighestBlockIfNeeded(header.Number, ts)

	latency := ts.Sub(s.state.HighestBlockTime(namespace))
	var attrs []attribute.KeyValue
	if namespace != "" {
		attrs = []attribute.KeyValue{
			{Key: "name", Value: attribute.StringValue(name)},
			{Key: "namespace", Value: attribute.StringValue(namespace)},
		}
	} else {
		attrs = []attribute.KeyValue{
			{Key: "name", Value: attribute.StringValue(name)},
		}
	}

	s.metrics.newBlockLatency.Record(ctx,
		latency.Seconds(),
		metric.WithAttributes(attrs...),
	)

	l.Debug("Received new header",
		zap.String("block", header.Number.String()),
		zap.String("id", id),
		zap.Duration("latency", latency),
	)
}

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleEventPrometheusObserve(_ context.Context, o metric.Observer) error {
	s.state.IterateNamespaces(func(namespace string, highestBlock *big.Int, highestBlockTime time.Time) {
		var attrs []attribute.KeyValue
		if namespace != "" {
			attrs = []attribute.KeyValue{
				{Key: "namespace", Value: attribute.StringValue(namespace)},
			}
		} else {
			attrs = []attribute.KeyValue{}
		}

		o.ObserveFloat64(
			s.metrics.secondsSinceLastBlock,
			time.Since(highestBlockTime).Seconds(),
			metric.WithAttributes(attrs...),
		)
	})

	s.state.IterateExecutionEndpoints(func(id string, ee *state.ExecutionEndpoint) {
		name, namespace := ee.Name()
		var attrs []attribute.KeyValue
		if namespace != "" {
			attrs = []attribute.KeyValue{
				{Key: "name", Value: attribute.StringValue(name)},
				{Key: "namespace", Value: attribute.StringValue(namespace)},
			}
		} else {
			attrs = []attribute.KeyValue{
				{Key: "name", Value: attribute.StringValue(name)},
			}
		}

		o.ObserveFloat64(
			s.metrics.secondsSinceLastBlock,
			time.Since(ee.HighestBlockTime()).Seconds(),
			metric.WithAttributes(attrs...),
		)
	})

	return nil
}
