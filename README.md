# Pulse2MQTT

Pulse2MQTT reads a Tibber Pulse on the local network and publishes meter readings and diagnostics to MQTT.

## Standalone Configuration

Copy `settings.default.yaml` to `settings.yaml`, configure the Pulse and MQTT connection, and start `pulse2mqtt`. Home Assistant MQTT Discovery remains disabled unless explicitly enabled in the configuration.

## Releases

Every push and pull request targeting `main` runs the tests and validates the GoReleaser configuration. Pushes to `main` additionally publish either a stable or an edge release.

- If no tag exists for the version in `include/version/version.go`, the workflow creates the tag and publishes a stable GitHub release such as `v0.5.0`.
- Further pushes with the same base version update the rolling `edge` prerelease.
- Edge builds use versions such as `0.5.0.48-edge`, where `48` is the GitHub Actions run number.
- Changing the base version to `0.6.0` causes the next push to publish the new stable `v0.6.0` release.

Stable releases are permanent. The `edge` release and its assets are replaced on every edge build. Packages and binary archives use fixed asset names. For example:

```text
https://github.com/SnisLab/pulse2mqtt/releases/download/edge/pulse2mqtt_linux_amd64.apk
https://github.com/SnisLab/pulse2mqtt/releases/download/edge/pulse2mqtt_linux_arm64.apk
https://github.com/SnisLab/pulse2mqtt/releases/download/edge/pulse2mqtt_linux_armv7.apk
```

The APK contains the binary and `settings.default.yaml`. The DEB additionally installs the systemd service and its package scripts.

## Home Assistant App

Home Assistant packaging is maintained separately in
[SnisLab/app-pulse2mqtt](https://github.com/SnisLab/app-pulse2mqtt) and is
distributed through
[SnisLab/home-assistant-apps](https://github.com/SnisLab/home-assistant-apps).
Application releases automatically notify the packaging repository after the
binary, GitHub release and container image have been published successfully.
The release workflow sends a `release-built` dispatch to
`SnisLab/app-pulse2mqtt`; configure the `APP_PULSE2MQTT_TOKEN` Actions secret
with a token that can write repository dispatches in that repository.

## Battery Profiles

The Pulse reports only the combined voltage of its two AA cells. Pulse2MQTT can estimate a percentage with selectable profiles for alkaline, LFB AA and rechargeable NiMH cells. Select `regulated_1_5v` for rechargeable cells with a regulated 1.5 V output; this profile reports 80% while the combined output is at least 2.8 V and 0% below that threshold.
