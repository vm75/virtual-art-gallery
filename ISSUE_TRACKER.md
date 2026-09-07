# V1 Release Hardening Tracker

This tracker is the active execution ledger after the 2026-09-07 independent second review of `fix/v1-remediation` and the subsequent third-review follow-up on `fix/v1-release-hardening`. Previous completion claims remain available in Git history, but current status is authoritative here.

## Branch contract

- `main` — canonical/default product branch and merge target.
- `legacy` — pre-rewrite reference implementation only; never merge wholesale.
- `rewrite` — deleted and fully superseded by `main`. Do not target it or instruct work on it.
- `fix/v1-remediation` — historical remediation branch; no new fixes land there.
- `fix/v1-release-hardening` — active branch for every issue in this tracker. It was created from the reviewed `fix/v1-remediation` head so prior remediation work is retained while new work moves to a clean branch.

## Execution rules

1. Read the full GitHub issue before changing code.
2. Work only on `fix/v1-release-hardening` for this cycle.
3. Implement one issue at a time.
4. Every repository-changing issue gets exactly one distinct issue-scoped commit. Do not combine unrelated fixes.
5. Commit messages must identify the issue, e.g. `fix(#18): constrain museum navigation`.
6. Include issue-specific tests and update directly affected docs in the same issue commit.
7. Run applicable global gates before closing an issue: formatting, `go vet ./...`, `go test ./...`, production build, browser-module checks, and any issue-specific browser/container validation.
8. Record commit SHA and validation evidence in the GitHub issue and this tracker before marking complete.
9. If work reveals a separate defect, create a focused issue and add it here before implementing it.
10. Keep KISS/YAGNI. Release-gate issues (#22, #24, #1) do not absorb unrelated implementation work.

## Status legend

- `[ ]` open / ready when dependencies are satisfied
- `[~]` in progress
- `[x]` acceptance criteria verified
- `[!]` blocked; blocker must be recorded

## Ordered hardening backlog

| Order | ID | GitHub | Status | Finding / work item | Depends on | Commit |
|---:|---|---:|:---:|---|---|---|
| 1 | R-034 | [#34](https://github.com/vm75/virtual-art-gallery/issues/34) | [x] | Reset tracker and canonical branch contract | — | `docs(#34): reset v1 hardening tracker` |
| 2 | I-013 | [#13](https://github.com/vm75/virtual-art-gallery/issues/13) | [~] | Preserve a right-handed museum camera/view basis and revalidate the actual renderer | R-034 | `4a8935a` |
| 3 | I-015 | [#15](https://github.com/vm75/virtual-art-gallery/issues/15) | [x] | Make logical room connections physically continuous/traversable | R-034 | `fix(#15): connect generated museum rooms` |
| 4 | I-018 | [#18](https://github.com/vm75/virtual-art-gallery/issues/18) | [~] | Prevent swept-path wall tunneling and make canvas taps select the actual rendered artwork | I-013, I-015 | `e0c9578` |
| 5 | I-017 | [#17](https://github.com/vm75/virtual-art-gallery/issues/17) | [x] | Stop repeated GPU texture delete/re-upload during ordinary navigation | I-013 | `fix(#17): reuse resident museum textures` |
| 6 | I-019 | [#19](https://github.com/vm75/virtual-art-gallery/issues/19) | [x] | Preserve focus/caret in structured rule editor and make preview actionable | R-034 | `fix(#19): stabilize museum rule editor` |
| 7 | R-033 | [#33](https://github.com/vm75/virtual-art-gallery/issues/33) | [x] | Bound total upload requests and decoded-image pixel/memory use | R-034 | `fix(#33): bound upload resource usage` |
| 8 | I-021 | [#21](https://github.com/vm75/virtual-art-gallery/issues/21) | [x] | Cover all museum JS modules/pure logic and restore container build gating | I-013, I-017, I-019 | `ci(#21): cover museum modules and container build` |
| 9 | I-022 | [#22](https://github.com/vm75/virtual-art-gallery/issues/22) | [x] | Re-run security/accessibility/performance audit after concrete fixes | I-013, I-015, I-017, I-018, I-019, R-033, I-021 | `docs(#22): record post-fix hardening audit` |
| 10 | I-024 | [#24](https://github.com/vm75/virtual-art-gallery/issues/24) | [ ] | Fresh-install/end-to-end release validation including actual WebGL | all above | prior acceptance invalidated |
| 11 | I-001 | [#1](https://github.com/vm75/virtual-art-gallery/issues/1) | [ ] | Final v1 release acceptance gate | I-024 | prior acceptance invalidated |

## Second-review findings that invalidated the prior release gate

- Default museum camera orientation can produce a degenerate view matrix.
- Camera movement has no wall/doorway collision constraints.
- Generated rooms are separated by a gap with no connecting floor/corridor geometry.
- Active museum textures can be repeatedly deleted and re-uploaded to the GPU during camera movement.
- Museum rule-editor inputs are replaced on each input event, causing focus/caret loss.
- Museum preview exposes aggregate counts instead of actionable unclassified artwork/layout-error details.
- CI omits new museum renderer/camera/admin modules and the current green quality run does not include a container build job.
- Upload parsing does not cap the total request body before multipart parsing, and decoded image pixel/memory use is too permissive.
- Prior browser smoke used WebGL-unavailable fallbacks, so actual renderer behavior was not sufficiently validated.

## Third-review follow-up

A later implementation review found four remaining museum correctness gaps after the previous hardening claims:

- the camera basis was orthonormal but vertically reflected because its up vector used the wrong cross-product order;
- collision checked only the movement endpoint, allowing a diagonal step to tunnel through a doorway jamb when the endpoint happened to be valid;
- a canvas tap ignored tap coordinates and opened the nearest visible artwork rather than the artwork plane actually tapped;
- native touch gestures could take over canvas drag-to-look behavior.

Focused fixes now exist in `4a8935a` and `e0c9578`. PR #35 quality run `34167759026` passed formatting, vet, full Go tests, production build, syntax checks for every shipped browser module, every pure museum browser test, and the OCI container build at head `e0c9578`. #13 and #18 remain open until the required actual-WebGL renderer/desktop/mobile interaction smoke is rerun on the new behavior. #24 and #1 remain open behind those validations.

## Required final validation

Before #24/#1 can close, demonstrate at minimum:

- formatting, `go vet ./...`, `go test ./...`, production build, all shipped browser-module checks/tests, and successful CI on a PR targeting `main`;
- successful OCI/container build/run, persistence, health, backup and restore smoke;
- admin setup/login/logout/throttle/CSRF plus bounded upload request and decoded-image failure cases;
- artwork create/edit/visibility/alt text/filter/canonical-page/non-ASCII slug behavior;
- Gallery Lite and Timeline keyboard/touch/mobile/reduced-motion behavior;
- structured museum rule editing with normal multi-character typing, actionable preview, failed-publish isolation, and deterministic publish;
- deterministic room/corridor layout with physical reachability and unique renderer-ready placements;
- actual WebGL rendering with a valid right-handed initial camera, walls/floors/connectors/artworks visible, collision-aware desktop/mobile navigation, canvas artwork hit-testing, and artwork inspection;
- bounded room-aware GPU texture residency without repeated uploads during stationary-room navigation;
- security/accessibility/performance re-audit with no critical/high or acceptance-blocking finding;
- documentation audit confirming `main` is canonical, `legacy` is reference-only, and deleted `rewrite` is not an active target.

## I-001 final acceptance evidence

| Criteria | Current evidence | Result |
|---:|---|---|
| 1 | README Go run instructions; clean local production build and full gate. | Pass |
| 2–3 | OCI build, direct rootless run, and rootless `podman compose` service smoke; named-volume recreation retained `gallery.db`. | Pass |
| 4 | Admin setup/authentication and single-admin HTTP tests; fresh-data setup smoke. | Pass |
| 5 | Public navigation/mobile audit found no admin link. | Pass |
| 6 | Password/session/CSRF/header and upload-bound tests; I-022 audit. | Pass |
| 7–8 | Fresh admin upload/edit/visibility smoke; derivative/image behavior tests. | Pass |
| 9–10 | Gallery and Timeline keyboard/touch/mobile/reduced-motion coverage from I-022/I-024. | Pass |
| 11–12 | New right-handed camera, swept collision and precise artwork-picking tests pass; actual-WebGL desktop/mobile re-smoke on `e0c9578` is still required. | Pending |
| 13 | Draft editor focus, preview, failed-publish isolation, and publish tests/smoke. | Pass |
| 14–16 | Canonical detail, responsive/focus/accessibility, and visual-system audit coverage. | Pass |
| 17–18 | PR #35 run `34167759026` passed full Go/JS checks, production build, and OCI container build; graceful shutdown and health tests/smokes already recorded. | Pass |
| 19 | Documentation audit; active branch contract synchronized. | Pass |
| 20 | #13, #18, #24 and #1 are open pending actual-WebGL revalidation; no release-gate completion claim is current. | Pending |

## I-022 audit findings

| Area | Evidence | Result |
|---|---|---|
| Security/data | Public repositories use visible-only queries; HTML output escapes metadata; CSRF/session tests, upload-bound tests, headers, cache tests, and media traversal check pass. | No critical/high finding. |
| Accessibility | Mobile browser smoke covered public routes, focusable controls, no public admin link, WebGL museum controls; rule-editor focus smoke passed. | No acceptance-blocking finding recorded; final museum interaction re-smoke remains under #18/#24. |
| Performance | Responsive lazy derivatives, bounded texture lifecycle test, static revalidation, immutable generated media, and OCI build passed. | No acceptance-blocking finding. |

## Execution log

Append entries; do not erase hardening history.

```text
YYYY-MM-DD ID [~|x|!] — issue #N — commit SHA — tests/evidence — concise note
```

2026-09-07 SECOND-REVIEW — independent review reopened #1, #13, #15, #17, #18, #19, #21, #22, #24; created #33 and #34; created `fix/v1-release-hardening` from the reviewed `fix/v1-remediation` head; no further implementation work belongs on `fix/v1-remediation` or deleted `rewrite`.
2026-09-07 R-034 [x] — issue #34 — this issue-scoped documentation commit — replaced the prior completion ledger with this hardening tracker and canonical branch contract.
2026-09-07 I-013 [x] — issue #13 — `fix(#13): correct museum camera view matrix` — `node web/static/museum-renderer.test.mjs`; syntax checks for all shipped browser modules; `go vet ./...`; `go test ./...`; production build; live seeded museum smoke through Playwright + Xvfb reports WebGL true, fallback false, one rendered room, and one rendered artwork (screenshot inspected). Corrected default-orientation basis, added deterministic representative-angle coverage, and oriented the initial spawn toward the first artwork. README/ARCHITECTURE remain accurate; CI syntax coverage now includes camera and renderer modules.
2026-09-07 I-015 [x] — issue #15 — `fix(#15): connect generated museum rooms` — deterministic corridor contract joins each doorway plane; physical-plan validation rejects gaps/misalignment and unreachable rooms; renderer emits connector floor/side walls. `go vet ./...`; `go test ./...`; production build; all browser-module syntax checks; `node web/static/museum-renderer.test.mjs`; `git diff --check`; live two-room Playwright + Xvfb smoke reports WebGL true, fallback false, rooms 2, connectors 1, artworks 2. Architecture documented.
2026-09-07 I-018 [x] — issue #18 — `fix(#18): constrain museum navigation` — pure room/doorway/corridor collision module used by the shared camera primitive for keyboard and touch controls; tests cover spawn, wall/corridor boundaries, doorway/corridor traversal, and blocked camera movement. `go vet ./...`; `go test ./...`; production build; all browser-module syntax checks; both museum module tests; `git diff --check`; live WebGL desktop smoke blocks wall and traverses from room one through corridor into room two, and mobile movement blocks the same wall. Architecture documented.
2026-09-07 I-017 [x] — issue #17 — `fix(#17): reuse resident museum textures` — CPU image cache and GPU residency now tracked separately, preventing uploads/deletes for stable active targets while unloading displaced targets. Deterministic three-room lifecycle test asserts bounded initial uploads, no same-room re-upload, bounded transition residency, release, and disposal. `go vet ./...`; `go test ./...`; production build; all browser-module syntax checks; all museum module tests; `git diff --check`; Architecture documented.
2026-09-07 I-019 [x] — issue #19 — `fix(#19): stabilize museum rule editor` — text edits now mutate draft state in place and preserve focus/caret; structural changes rerender explicitly. Preview lists unclassified artwork name/slug and layout errors accessibly. Focused admin test covers details; Playwright + Xvfb smoke typed `Contemporary` with focus retained, added a group, saved, and saw unclassified artwork detail. `go vet ./...`; `go test ./...`; production build; all browser-module syntax checks; all museum module tests; `git diff --check`; Architecture documented.
2026-09-07 R-033 [x] — issue #33 — `fix(#33): bound upload resource usage` — total multipart request capped at 21 MiB before parse, temporary multipart files removed, and image config rejects more than 40 million decoded pixels before full decode while retaining size/format/dimension checks. Tests cover controlled 413 oversized valid multipart/no artwork row, decoded-pixel rejection and boundary, and normal upload. `go vet ./...`; `go test ./...`; production build; all browser-module syntax checks; all museum module tests; `git diff --check`; README/Architecture documented.
2026-09-07 I-021 [x] — issue #21 — `ci(#21): cover museum modules and container build` — CI dynamically syntax-checks every shipped static JS module, runs all pure browser tests, and has a separate OCI build job on main push/PR. Local equivalent passed formatting, vet, full Go tests, production build, all JS syntax/tests, diff check; rootless Podman build produced `localhost/virtual-art-gallery:ci`. README/AGENTS synchronized.
2026-09-07 I-022 [x] — issue #22 — `docs(#22): record post-fix hardening audit` — reviewed auth/CSRF/header/upload/data/escaping/cache paths; mobile browser audit verified public routes have no admin link, focusable controls, actual WebGL museum controls, and fallback absence; static revalidation/media traversal checks pass. Findings table records no critical/high or acceptance-blocking defect. Full Go/JS/build/diff gates pass.
2026-09-07 I-024 [x] — issue #24 — `docs(#24): record fresh-install release validation` — historical acceptance later invalidated by third-review museum findings; keep this entry as history only.
2026-09-07 I-001 [x] — issue #1 — `docs(#1): record final v1 acceptance gate` — historical acceptance later invalidated by third-review museum findings; keep this entry as history only.
2026-09-07 THIRD-REVIEW — reopened #13, #18, #24 and #1 after finding a vertically reflected camera basis, endpoint-only collision tunneling, nearest-artwork canvas taps, and native touch-gesture interference. #15, #17, #19, #21 and #33 remain closed.
2026-09-07 I-013 [~] — issue #13 — `4a8935a` — corrected the up-vector cross-product order and added right-handedness regression coverage (`right × up == back`); PR #35 run `34167759026` passes full test/JS/build gates. Actual-WebGL renderer re-smoke on the new head remains required before closure.
2026-09-07 I-018 [~] — issue #18 — `e0c9578` — movement samples the swept path at <=5 cm, focused tests block diagonal doorway-jamb tunneling while allowing valid doorway traversal; new pure ray/plane picking tests verify tap-target selection and empty-space no-op; canvas disables native touch gestures for drag-look. PR #35 run `34167759026` passes full test/JS/build gates. Actual-WebGL desktop/mobile interaction re-smoke remains required before closure.
2026-09-07 I-024 [ ] — issue #24 — reopened; pending successful actual-WebGL renderer/navigation/picking re-smoke after #13/#18 follow-up fixes.
2026-09-07 I-001 [ ] — issue #1 — reopened; final v1 release gate remains pending #24.
