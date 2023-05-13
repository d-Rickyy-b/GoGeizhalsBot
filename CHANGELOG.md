# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
### Changed
### Fixed

## [2.2.0] - 2023-05-13
### Added
- Implement maxpriceagents config
- Implement httpmaxtries config
- Take user to price agents when too many already exist
- Implement http client timeout of 10 seconds
### Changed
- ci: update changelog reminder to more active action
- Update dependencies

## [2.1.2] - 2023-04-02
### Fixed
- Use proper count for 429 response codes in metrics

## [2.1.1] - 2023-04-02
### Fixed
- Only run a single background price updater

## [2.1.0] - 2023-04-01
### Added
- Add prometheus metrics for HTTP 429 status codes
- Make update interval for price agents configurable

### Changed
- Refactored price history code
- Changed package name to `github.com/d-Rickyy-b/gogeizhalsbot`
- Use go 1.18 as new minimum version

### Fixed
- wishlistURLPattern did not match certain urls
- implement current api of gotgbot v2.0.0-rc.15

## [2.0.1] - 2022-08-27
### Fixed
- Fixed a bug with Notifications not being sent 

## [2.0.0] - 2022-08-25
### Added
- Store prices per location per entity
- Allow skinflint.co.uk links
- Ability to disable price agents (no UI yet)
 
### Changed
- Display price agent name (& link) in price history graph caption
- Display currency in price history graph
- Send new message with price agent details when clicking on the button for the price notification message
- Ability to switch graph theme (dark/light)

### Fixed
- Improve graph visibility

## [1.1.0] - 2022-03-28

### Added
- Support for prometheus database 

## [1.0.0] - 2022-02-14
Initial release! First stable version of GoGeizhalsBot is published as v1.0.0 

[unreleased]: https://github.com/d-Rickyy-b/GoGeizhalsBot/compare/v2.2.0...HEAD
[2.2.0]: https://github.com/d-Rickyy-b/GoGeizhalsBot/compare/v2.1.2...v2.2.0
[2.1.2]: https://github.com/d-Rickyy-b/GoGeizhalsBot/compare/v2.1.1...v2.1.2
[2.1.1]: https://github.com/d-Rickyy-b/GoGeizhalsBot/compare/v2.1.0...v2.1.1
[2.1.0]: https://github.com/d-Rickyy-b/GoGeizhalsBot/compare/v2.0.1...v2.1.0
[2.0.1]: https://github.com/d-Rickyy-b/GoGeizhalsBot/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/d-Rickyy-b/GoGeizhalsBot/compare/v1.1.0...v2.0.0
[1.1.0]: https://github.com/d-Rickyy-b/GoGeizhalsBot/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/d-Rickyy-b/GoGeizhalsBot/tree/v1.0.0
