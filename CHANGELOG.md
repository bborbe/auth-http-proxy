# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v3.6.18

- Add osv-scanner, gosec, golines as indirect dependencies
- Add Makefile fix target for batch dependency updates

## v3.6.17

- security: bump github.com/bborbe/kv to v1.21.1; transitively removes vulnerable github.com/containerd/containerd v1.7.29 (CVE-2026-46680 / GHSA-fqw6-gf59-qr4w) and github.com/google/osv-scanner/v2 v2.3.4 from the dep tree

## v3.6.16

- Update bborbe/errors to v1.5.13, bborbe/http to v1.26.12
- Update onsi/ginkgo/v2 to v2.29.0 and gomega to v1.41.0
- Update kisielk/errcheck to v1.20.0 and sentry-go to v0.46.2
- Security: bump golang.org/x/net to v0.55.0 (CVE-2026-46680)
- Clean up unused indirect dependencies from go.mod

## v3.6.15

- security: bump github.com/containerd/containerd to v1.7.32 (CVE-2026-46680, GHSA-fqw6-gf59-qr4w)
- security: bump golang.org/x/net to v0.55.0 (GO-2026-5025..5030)

## v3.6.14

- security: bump github.com/go-git/go-git/v5 to v5.19.1 (CVE-2026-45570, CVE-2026-45571)

## v3.6.13

- security: suppress docker/docker CVE-2026-41567 / CVE-2026-42306 / CVE-2026-41568 / GHSA-x86f-5xw2-fm2r / GHSA-rg2x-37c3-w2rh / GHSA-vp62-88p7-qqf5 in .trivyignore/.osv-scanner.toml (no upstream fix; docker/docker module path unpatched, latest is v28.5.2)

## v3.6.12

- security: bump github.com/go-git/go-git/v5 to v5.19.0 (CVE-2026-45022)
- security: bump Go to 1.26.3 (GO-2026-4918, GO-2026-4971, GO-2026-4980, GO-2026-4982)
- chore: remove stale unused ignore entries from .osv-scanner.toml
- security: suppress docker/docker advisories (GHSA-pxq6-2prw-chj9, GHSA-x744-4wpc-v9h2) in .trivyignore/.osv-scanner.toml (no upstream fix; latest is v28.5.2, advisory wants >= v29.3.1)

## v3.6.11

- chore: Bump go.opentelemetry.io/otel* to v1.43.0 (CVE-2026-29181)
- chore: Verified github.com/go-git/go-git/v5 at v5.18.0 (CVE-2026-41506, no change needed)
- chore: github.com/docker/docker remains at v28.5.2+incompatible (Dependabot advisory: v29.3.1 not yet available from module proxy; OSV scanner confirms filtered as indirect dep with no fix available)

## v3.6.10

- Update ginkgo/v2 to v2.28.2
- Update golang.org/x/vuln to v1.3.0
- Update golang.org/x/telemetry
- Remove pinned anthropic-sdk-go replace directive

## v3.6.9

- Update go dependencies (bborbe/errors, bborbe/http, golang.org/x stdlib)
- Bump golang.org/x/vuln v1.2.0, golangci-lint v2.11.4, getsentry v0.45.0
- Simplify vulncheck Makefile target with inline OSV ignore list

## v3.6.8

- Update Go 1.26.2
- Update counterfeiter v6.12.2

## v3.6.7

- update go dependencies

## v3.6.6

- downgrade multiple Go dependencies to resolve compatibility issues
- add replace directive for github.com/denis-tingaikin/go-header v0.5.0

## v3.6.5

- Update dependencies to fix security vulnerabilities (go-git/v5 v5.17.2, buildkit v0.29.0, go-sdk v1.4.1)

## v3.6.4

- Update go-git/go-git to v5.17.1 (fix security vulnerabilities)

## v3.6.3

- Add SameSite=Lax attribute to auth cookie for CSRF protection
- Suppress gosec G124 false positive on cookie setup
- Update bborbe/errors to v1.5.8 and bborbe/http to v1.26.8
- Update indirect dependencies (deps.dev, cloud.google.com, charm.land)

## v3.6.2
- Update Go to 1.26.1
- Update dependencies (grpc v1.79.3, gosec v2.24.7, openai-go v3.23.0, etc.)
- Fix G120: limit request body size before form parsing

## v3.6.1

- Fix SSRF vulnerability in forward handler (gosec G704)
- Update Alpine to 3.23
- Update dependencies

## v3.6.0

- Update Go to 1.26.0
- Update all dependencies (errors v1.5.2, http v1.26.1, etc.)
- Modernize Dockerfile with BuildKit cache mounts
- Remove vendor mode from build
- Add OSV scanner to CI

## v3.5.2

- remove vendor
- go mod update

## v3.5.1

- go mod update

## v3.5.0

- set username in X-Forwarded-User header
- go mod update

## v3.4.3

- go mod update

## v3.4.2

- go mod update
- update Dockerimages

## v3.4.1

- go mod update

## v3.4.0

- inline http_handler code 
- go mod update

## v3.3.2

- go mod update

## v3.3.1

- go mod upgrade

## v3.3.0

- Add multi tags

## v3.2.1

- Fix CacheVerifier

## v3.2.0

- Add Dockerfile again

## v3.1.0

- Replace deps with go modules

## v3.0.0

- Remove auth support 
- Refactoring

## v2.1.3

- Use multistage dockerfile

## v2.1.2

- Fix Security Bug (https://github.com/jtblin/go-ldap-client/issues/16)

## v2.1.1

- Update dependencies

## v2.1.0

- Close Ldap connections on failure 

## v2.0.0

- Replace glide with deps
- Add Jenkinsfile
- Remove auth
