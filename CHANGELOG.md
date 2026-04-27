
## [2.1.1](https://github.com/PlushGuardian/obfuscation-server-exporter/compare/v2.1.0...v2.1.1) (2026-04-27)


### Bug Fixes

* use GH_RELEASE_PAT in pipeline ([3195616](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/3195616393419554eb17f93c8db726f67d4510ec))
* use GH_RELEASE_PAT in pipeline instead of GITHUB_TOKEN to avoid authentication issues ([9de186e](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/9de186e32c2fb342189c8f1eff30728a2d8ab95d))

# [2.1.0](https://github.com/PlushGuardian/obfuscation-server-exporter/compare/v2.0.0...v2.1.0) (2026-04-27)


### Features

* add system collector with two metrics ([#19](https://github.com/PlushGuardian/obfuscation-server-exporter/issues/19)) ([597b025](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/597b025f4fc639215b4aaea279ecd0c3dcaefafd))


# v2.0.0

The main focus of this pull request was to prepare the repository for implementation of two additional exporters. Thus, the code was restructured, the exporter renamed, and API calls to external URL of 3X-UI panel removed.

### Added

- viper: added viper for more streamlined config and CLI management.

- Added logging

- Issue Templates: Introduced a feature request issue template located at .github/issue_templates/feature_request.yaml to streamline user feedback and suggestions

- .gitignore: Expanded the .gitignore file to cover a wider range of development artifacts, environment files, and local exporter binaries

### Changed

- Core Naming & Module: Rebranded the project from "x-ui-exporter" to "obfs-exporter", which involved renaming the Go module, all associated binaries, user contexts, and configuration paths

- API Endpoint Logic: Modified the API calls to target localhost instead of an external panel URL, requiring a fundamental rework of the configuration structure and related logic

- Configuration Overhaul: The monolithic CLI configuration struct has been deprecated. It is now replaced by a dedicated Config struct, broken down into OBFSExporterConfig and ThreeXUIConfig sections for clearer separation of concerns. Additionally, the configuration changed, and now the exporter settings are located under its name.

- go.mod & go.sum: Updated the module path and bumped many dependency versions to their latest releases, such as prometheus/client_golang (to v1.23.2) and golang.org/x/sys (to v0.35.0)

- Installation Script: The install.sh script was updated to reflect the new project name, binary names, user, and directory structures

- Dockerfile: The Dockerfile was modified to build and copy the new obfs-exporter binary and to update the container's entrypoint accordingly

### Removed

- Deprecated Code: The old api/api.go and config/config.go files and their associated config-example.yaml were completely removed

- Removed unused dependencies

### Fixed

- No specific bug fixes were explicitly mentioned in this pull request. The focus was on the architectural pivot and rebranding.
