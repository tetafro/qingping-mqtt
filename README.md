# QingPing MQTT

[![Codecov](https://codecov.io/gh/tetafro/qingping-mqtt/branch/master/graph/badge.svg)](https://codecov.io/gh/tetafro/qingping-mqtt)
[![Go Report](https://goreportcard.com/badge/github.com/tetafro/qingping-mqtt)](https://goreportcard.com/report/github.com/tetafro/qingping-mqtt)
[![CI](https://github.com/tetafro/qingping-mqtt/actions/workflows/push.yml/badge.svg)](https://github.com/tetafro/qingping-mqtt/actions)

Expose metrics from Qingping Lite air quality monitor.

MQTT broker + Prometheus exporter.

## Build and run

```sh
make build run
```

Run in Docker
```sh
make docker
docker run --rm -it -p 8080:8080 -p 1883:1883 \
    ghcr.io/tetafro/qingping-mqtt
```

Check metrics
```sh
curl http://localhost:8080/metrics
```

## Links

- [Device overview](https://www.qingping.co/air-monitor-lite/overview)
- [QingPing developer docs](https://developer.qingping.co/introductions)
