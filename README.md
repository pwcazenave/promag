# AirGradient-Prometheus shim

Server to receive POST requests from AirGradient and re-serve them as Prometheus metrics. Data are stored in a redis DB from the AirGradient unit(s) and metrics can be accesses on a per-target basis if needed.

# Build
podman build -t localhost/promag:latest -f Containerfile

# Run
## Compose
This will set up redis and promag servers

```
podman compose -f compose.yml
```

## Standalone

```
podman run -dt library.docker.io/redis:latest --name redis
podman run -dt localhost/promag:latest --name promag
```

## Environment variables for promag
- `REDIS_HOST`: defaults to `localhost`
- `REDIS_PASSWORD`: defaults to `""`
- `REDIS_PORT`: defaults to `6379`
- `REDIS_DB`: defaults to `0`
- `PROM_HOST`: defaults to `""`
- `PROM_PORT`: defaults to `9000`

