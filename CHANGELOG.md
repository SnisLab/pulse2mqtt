# Changelog

## 0.5.0

- Add startup validation for MQTT and Pulse configuration.
- Support selecting the configuration file with `PULSE2MQTT_CONFIG`.
- Return data and metrics results explicitly instead of storing them globally.
- Handle invalid SML data without panicking.
- Shut down background workers cleanly on SIGINT and SIGTERM.
- Add release automation with one immutable release per push.
- Trigger and verify the Home Assistant app release after the main release.
- Add unit tests for data parsing, HTTP handling, configuration, MQTT messages and shutdown behavior.

## 0.2.3

- Update dependencies and Go tooling.

## 0.2.2

- Replace the internal logger with `DjSni/go-log`.

## 0.2.0

- Improve Debian install and removal scripts.
- Improve GoReleaser packaging.

## 0.1.5

- Improve Debian package installation handling.

## 0.1.4

- Update the SML dependency.

## 0.1.3

- Update the application release version.

## 0.1.2

- Update the application release version.

## 0.1.1

- Update the application release version.

## 0.1.0

- Initial Pulse2MQTT release.
