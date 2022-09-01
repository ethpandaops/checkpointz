# Checkpointz

A beacon chain aware Checkpoint Sync provider.

Checkpointz exists to reduce the operational burden of running a checkpoint sync endpoint. Checkpointz only serves a subset of the [beacon APIs](https://ethereum.github.io/beacon-APIs/#/) that are required for all consensus clients to checkpoint sync.

> :warning: **Checkpointz is still in heavy development** - use with caution

## Features
- Resource reduction
  - Adds HTTP cache-control headers depending on the content
- DOS protection
  - Never routes an incoming request directly to an upstream beacon node
  - Caches all requests
- Support for multiple upstream beacon nodes
  - Only serves a new finalized epoch once 50%+ of upstream beacon nodes agree
- Extensive Prometheus metrics

## Future features
- Web UI
  - Public-facing: shows information about the state of the provider, along with all state roots that the instance is aware of for cross-checking against instances.
  - Internal: shows information about the internal instance, health checks, etc.

## What is checkpoint sync?
Checkpoint sync is an operation that lets fresh beacon nodes jump to the head of the chain by fetching the state from a trusted & synced beacon node. 

More info: https://notes.ethereum.org/sWeLohipS9GdgMugYn9VkQ
## Usage
Checkpointz requires a config file. An example file can be found [here](https://github.com/samcm/checkpointz/blob/master/example_config.yaml).

```
Checkpoint sync provider for Ethereum beacon nodes

Usage:
  checkpointz [flags]

Flags:
      --config string   config file (default is config.yaml) (default "config.yaml")
  -h, --help            help for checkpointz
```

## Configuration

Checkpointz relies entirely on a single `yaml` config file.

| Name | Default | Description |
| --- | --- | --- |
| global.listenAddr | `:5555` | The address the main http server will listen on |
| global.logging | `warn` | Log level (`panic`, `fatal`, `warn`, `info`, `debug`, `trace`) |
| global.metricsAddr | `:9090` | The address the metrics server will listen on |
| checkpointz.caches.blocks.max_items | `200` | Controls the amount of "block" items that can be stored by Checkpointz (minimum 3) |
| checkpointz.caches.states.max_items | `5` | Controls the amount of "state" items that can be stored by Checkpointz (minimum 3). These starts a very large and this value will directly relate to memory usage. Anything higher than 10 is not recommended |
| checkpointz.historical_epoch_count | `20` | Controls the amount of historical epoch boundaries that Checkpointz will fetch and serve. |
| checkpointz.frontend.enabled | `true` | if the frontend should be enabled |
| checkpointz.frontend.brand_image_url |  | The brand logo to display on the frontend |
| checkpointz.frontend.brand_name | `Checkpointz` | The name of the brand to display on the frontend |
| checkpointz.frontend.public_url |  | The public URL of where the frontend will be served from |
| beacon.upstreams[].name |  | Shown in the frontend |
| beacon.upstreams[].address |  | The address of your beacon node. Note: NOT shown in the frontend |
| beacon.upstreams[].dataProvider |  | If true, Checkpointz will use this instance to fetch beacon blocks/state. If false, will only be used for finality checkpoints |

### Simple example

```
# use defaults and add a single beacon upstream node

beacon:
  upstreams:
  - name: remote
    address: http://localhost:5052
    dataProvider: true
```

### Disabled frontend

```
checkpointz:
  frontend:
    enabled: false

beacon:
  upstreams:
  - name: remote
    address: http://localhost:5052
    dataProvider: true
```

### Full example

```
global:
  # The address the main http server will listen on
  listenAddr: ":5555"
  # Log level (panic, fatal, warn, info, debug, trace)
  logging: "debug"
  # The address the metrics server will listen on
  metricsAddr: ":9090"

checkpointz:
  caches:
    blocks:
      # Controls the amount of "block" items that can be stored by Checkpointz (minimum 3)
      max_items: 200
    states:
      # Controls the amount of "state" items that can be stored by Checkpointz (minimum 3)
      # These starts a very large and this value will directly relate to memory usage. Anything higher than 
      # 10 is not recommended.
      max_items: 5
  historical_epoch_count: 20 # Controls the amount of historical epoch boundaries that Checkpointz will fetch and serve.
  frontend:
    # if the frontend should be enabled
    enabled: true
    # The brand logo to display on the frontend (optional)
    # brand_image_url: https://www.cdn.com/logo.png
    # The name of the brand to display on the frontend (optional)
    # brand_name: Brandname
    # The public URL of where the frontend will be served from (optional)
    # public_url: https://www.domain.com

beacon:
  # Upstreams configures the upstream beacon nodes to use.
  upstreams:
    # Shown in the frontend
  - name: remote
    # The address of your beacon node. Note: NOT shown in the frontend
    address: http://localhost:5052
    # If true, Checkpointz will use this instance to fetch beacon blocks/state. If false, will only be used for finality checkpoints.
    dataProvider: true
```

## Getting Started

### Download a release
Download the latest release from the [Releases page](https://github.com/samcm/checkpointz/releases). Extract and run with:
```
./checkpointz --config your-config.yaml
```

### Frontend

A basic frontend is provided in this project in [`./web`](https://github.com/samcm/checkpointz/blob/master/example_config.yaml) directory which needs to be built before it can be served by the server, eg. `http://localhost:5555`.

The frontend can be built with the following command;
```bash
# install node modules and build
make build-web
```

Building frontend requires `npm` and `NodeJS` to be installed.

### Docker
Available as a docker image at [samcm/checkpointz](https://hub.docker.com/r/samcm/checkpointz/tags)
#### Images
- `latest` - distroless, multiarch
- `latest-debian` - debian, multiarch
- `$version` - distroless, multiarch, pinned to a release (i.e. `0.4.0`)
- `$version-debian` - debian, multiarch, pinned to a release (i.e. `0.4.0-debian`)

**Quick start**
```
docker run -d  --name checkpointz -v $HOST_DIR_CHANGE_ME/config.yaml:/opt/checkpointz/config.yaml -p 9090:9090 -p 5555:5555 -it samcm/checkpointz:latest --config /opt/checkpointz/config.yaml;
docker logs -f checkpointz;
```


### Kubernetes via Helm
[Read more](https://github.com/skylenet/ethereum-helm-charts/tree/master/charts/checkpointz)
```
helm repo add ethereum-helm-charts https://skylenet.github.io/ethereum-helm-charts

helm install checkpointz ethereum-helm-charts/checkpointz -f your_values.yaml
```


**Building yourself (requires Go)**

1. Clone the repo
   ```sh
   go get github.com/samcm/checkpointz
   ```
2. Change directories
   ```sh
   cd ./checkpointz
   ```
3. Build the binary
   ```sh  
    go build -o checkpointz .
   ```
4. Run the exporter
   ```sh  
    ./checkpointz
   ```

## Contributing

Contributions are greatly appreciated! Pull requests will be reviewed and merged promptly if you're interested in improving the exporter! 

1. Fork the project
2. Create your feature branch:
    - `git checkout -b feat/new-metric-profit`
3. Commit your changes:
    - `git commit -m 'feat(profit): Export new metric: profit`
4. Push to the branch:
    -`git push origin feat/new-metric-profit`
5. Open a pull request

## Contact

Sam - [@samcmau](https://twitter.com/samcmau)
