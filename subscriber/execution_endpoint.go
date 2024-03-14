package subscriber

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/flashbots/node-monitor/config"
	"github.com/flashbots/node-monitor/logutils"
	"go.uber.org/zap"
)

type ExecutionEndpoint struct {
	id                  string
	resubscribeInterval time.Duration
	uri                 string

	client       *ethclient.Client
	subscription ethereum.Subscription

	done    chan struct{}
	headers chan *ethtypes.Header

	handler func(context.Context, string, time.Time, *ethtypes.Header)
	ticker  *time.Ticker
}

var (
	ErrAlreadySubscribed = errors.New("already subscribed")
)

func NewExecutionEndpoint(cfg *config.Config, id, uri string) (
	*ExecutionEndpoint, error,
) {
	parsed, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "" {
		parsed.Scheme = "ws"
	}

	return &ExecutionEndpoint{
		id:                  id,
		resubscribeInterval: cfg.Eth.ResubscribeInterval,
		uri:                 parsed.String(),

		done:    make(chan struct{}),
		headers: make(chan *ethtypes.Header),
	}, nil
}

func (sub *ExecutionEndpoint) NodeID() string {
	return sub.id
}

func (sub *ExecutionEndpoint) URI() string {
	return sub.uri
}

func (sub *ExecutionEndpoint) IsSubscribed() bool {
	return sub.client != nil && sub.subscription != nil
}

func (sub *ExecutionEndpoint) Subscribe(
	ctx context.Context,
	handler func(ctx context.Context, nodeID string, ts time.Time, header *ethtypes.Header),
) error {
	if sub.handler != nil {
		return ErrAlreadySubscribed
	}
	sub.handler = handler

	sub.subscribe(ctx)

	go sub.run(ctx)

	return nil
}

func (sub *ExecutionEndpoint) Unsubscribe() {
	sub.done <- struct{}{}
}

func (sub *ExecutionEndpoint) subscribe(ctx context.Context) (success bool) {
	if sub.IsSubscribed() {
		panic("must never happen: double subscription attempt")
	}

	l := logutils.LoggerFromContext(ctx)
	if sub.client == nil {
		client, err := ethclient.Dial(sub.uri)
		if err != nil {
			l.Error("Failed to connect to execution endpoint websocket",
				zap.String("id", sub.id),
				zap.Error(err),
			)
			return false
		}
		l.Info("Connected to execution endpoint websocket",
			zap.String("id", sub.id),
		)
		sub.client = client
	}

	if sub.subscription == nil {
		subscription, err := sub.client.SubscribeNewHead(ctx, sub.headers)
		if err != nil {
			l.Error("Failed to subscribe to new headers",
				zap.String("id", sub.id),
				zap.Error(err),
			)
			return false
		}
		l.Info("Subscribed to execution endpoint's new headers",
			zap.String("id", sub.id),
		)
		sub.subscription = subscription
	}

	return true
}

func (sub *ExecutionEndpoint) run(ctx context.Context) {
	l := logutils.LoggerFromContext(ctx)

	for {
		if !sub.IsSubscribed() {
			sub.ticker = time.NewTicker(sub.resubscribeInterval)

			// (re-)subscription loop
		loopResubscribe:
			for {
				select {
				case <-sub.ticker.C:
					if sub.subscribe(ctx) {
						sub.ticker.Stop()
						sub.ticker = nil
						break loopResubscribe
					}

				case <-sub.done:
					l.Debug("Stopping (re-)subscription loop",
						zap.String("id", sub.id),
					)
					sub.ticker.Stop()
					return
				}
			}
		}

		// event loop
	loopEvent:
		for {
			select {
			case head := <-sub.headers:
				sub.handler(ctx, sub.id, time.Now(), head)

			case err := <-sub.subscription.Err():
				l.Warn("Execution endpoint subscription error",
					zap.String("id", sub.id),
					zap.Error(err),
				)
				sub.subscription.Unsubscribe()
				sub.subscription = nil
				break loopEvent

			case <-sub.done:
				l.Debug("Stopping execution endpoint subscriber",
					zap.String("id", sub.id),
				)
				sub.subscription.Unsubscribe()
				return
			}
		}
	}
}
