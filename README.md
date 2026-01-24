# QingPing MQTT

[![Codecov](https://codecov.io/gh/tetafro/qingping-mqtt/branch/master/graph/badge.svg)](https://codecov.io/gh/tetafro/qingping-mqtt)
[![Go Report](https://goreportcard.com/badge/github.com/tetafro/qingping-mqtt)](https://goreportcard.com/report/github.com/tetafro/qingping-mqtt)
[![CI](https://github.com/tetafro/qingping-mqtt/actions/workflows/push.yml/badge.svg)](https://github.com/tetafro/qingping-mqtt/actions)

Expose metrics from a [Qingping Lite air quality monitor](https://www.qingping.co/air-monitor-lite/overview).

MQTT broker + Prometheus exporter.

## Set up the device

To set up your device to send metrics to a remote MQTT broker follow the steps:

1. Create a QingPing account in the [Qingping+](https://play.google.com/store/apps/details?id=com.cleargrass.app.air) app.
1. Add your device to the app.
1. Log in with the account to the [developers portal](https://developer.qingping.co).
1. Go to "Private Configuration" and add a configuration:

    - Private Type: Self-built MQTT
    - Host/Port: listen address of `qingping-mqtt`
    - Up topic: `qingping/your-device-name/up`
    - Down topic: `qingping/your-device-name/down`

1. Go to "Push Configuration" and add the configuration from the previous step for
    your device.

## Run

Download and run a pre-built binary ([releases](https://github.com/tetafro/qingping-mqtt/releases))
```sh
./qingping-mqtt \
    -http-addr 0.0.0.0:8080 \
    -mqtt-addr 0.0.0.0:1883 \
```

Or run in Docker ([image tag](https://github.com/tetafro/qingping-mqtt/pkgs/container/qingping-mqtt))
```sh
docker run -d -p 8080:8080 -p 1883:1883 \
    ghcr.io/tetafro/qingping-mqtt
```

## Build from source

Binary
```sh
make build
```

Docker image
```sh
make docker
```
