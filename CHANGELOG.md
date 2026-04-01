# Changelog

All notable changes to this project will be documented in this file.

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
