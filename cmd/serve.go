package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/flashbots/node-monitor/config"
	"github.com/flashbots/node-monitor/server"
	"github.com/flashbots/node-monitor/utils"
	"github.com/urfave/cli/v2"
)

const (
	categoryEth    = "ETHEREUM:"
	categoryServer = "SERVER:"
)

var (
	ErrUnexpectedExecutionEndpoint = errors.New("unexpected execution endpoint rpc (must look like `id=127.0.0.1:8546`)")
)

func CommandServe(cfg *config.Config) *cli.Command {
	executionEndpoints := &cli.StringSlice{}
	externalExecutionEndpoints := &cli.StringSlice{}

	ethFlags := []cli.Flag{
		&cli.StringSliceFlag{
			Category:    categoryEth,
			Destination: executionEndpoints,
			EnvVars:     []string{"NODE_MONITOR_ETH_EL_ENDPOINTS"},
			Name:        "eth-el-endpoint",
			Usage:       "eth execution endpoints (websocket) in the format of `[namespace:]id=hostname:port`",
		},

		&cli.StringSliceFlag{
			Category:    categoryEth,
			Destination: externalExecutionEndpoints,
			EnvVars:     []string{"NODE_MONITOR_ETH_EXT_EL_ENDPOINTS"},
			Name:        "eth-ext-el-endpoint",
			Usage:       "external eth execution endpoints (websocket) in the format of `[namespace:]id=hostname:port`",
		},

		&cli.DurationFlag{
			Category:    categoryEth,
			Destination: &cfg.Eth.ResubscribeInterval,
			EnvVars:     []string{"NODE_MONITOR_RESUBSCRIBE_INTERVAL"},
			Name:        "resubscribe-interval",
			Usage:       "an `interval` at which the monitor will try to (re-)subscribe to node events",
			Value:       15 * time.Second,
		},
	}

	serverFlags := []cli.Flag{
		&cli.StringFlag{
			Category:    categoryServer,
			Destination: &cfg.Server.ListenAddress,
			EnvVars:     []string{"NODE_MONITOR_LISTEN_ADDRESS"},
			Name:        "listen-address",
			Usage:       "`host:port` for the server to listen on",
			Value:       "0.0.0.0:8080",
		},

		&cli.StringFlag{
			Category:    categoryServer,
			Destination: &cfg.Server.Name,
			EnvVars:     []string{"NODE_MONITOR_SERVER_NAME"},
			Name:        "server-name",
			Usage:       "service `name` to report in prometheus metrics",
			Value:       "node-monitor",
		},
	}

	flags := slices.Concat(
		ethFlags,
		serverFlags,
	)

	return &cli.Command{
		Name:  "serve",
		Usage: "run the monitor server",
		Flags: flags,

		Before: func(ctx *cli.Context) error {
			executionEndpoints := slices.Concat(
				executionEndpoints.Value(),
				externalExecutionEndpoints.Value(),
			)
			for idx, ee := range executionEndpoints {
				ee = strings.TrimSpace(ee)
				parts := strings.Split(ee, "=")
				if len(parts) != 2 {
					return fmt.Errorf("%w: %s", ErrUnexpectedExecutionEndpoint, ee)
				}
				for idx, part := range parts {
					parts[idx] = strings.TrimSpace(part)
				}
				id := parts[0]
				uri := parts[1]
				parsed, err := utils.ParseRawURI(uri)
				if err != nil {
					return err
				}
				if parsed.Scheme == "" {
					parsed.Scheme = "ws"
				}
				executionEndpoints[idx] = fmt.Sprintf("%s=%s", id, parsed.String())
			}
			cfg.Eth.ExecutionEndpoints = executionEndpoints
			return nil
		},

		Action: func(_ *cli.Context) error {
			s, err := server.New(cfg)
			if err != nil {
				return err
			}
			return s.Run()
		},
	}
}
