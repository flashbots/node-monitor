package subscriber

import (
	"context"
	"errors"
	"math/rand"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/flashbots/node-monitor/config"
	"github.com/flashbots/node-monitor/logutils"
	"go.uber.org/zap"
)

type ELEndpoint struct {
	group string
	name  string

	resubInterval time.Duration
	uri           string

	client       *ethclient.Client
	subscription ethereum.Subscription

	done    chan struct{}
	headers chan *ethtypes.Header

	handler func(ctx context.Context, gname, ename string, ts time.Time, header *ethtypes.Header)
	ticker  *time.Ticker
}

var (
	ErrAlreadySubscribed = errors.New("already subscribed")
)

func NewELEndpoint(cfg *config.Config, group, name, uri string) (
	*ELEndpoint, error,
) {
	parsed, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	return &ELEndpoint{
		group: group,
		name:  name,

		resubInterval: cfg.Eth.ResubscribeInterval,
		uri:           parsed.String(),

		done:    make(chan struct{}),
		headers: make(chan *ethtypes.Header),
	}, nil
}

func (e *ELEndpoint) Name() string {
	return e.name
}

func (e *ELEndpoint) Group() string {
	return e.group
}

func (e *ELEndpoint) URI() string {
	return e.uri
}

func (e *ELEndpoint) IsSubscribed() bool {
	return e.client != nil && e.subscription != nil
}

func (e *ELEndpoint) Subscribe(
	ctx context.Context,
	handler func(ctx context.Context, group, name string, ts time.Time, header *ethtypes.Header),
) {
	if e.handler != nil {
		panic("must never happen: double subscription attempt")
	}
	e.handler = handler

	go e.run(ctx)
}

func (e *ELEndpoint) Unsubscribe() {
	e.done <- struct{}{}
}

func (e *ELEndpoint) subscribe(ctx context.Context) (success bool) {
	if e.IsSubscribed() {
		panic("must never happen: double subscription attempt")
	}

	l := logutils.LoggerFromContext(ctx)

	if e.client == nil {
		client, err := ethclient.Dial(e.uri)
		if err != nil {
			l.Error("Failed to connect to execution endpoint websocket",
				zap.String("endpoint_group", e.group),
				zap.String("endpoint_name", e.name),
				zap.Error(err),
			)
			return false
		}
		l.Debug("Connected to execution endpoint websocket",
			zap.String("endpoint_group", e.group),
			zap.String("endpoint_name", e.name),
		)
		e.client = client
	}

	if e.subscription == nil {
		subscription, err := e.client.SubscribeNewHead(ctx, e.headers)
		if err != nil {
			l.Error("Failed to subscribe to new headers",
				zap.String("endpoint_group", e.group),
				zap.String("endpoint_name", e.name),
				zap.Error(err),
			)
			return false
		}
		l.Info("Subscribed to execution endpoint's new headers",
			zap.String("endpoint_group", e.group),
			zap.String("endpoint_name", e.name),
		)
		e.subscription = subscription
	}

	return true
}

func (e *ELEndpoint) run(ctx context.Context) {
	l := logutils.LoggerFromContext(ctx)

	for {
		if !e.IsSubscribed() {
			// +/- 10% jitter
			intInterval := int64(e.resubInterval)
			interval := time.Duration(
				intInterval + rand.Int63n(intInterval/5) - intInterval/10,
			).Round(time.Millisecond)
			e.ticker = time.NewTicker(interval)

			l.Info("Will (re-)subscribe to execution endpoint",
				zap.Float64("delay_sec", interval.Seconds()),
				zap.String("endpoint_group", e.group),
				zap.String("endpoint_name", e.name),
			)

			// (re-)subscription loop
		loopResubscribe:
			for {
				select {
				case <-e.ticker.C:
					if e.subscribe(ctx) {
						e.ticker.Stop()
						e.ticker = nil
						break loopResubscribe
					}

				case <-e.done:
					l.Debug("Stopping (re-)subscription loop",
						zap.String("endpoint_group", e.group),
						zap.String("endpoint_name", e.name),
					)
					e.ticker.Stop()
					e.ticker = nil
					return
				}
			}
		}

		// event loop
	loopEvent:
		for {
			select {
			case header := <-e.headers:
				l.Debug("Got header",
					zap.Any("header", header),
					zap.String("endpoint_group", e.group),
					zap.String("endpoint_name", e.name),
				)
				e.handler(ctx, e.group, e.name, time.Now(), header)

			case err := <-e.subscription.Err():
				l.Warn("Execution endpoint subscription error",
					zap.String("endpoint_group", e.group),
					zap.String("endpoint_name", e.name),
					zap.Error(err),
				)
				e.subscription.Unsubscribe()
				e.subscription = nil
				break loopEvent

			case <-e.done:
				l.Debug("Stopping execution endpoint subscriber",
					zap.String("endpoint_group", e.group),
					zap.String("endpoint_name", e.name),
				)
				e.subscription.Unsubscribe()
				return
			}
		}
	}
}
