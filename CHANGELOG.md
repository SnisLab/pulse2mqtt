# Changelog

## 0.5.3.17

### Fixed

- fix: handle Pulse SML transport framing (2e30c60)


## 0.5.3.16

### Fixed

- fix: parse modern Pulse SML payloads (4043788)


## 0.5.3.15

### Added

- feat: detect Pulse API at startup (336c490)

### Fixed

- fix: raise Pulse probe failures (2fabf2e)
- fix: cache detected Tibber endpoints (bb8d5ad)
- fix: support new Tibber Pulse endpoints (2892f24)
- fix: keep changelog heading at top (f4bce01)


## 0.5.1.14

### Fixed

- fix: preserve pulse power scaling (8389b59)

## 0.5.0.13

- ci: update changelog during releases (29a927b)
- docs: document 0.4 release history (f3b7758)
- docs: complete release history (d106879)
- docs: add project changelog (6de514e)

## 0.5.0

- Add startup validation for MQTT and Pulse configuration.
- Support selecting the configuration file with `PULSE2MQTT_CONFIG`.
- Return data and metrics results explicitly instead of storing them globally.
- Handle invalid SML data without panicking.
- Shut down background workers cleanly on SIGINT and SIGTERM.
- Add release automation with one immutable release per push.
- Trigger and verify the Home Assistant app release after the main release.
- Add unit tests for data parsing, HTTP handling, configuration, MQTT messages and shutdown behavior.

## 0.4.5

- Shut down background workers cleanly on SIGINT and SIGTERM.
- Add retry-safe release handling and verify published release assets.
- Add Actionlint and race detection to continuous integration.
- Support explicit configuration paths and validate battery profiles.
- Improve MQTT connection option and Home Assistant discovery tests.

## 0.4.4

- Add tests for MQTT data and metrics payloads.
- Test disconnected MQTT clients without requiring a broker.

## 0.4.3

- Add HTTP tests for Pulse authentication, node selection and error responses.

## 0.4.2

- Handle invalid meter data safely without panicking.

## 0.4.1

- Validate the configuration before starting MQTT and worker processes.
- Report missing or invalid configuration files clearly.

## 0.4.0

- Split Home Assistant packaging from the Pulse2MQTT application repository.
- Publish the Home Assistant app using a versioned application image.
- Add selectable AA battery profiles and an estimated battery entity.

## 0.3.1

- Fix Home Assistant app startup.

## 0.3.0

- Add the Home Assistant app and MQTT Discovery integration.

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
