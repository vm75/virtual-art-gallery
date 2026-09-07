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
| I-001 | [#1](https://github.com/vm75/virtual-art-gallery/issues/1) | [ ] | V1 rewrite scope and release acceptance gate | planning baseline |
| I-002 | [#2](https://github.com/vm75/virtual-art-gallery/issues/2) | [ ] | Bootstrap Go app, config, health, shutdown, and test harness | I-001 |
| I-003 | [#3](https://github.com/vm75/virtual-art-gallery/issues/3) | [ ] | SQLite store and migration runner | I-002 |
| I-004 | [#4](https://github.com/vm75/virtual-art-gallery/issues/4) | [ ] | Artwork domain, persistence, and public read API | I-003 |
| I-005 | [#5](https://github.com/vm75/virtual-art-gallery/issues/5) | [ ] | Image upload, storage, and derivative pipeline | I-003, I-004 |
| I-006 | [#6](https://github.com/vm75/virtual-art-gallery/issues/6) | [ ] | Single-admin first-use setup, login, sessions, and CSRF | I-003 |
| I-007 | [#7](https://github.com/vm75/virtual-art-gallery/issues/7) | [ ] | Admin artwork create/edit/upload/visibility UI | I-004, I-005, I-006 |
| I-008 | [#8](https://github.com/vm75/virtual-art-gallery/issues/8) | [ ] | Tags, surfaces, mediums metadata management and filtering | I-004, I-006 |
| I-009 | [#9](https://github.com/vm75/virtual-art-gallery/issues/9) | [ ] | Shared visual tokens, responsive shell, and public home | I-002 |
| I-010 | [#10](https://github.com/vm75/virtual-art-gallery/issues/10) | [ ] | Canonical artwork detail pages and deep-link contract | I-004, I-005, I-009 |
| I-011 | [#11](https://github.com/vm75/virtual-art-gallery/issues/11) | [ ] | Gallery Lite responsive frontend | I-004, I-005, I-009, I-010 |
| I-012 | [#12](https://github.com/vm75/virtual-art-gallery/issues/12) | [ ] | Horizontal chronological Timeline frontend | I-004, I-005, I-009, I-010 |
| I-013 | [#13](https://github.com/vm75/virtual-art-gallery/issues/13) | [ ] | Museum renderer baseline adapted selectively from legacy | I-002, I-005 |
| I-014 | [#14](https://github.com/vm75/virtual-art-gallery/issues/14) | [ ] | Museum rule schema, validation, and grouping contract | I-004, I-008 |
| I-015 | [#15](https://github.com/vm75/virtual-art-gallery/issues/15) | [ ] | Deterministic dynamic room/layout generator | I-014 |
| I-016 | [#16](https://github.com/vm75/virtual-art-gallery/issues/16) | [ ] | Explicit metadata-driven artwork placement | I-013, I-015 |
| I-017 | [#17](https://github.com/vm75/virtual-art-gallery/issues/17) | [ ] | Room-aware museum texture loading and culling lifecycle | I-013, I-016 |
| I-018 | [#18](https://github.com/vm75/virtual-art-gallery/issues/18) | [ ] | Museum desktop/mobile controls and artwork inspection | I-010, I-013, I-016 |
| I-019 | [#19](https://github.com/vm75/virtual-art-gallery/issues/19) | [ ] | Admin museum rules draft, preview, and publish workflow | I-006, I-014, I-015, I-016 |
| I-020 | [#20](https://github.com/vm75/virtual-art-gallery/issues/20) | [ ] | Portable non-root Containerfile and Compose.yml | I-002, I-003, I-005, I-006 |
| I-021 | [#21](https://github.com/vm75/virtual-art-gallery/issues/21) | [ ] | CI and repeatable quality gates | I-002 and incremental features |
| I-022 | [#22](https://github.com/vm75/virtual-art-gallery/issues/22) | [ ] | Security, accessibility, and performance hardening | I-007, I-011, I-012, I-018, I-019 |
| I-023 | [#23](https://github.com/vm75/virtual-art-gallery/issues/23) | [ ] | Backup/restore and operational readiness | I-003, I-005, I-020 |
| I-024 | [#24](https://github.com/vm75/virtual-art-gallery/issues/24) | [ ] | End-to-end fresh-install acceptance and v1 release | all v1 items |

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

## Follow-up issues

Add newly approved work here with its GitHub issue number, dependency, and reason. Do not add wishlist work merely because it was noticed while implementing another issue.

| ID | GitHub | Status | Work item | Depends on | Reason |
|---|---:|:---:|---|---|---|
