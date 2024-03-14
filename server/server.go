package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flashbots/node-monitor/config"
	"github.com/flashbots/node-monitor/httplogger"
	"github.com/flashbots/node-monitor/logutils"
	"github.com/flashbots/node-monitor/prometheus"
	"github.com/flashbots/node-monitor/state"
	"github.com/flashbots/node-monitor/subscriber"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	cfg   *config.Config
	log   *zap.Logger
	meter otelapi.Meter

	metrics *metrics
	state   *state.State

	subs map[string]*subscriber.ExecutionEndpoint
}

var (
	ErrExecutionEndpointDuplicateId       = errors.New("duplicate execution endpoint id")
	ErrExecutionEndpointFailedToSubscribe = errors.New("failed to subscribe to execution endpoint ws rpc")
	ErrExecutionEndpointFailedToRegister  = errors.New("failed to register execution endpoint")
	ErrPrometheusFailedToCreateMeter      = errors.New("failed to create prometheus meter")
	ErrPrometheusFailedToSetupMetrics     = errors.New("failed to setup prometheus metrics")
)

func New(cfg *config.Config) (*Server, error) {
	l := zap.L()
	ctx := logutils.ContextWithLogger(context.Background(), l)

	meter, err := prometheus.NewMeter(ctx, cfg.Server.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: %w",
			ErrPrometheusFailedToCreateMeter, err,
		)
	}

	state := state.New()
	subs := make(map[string]*subscriber.ExecutionEndpoint, len(cfg.Eth.ExecutionEndpoints))
	for _, rpc := range cfg.Eth.ExecutionEndpoints {
		parts := strings.Split(rpc, "=")
		id := parts[0]
		uri := parts[1]
		if _, exists := subs[id]; exists {
			return nil, fmt.Errorf("%w: %s",
				ErrExecutionEndpointDuplicateId, id,
			)
		}
		sub, err := subscriber.NewExecutionEndpoint(cfg, id, uri)
		if err != nil {
			return nil, fmt.Errorf("%w: %w",
				ErrExecutionEndpointFailedToSubscribe, err,
			)
		}
		subs[id] = sub
		if err := state.RegisterExecutionEndpoint(id); err != nil {
			return nil, fmt.Errorf("%w: %w",
				ErrExecutionEndpointFailedToRegister, err,
			)
		}
	}

	return &Server{
		cfg:   cfg,
		log:   l,
		meter: meter,

		metrics: &metrics{},
		state:   state,

		subs: subs,
	}, nil
}

func (s *Server) Run() error {
	l := s.log
	ctx := logutils.ContextWithLogger(context.Background(), l)

	if err := s.metrics.setup(s.meter, s.handleEventPrometheusObserve); err != nil {
		return fmt.Errorf("%w: %w",
			ErrPrometheusFailedToSetupMetrics, err,
		)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleHealthcheck)
	mux.Handle("/metrics", promhttp.Handler())
	handler := httplogger.Middleware(l, mux)

	srv := &http.Server{
		Addr:              s.cfg.Server.ListenAddress,
		Handler:           handler,
		MaxHeaderBytes:    1024,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	go func() {
		terminator := make(chan os.Signal, 1)
		signal.Notify(terminator, os.Interrupt, syscall.SIGTERM)
		stop := <-terminator

		l.Info("Stop signal received; shutting down...", zap.String("signal", stop.String()))

		for _, sub := range s.subs {
			sub.Unsubscribe()
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			l.Error("HTTP server shutdown failed",
				zap.Error(err),
			)
		}
	}()

	for _, sub := range s.subs {
		if err := sub.Subscribe(ctx, s.handleEventEthNewHeader); err != nil {
			return fmt.Errorf("%w: %w",
				ErrExecutionEndpointFailedToSubscribe, err,
			)
		}
	}

	l.Info("Starting up the monitor server...",
		zap.String("server_listen_address", s.cfg.Server.ListenAddress),
	)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Error("Monitor server failed", zap.Error(err))
	}
	l.Info("Monitor server is down")

	return nil
}
