# QingPing MQTT

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
