# AirGradient-Prometheus exporter

Server to receive POST requests from AirGradient and re-serve them as Prometheus metrics. Data are stored in redis from the AirGradient unit(s) and metrics can be accessed on a per-target basis if needed.

# Build

## Docker
```
docker build -t localhost/promag:latest -f Containerfile
```

## Go binary
```
go mod download
go build -o promag .
```

# Run
## Docker
### Compose
This will set up redis and promag servers

```
docker compose -f compose.yml up -d
```

To stop:

```
docker compose -f compose.yml down
```

### Standalone

```
docker run -dt docker.io/library/redis:latest --name redis
docker run -dt localhost/promag:latest --port 9000:9000 --name promag
```

## Standalone
### Manually
The binary expects a redis instance running on `localhost:6379` with no authentication. Use the environment variables `REDIS_HOST`, `REDIS_PASSWORD`, `REDIS_PORT` and `REDIS_DB` to adjust your configuration. For example:

```
REDIS_HOST=my.redis.host.local REDIS_DB=1 ./promag
```

### systemd
The following is a systemd unit to run `promag`:

```ini
[Unit]
Description=AirGradient Prometheus Exporter
Documentation=https://github.com/pwcazenave/promag
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
# Uncomment if you wish to point promag at your redis host
Environment="REDIS_HOST=my.redis.host.local"
User=promag
Group=promag
ExecStart=/usr/bin/promag
Restart=always

[Install]
WantedBy=multi-user.target
```

# Environment variables
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
    metrics_path: /probe
    scrape_interval: 10s
    static_configs:
      - targets:
        - '123456'  # the unique CHIP_ID of the AirGradient sensor
        - '789012'  # add as many as you have AirGradient sensors
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: 'airgradient-exporter.local'  # the DNS entry for the airgradient exporter
```

The `targets` should contain the unique `CHIP_ID`(s) of your AirGradient sensor(s). This is the last six digits of the wireless access point that is created by the AirGradient firmware when it's being configured.

# AirGradient sketch
In the AirGradient sketch ([]`DIY_BASIC.ino`](https://github.com/airgradienthq/arduino/blob/master/examples/DIY_BASIC/DIY_BASIC.ino)), set the `APIROOT` to the location where this exporter is running (e.g. http://airgradient-exporter.local), then follow the instructions as per [the documentation](https://www.airgradient.com/open-airgradient/instructions/diy-v4/#software).
