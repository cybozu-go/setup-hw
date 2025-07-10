# Change Log

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [1.22.1] - 2025-07-10

### Fixed

- racadm command does not exist [#141](https://github.com/cybozu-go/setup-hw/pull/141)

## [1.22.0] - 2025-07-07

### Changed

- Update dependencies in [#139](https://github.com/cybozu-go/setup-hw/pull/139)
    - Update go module dependencies
    - Update OSMA to v11.1.0.0
    - Downgrade Ubuntu to 22.04

## [1.21.0] - 2025-06-12

### Changed

- Update for new F/W in [#135](https://github.com/cybozu-go/setup-hw/pull/135)
    - Update DELL F/W ISO-Image 25.06.02

## [1.20.2] - 2025-04-09

### Changed

- Update dependencies in [#133](https://github.com/cybozu-go/setup-hw/pull/133)
    - Update go module dependencies

## [1.20.1] - 2025-02-19

### Changed

- Fix bug of auto accepting EULA ([#130](https://github.com/cybozu-go/setup-hw/pull/130))

## [1.20.0] - 2025-02-14

### Added

- Add auto accepting EULA of DELL ([#128](https://github.com/cybozu-go/setup-hw/pull/128))

## [1.19.0] - 2025-01-20

### Added

- Support RedFish 1.20.1 ([#126](https://github.com/cybozu-go/setup-hw/pull/126))
- Installation procedure document ([#126](https://github.com/cybozu-go/setup-hw/pull/126))

## [1.18.0] - 2024-12-05

### Added

- Support R6615, R7615 ([#119](https://github.com/cybozu-go/setup-hw/pull/119))

## [1.17.2] - 2024-11-28

### Changed

- Update dependencies ([#116](https://github.com/cybozu-go/setup-hw/pull/116))
  - Update go module dependencies
  - Update ubuntu version in Dockerfile
- Fix release workflow ([#118](https://github.com/cybozu-go/setup-hw/pull/118))

## [1.17.1] - 2024-07-19

### Changed

- Update release procedure to use gh command ([#114](https://github.com/cybozu-go/setup-hw/pull/114))
- Update dependencies ([#113](https://github.com/cybozu-go/setup-hw/pull/113))
  - Update actions/setup-go
  - Update go module dependencies

## [1.17.0] - 2024-06-13

### Added

- Support Redfish API v1.18.0 ([#111](https://github.com/cybozu-go/setup-hw/pull/111))

### Changed

- Bump golang.org/x/net from 0.20.0 to 0.23.0 ([#108](https://github.com/cybozu-go/setup-hw/pull/108))
- Merge base rules ([#110](https://github.com/cybozu-go/setup-hw/pull/110))

## [1.16.1] - 2024-03-15

### Changed

- Update dependencies ([#105](https://github.com/cybozu-go/setup-hw/pull/105))
  - Update Dockerfile
  - Update go module dependencies

## [1.16.0] - 2024-01-24

### Breaking Changes

#### Migrate image registry

We migrated the image repository of setup-hw to `ghcr.io`.
From setup-hw v1.16.0, please use the following image.

- https://github.com/cybozu-go/setup-hw/pkgs/container/setup-hw

The [quay.io/cybozu/setup-hw](https://quay.io/repository/cybozu/setup-hw) will not be updated in the future.

### Changed

- Migrate to ghcr.io (#103)

## [1.15.2] - 2024-01-12

### Changed

- Remove section for executing resetDell periodically ([#101](https://github.com/cybozu-go/setup-hw/pull/101))

## [1.15.1] - 2023-11-29

### Changed

- Update go module dependencies and actions ([#99](https://github.com/cybozu-go/setup-hw/pull/99))

## [1.15.0] - 2023-11-02

### Added

- Add "repair" BMC user in [#96](https://github.com/cybozu-go/setup-hw/pull/96)
    - **Breaking change:** This user is mandatory for iDRAC.

## [1.14.3] - 2023-08-15

### Changed

- Disable "iDRAC path through I/F" in [#93](https://github.com/cybozu-go/setup-hw/pull/93)
  - Factory setting has been changed that iDRAC path through I/F enabled, but it gives problem to neco setup.

## [1.14.2] - 2023-07-24

### Changed

- Update dependencies in [#92](https://github.com/cybozu-go/setup-hw/pull/92)
   - Add ubuntu package curl, Ubuntu22.04 base image does not have curl package

## [1.14.1] - 2023-07-20

### Changed

- Update dependencies in [#90](https://github.com/cybozu-go/setup-hw/pull/90)
   - Add ubuntu package kmod

## [1.14.0] - 2023-07-19

### Changed

- Update dependencies in [#88](https://github.com/cybozu-go/setup-hw/pull/88)
    - Change Container image Ubuntu:22.04 from CentOS:7
    - Upgrade OpenManage Server Administrator v11.0.0.0 from v10.2.0.0
    - Add rules for Redfish version 1.17.0
    - Add test case of actual collected data from iDRAC9 / Redfish v1.17.0

## [1.13.2] - 2023-02-24

### Changed

- Update dependencies in [#84](https://github.com/cybozu-go/setup-hw/pull/84)
    - Upgrade direct dependencies in go.mod
    - Update testing/releasing environments
- Generate statically linked binaries in [#84](https://github.com/cybozu-go/setup-hw/pull/84)

## [1.13.1] - 2023-01-13

### Changed

- Revert "Change system profile setting to `Performance`" ([#82](https://github.com/cybozu-go/setup-hw/pull/82))

## [1.13.0] - 2022-12-19

### Changed

- Change system profile setting to `Performance` ([#80](https://github.com/cybozu-go/setup-hw/pull/80))

## [1.12.1] - 2022-10-26

### Changed

- Update dependencies ([#78](https://github.com/cybozu-go/setup-hw/pull/78))
    - Upgrade direct dependencies in go.mod
    - Update Golang to 1.19

## [1.12.0] - 2022-04-21

### Changed

- Change NPS settings for R6525 ([#73](https://github.com/cybozu-go/setup-hw/pull/73))

## [1.11.0] - 2022-04-15

### Changed

- Update go module dependencies and actions ([#71](https://github.com/cybozu-go/setup-hw/pull/71))

## [1.10.1] - 2022-02-09

### Changed

- Disable iDRAC.WebServer.HostHeaderCheck ([#69](https://github.com/cybozu-go/setup-hw/pull/69))

## [1.10.0] - 2022-01-04

### Changed

- Update dependencies ([#67](https://github.com/cybozu-go/setup-hw/pull/67))

### Added

- Add BIOS settings for new equipment ([#65](https://github.com/cybozu-go/setup-hw/pull/65))
- Add rules for Redfish version 1.11.0 ([#66](https://github.com/cybozu-go/setup-hw/pull/66))

## [1.9.2] - 2021-09-15

### Changed

- update golang to 1.17 ([#63](https://github.com/cybozu-go/setup-hw/pull/63))

## [1.9.1] - 2021-05-31

### Changed

- Update dependencies ([#61](https://github.com/cybozu-go/setup-hw/pull/61))

## [1.9.0] - 2021-05-20

### Added

- add support command for automated ISO reboot ([#57](https://github.com/cybozu-go/setup-hw/pull/57))

## [1.8.0] - 2021-05-18

### Added

- add support command for automated firmware update ([#55](https://github.com/cybozu-go/setup-hw/pull/55))

## [1.7.2] - 2021-01-29

### Changed
- update dependencies

## [1.7.1] - 2021-01-05

### Changed
- setup-hw: Wait for iDRAC to get ready ([#46](https://github.com/cybozu-go/setup-hw/pull/46))

## [1.7.0] - 2020-07-28

### Removed
- Purge settings for sshpkauth ([#42](https://github.com/cybozu-go/setup-hw/pull/42))

## [1.6.10] - 2020-06-29

### Added
- Reset iDRAC at startup ([#40](https://github.com/cybozu-go/setup-hw/pull/40))

## [1.6.9] - 2020-03-16

### Added
- Support Redfish version 1.6.0 ([#36](https://github.com/cybozu-go/setup-hw/pull/36))

## [1.6.8] - 2019-08-27

### Added
- monitor-hw: skip iDRAC reset if `no-reset` file exists ([#31](https://github.com/cybozu-go/setup-hw/pull/31))

### Changed
- setup-hw: hide raw passwords ([#32](https://github.com/cybozu-go/setup-hw/pull/32))

## [1.6.7] - 2019-08-19

### Changed
- Enable TPM 2.0 on Dell servers ([#29](https://github.com/cybozu-go/setup-hw/pull/29))

## [1.6.6] - 2019-07-16

### Changed
- Update Dell EMC System from 19.01.00 to 19.07.00 ([#25](https://github.com/cybozu-go/setup-hw/pull/25))

## [1.6.5] - 2019-06-25

### Added
- A utility to collect Redfish data ([#23](https://github.com/cybozu-go/setup-hw/pull/23))

### Changed
- Update the rule file for Dell 14G ([#24](https://github.com/cybozu-go/setup-hw/pull/24))

## [1.6.4] - 2019-06-11

### Changed
- Fix `monitor-hw` for Dell servers ([#21](https://github.com/cybozu-go/setup-hw/pull/21))

## [1.6.3] - 2019-06-10

### Changed
- Fix Redfish API version detection again ([#20](https://github.com/cybozu-go/setup-hw/pull/20))

## [1.6.2] - 2019-06-07

### Changed
- Fix Redfish API version detection ([#19](https://github.com/cybozu-go/setup-hw/pull/19))

## [1.6.1] - 2019-06-07

### Changed
- Fix connection leaks introduced in 1.6.0 ([#18](https://github.com/cybozu-go/setup-hw/pull/18))

## [1.6.0] - 2019-06-06

### Changed
- Introduce timeouts for RedFish traversal ([#16](https://github.com/cybozu-go/setup-hw/pull/16))
- Dynamically detect RedFish version ([#17](https://github.com/cybozu-go/setup-hw/pull/17))

## [1.5.0] - 2019-05-20

### Added
- Add default data for mock client ([#14](https://github.com/cybozu-go/setup-hw/pull/14))

## [1.4.0] - 2019-05-14

### Added
- Add mock client for QEMU ([#13](https://github.com/cybozu-go/setup-hw/pull/13))

## [1.3.0] - 2019-04-26

### Added
- Support Redfish version 1.2.0 and 1.4.0 ([#12](https://github.com/cybozu-go/setup-hw/pull/12))

[Unreleased]: https://github.com/cybozu-go/setup-hw/compare/v1.22.1...HEAD
[1.22.1]: https://github.com/cybozu-go/setup-hw/compare/v1.22.0...v1.22.1
[1.22.0]: https://github.com/cybozu-go/setup-hw/compare/v1.21.0...v1.22.0
[1.21.0]: https://github.com/cybozu-go/setup-hw/compare/v1.20.2...v1.21.0
[1.20.2]: https://github.com/cybozu-go/setup-hw/compare/v1.20.1...v1.20.2
[1.20.1]: https://github.com/cybozu-go/setup-hw/compare/v1.20.0...v1.20.1
[1.20.0]: https://github.com/cybozu-go/setup-hw/compare/v1.19.0...v1.20.0
[1.19.0]: https://github.com/cybozu-go/setup-hw/compare/v1.18.0...v1.19.0
[1.18.0]: https://github.com/cybozu-go/setup-hw/compare/v1.17.2...v1.18.0
[1.17.2]: https://github.com/cybozu-go/setup-hw/compare/v1.17.1...v1.17.2
[1.17.1]: https://github.com/cybozu-go/setup-hw/compare/v1.17.0...v1.17.1
[1.17.0]: https://github.com/cybozu-go/setup-hw/compare/v1.16.1...v1.17.0
[1.16.1]: https://github.com/cybozu-go/setup-hw/compare/v1.16.0...v1.16.1
[1.16.0]: https://github.com/cybozu-go/setup-hw/compare/v1.15.2...v1.16.0
[1.15.2]: https://github.com/cybozu-go/setup-hw/compare/v1.15.1...v1.15.2
[1.15.1]: https://github.com/cybozu-go/setup-hw/compare/v1.15.0...v1.15.1
[1.15.0]: https://github.com/cybozu-go/setup-hw/compare/v1.14.3...v1.15.0
[1.14.3]: https://github.com/cybozu-go/setup-hw/compare/v1.14.2...v1.14.3
[1.14.2]: https://github.com/cybozu-go/setup-hw/compare/v1.14.1...v1.14.2
[1.14.1]: https://github.com/cybozu-go/setup-hw/compare/v1.14.0...v1.14.1
[1.14.0]: https://github.com/cybozu-go/setup-hw/compare/v1.13.2...v1.14.0
[1.13.2]: https://github.com/cybozu-go/setup-hw/compare/v1.13.1...v1.13.2
[1.13.1]: https://github.com/cybozu-go/setup-hw/compare/v1.13.0...v1.13.1
[1.13.0]: https://github.com/cybozu-go/setup-hw/compare/v1.12.1...v1.13.0
[1.12.1]: https://github.com/cybozu-go/setup-hw/compare/v1.12.0...v1.12.1
[1.12.0]: https://github.com/cybozu-go/setup-hw/compare/v1.11.0...v1.12.0
[1.11.0]: https://github.com/cybozu-go/setup-hw/compare/v1.10.1...v1.11.0
[1.10.1]: https://github.com/cybozu-go/setup-hw/compare/v1.10.0...v1.10.1
[1.10.0]: https://github.com/cybozu-go/setup-hw/compare/v1.9.2...v1.10.0
[1.9.2]: https://github.com/cybozu-go/setup-hw/compare/v1.9.1...1.9.2
[1.9.1]: https://github.com/cybozu-go/setup-hw/compare/v1.9.0...1.9.1
[1.9.0]: https://github.com/cybozu-go/setup-hw/compare/v1.8.0...1.9.0
[1.8.0]: https://github.com/cybozu-go/setup-hw/compare/v1.7.2...1.8.0
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
