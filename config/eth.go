package config

import "time"

type Eth struct {
	ExecutionEndpoints         []string      `yaml:"execution_endpoints"`
	ExternalExecutionEndpoints []string      `yaml:"external_execution_endpoints"`
	ResubscribeInterval        time.Duration `yaml:"resubscribe_interval"`
}
