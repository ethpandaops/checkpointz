# eth-proxy

Reverse proxy for ethereum nodes. Warning: This is very experimental.

### Features:
- Status endpoint for all your beacon nodes
- Beacon chain API path based allow list. Allows you to restrict which API endpoints you're exposing.
- Execution JSON RPC method allow list


## Endpoints

**Status**: Shows you the configured nodes and some additional information about them.

```sh
curl http://localhost:5555/status
```

Example response:

```json
{
    "beacon_nodes": {
        "node1": {
            "version": "Lighthouse/v2.3.2-rc.0-828d5bc/x86_64-linux",
            "syncing": {
                "head_slot": "366025",
                "sync_distance": "0",
                "is_syncing": false,
                "is_optimistic": false
            },
            "peer_count": {
                "disconnected": 561,
                "connected": 110
            },
            "last_check": 1658315108
        },
        "node2": {
            "version": "Lodestar/v0.38.0-dev.b5e24f7138",
            "syncing": {
                "head_slot": "366025",
                "sync_distance": "0",
                "is_syncing": false,
                "is_optimistic": false
            },
            "peer_count": {
                "disconnected": 0,
                "connected": 50
            },
            "last_check": 1658315108
        }
    },
    "execution_nodes": {
        "node1": {
            "head_block": 12630114,
            "chain_id": 3,
            "peer_count": 9,
            "is_syncing": false,
            "last_check": 1658315108
        },
        "node2": {
            "head_block": 12630114,
            "chain_id": 3,
            "peer_count": 12,
            "is_syncing": false,
            "last_check": 1658315108
        }
    }
}
```

Reverse proxy to a specific node by name
```sh
# Beacon HTTP API
curl -X GET 'http://localhost:5555/proxy/beacon/node1/eth/v1/node/identity'

# Execution JSON RPC API
curl -X POST 'http://localhost:5555/proxy/execution/node1/' \
     --header 'Content-Type: application/json' --data-raw '{
        "jsonrpc":"2.0",
        "method":"eth_blockNumber",
        "params":[],
        "id":1
}'
```


### Building and running

Adjust the configuration file for your needs. An example can be seen in [example_config.yaml](example_config.yaml)


```sh
go build -o bin/eth-proxy ./cmd/eth-proxy && ./bin/eth-proxy --config example_config.yaml
```
