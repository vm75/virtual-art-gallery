# V1 Remediation Tracker

This file replaces the original implementation-completion ledger after an independent review of `main` found acceptance-blocking gaps. Historical completion claims from the previous tracker are intentionally not carried forward as current status.

## Branch contract

- `main` — canonical/default product branch.
- `legacy` — pre-rewrite reference implementation only; never merge wholesale.
- `rewrite` — deleted. Historical issue text mentioning it is superseded by current issue comments and this tracker.
- `fix/v1-remediation` — active branch for all work in this tracker. Merge to `main` only after the release gates pass.

Review baseline: `main` commit `1a281221ab41e66af430856e14b3c4e3532c899e` (`feat: complete virtual art gallery v1 rewrite`).

## Agent execution rules

1. Read `AGENTS.md`, `README.md`, `ARCHITECTURE.md`, `IMPLEMENTATION_PLAN.md`, this tracker, and the full linked GitHub issue before changing code.
2. Work only on `fix/v1-remediation` for this corrective cycle.
3. Select the next open item whose dependencies are complete. Parallel work is allowed only when files and architectural boundaries do not conflict.
4. Implement one issue at a time. **Each issue that changes repository content must have its own distinct issue-scoped commit.** Do not combine unrelated tracker items in one commit.
5. Commit messages should identify the issue, for example `fix(#26): enforce sqlite foreign keys per connection`.
6. Run issue-specific tests plus applicable global quality gates before marking complete.
7. Update affected docs/tests in the same issue commit. Do not defer documentation to a later bulk cleanup except where #31 explicitly owns the branch/workflow transition.
8. Record the commit SHA and validation evidence here and in the GitHub issue before closing it.
9. Do not mark an issue complete merely because code exists; every acceptance criterion must be demonstrated.
10. Keep KISS and YAGNI. If a required fix reveals genuinely separate work, create a focused GitHub issue and add it here before implementing it.

## Status legend

- `[ ]` open / ready when dependencies are satisfied
- `[~]` in progress
- `[x]` acceptance criteria verified
- `[!]` blocked; blocker must be recorded

## Ordered remediation backlog

