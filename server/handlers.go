package server

import (
	"context"
	"net/http"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/flashbots/node-monitor/logutils"
	"github.com/flashbots/node-monitor/state"
	"github.com/flashbots/node-monitor/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

const (
	defaultTargetGroup   = "__default"
	groupVirtualEndpoint = "__group"

	keyTargetEndpoint = "node_monitor_target_endpoint"
	keyTargetGroup    = "node_monitor_target_group"
	keyTargetID       = "node_monitor_target_id"
)

func (s *Server) handleEventEthNewHeader(
	ctx context.Context,
	gname, ename string,
	ts time.Time,
	header *ethtypes.Header,
) {
	l := logutils.LoggerFromContext(ctx)

	block := header.Number

	g := s.state.ExecutionGroup(gname)
	e := g.Endpoint(ename)

	e.RegisterBlock(block, ts)
	latency_s := g.RegisterBlockAndGetLatency(block, ts).Seconds()

	attrs := []attribute.KeyValue{
		{Key: keyTargetEndpoint, Value: attribute.StringValue(ename)},
		{Key: keyTargetGroup, Value: attribute.StringValue(normalisedGroup(gname))},
		{Key: keyTargetID, Value: attribute.StringValue(utils.MakeELEndpointID(gname, ename))},
	}

	s.metrics.newBlockLatency.Record(ctx,
		latency_s,
		metric.WithAttributes(attrs...),
	)

	l.Debug("Received new header",
		zap.Float64("latency_s", latency_s),
		zap.String("block", block.String()),
		zap.String("endpoint_group", gname),
		zap.String("endpoint_name", ename),
	)
}

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleEventPrometheusObserve(_ context.Context, o metric.Observer) error {
	s.state.IterateELGroupsRO(func(gname string, g *state.ELGroup) {
		attrs := []attribute.KeyValue{
			{Key: keyTargetEndpoint, Value: attribute.StringValue(groupVirtualEndpoint)},
			{Key: keyTargetGroup, Value: attribute.StringValue(normalisedGroup(gname))},
		}

		blockGroup, tsBlockGroup := g.TimeSinceHighestBlock()

		// group's highest block
		o.ObserveInt64(s.metrics.highestBlock, blockGroup, metric.WithAttributes(attrs...))

		// group's time since last block
		o.ObserveFloat64(s.metrics.timeSinceLastBlock, tsBlockGroup.Seconds(), metric.WithAttributes(attrs...))

		g.IterateEndpointsRO(func(ename string, e *state.ELEndpoint) {
			attrs := []attribute.KeyValue{
				{Key: keyTargetEndpoint, Value: attribute.StringValue(ename)},
				{Key: keyTargetGroup, Value: attribute.StringValue(normalisedGroup(gname))},
				{Key: keyTargetID, Value: attribute.StringValue(utils.MakeELEndpointID(gname, ename))},
			}

			blockEndpoint, tsBlockEndpoint := e.TimeSinceHighestBlock()

			// endpoint's highest block
			o.ObserveInt64(s.metrics.highestBlock, blockEndpoint, metric.WithAttributes(attrs...))

			// endpoint's highest block lag
			var lag int64
			if blockGroup != 0 && blockEndpoint != 0 {
				lag = blockGroup - blockEndpoint
			}
			o.ObserveInt64(s.metrics.highestBlockLag, lag, metric.WithAttributes(attrs...))

			// endpoint's time since last block
			o.ObserveFloat64(s.metrics.timeSinceLastBlock, tsBlockEndpoint.Seconds(), metric.WithAttributes(attrs...))
		})
	})

	return nil
}

func normalisedGroup(gname string) string {
	if gname == "" {
		return defaultTargetGroup
	}
	return gname
}
