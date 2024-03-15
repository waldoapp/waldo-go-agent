# Changelog

All notable changes to this project will be documented in this file. The format
is based on [Keep a Changelog].

This project adheres to [Semantic Versioning].

## [Unreleased]

## [2.3.1] - 2024-03-15

### Changed

- Fix encoding of upload errors.

## [2.3.0] - 2024-02-26

### Added

- Added more upload metadata for Waldo Core API.

## [2.2.0] - 2024-02-22

### Added

- Added support for Waldo Core API.

### Changed

- Upgraded to use Go 1.22.

## [2.1.3] - 2023-12-14

### Changed

- Upgraded to use Go 1.21.

## [2.1.2] - 2023-09-01

### Changed

- Added retry count to upload query parameters.

## [2.1.1] - 2023-08-31

### Changed

- Explicitly removed timeouts from HTTP client.

## [2.1.0] - 2023-01-30

### Added

- Added new `--git_commit` options to `trigger` verb.
- Added support for network retry.

## [2.0.2] - 2022-11-10

### Changed

- Removed support for uploading `.ipa` builds.

## [2.0.1] - 2022-06-10

### Fixed

- Fixed erroneous reporting of git pull request branch and commit info when
  running on Xcode Cloud.

## [2.0.0] - 2022-04-26

Initial public release of rewritten agent.

[Unreleased]:   https://github.com/waldoapp/waldo-go-agent/compare/2.3.1...HEAD
[2.3.1]:        https://github.com/waldoapp/waldo-go-agent/compare/2.3.0...2.3.1
[2.3.0]:        https://github.com/waldoapp/waldo-go-agent/compare/2.2.0...2.3.0
[2.2.0]:        https://github.com/waldoapp/waldo-go-agent/compare/2.1.3...2.2.0
[2.1.3]:        https://github.com/waldoapp/waldo-go-agent/compare/2.1.2...2.1.3
[2.1.2]:        https://github.com/waldoapp/waldo-go-agent/compare/2.1.1...2.1.2
[2.1.1]:        https://github.com/waldoapp/waldo-go-agent/compare/2.1.0...2.1.1
[2.1.0]:        https://github.com/waldoapp/waldo-go-agent/compare/2.0.2...2.1.0
[2.0.2]:        https://github.com/waldoapp/waldo-go-agent/compare/2.0.1...2.0.2
[2.0.1]:        https://github.com/waldoapp/waldo-go-agent/compare/2.0.0...2.0.1
[2.0.0]:        https://github.com/waldoapp/waldo-go-agent/compare/1a5f9ae...2.0.0

[Keep a Changelog]:     https://keepachangelog.com
[Semantic Versioning]:  https://semver.org