| Order | ID | GitHub | Status | Work item | Depends on | Commit |
|---:|---|---:|:---:|---|---|---|
| 1 | R-031 | [#31](https://github.com/vm75/virtual-art-gallery/issues/31) | [x] | Update branch/documentation contract from rewrite to main + remediation flow | — | 6570a6c (rebased) |
| 2 | I-021 | [#21](https://github.com/vm75/virtual-art-gallery/issues/21) | [ ] | CI and repeatable quality gates: target `main`/PRs correctly | R-031 | — |
| 3 | I-006 | [#6](https://github.com/vm75/virtual-art-gallery/issues/6) | [ ] | Fix single-admin login throttle identity and bounded cleanup | — | — |
| 4 | R-026 | [#26](https://github.com/vm75/virtual-art-gallery/issues/26) | [ ] | Enforce SQLite foreign keys on every connection | — | — |
| 5 | R-025 | [#25](https://github.com/vm75/virtual-art-gallery/issues/25) | [ ] | Version static assets or stop immutable caching stable URLs | — | — |
| 6 | R-027 | [#27](https://github.com/vm75/virtual-art-gallery/issues/27) | [ ] | Add artwork alt text editing and accessible image-link fallback | — | — |
| 7 | R-028 | [#28](https://github.com/vm75/virtual-art-gallery/issues/28) | [ ] | Make artwork slugs robust for non-ASCII titles | — | — |
| 8 | R-029 | [#29](https://github.com/vm75/virtual-art-gallery/issues/29) | [ ] | Improve derivative resizing quality and pixel bounds | — | — |
| 9 | I-015 | [#15](https://github.com/vm75/virtual-art-gallery/issues/15) | [ ] | Fix deterministic room geometry/capacity and placement-location collisions | — | — |
| 10 | I-016 | [#16](https://github.com/vm75/virtual-art-gallery/issues/16) | [ ] | Emit/validate renderer-ready artwork transforms and unique placements | I-015 | — |
| 11 | I-013 | [#13](https://github.com/vm75/virtual-art-gallery/issues/13) | [ ] | Render actual generated 3D rooms/walls/artworks with spatial camera | I-015, I-016, R-029 | — |
| 12 | I-017 | [#17](https://github.com/vm75/virtual-art-gallery/issues/17) | [ ] | Implement real room/spatial-aware texture loading/culling lifecycle | I-013, I-016, R-029 | — |
| 13 | I-018 | [#18](https://github.com/vm75/virtual-art-gallery/issues/18) | [ ] | Make desktop/mobile controls operate on the real 3D scene and select rendered art | I-013, I-016 | — |
| 14 | I-019 | [#19](https://github.com/vm75/virtual-art-gallery/issues/19) | [ ] | Build form-based museum rule editor and reject invalid generated layouts at publish | I-015, I-016 | — |
| 15 | R-030 | [#30](https://github.com/vm75/virtual-art-gallery/issues/30) | [ ] | Bring admin UI into responsive Material-inspired visual system | I-019, R-027 | — |
| 16 | I-022 | [#22](https://github.com/vm75/virtual-art-gallery/issues/22) | [ ] | Re-run security/accessibility/performance hardening after concrete fixes | I-006, R-025, R-026, R-027, R-028, R-029, R-030, I-017, I-018, I-019 | — |
| 17 | I-024 | [#24](https://github.com/vm75/virtual-art-gallery/issues/24) | [ ] | End-to-end fresh-install acceptance and v1 release validation | all items above | — |
| 18 | I-001 | [#1](https://github.com/vm75/virtual-art-gallery/issues/1) | [ ] | Final v1 release acceptance gate | I-024 | — |

Previously completed issues not listed above remain historical unless a remediation issue changes their behavior. They are still subject to regression verification during I-022/I-024.

## Required validation before final gate

At minimum, the final acceptance run must demonstrate:

- `gofmt` check, `go vet ./...`, `go test ./...`, production Go build, browser-module syntax checks, and CI success for the actual branch model;
- clean fresh install, one-admin setup/login/logout and stable throttling behavior;
- artwork upload/edit/visibility/alt text, filtering, canonical pages, non-ASCII-title slug behavior, safe derivatives and media delivery;
- Gallery Lite and Timeline keyboard/touch/mobile/reduced-motion behavior;
- museum rule builder draft/preview/publish including rejection of collisions/capacity-invalid layouts;
- deterministic generated rooms with unique renderer-ready artwork transforms;
- actual 3D walls/floor/artworks, spatial camera/navigation/collision, 3D artwork selection, desktop and mobile controls;
- bounded room/spatial-aware texture residency and cleanup;
- safe static caching across upgrades and SQLite foreign-key enforcement;
- non-root OCI build/run, Docker/Compose semantics where available, rootless Podman semantics, persistence, backup and restore;
- documentation audit for `README.md`, `ARCHITECTURE.md`, `AGENTS.md`, `DOCKERHUB.md`, `IMPLEMENTATION_PLAN.md`, and this tracker.

## Execution log
Append entries; do not erase prior remediation history.

```text
YYYY-MM-DD ID [~|x|!] — issue #N — commit SHA — tests/evidence — concise note
```

2026-09-07 REMEDIATION — review of main `1a281221...` reopened #1, #6, #13, #15, #16, #17, #18, #19, #21, #22, #24; created #25-#31; created branch `fix/v1-remediation`; replaced previous completion ledger with this remediation tracker.
2026-09-07 R-031 [x] — issue #31 — 6570a6c (rebased) — `git diff --check`; README/AGENTS/IMPLEMENTATION_PLAN/DOCKERHUB and tracker now document main + remediation flow.
