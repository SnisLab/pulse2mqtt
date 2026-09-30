# Changelog

## 0.6.4.31

### Added

- feat: publish Pulse bridge diagnostics (08f43e1)


## 0.6.2.30

### Fixed

- fix: remove redundant modern data retry (24ac84a)


## 0.6.1.29

### Added

- feat: salvage CRC-valid SML readings (e008d6d)

### Changed

- refactor: use strict SML transport parsing (31799cb)


## 0.5.9.28

### Fixed

- fix: keep modern data on HTTP polling (f119a3c)


## 0.5.8.27

### Added

- feat: use Pulse WebSocket data stream (dd88fee)

### Fixed

- fix: serialize Pulse HTTP requests (27aa774)


## 0.5.6.26

### Fixed

- fix: retry invalid modern Pulse data (6f80aa4)


## 0.5.5.25

- Maintenance release.


## 0.5.4.24

- Maintenance release.


## 0.5.4.23

### Changed

- refactor: separate legacy and modern Pulse paths (b229545)


## 0.5.3.22

- Maintenance release.


## 0.5.3.21

- Maintenance release.


## 0.5.3.20

### Fixed

- fix: parse modern Pulse timestamp fields (712344c)


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
