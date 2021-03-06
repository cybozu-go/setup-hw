# Change Log

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [1.7.2] - 2021-01-29

### Changed
- update dependencies

## [1.7.1] - 2021-01-05

### Changed
- setup-hw: Wait for iDRAC to get ready (#46)

## [1.7.0] - 2020-07-28

### Removed
- Purge settings for sshpkauth (#42)

## [1.6.10] - 2020-06-29

### Added
- Reset iDRAC at startup (#40)

## [1.6.9] - 2020-03-16

### Added
- Support Redfish version 1.6.0 (#36)

## [1.6.8] - 2019-08-27

### Added
- monitor-hw: skip iDRAC reset if `no-reset` file exists (#31)

### Changed
- setup-hw: hide raw passwords (#32)

## [1.6.7] - 2019-08-19

### Changed
- Enable TPM 2.0 on Dell servers (#29)

## [1.6.6] - 2019-07-16

### Changed
- Update Dell EMC System from 19.01.00 to 19.07.00 (#25)

## [1.6.5] - 2019-06-25

### Added
- A utility to collect Redfish data (#23)

### Changed
- Update the rule file for Dell 14G (#24)

## [1.6.4] - 2019-06-11

### Changed
- Fix `monitor-hw` for Dell servers (#21)

## [1.6.3] - 2019-06-10

### Changed
- Fix Redfish API version detection again (#20)

## [1.6.2] - 2019-06-07

### Changed
- Fix Redfish API version detection (#19)

## [1.6.1] - 2019-06-07

### Changed
- Fix connection leaks introduced in 1.6.0 (#18)

## [1.6.0] - 2019-06-06

### Changed
- Introduce timeouts for RedFish traversal (#16)
- Dynamically detect RedFish version (#17)

## [1.5.0] - 2019-05-20

### Added
- Add default data for mock client (#14)

## [1.4.0] - 2019-05-14

### Added
- Add mock client for QEMU (#13)

## [1.3.0] - 2019-04-26

### Added
- Support Redfish version 1.2.0 and 1.4.0 (#12)

[Unreleased]: https://github.com/cybozu-go/setup-hw/compare/v1.7.2...HEAD
[1.7.2]: https://github.com/cybozu-go/setup-hw/compare/v1.7.1...1.7.2
[1.7.1]: https://github.com/cybozu-go/setup-hw/compare/v1.7.0...1.7.1
[1.7.0]: https://github.com/cybozu-go/setup-hw/compare/v1.6.10...1.7.0
[1.6.10]: https://github.com/cybozu-go/setup-hw/compare/v1.6.9...v1.6.10
[1.6.9]: https://github.com/cybozu-go/setup-hw/compare/v1.6.8...v1.6.9
[1.6.8]: https://github.com/cybozu-go/setup-hw/compare/v1.6.7...v1.6.8
[1.6.7]: https://github.com/cybozu-go/setup-hw/compare/v1.6.6...v1.6.7
[1.6.6]: https://github.com/cybozu-go/setup-hw/compare/v1.6.5...v1.6.6
[1.6.5]: https://github.com/cybozu-go/setup-hw/compare/v1.6.4...v1.6.5
[1.6.4]: https://github.com/cybozu-go/setup-hw/compare/v1.6.3...v1.6.4
[1.6.3]: https://github.com/cybozu-go/setup-hw/compare/v1.6.2...v1.6.3
[1.6.2]: https://github.com/cybozu-go/setup-hw/compare/v1.6.1...v1.6.2
[1.6.1]: https://github.com/cybozu-go/setup-hw/compare/v1.6.0...v1.6.1
[1.6.0]: https://github.com/cybozu-go/setup-hw/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/cybozu-go/setup-hw/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/cybozu-go/setup-hw/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/cybozu-go/setup-hw/compare/e370989b320534a6af5b9b83d921f6312af40b42...v1.3.0
