---
status: completed
summary: Bumped containerd to v1.7.32 and golang.org/x/net to v0.55.0 to patch CVE-2026-46680
container: auth-http-proxy-exec-005-update-containerd-1-7-32
dark-factory-version: v0.164.0
created: "2026-05-22T18:00:00Z"
queued: "2026-05-22T17:11:43Z"
started: "2026-05-22T17:39:27Z"
completed: "2026-05-22T17:40:58Z"
---

<summary>
- Bumps `github.com/containerd/containerd` from v1.7.29 to v1.7.32 (CVE-2026-46680 / GHSA-fqw6-gf59-qr4w, High)
- Bumps `golang.org/x/net` from v0.53.0 to v0.55.0 as collateral fix (GO-2026-5025..5030 surface in govulncheck after containerd bump)
- Resolves Dependabot containerd advisory for bborbe/auth-http-proxy (2026-05-21)
- containerd and x/net are indirect deps — must stay `// indirect`
- `make precommit` exits 0 after the change
- CHANGELOG `## Unreleased` documents both bumps
</summary>

<objective>
Patch CVE-2026-46680 (containerd user-ID handling bypass allows runAsNonRoot evasion) by upgrading `github.com/containerd/containerd` to v1.7.32.
</objective>

<context>
Read `CLAUDE.md` for project conventions.
Read `docs/dod.md` for the Definition of Done.

Current `go.mod`:
- `github.com/containerd/containerd v1.7.29 // indirect` — has fix: bump to v1.7.32 (vulnerable range: >= 1.7.27, < 1.7.32)
- `golang.org/x/net v0.53.0 // indirect` — govulncheck reports GO-2026-5025..5030; latest fixed: v0.55.0

Advisories:
- https://github.com/advisories/GHSA-fqw6-gf59-qr4w (containerd)
- CVE: CVE-2026-46680, severity High, CVSS 7.3
- govulncheck GO-2026-5025, 5026, 5027, 5028, 5029, 5030 (golang.org/x/net)

Note: a prior run of this prompt with only the containerd bump caused `make precommit` to fail because govulncheck surfaced the above x/net advisories. Bumping x/net to v0.55.0 alongside containerd is the correct collateral fix per the runbook section "Common collateral findings".
</context>

<requirements>
1. Bump containerd and x/net:
   ```bash
   go get github.com/containerd/containerd@v1.7.32
   go get golang.org/x/net@v0.55.0
   go mod tidy
   ```

2. Run `make precommit`. If it fails because trivy or osv-scanner reports NEW advisory IDs on `github.com/docker/docker` (existing suppression pattern) that are NOT yet in the ignore files, append them:
   - CVE-IDs → `.trivyignore` under the existing `# github.com/docker/docker indirect dep, no fix available via Go modules` block.
   - GHSA-IDs → new `[[IgnoredVulns]]` blocks in `.osv-scanner.toml` with `reason = "github.com/docker/docker indirect dep, no fix available"`.
   - Re-run `make precommit`. Cap at 3 iterations. If still failing after 3 iterations, stop and report blocker with the remaining unsuppressed advisory IDs in the prompt summary.

3. If the scanner reports advisories on any package OTHER than `github.com/docker/docker`, `github.com/containerd/containerd`, or `golang.org/x/net` AND a fix version is available, bump that package to the fixed version using `go get pkg@version && go mod tidy`. If no fix is available, stop and report blocker — do NOT suppress non-docker/docker packages.

4. Do NOT add ignore entries for any advisory that has a fix available. Only docker/docker is allowed to be suppressed.

5. Update `CHANGELOG.md` under `## Unreleased` (create the section as the first one below the top-of-file header if it does not exist):
   ```
   - security: bump github.com/containerd/containerd to v1.7.32 (CVE-2026-46680, GHSA-fqw6-gf59-qr4w)
   - security: bump golang.org/x/net to v0.55.0 (GO-2026-5025..5030)
   ```
   If additional docker/docker IDs were added in step 2, append another line listing them. If additional collateral bumps were done in step 3, list them too.

6. Verify:
   - `go list -m github.com/containerd/containerd` reports `v1.7.32`
   - `go list -m golang.org/x/net` reports `v0.55.0`
   - `make precommit` exits 0
</requirements>

<constraints>
- Only edit: `go.mod`, `go.sum`, `.trivyignore`, `.osv-scanner.toml`, `CHANGELOG.md`
- Do NOT edit `Makefile` — VULNCHECK_IGNORE list is not part of this scope
- Do NOT bump deps unrelated to security fixes
- Do NOT promote any indirect dep to direct (containerd, docker/docker, x/net must stay `// indirect`)
- Do NOT add a `replace` or `exclude` directive
- Do NOT commit — dark-factory handles git
- Existing tests must still pass
</constraints>

<verification>
```bash
go list -m github.com/containerd/containerd   # must print v1.7.32
make precommit                                 # must exit 0
```
</verification>
