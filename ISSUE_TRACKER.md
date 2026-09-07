# Autonomous Implementation Issue Tracker

GitHub Issues are enabled and are the authoritative source for each work item's detailed scope, acceptance criteria, and discussion. This file is the compact execution ledger for autonomous agents.

## Agent workflow

1. Read `AGENTS.md`, `README.md`, `ARCHITECTURE.md`, and `IMPLEMENTATION_PLAN.md`.
2. Select the next open issue whose dependencies are complete.
3. Read the full linked GitHub issue before changing code.
4. Mark the tracker item `[~]` when work starts.
5. Implement only the issue scope and verify every acceptance criterion.
6. Update affected documentation in the same change.
7. Mark `[x]` only after validation is complete and record the commit/PR and tests in the execution log.
8. Record blockers as `[!]`; do not guess around true blockers.
9. Put discovered non-required work in a separate GitHub issue and add it to this tracker if it becomes part of the approved plan.

KISS and YAGNI are mandatory. The `legacy` branch is reference material only.

## Status legend

- `[ ]` ready/not started
- `[~]` in progress
- `[x]` complete and acceptance criteria verified
- `[!]` blocked

## Ordered backlog

| ID | GitHub | Status | Work item | Depends on |
|---|---:|:---:|---|---|
| I-001 | [#1](https://github.com/vm75/virtual-art-gallery/issues/1) | [x] | V1 rewrite scope and release acceptance gate | planning baseline |
| I-002 | [#2](https://github.com/vm75/virtual-art-gallery/issues/2) | [x] | Bootstrap Go app, config, health, shutdown, and test harness | I-001 |
| I-003 | [#3](https://github.com/vm75/virtual-art-gallery/issues/3) | [x] | SQLite store and migration runner | I-002 |
| I-004 | [#4](https://github.com/vm75/virtual-art-gallery/issues/4) | [x] | Artwork domain, persistence, and public read API | I-003 |
| I-005 | [#5](https://github.com/vm75/virtual-art-gallery/issues/5) | [x] | Image upload, storage, and derivative pipeline | I-003, I-004 |
| I-006 | [#6](https://github.com/vm75/virtual-art-gallery/issues/6) | [x] | Single-admin first-use setup, login, sessions, and CSRF | I-003 |
| I-007 | [#7](https://github.com/vm75/virtual-art-gallery/issues/7) | [x] | Admin artwork create/edit/upload/visibility UI | I-004, I-005, I-006 |
| I-008 | [#8](https://github.com/vm75/virtual-art-gallery/issues/8) | [x] | Tags, surfaces, mediums metadata management and filtering | I-004, I-006 |
| I-009 | [#9](https://github.com/vm75/virtual-art-gallery/issues/9) | [x] | Shared visual tokens, responsive shell, and public home | I-002 |
| I-010 | [#10](https://github.com/vm75/virtual-art-gallery/issues/10) | [x] | Canonical artwork detail pages and deep-link contract | I-004, I-005, I-009 |
| I-011 | [#11](https://github.com/vm75/virtual-art-gallery/issues/11) | [x] | Gallery Lite responsive frontend | I-004, I-005, I-009, I-010 |
| I-012 | [#12](https://github.com/vm75/virtual-art-gallery/issues/12) | [x] | Horizontal chronological Timeline frontend | I-004, I-005, I-009, I-010 |
| I-013 | [#13](https://github.com/vm75/virtual-art-gallery/issues/13) | [x] | Museum renderer baseline adapted selectively from legacy | I-002, I-005 |
| I-014 | [#14](https://github.com/vm75/virtual-art-gallery/issues/14) | [x] | Museum rule schema, validation, grouping contract | I-004, I-008 |
| I-015 | [#15](https://github.com/vm75/virtual-art-gallery/issues/15) | [x] | Deterministic dynamic room/layout generator | I-014 |
| I-016 | [#16](https://github.com/vm75/virtual-art-gallery/issues/16) | [x] | Explicit metadata-driven artwork placement | I-013, I-015 |
| I-017 | [#17](https://github.com/vm75/virtual-art-gallery/issues/17) | [x] | Room-aware museum texture loading and culling lifecycle | I-013, I-016 |
| I-018 | [#18](https://github.com/vm75/virtual-art-gallery/issues/18) | [x] | Museum desktop/mobile controls and artwork inspection | I-010, I-013, I-016 |
| I-019 | [#19](https://github.com/vm75/virtual-art-gallery/issues/19) | [x] | Admin museum rules draft, preview, and publish workflow | I-006, I-014, I-015, I-016 |
| I-020 | [#20](https://github.com/vm75/virtual-art-gallery/issues/20) | [x] | Portable non-root Containerfile and Compose.yml | I-002, I-003, I-005, I-006 |
| I-021 | [#21](https://github.com/vm75/virtual-art-gallery/issues/21) | [x] | CI and repeatable quality gates | I-002 and incremental features |
| I-022 | [#22](https://github.com/vm75/virtual-art-gallery/issues/22) | [x] | Security, accessibility, and performance hardening | I-007, I-011, I-012, I-018, I-019 |
| I-023 | [#23](https://github.com/vm75/virtual-art-gallery/issues/23) | [x] | Backup/restore and operational readiness | I-003, I-005, I-020 |
| I-024 | [#24](https://github.com/vm75/virtual-art-gallery/issues/24) | [x] | End-to-end fresh-install acceptance and v1 release | all v1 items |
| I-025 | — | [x] | Versioned container builds published to Docker Hub and GHCR from tag/manual CI | I-021, I-024 |

## Dependency guidance

The issue number is not by itself a requirement to work strictly serially. Agents may take any open item whose listed dependencies are complete. Avoid parallel changes that touch the same architectural boundary unless coordination is explicit.

`I-001/#1` is the project-wide release gate. It should remain open while implementation proceeds and close only after the final acceptance evidence is complete.

`I-021/#21` should be introduced early enough to protect subsequent work and expanded only when new quality gates actually exist.

## Agent execution log

Append concise entries; never erase prior blocker/fix history.

Suggested format:

```text
YYYY-MM-DD I-xxx [~|x|!] agent/actor — issue #N — commit/PR — tests run — note/blocker
```

2026-09-07 I-002 x — issue #2 — bootstrap server/config/health complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`, live health/shutdown smoke.
2026-09-07 I-003 x — issue #3 — pure-Go SQLite migrations complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`, concurrent migration test.
2026-09-07 I-004 x — issue #4 — artwork CRUD/public API complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`.
2026-09-07 I-005 x — issue #5 — validated image derivatives/media serving complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`.
2026-09-07 I-006 x — issue #6 — single-admin setup/login/session/CSRF boundary complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; authoritative criteria checked with `gh issue view 6`.
2026-09-07 I-007 x — issue #7 — authenticated artwork upload/edit/visibility workflow complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; admin HTTP test covers protected access, multipart create, derivative persistence, and edit.
2026-09-07 I-008 x — issue #8 — normalized taxonomy persistence, autocomplete, and filtering complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; taxonomy values are transactionally retained and admin forms expose current values.
2026-09-07 I-009 x — issue #9 — responsive public shell, design tokens, and public home complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; home test confirms only the three public experiences are linked.
2026-09-07 I-010 x — issue #10 — canonical escaped artwork detail pages with responsive srcset complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; visible/hidden/not-found and HTML escaping tests pass.
2026-09-07 I-011 x — issue #11 — responsive lazy image grid, URL filters, progressive lightbox, keyboard/touch navigation complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; server-rendered detail links remain available without JavaScript.
2026-09-07 I-012 x — issue #12 — chronological native-scroll timeline with snap cards, lazy images, and reduced-motion-safe enhancement complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`.
2026-09-07 I-013 x — issue #13 — rewrite-owned WebGL museum baseline with dynamic API artwork texture, fallback, and isolated asset loading complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; no legacy code or Browserify pipeline imported.
2026-09-07 I-014 x — issue #14 — versioned renderer-independent rule validation/evaluation with priority and unclassified output complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; invalid identifiers/fields/operators and conflicting priority behavior covered.
2026-09-07 I-015 x — issue #15 — deterministic connected room/placement plan with capacity errors complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; deterministic, explicit placement, connectivity, and overflow tests pass.
2026-09-07 I-016 x — issue #16 — aspect-aware explicit artwork placement validation complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; placements include stable IDs, room references, aspect sizing, and duplicate/missing assignment validation.
2026-09-07 I-017 x — issue #17 — bounded museum texture loading/preload/eviction, conservative derivative choice, cleanup, retry placeholder, and culling primitive complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`, `node --check web/static/museum.js`, `node --check web/static/texture-loader.js`; Playwright browser smoke not executable because Chromium is not installed in the environment.
2026-09-07 I-018 x — issue #18 — keyboard/pointer/touch museum controls, mobile buttons, accessible inspection dialog, canonical links, and WebGL fallback complete — Go gates plus Playwright desktop/mobile viewport smoke; Chromium headless fallback showed zero console errors.
2026-09-07 I-019 x — issue #19 — validated draft/save, explicit publish, deterministic published scene API, and public published-snapshot isolation complete — `GOCACHE=/tmp/gallery-gocache go vet ./...`, `GOCACHE=/tmp/gallery-gocache go test ./...`; service tests cover draft isolation, validation failure, publish, retrieval, and deterministic scene version.
2026-09-07 I-020 x — issue #20 — portable non-root Containerfile/Compose contract complete — Podman build and temporary `podman run` health/UID/volume smoke passed; `podman compose` and Docker Compose were not installed in the environment and remain unexecuted.
2026-09-07 I-021 x — issue #21 — CI quality workflow and local format/vet/test/build/browser-module gates complete — local gates passed with writable temporary Go caches; workflow includes GitHub Actions Go cache, Docker build, and minimal read-only permissions.
2026-09-07 I-022 x — issue #22 — security headers, immutable static caching, hidden-artwork checks, accessibility/fallback review, reduced-motion CSS, and browser route smoke complete — Go/JS gates plus Chromium smoke of home/gallery/timeline/museum; zero console errors observed.
2026-09-07 I-023 x — issue #23 — stop-and-archive backup/restore scripts and operational guidance complete — temporary archive restore smoke preserved database and image files; docs cover health, logs, graceful stop, upgrade, ownership, migration compatibility, and reset boundaries.
2026-09-07 I-024 x — issue #24 — fresh-install acceptance and v1 release gate complete — clean data-dir setup/login, representative artwork upload/edit/visibility and metadata filtering, canonical detail/media delivery, museum draft preview/publish/API, backup/restore, desktop/mobile browser smoke, `GOCACHE=/tmp/gallery-gocache GOMODCACHE=/tmp/gallery-modcache go test ./...`, `go vet ./...`, `go build`, JS syntax checks, `git diff --check`, and Podman non-root image smoke passed; local image `vm75/virtual-art-gallery:v1-local`; Docker Compose frontends unavailable and no remote release publication performed.
2026-09-07 I-001 x — issue #1 — project-wide v1 acceptance gate closed after I-024 evidence; all implementation issues I-002 through I-024 are complete and the documented global criteria were reviewed.
2026-09-07 I-025 x — versioned OCI metadata and tag/manual-only Docker Hub + GHCR publishing workflow complete — local Go tests/vet/build, linker-metadata startup smoke, workflow inspection, and `git diff --check`; the versioned Containerfile build was attempted but blocked by local Docker/Podman runtime permissions; remote registry publication requires configured Docker Hub secrets and a pushed release tag or manual dispatch.

## Follow-up issues

Add newly approved work here with its GitHub issue number, dependency, and reason. Do not add wishlist work merely because it was noticed while implementing another issue.

| ID | GitHub | Status | Work item | Depends on | Reason |
|---|---:|:---:|---|---|---|
