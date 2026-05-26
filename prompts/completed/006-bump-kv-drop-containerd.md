---
status: completed
summary: Bumped github.com/bborbe/kv to v1.21.1, transitively removing vulnerable github.com/containerd/containerd v1.7.29 and github.com/google/osv-scanner/v2 from the dep tree; updated CHANGELOG.md
container: auth-http-proxy-exec-006-bump-kv-drop-containerd
dark-factory-version: v0.173.0
created: "2026-05-26T15:38:32Z"
queued: "2026-05-26T15:38:32Z"
started: "2026-05-26T15:38:37Z"
completed: "2026-05-26T15:46:18Z"
---
<summary>
- Bumps `github.com/bborbe/kv` from v1.19.6 to v1.21.1 in `go.mod`
- Side effect: `bborbe/kv` v1.21.1 no longer depends on `google/osv-scanner/v2` (which was the sole source of `github.com/containerd/containerd v1.7.29` in this repo's dep tree)
- After `go mod tidy`, both `osv-scanner/v2` and `containerd` are removed from `go.mod` / `go.sum`
- Patches CVE-2026-46680 / GHSA-fqw6-gf59-qr4w (containerd, High, Dependabot alert #34) by dependency-tree removal — cleaner than pinning
- `make precommit` exits 0
- CHANGELOG `## Unreleased` documents the bump and its security impact
</summary>

<objective>
Eliminate the vulnerable `github.com/containerd/containerd v1.7.29` indirect dep from `bborbe/auth-http-proxy` by bumping its only parent (`bborbe/kv`) to a version (v1.21.1) that no longer pulls the containerd dep chain. Resolves Dependabot alert #34 (CVE-2026-46680).
</objective>

<context>
Read `CLAUDE.md` for project conventions.
Read `docs/dod.md` for the Definition of Done.

Current state (verified before writing this prompt):
- `go list -m github.com/containerd/containerd` → `v1.7.29` (vulnerable range: >= 1.7.27, < 1.7.32)
- `containerd` is NOT pinned in `go.mod` directly; it is pulled transitively
- `go mod graph` shows: only `github.com/bborbe/kv v1.19.6` (via `google/osv-scanner/v2 v2.3.4`) pulls containerd
- `osv-scanner/v2` is `// indirect` in `auth-http-proxy/go.mod` — not directly imported by any `.go` file in this repo

Verification done in `/tmp` with a throwaway module:
- `go get github.com/bborbe/kv@v1.21.1` → tree contains NO `osv-scanner/v2` and NO `containerd`
- `bborbe/kv` v1.21.1 is the cleanest fix: dependency-tree removal, no explicit pin needed

Prior prompt `005-update-containerd-1.7.32.md` (completed 2026-05-22) bumped containerd to v1.7.32 via explicit `// indirect` pin, but commit `bdc3314 update dependencies` (2026-05-26 07:57) ran `go mod tidy` which removed the pin (because nothing was importing it), and MVS resolved containerd back to v1.7.29. Pinning is fragile here; removing the parent dep chain is the durable fix.

Advisory: https://github.com/advisories/GHSA-fqw6-gf59-qr4w
CVE: CVE-2026-46680, severity High

Existing docker/docker suppressions in `.trivyignore` and `.osv-scanner.toml` MUST remain intact.
</context>

<requirements>
1. Bump kv:
   ```bash
   go get github.com/bborbe/kv@v1.21.1
   go mod tidy
   ```

2. Confirm containerd and osv-scanner/v2 were removed from the dep tree:
   ```bash
   go list -m github.com/containerd/containerd 2>&1 | grep -q "not a known dependency"
   go list -m github.com/google/osv-scanner/v2 2>&1 | grep -q "not a known dependency"
   ```
   - If either still appears in `go list -m`, run `go mod graph | grep -E "(containerd/containerd|osv-scanner)"` to discover the other parent and STOP — report blocker with the parent path in the prompt summary. Do NOT pin containerd explicitly without approval; the goal is dependency-tree removal.

3. Run `make precommit`. If trivy or osv-scanner reports NEW advisory IDs on `github.com/docker/docker` that are NOT yet in the ignore files, append them:
   - CVE-IDs → `.trivyignore` under the existing `# github.com/docker/docker indirect dep, no fix available via Go modules` block
   - GHSA-IDs → new `[[IgnoredVulns]]` blocks in `.osv-scanner.toml` with `reason = "github.com/docker/docker indirect dep, no fix available"`
   - Re-run `make precommit`. Cap at 3 iterations. If still failing after 3, stop and report blocker.

4. If the scanner reports advisories on any package OTHER than `github.com/docker/docker` AND a fix version is available, bump that package to the fixed version using `go get pkg@version && go mod tidy`. List bumps in CHANGELOG. If no fix is available for a non-docker/docker package, stop and report blocker — do NOT suppress.

5. Do NOT remove any existing entries from `.trivyignore` or `.osv-scanner.toml`.

6. Update `CHANGELOG.md` under `## Unreleased` (create the section as the first one below the top-of-file header if missing):
   ```
   - security: bump github.com/bborbe/kv to v1.21.1; transitively removes vulnerable github.com/containerd/containerd v1.7.29 (CVE-2026-46680 / GHSA-fqw6-gf59-qr4w) and github.com/google/osv-scanner/v2 v2.3.4 from the dep tree
   ```
   If additional docker/docker IDs were added in step 3 or collateral bumps happened in step 4, list them on separate lines.

7. Final verification:
   - `grep -E '^\s*github\.com/bborbe/kv\s+v1\.21\.1' go.mod` succeeds
   - `! grep -q 'containerd/containerd' go.mod`
   - `! grep -q 'osv-scanner/v2' go.mod`
   - `make precommit` exits 0
</requirements>

<constraints>
- Only edit: `go.mod`, `go.sum`, `.trivyignore`, `.osv-scanner.toml`, `CHANGELOG.md`
- Do NOT pin containerd explicitly (the whole point is dependency-tree removal)
- Do NOT add `replace` or `exclude` directives
- Do NOT promote any indirect dep to direct
- Do NOT remove existing docker/docker suppressions
- Do NOT bump deps unrelated to security fixes (except collateral fixes per step 4)
- Do NOT commit — dark-factory handles git
- Existing tests must still pass
</constraints>

<verification>
```bash
grep -E '^\s*github\.com/bborbe/kv\s+v1\.21\.1' go.mod
! grep -q 'containerd/containerd' go.mod
! grep -q 'osv-scanner/v2' go.mod
make precommit
```
</verification>
