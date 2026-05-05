---
status: committing
summary: 'Verified Go security dep updates: otel at v1.43.0 (CVE-2026-29181 fixed), go-git at v5.18.0 (CVE-2026-41506 fixed), docker/docker remains at v28.5.2 (v29.3.1 unavailable from proxy, filtered by osv-scanner); make precommit passed with exit code 0.'
container: auth-http-proxy-001-update-go-deps-security
dark-factory-version: v0.148.4-3-gc45254a
created: "2026-05-05T17:38:46Z"
queued: "2026-05-05T17:38:46Z"
started: "2026-05-05T19:26:26Z"
completed: "2026-05-05T19:10:57Z"
lastFailReason: 'validate completion report: completion report status: partial'
---

<summary>
- Go dependencies updated to latest allowed versions via `updater`
- `go.opentelemetry.io/otel` core (otel, otel/sdk, otel/trace, otel/metric) bumped to v1.41.0+ (CVE-2026-29181)
- `github.com/docker/docker` bumped if a fixed version is available on proxy.golang.org; otherwise the blocker is documented and the run still completes
- `github.com/go-git/go-git/v5` confirmed at >= v5.18.0 (CVE-2026-41506)
- make precommit passes cleanly
- `## Unreleased` section in CHANGELOG.md lists what changed
</summary>

<objective>
Update Go module dependencies to resolve Dependabot security advisories on `docker/docker` and `go.opentelemetry.io/otel`, and verify `go-git/v5` is at the patched version.
</objective>

<context>
Read CLAUDE.md for project conventions.
Read `docs/dod.md` for the Definition of Done criteria.

Current state of `go.mod` (all three deps are `// indirect`):
- `github.com/docker/docker v28.5.2+incompatible` — vulnerable, advisory: bump to >= v29.3.1
- `go.opentelemetry.io/otel v1.40.0` — vulnerable (CVE-2026-29181, high), bump to >= v1.41.0 (covers all `go.opentelemetry.io/otel*` modules)
- `github.com/go-git/go-git/v5 v5.18.0` — already at fixed version (CVE-2026-41506); verify only

`updater` is pre-installed in the claude-yolo container.
</context>

<requirements>
1. Run `updater --verbose --yes go` in the **foreground** (do NOT background this command).
2. If `updater` fails on any rename, follow recovery: `grep -r '<stale-identifier>' --exclude-dir=vendor`, fix all occurrences, re-run `make generate`, `make test`. Common rename patterns from prior runs: `*Id` → `*ID`, `*Url` → `*URL`, `HttpClient` → `HTTPClient`.
3. **Docker advisory (best-effort, do NOT block on this):**
   - Check the latest available version: `curl -s https://proxy.golang.org/github.com/docker/docker/@latest`
   - If the proxy reports a version `>= v29.3.1`, run `go get github.com/docker/docker@latest && go mod tidy`. `+incompatible` suffix is expected; transition `// indirect` → direct is acceptable.
   - If the proxy still reports a version `< v29.3.1` (as of 2026-05-05 the latest was v28.5.2+incompatible), the advisory is **unfixable today** — leave `go.mod` as-is and document the blocker in the CHANGELOG entry. This is NOT a partial outcome; the project's `osv-scanner` config already filters this advisory as "indirect dep, no fix available". Report status `completed` in the DARK-FACTORY-REPORT.
4. **otel advisory (CVE-2026-29181):** Bump the core modules. If `go.opentelemetry.io/otel` is `< v1.41.0`, run `go get go.opentelemetry.io/otel@latest go.opentelemetry.io/otel/sdk@latest go.opentelemetry.io/otel/trace@latest go.opentelemetry.io/otel/metric@latest && go mod tidy`. **Exporter submodules** (`otelhttp`, `otlpmetricgrpc`, `otlptracegrpc`, `prometheus`, `jaeger`) may stay below v1.41.0 — they are pinned by transitive deps (`moby/buildkit@v0.29.0`) and the CVE only affects the core modules. This is acceptable; do NOT force-bump exporters.
5. Verify `go.mod` shows `github.com/go-git/go-git/v5 >= v5.18.0`. If a regression dropped it below, run `go get github.com/go-git/go-git/v5@latest && go mod tidy`.
6. Run `make precommit` — must pass with exit code 0.
7. Update `CHANGELOG.md`:
   - If a `## Unreleased` heading does not exist, insert one above the most recent version section (currently `## v3.6.10`).
   - Under `## Unreleased`, add one bullet per bumped dep (format: `- Bump github.com/docker/docker to vX.Y.Z (Dependabot advisory)`).
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git
- Do NOT run `updater` as a background task — use foreground with `--verbose`
- Existing tests must still pass
- No `exclude` or `replace` directives in go.mod
- Do NOT hand-edit version numbers in `go.mod` — let `updater` / `go get` write them
</constraints>

<verification>
Run `make precommit` — must pass with exit code 0.
Run `go list -m go.opentelemetry.io/otel go.opentelemetry.io/otel/sdk go.opentelemetry.io/otel/trace go.opentelemetry.io/otel/metric` — each must be >= v1.41.0.
Run `go list -m github.com/go-git/go-git/v5` — version must be >= v5.18.0.
For docker/docker: report the version found and whether the proxy has v29.3.1+ available. Do NOT fail the run if v29.3.1 is unavailable upstream (see requirement 3).
</verification>
