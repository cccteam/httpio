# Changelog

## [0.4.1](https://github.com/cccteam/httpio/compare/v0.4.0...v0.4.1) (2024-09-23)


### Features

* Enhance read performance using tee reader ([#61](https://github.com/cccteam/httpio/issues/61)) ([5d2c58e](https://github.com/cccteam/httpio/commit/5d2c58e33a54fafcbc9c0dd23a0519327e8aa7b6))
* Implement columnset package ([#61](https://github.com/cccteam/httpio/issues/61)) ([5d2c58e](https://github.com/cccteam/httpio/commit/5d2c58e33a54fafcbc9c0dd23a0519327e8aa7b6))


### Bug Fixes

* Fix bug in Decoder permission checking ([#61](https://github.com/cccteam/httpio/issues/61)) ([5d2c58e](https://github.com/cccteam/httpio/commit/5d2c58e33a54fafcbc9c0dd23a0519327e8aa7b6))

## [0.4.0](https://github.com/cccteam/httpio/compare/v0.3.1...v0.4.0) (2024-09-17)


### ⚠ BREAKING CHANGES

* Fix Enforcer interface caused by breaking change in implementation package ([#60](https://github.com/cccteam/httpio/issues/60))

### Features

* Match Column sort to Struct field order. ([#58](https://github.com/cccteam/httpio/issues/58)) ([568bf2b](https://github.com/cccteam/httpio/commit/568bf2b84f18280c990fd72e228ad67d40c7f584))


### Bug Fixes

* Fix Enforcer interface caused by breaking change in implementation package ([#60](https://github.com/cccteam/httpio/issues/60)) ([be1efc4](https://github.com/cccteam/httpio/commit/be1efc449544a6717f356926a51508b8751cde44))


### Code Refactoring

* Move resourceset package to this repository ([#60](https://github.com/cccteam/httpio/issues/60)) ([be1efc4](https://github.com/cccteam/httpio/commit/be1efc449544a6717f356926a51508b8751cde44))
* Remove patcher to a different repository ([#60](https://github.com/cccteam/httpio/issues/60)) ([be1efc4](https://github.com/cccteam/httpio/commit/be1efc449544a6717f356926a51508b8751cde44))
* Rename patching package to patchset ([#60](https://github.com/cccteam/httpio/issues/60)) ([be1efc4](https://github.com/cccteam/httpio/commit/be1efc449544a6717f356926a51508b8751cde44))

## [0.3.1](https://github.com/cccteam/httpio/compare/v0.3.0...v0.3.1) (2024-09-11)


### Features

* **patching:** Implement Column rendering for patchset ([#56](https://github.com/cccteam/httpio/issues/56)) ([9bedabd](https://github.com/cccteam/httpio/commit/9bedabd62b3a923c2f5a6655534e51efc187d3e3))
* **patching:** Implement Primary Key handling ([#56](https://github.com/cccteam/httpio/issues/56)) ([9bedabd](https://github.com/cccteam/httpio/commit/9bedabd62b3a923c2f5a6655534e51efc187d3e3))

## [0.3.0](https://github.com/cccteam/httpio/compare/v0.2.5...v0.3.0) (2024-08-30)


### ⚠ BREAKING CHANGES

* Replaced Decoder with a new Generic Decoder that supports PatchSets and permission enforcement (50)

### Features

* Add Log function ([#49](https://github.com/cccteam/httpio/issues/49)) ([4d6dfda](https://github.com/cccteam/httpio/commit/4d6dfdad92a8e33ab67987c53c1d058505ec7ac7))
* Implement a new generic Decoder (50) ([c9a38bd](https://github.com/cccteam/httpio/commit/c9a38bd0fe856c0cc1fe0c0256eba35ae4fa3ac2))
* Replaced Decoder with a new Generic Decoder that supports PatchSets and permission enforcement (50) ([c9a38bd](https://github.com/cccteam/httpio/commit/c9a38bd0fe856c0cc1fe0c0256eba35ae4fa3ac2))


### Bug Fixes

* Fix bugs in Diff method of the Patcher ([#53](https://github.com/cccteam/httpio/issues/53)) ([4baa421](https://github.com/cccteam/httpio/commit/4baa421a0efc5b4c75238eb9eb8f36fd2b36c6d3))

## [0.2.5](https://github.com/cccteam/httpio/compare/v0.2.4...v0.2.5) (2024-07-03)


### Code Upgrade

* Update Go version to 1.22.5 to address GO-2024-2963 ([#47](https://github.com/cccteam/httpio/issues/47)) ([531f9c9](https://github.com/cccteam/httpio/commit/531f9c93aea98158deeef60cc361a6889ee32e00))
* Update workflows and add semantic pull request workflow (39) ([e3cd9d8](https://github.com/cccteam/httpio/commit/e3cd9d8216fb0cc1e211670405f6089a611edd77))

## [0.2.4](https://github.com/cccteam/httpio/compare/v0.2.3...v0.2.4) (2024-06-10)


### Code Upgrade

* Go version 1.22.3 and dependencies ([#37](https://github.com/cccteam/httpio/issues/37)) ([3ae6b17](https://github.com/cccteam/httpio/commit/3ae6b174343d38a13d9a2e411ed9b29ba806d197))
* Go version 1.22.4 for vulnerability GO-2024-2887 ([#43](https://github.com/cccteam/httpio/issues/43)) ([33ef042](https://github.com/cccteam/httpio/commit/33ef042f727f84ad94fe777f3b21685880663756))

## [0.2.3](https://github.com/cccteam/httpio/compare/v0.2.2...v0.2.3) (2024-04-05)


### Code Upgrade

* Upgrade to go1.22.2 and x/net to v0.24 (fix vulnerabilities) ([#33](https://github.com/cccteam/httpio/issues/33)) ([03ec4bb](https://github.com/cccteam/httpio/commit/03ec4bbbf06ff25a4678550cb6cedc0e8de289a7))

## [0.2.2](https://github.com/cccteam/httpio/compare/v0.2.1...v0.2.2) (2024-03-30)


### Bug Fixes

* Bug in response when no message is specified ([#28](https://github.com/cccteam/httpio/issues/28)) ([0f2172e](https://github.com/cccteam/httpio/commit/0f2172ec726d01caa5ada6a8d6e1ed40da34f709))

## [0.2.1](https://github.com/cccteam/httpio/compare/v0.2.0...v0.2.1) (2024-03-07)


### Code Upgrade

* Go version 1.22.1 and dependencies ([#26](https://github.com/cccteam/httpio/issues/26)) ([5321011](https://github.com/cccteam/httpio/commit/53210113b126bf8778b29ef85832edf712930863))

## [0.2.0](https://github.com/cccteam/httpio/compare/v0.1.1...v0.2.0) (2024-02-24)


### ⚠ BREAKING CHANGES

* Encoder methods were refactored ([#21](https://github.com/cccteam/httpio/issues/21))

### Features

* Add client error message handling ([#21](https://github.com/cccteam/httpio/issues/21)) ([64b2edb](https://github.com/cccteam/httpio/commit/64b2edb7de7ae9b2b1a3a07df01cfc1d8ec81e6d))


### Code Refactoring

* Encoder methods were refactored ([#21](https://github.com/cccteam/httpio/issues/21)) ([64b2edb](https://github.com/cccteam/httpio/commit/64b2edb7de7ae9b2b1a3a07df01cfc1d8ec81e6d))

## [0.1.1](https://github.com/cccteam/httpio/compare/v0.1.0...v0.1.1) (2023-11-28)


### Features

* Add support for encoding.TextUnmarshaler interface ([#17](https://github.com/cccteam/httpio/issues/17)) ([8ca0b51](https://github.com/cccteam/httpio/commit/8ca0b51652f6887c70751296f9fd3076b9cdebfc))

## [0.1.0](https://github.com/cccteam/httpio/compare/v0.0.2...v0.1.0) (2023-08-08)


### ⚠ BREAKING CHANGES

* Change the parameter order to align with other packages with similar api

### Features

* Add UUID Parameter support ([#13](https://github.com/cccteam/httpio/issues/13)) ([6e880fc](https://github.com/cccteam/httpio/commit/6e880fc72ac958b41c3ea1e9f8676aeccf97eec9))


### Code Refactoring

* Change the parameter order to align with other packages with similar api ([92469f6](https://github.com/cccteam/httpio/commit/92469f6abd451b92a10a3bc51dc235cf5daf85df))

## [0.0.2](https://github.com/cccteam/httpio/compare/v0.0.1...v0.0.2) (2023-07-07)


### Features

* chi UrlParam type parsing ([#10](https://github.com/cccteam/httpio/issues/10)) ([c2ba993](https://github.com/cccteam/httpio/commit/c2ba9931905d3e9894b9c63821aaf39e696d69fd))

## [0.0.1](https://github.com/cccteam/httpio/compare/v0.0.2...v0.0.1) (2023-05-25)


### Features

* Add additional unit tests and change decoder interface ([#5](https://github.com/cccteam/httpio/issues/5)) ([3f16dc5](https://github.com/cccteam/httpio/commit/3f16dc5c19168790261a8ccfaaf4118b310c4219))


### Continuous Integration

* Add missing manifest file ([63c78c2](https://github.com/cccteam/httpio/commit/63c78c20b2d88d15343af8865f3fe9da316bb9f7))

## [0.0.2](https://github.com/cccteam/httpio/compare/httpio-v0.0.1...httpio-v0.0.2) (2023-05-24)


### Features

* Add additional unit tests and change decoder interface ([#5](https://github.com/cccteam/httpio/issues/5)) ([3f16dc5](https://github.com/cccteam/httpio/commit/3f16dc5c19168790261a8ccfaaf4118b310c4219))

## 0.0.1 (2023-05-19)


### Continuous Integration

* Add missing manifest file ([63c78c2](https://github.com/cccteam/httpio/commit/63c78c20b2d88d15343af8865f3fe9da316bb9f7))
