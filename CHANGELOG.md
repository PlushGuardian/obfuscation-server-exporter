## [2.2.0](https://github.com/PlushGuardian/obfuscation-server-exporter/compare/v2.1.2...v2.2.0) (2026-04-29)

* feat: add system metrics (#30)
  Detailed metrics added:

* system_swap_total_bytes and system_swap_used_bytes
* system_cpu_usage_percent
* system_load1, system_load5, system_load15
* system_disk_total_bytes and system_disk_used_bytes
* system_uptime_seconds
* aggregated network metrics: system_network_(receive|transmit)_(bytes|packets|errors)_total
* per-interface network metrics
* protocol statistics system_network_protocol_total
* Add new categories for changelog and make it pull from the commit message (#33)
  * docs: add new categories for changelog and make it pull from the commit message
* Add system metrics (#30)
  * feat: add swap metrics system_swap_total_bytes and system_swap_used_bytes

* feat: add cpu metric system_cpu_usage_percent

* feat: add load average metrics system_load*

* feat: add disk metrics system_disk_total_bytes and system_disk_used_bytes

* feat: add system_uptime_seconds metric

* feat: add aggregated network metrics system_network_(receive|transmit)_(bytes|packets|errors)_total

* feat: add per-interface network metrics system_network_interface_(receive|transmit)_(bytes|packets|errors)_total

* feat: add protocol statistics system_network_protocol_total
* fix variable name in ci
* fix variable name so that releases can be created properly (#29)
  * fix: fix variable name so that releases can be created properly

* fix variable name

* come comment under jobs

* fix uses file reference

* fix variable naming

* fix small errors

* change token

* add two variables to output

* add debug values

* fix naming

* test dry run

* print all output

* let's play a game of lying

* MORE TRICKING

* try the --no-verify flag

* try another option with --no-ci

* unset github vars and use --branches

* try bot slander

* try another refs branch

* Update reusable-semantic-release.yaml

* Update reusable-semantic-release.yaml

* Update ci-pull-request.yaml

* Update reusable-semantic-release.yaml

* Update reusable-semantic-release.yaml

* Update reusable-semantic-release.yaml

* Update ci-pull-request.yaml

* Update reusable-semantic-release.yaml

* Update reusable-semantic-release.yaml

* fix: make only one message from bot with semantic release appear

* Update reusable-comment-on-pr.yaml

* Update ci-pull-request.yaml

* Update ci-pull-request.yaml

* Update reusable-comment-on-pr.yaml

* Update ci-pull-request.yaml

* Update ci-pull-request.yaml
* restructure workflow files (#28)
  * chore: restructure workflow files

* add space
* return golang-lint (#27)
  * fix: return golang-lint

* bump version

* fix: bump golang-lint version to 2.10.1

* fix: use obfsExporterLogger for all general errors

* fix: add helper for closing the body

* chore: remove use-outputs

* fix: add .golangci.yml to skip workflow directory

* migrate golangci to new structure

* add version to to golang lint config
* Update ci-release.yaml
* Update reusable-docker-build-and-push.yaml
* chore: add result output at the end of the file in semantic release
* chore: fix releases not creating (#32)

## [2.1.2](https://github.com/PlushGuardian/obfuscation-server-exporter/compare/v2.1.1...v2.1.2) (2026-04-27)


### Bug Fixes

* add tab in workflows/release.yaml for proper readability of json ([bf52f64](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/bf52f64f710c96374624ce31f654668b7b171d4a))
* check varialbe output ([c553a78](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/c553a781a2adc39587d0dc91eb381960abbeb653))
* fix binaries job name ([a5ed787](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/a5ed7870be0bc9597b6906695b260148a1a6f03c))
* fix step names in semantic-release ([#26](https://github.com/PlushGuardian/obfuscation-server-exporter/issues/26)) ([e6972a9](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/e6972a9964d59df2c85c00ea926af40695a6b73f))
* fix variable references in semantic-release.yaml ([91a4f2a](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/91a4f2addaaebb30110e762d588e1657a02f1a48))
* make all build related jobs run right after merge ([057f7a9](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/057f7a9b26aedcab48ed88f0dc7f1cdb544ed836))
* remove github-token from name ([#22](https://github.com/PlushGuardian/obfuscation-server-exporter/issues/22)) ([382cdf3](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/382cdf3e11dd262dbd844012fea206095a58dba1))
* try to add | instead of > for variables  ([#24](https://github.com/PlushGuardian/obfuscation-server-exporter/issues/24)) ([1926caf](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/1926caf22c9f797e727a4e28f18920358b80fe0c))
* update output step ([11e4d0a](https://github.com/PlushGuardian/obfuscation-server-exporter/commit/11e4d0a09ae8bbd2e7e3b848b8b004c2aad5272f))

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
