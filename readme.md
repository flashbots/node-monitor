# node-monitor

A monitor that subscribes to several ethereum nodes, keeps track of the block
headers received by them, and reports the cumulative statistics via prometheus.

## TL;DR

```shell
export NODE_MONITOR_ETH_EXT_EL_ENDPOINTS=infura=wss://mainnet.infura.io/ws/v3/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

node-monitor serve --server-geth-ws-rpc local=127.0.0.1:8546
```

```shell
curl -sS 127.0.0.1:8080/metrics | grep node_monitor
```

```text
# HELP node_monitor_new_block_latency_seconds Statistics on how late a node receives blocks compared to the earliest observed ones
# TYPE node_monitor_new_block_latency_seconds histogram
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="0.005"} 0
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="0.01"} 0
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="0.025"} 0
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="0.05"} 0
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="0.075"} 0
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="0.1"} 0
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="0.25"} 0
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="0.5"} 0
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="0.75"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="1"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="1.5"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="3"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="6"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="12"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version="",le="+Inf"} 3
node_monitor_new_block_latency_seconds_sum{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version=""} 1.728523375
node_monitor_new_block_latency_seconds_count{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version=""} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="0.005"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="0.01"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="0.025"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="0.05"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="0.075"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="0.1"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="0.25"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="0.5"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="0.75"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="1"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="1.5"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="3"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="6"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="12"} 3
node_monitor_new_block_latency_seconds_bucket{instance_name="local",otel_scope_name="node-monitor",otel_scope_version="",le="+Inf"} 3
node_monitor_new_block_latency_seconds_sum{instance_name="local",otel_scope_name="node-monitor",otel_scope_version=""} 0
node_monitor_new_block_latency_seconds_count{instance_name="local",otel_scope_name="node-monitor",otel_scope_version=""} 3
# HELP node_monitor_time_since_last_block_seconds Time passed since last block was received
# TYPE node_monitor_time_since_last_block_seconds gauge
node_monitor_time_since_last_block_seconds{instance_name="__global",otel_scope_name="node-monitor",otel_scope_version=""} 6.0159025
node_monitor_time_since_last_block_seconds{instance_name="infura",otel_scope_name="node-monitor",otel_scope_version=""} 5.50802725
node_monitor_time_since_last_block_seconds{instance_name="local",otel_scope_name="node-monitor",otel_scope_version=""} 6.015933375
```
