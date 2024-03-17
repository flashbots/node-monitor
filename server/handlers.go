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
	defaultGroup  = "__default"
	groupEndpoint = "__group"

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
	latency := g.RegisterBlockAndGetLatency(block, ts)

	var attrs []attribute.KeyValue
	if gname != "" {
		attrs = []attribute.KeyValue{
			{Key: keyTargetEndpoint, Value: attribute.StringValue(ename)},
			{Key: keyTargetGroup, Value: attribute.StringValue(gname)},
			{Key: keyTargetID, Value: attribute.StringValue(utils.MakeELEndpointID(gname, ename))},
		}
	} else {
		attrs = []attribute.KeyValue{
			{Key: keyTargetEndpoint, Value: attribute.StringValue(ename)},
			{Key: keyTargetGroup, Value: attribute.StringValue(defaultGroup)},
			{Key: keyTargetID, Value: attribute.StringValue(utils.MakeELEndpointID(gname, ename))},
		}
	}

	s.metrics.newBlockLatency.Record(ctx,
		latency.Seconds(),
		metric.WithAttributes(attrs...),
	)

	l.Debug("Received new header",
		zap.String("block", block.String()),
		zap.String("endpoint_group", gname),
		zap.String("endpoint_name", ename),
		zap.Duration("latency_s", latency),
	)
}

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleEventPrometheusObserve(_ context.Context, o metric.Observer) error {
	s.state.IterateELGroupsRO(func(gname string, g *state.ELGroup) {
		var attrs []attribute.KeyValue
		if gname != "" {
			attrs = []attribute.KeyValue{
				{Key: keyTargetEndpoint, Value: attribute.StringValue(groupEndpoint)},
				{Key: keyTargetGroup, Value: attribute.StringValue(gname)},
			}
		} else {
			attrs = []attribute.KeyValue{
				{Key: keyTargetGroup, Value: attribute.StringValue(defaultGroup)},
				{Key: keyTargetEndpoint, Value: attribute.StringValue(groupEndpoint)},
			}
		}

		b, t := g.TimeSinceHighestBlock()

		// group's highest block
		o.ObserveInt64(s.metrics.highestBlock, b, metric.WithAttributes(attrs...))

		// group's time since last block
		o.ObserveFloat64(s.metrics.timeSinceLastBlock, t.Seconds(), metric.WithAttributes(attrs...))

		g.IterateEndpointsRO(func(ename string, e *state.ELEndpoint) {
			if gname != "" {
				attrs = []attribute.KeyValue{
					{Key: keyTargetEndpoint, Value: attribute.StringValue(ename)},
					{Key: keyTargetGroup, Value: attribute.StringValue(gname)},
					{Key: keyTargetID, Value: attribute.StringValue(utils.MakeELEndpointID(gname, ename))},
				}
			} else {
				attrs = []attribute.KeyValue{
					{Key: keyTargetEndpoint, Value: attribute.StringValue(ename)},
					{Key: keyTargetGroup, Value: attribute.StringValue(defaultGroup)},
					{Key: keyTargetID, Value: attribute.StringValue(utils.MakeELEndpointID(gname, ename))},
				}
			}

			b, t := e.TimeSinceHighestBlock()

			// endpoint's highest block
			o.ObserveInt64(s.metrics.highestBlock, b, metric.WithAttributes(attrs...))

			// endpoint's time since last block
			o.ObserveFloat64(s.metrics.timeSinceLastBlock, t.Seconds(), metric.WithAttributes(attrs...))
		})
	})

	return nil
}
