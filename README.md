# Pulse2MQTT

Pulse2MQTT reads a Tibber Pulse on the local network and publishes meter
readings and diagnostics to MQTT. Home Assistant MQTT Discovery can be enabled
as an optional MQTT feature, but the application has no Home Assistant runtime
dependency.

## Standalone Usage

Copy `settings.default.yaml` to `settings.yaml`, configure the Pulse and MQTT
connection, and start the binary:

```shell
./pulse2mqtt
```

Set `PULSE2MQTT_CONFIG` to use an explicit configuration file:

```shell
PULSE2MQTT_CONFIG=/etc/pulse2mqtt/settings.yaml ./pulse2mqtt
```

## Docker

Published images are available from GHCR:

```shell
docker run --rm \
  --name pulse2mqtt \
  --volume ./settings.yaml:/config/settings.yaml:ro \
  ghcr.io/snislab/pulse2mqtt:latest
```

Versioned tags such as `ghcr.io/snislab/pulse2mqtt:0.4.0` are recommended for
reproducible deployments.

## Home Assistant

Home Assistant packaging is maintained separately in
[SnisLab/app-pulse2mqtt](https://github.com/SnisLab/app-pulse2mqtt) and is
distributed through
[SnisLab/home-assistant-apps](https://github.com/SnisLab/home-assistant-apps).
Application releases automatically notify the packaging repository after the
binary, GitHub release and container image have been published successfully.

## Releases

Push a semantic version tag such as `v0.4.0`. The release workflow runs tests,
creates GitHub release assets and Debian packages, publishes the multi-platform
GHCR image, and dispatches the immutable release metadata to
`app-pulse2mqtt`.
