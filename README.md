# Pulse2MQTT

Pulse2MQTT reads a Tibber Pulse on the local network and publishes meter readings and diagnostics to MQTT.

## Standalone Configuration

Copy `settings.default.yaml` to `settings.yaml`, configure the Pulse and MQTT connection, and start `pulse2mqtt`. Home Assistant MQTT Discovery remains disabled unless explicitly enabled in the configuration.

To use a configuration at a different path, set `PULSE2MQTT_CONFIG`:

```shell
PULSE2MQTT_CONFIG=/config/settings.yaml ./pulse2mqtt
```

## Releases

Every push and pull request targeting `main` runs the tests and validates the GoReleaser configuration. Every push to `main` additionally publishes an immutable release.

- Release versions consist of the base version from `include/version/version.go` and the GitHub Actions run number, for example `0.4.0.37`.
- Each release gets its own tag and GitHub release, for example `v0.4.0.37`.
- Changing the base version to `0.5.0` causes subsequent releases to use versions such as `0.5.0.38`.

Releases are permanent. Packages and binary archives use fixed asset names within each release. For example:

```text
https://github.com/SnisLab/pulse2mqtt/releases/download/v0.4.0.37/pulse2mqtt_linux_amd64.apk
https://github.com/SnisLab/pulse2mqtt/releases/download/v0.4.0.37/pulse2mqtt_linux_arm64.apk
https://github.com/SnisLab/pulse2mqtt/releases/download/v0.4.0.37/pulse2mqtt_linux_armv7.apk
```

The APK contains the binary and `settings.default.yaml`. The DEB additionally installs the systemd service and its package scripts.

## Home Assistant App

Home Assistant packaging is maintained separately in
[SnisLab/app-pulse2mqtt](https://github.com/SnisLab/app-pulse2mqtt) and is
distributed through
[SnisLab/home-assistant-apps](https://github.com/SnisLab/home-assistant-apps).
Releases automatically notify the packaging repository after
the binaries and GitHub release have been published successfully. The release
workflow sends a `release-built` dispatch to
`SnisLab/app-pulse2mqtt`; configure the `APP_PULSE2MQTT_TOKEN` Actions secret
with a token that can write repository dispatches in that repository.

## Battery Profiles

The Pulse reports only the combined voltage of its two AA cells. Pulse2MQTT can estimate a percentage with selectable profiles for alkaline, LFB AA and rechargeable NiMH cells. Select `regulated_1_5v` for rechargeable cells with a regulated 1.5 V output; this profile reports 80% while the combined output is at least 2.8 V and 0% below that threshold.
