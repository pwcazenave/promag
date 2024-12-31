- [AirGradient-Prometheus exporter](#airgradient-prometheus-exporter)
- [Build](#build)
  - [Docker](#docker)
  - [Go binary](#go-binary)
- [Run](#run)
  - [Docker](#docker-1)
    - [Compose](#compose)
    - [Standalone](#standalone)
  - [Standalone](#standalone-1)
    - [Manually](#manually)
    - [systemd](#systemd)
- [Environment variables](#environment-variables)
- [Prometheus](#prometheus)
- [AirGradient sketch](#airgradient-sketch)
- [Grafana](#grafana)
- [Home Assistant](#home-assistant)

# AirGradient-Prometheus exporter

Server to receive POST requests from AirGradient and re-serve them as Prometheus metrics. Data are stored in redis from the AirGradient unit(s) and metrics can be accessed on a per-target basis if needed.

Optionally allows data to be ingested to Home Assistant via the `rest` sensor.

# Build

## Docker
```bash
docker build -t localhost/promag:latest -f Dockerfile
```

## Go binary
```bash
go mod download
go build -o promag .
```

# Run
## Docker
### Compose
This will set up redis and promag servers

```bash
docker compose -f compose.yml up -d
```

To stop:

```bash
docker compose -f compose.yml down
```

### Standalone

```bash
docker run -dt docker.io/library/redis:latest --port 6379:6379 --name redis
docker run -dt localhost/promag:latest --port 9000:9000 --name promag
```

## Standalone
### Manually
The binary expects a redis instance running on `localhost:6379` with no authentication. Use the environment variables `REDIS_HOST`, `REDIS_PASSWORD`, `REDIS_PORT` and `REDIS_DB` to adjust your configuration. For example:

```bash
REDIS_HOST=my.redis.host.local REDIS_DB=1 ./promag
```

### systemd
Copy the binary to `/usr/bin/promag`. Create the following systemd unit to run `promag` as `/etc/systemd/system/promag.service`:

```ini
[Unit]
Description=AirGradient Prometheus Exporter
Documentation=https://github.com/pwcazenave/promag
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
# Uncomment if you wish to point promag at your redis host
#Environment="REDIS_HOST=my.redis.host.local"
User=promag
Group=promag
ExecStart=/usr/bin/promag
Restart=always

[Install]
WantedBy=multi-user.target
```

Create a user:

```bash
useradd --user-group promag
```

Enable and run the systemd unit:

```bash
systemctl enable --now promag.service
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
In the AirGradient sketch ([`BASIC.ino`](https://github.com/airgradienthq/arduino/blob/master/examples/BASIC/BASIC.ino)), set the `APIROOT` to the location where this exporter is running (e.g. http://airgradient-exporter.local), then follow the instructions as per [the documentation](https://www.airgradient.com/open-airgradient/instructions/diy-v4/#software).

# Grafana
Import the `grafana_dashboard.json` file into Grafana to visualise the AirGradient sensor CO2 and particulate data.

* Log in as admin to your Grafana instance
* Click the + in the corner
* Select Import Dashboard
* Paste the contents of `grafana_dashboard.json` into the box headed Import via dashboard JSON model
* Click Load

![image](https://github.com/pwcazenave/promag/assets/531784/1b094e63-1eaa-4006-83ff-1b3114ffa23f)

# Home Assistant
There is an additional endpoint (`/json`) which returns the AirGradient data more or less as it came from the AirGradient sensor POST request:

```json
{
  "wifi": -59,
  "rco2": 720,
  "pm02": 2,
  "atmp": 11.6,
  "rhum": 78
}
```

This can be used with a `rest` sensor in Home Assistant via `configuration.yml`:

```yaml
  - platform: rest
    resource: https://airgradient-exporter.local/json?target=123456
    name: Air Quality Sensor - CO2
    scan_interval: 60
    value_template: "{{ value_json.rco2 }}"
    unit_of_measurement: "ppm"
```
