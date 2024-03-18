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

	keyTargetName  = "node_monitor_target_name"
	keyTargetGroup = "node_monitor_target_group"
	keyTargetID    = "node_monitor_target_id"
)

func (s *Server) handleEventEthNewHeader(
	ctx context.Context,
	gname, ename string,
	ts time.Time,
	header *ethtypes.Header,
) {
	l := logutils.LoggerFromContext(ctx)

	block := header.Number
	blockStr := block.String()

	g := s.state.ExecutionGroup(gname)
	e := g.Endpoint(ename)

	e.RegisterBlock(block, ts)
	latency := g.RegisterBlockAndGetLatency(block, ts)
	latency_s := latency.Seconds()

	switch latency {
	case time.Duration(0):
		l.Info("New block timestamp",
			zap.String("block", blockStr),
			zap.String("endpoint_group", gname),
			zap.String("endpoint_name", ename),
			zap.Time("ts", ts),
		)
	default:
		l.Debug("Received new header",
			zap.Float64("latency_s", latency_s),
			zap.String("block", blockStr),
			zap.String("endpoint_group", gname),
			zap.String("endpoint_name", ename),
			zap.Time("ts", ts),
		)
	case state.Infinity:
		l.Info("Skipping reporting block-latency on a very late block",
			zap.String("block", blockStr),
			zap.String("endpoint_group", gname),
			zap.String("endpoint_name", ename),
		)
		// don't bias the histogram
		return
	}

	attrs := []attribute.KeyValue{
		{Key: keyTargetName, Value: attribute.StringValue(ename)},
		{Key: keyTargetGroup, Value: attribute.StringValue(normalisedGroup(gname))},
		{Key: keyTargetID, Value: attribute.StringValue(utils.MakeELEndpointID(gname, ename))},
	}
	s.metrics.newBlockLatency.Record(ctx,
		latency_s,
		metric.WithAttributes(attrs...),
	)
}

func (s *Server) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleEventPrometheusObserve(_ context.Context, o metric.Observer) error {
	s.state.IterateELGroupsRO(func(gname string, g *state.ELGroup) {
		// don't report groups that did't progress yet
		if g.HighestBlock().Sign() == 0 {
			return
		}

		attrs := []attribute.KeyValue{
			{Key: keyTargetName, Value: attribute.StringValue(groupVirtualEndpoint)},
			{Key: keyTargetGroup, Value: attribute.StringValue(normalisedGroup(gname))},
		}

		blockGroup, tsBlockGroup := g.TimeSinceHighestBlock()

		// group's highest block
		o.ObserveInt64(s.metrics.highestBlock, blockGroup, metric.WithAttributes(attrs...))

		// group's time since last block
		o.ObserveFloat64(s.metrics.timeSinceLastBlock, tsBlockGroup.Seconds(), metric.WithAttributes(attrs...))

		g.IterateEndpointsRO(func(ename string, e *state.ELEndpoint) {
			// don't report endpoints that did't progress yet
			if e.HighestBlock().Sign() == 0 {
				return
			}

			attrs := []attribute.KeyValue{
				{Key: keyTargetName, Value: attribute.StringValue(ename)},
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
