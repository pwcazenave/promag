# AirGradient-Prometheus shim

Server to receive POST requests from AirGradient and re-serve them as Prometheus metrics. Data are stored in a redis DB from the AirGradient unit(s) and metrics can be accesses on a per-target basis if needed.

# Build
docker build -t localhost/promag:latest -f Containerfile

# Run
## Compose
This will set up redis and promag servers

```
docker compose -f compose.yml up -d
```

To stop:

```
docker compose -f compose.yml down
```

## Standalone

```
docker run -dt docker.io/library/redis:latest --name redis
docker run -dt localhost/promag:latest --port 9000:9000 --name promag
```

## Environment variables for promag
- `REDIS_HOST`: defaults to `localhost`
- `REDIS_PASSWORD`: defaults to `""`
- `REDIS_PORT`: defaults to `6379`
- `REDIS_DB`: defaults to `0`
- `PROM_HOST`: defaults to `""`
- `PROM_PORT`: defaults to `9000`

# Prometheus
Configure your prometheus server to scrape the metrics as below:

```yaml
scrape_configs:
  - job_name: 'airgradient'
    metrics_path: /probe?target=airgradientid
    scrape_interval: 30s
    static_configs:
      - targets:
        - 'container-host-ip:9000'
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9115  # The promag exporter's real hostname:port
```
