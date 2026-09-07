# Autonomous Implementation Issue Tracker

## Repository issue status

**BLOCKED (repository setting): GitHub Issues are currently disabled for this repository.**

The GitHub API returned HTTP 410 when attempting to create the v1 epic. The connected GitHub tool does not expose the repository setting required to enable Issues. Until Issues are enabled, this file is the authoritative executable backlog. Every entry below is deliberately written as an issue-ready specification and can be copied to GitHub Issues without redesign.

When GitHub Issues are enabled:

1. Create one GitHub issue per `I-xxx` entry using the exact title/scope/acceptance criteria.
2. Add the resulting `#number` beside the tracker ID.
3. Keep this file as the agent execution ledger; GitHub Issues remain the discussion/history surface.

## Status legend

- `[ ]` ready/not started
- `[~]` in progress
- `[x]` complete and acceptance criteria verified
- `[!]` blocked; record reason under the item

Agents must update status, completion commit/PR, validation commands, and newly discovered follow-up work. Do not mark an item complete solely because code was written.

## Ordered backlog

| ID | Status | Work item | Depends on |
|---|---|---|---|
| I-001 | [ ] | V1 rewrite scope and release acceptance gate | planning baseline |
| I-002 | [ ] | Bootstrap Go app, config, health, shutdown, tests | I-001 |
| I-003 | [ ] | SQLite store and migration runner | I-002 |
| I-004 | [ ] | Artwork domain, persistence, and public read API | I-003 |
| I-005 | [ ] | Image upload/storage/derivative pipeline | I-003, I-004 |
| I-006 | [ ] | Single-admin first-use setup, login, sessions, CSRF | I-003 |
| I-007 | [ ] | Admin artwork create/edit/upload/visibility UI | I-004, I-005, I-006 |
| I-008 | [ ] | Tags, surfaces, mediums metadata management and filters | I-004, I-006 |
| I-009 | [ ] | Shared visual tokens, responsive shell, public home | I-002 |
| I-010 | [ ] | Canonical artwork detail pages and deep-link contract | I-004, I-005, I-009 |
| I-011 | [ ] | Gallery Lite responsive frontend | I-004, I-005, I-009, I-010 |
| I-012 | [ ] | Horizontal chronological Timeline frontend | I-004, I-005, I-009, I-010 |
| I-013 | [ ] | Museum renderer baseline adapted selectively from legacy | I-002, I-005 |
| I-014 | [ ] | Museum rule schema, validation, grouping contract | I-004, I-008 |
| I-015 | [ ] | Deterministic dynamic room/layout generator | I-014 |
| I-016 | [ ] | Explicit metadata-driven artwork placement | I-013, I-015 |
| I-017 | [ ] | Room-aware museum texture loading/culling lifecycle | I-013, I-016 |
| I-018 | [ ] | Museum desktop/mobile controls and artwork inspection | I-010, I-013, I-016 |
| I-019 | [ ] | Admin museum rules draft/preview/publish workflow | I-006, I-014, I-015, I-016 |
| I-020 | [ ] | Portable Containerfile and Compose.yml | I-002, I-003, I-005, I-006 |
| I-021 | [ ] | CI and repeatable quality gates | I-002 and incremental features |
| I-022 | [ ] | Security, accessibility, and performance hardening | I-007, I-011, I-012, I-018, I-019 |
| I-023 | [ ] | Backup/restore and operational readiness | I-003, I-005, I-020 |
| I-024 | [ ] | End-to-end fresh-install acceptance and v1 release | all v1 items |

---

## I-001 — V1 rewrite scope and release acceptance gate

### Scope

Maintain the v1 boundary defined in `IMPLEMENTATION_PLAN.md`, keep KISS/YAGNI constraints visible, and act as the final release gate rather than an implementation bucket.

### Acceptance criteria

- [ ] All required I-002 through I-024 work is complete or explicitly removed from v1 by an intentional documented scope change.
- [ ] All 20 global acceptance criteria in `IMPLEMENTATION_PLAN.md` are checked with evidence.
- [ ] No critical/high security defect or acceptance-blocking defect remains.
- [ ] Fresh-install, persistent-data, public gallery, admin, museum, and container smoke results are recorded here.
- [ ] All required docs match shipped behavior.

---

## I-002 — Bootstrap Go app, config, health, shutdown, and test harness

### Scope

Create the minimal Go module and single-process HTTP application. Establish configuration parsing, logging, routing, graceful shutdown, and baseline test conventions without adding product features.

### Acceptance criteria

- [ ] `go run` starts an HTTP server using documented defaults and configurable listen address/data directory.
- [ ] Invalid configuration fails fast with actionable errors.
- [ ] `/healthz` returns a simple success response when process is live.
- [ ] SIGINT/SIGTERM triggers bounded graceful shutdown.
- [ ] `go fmt ./...`, `go vet ./...`, and `go test ./...` are documented and pass.
- [ ] Repository/package layout remains small; no framework/router dependency unless justified.
- [ ] README/ARCHITECTURE/AGENTS are updated with actual commands/config.

---

## I-003 — SQLite store and migration runner

### Scope

Add embedded persistence using SQLite in the configured data directory and a minimal forward migration mechanism.

### Acceptance criteria

- [ ] Uses a pure-Go SQLite driver unless a documented blocker proves CGO simpler/necessary.
- [ ] Database is created under the configured persistent data directory.
- [ ] Ordered migrations run automatically once and are safe on repeated startup.
- [ ] Migration version is recorded transactionally.
- [ ] Concurrent-start behavior does not corrupt the database.
- [ ] Tests cover fresh DB, repeat startup, and at least one migration application.
- [ ] No ORM is introduced.
- [ ] Architecture documents actual persistence/migration contract.

---

## I-004 — Artwork domain, persistence, and public read API

### Scope

Define the canonical artwork record and storage/retrieval behavior independent of UI.

Required metadata: name, date, tags, surface, medium. Also include stable ID/slug, visibility, image references/metadata needed by later issues, and timestamps where useful.

### Acceptance criteria

- [ ] Artwork can be created/read/updated at the service/store layer with required fields validated.
- [ ] Slugs/IDs are stable and collisions handled deterministically.
- [ ] Hidden/draft artwork is excluded from public reads by default.
- [ ] Public list/detail JSON endpoints expose documented fields and responsive image placeholders/contracts needed later.
- [ ] Filtering contract supports tags/surface/medium and date ordering without duplicating per-frontend APIs.
- [ ] Persistence and HTTP tests cover validation, visibility, filtering, not-found, and ordering.
- [ ] API/data model is documented in ARCHITECTURE.

---

## I-005 — Image upload, storage, and derivative pipeline

### Scope

Provide safe local artwork image storage under the application data directory. Retain original uploads and generate display derivatives for grid, timeline, museum, and detail usage.

### Acceptance criteria

- [ ] Upload validation checks configured size limit, actual decodable image content, supported format, and sane dimensions.
- [ ] Storage names are application-generated and cannot traverse directories or overwrite arbitrary files.
- [ ] Original dimensions/aspect ratio are recorded.
- [ ] At least thumbnail, medium/timeline, museum texture, and large/detail derivatives are produced or a documented smaller set demonstrably covers all uses.
- [ ] Derivative dimensions/formats preserve aspect ratio and avoid unnecessary full-original delivery.
- [ ] Failed processing leaves no half-committed artwork/image state.
- [ ] Image responses use useful content types/cache headers.
- [ ] Tests cover invalid content, oversize input, successful derivatives, cleanup/error behavior.
- [ ] README/ARCHITECTURE document storage and supported formats/limits.

---

## I-006 — Single-admin first-use setup, login, sessions, and CSRF

### Scope

Implement exactly one administrator account. `/admin/` on a fresh DB performs initial username/password setup; after setup, it becomes login/admin entry. No user management.

### Acceptance criteria

- [ ] Fresh install permits creation of one admin credential only.
- [ ] Concurrent/direct attempts cannot create a second admin.
- [ ] Password is stored only as a modern salted password hash with documented parameters.
- [ ] Login failure is generic and rate-limited or otherwise reasonably throttled without adding infrastructure.
- [ ] Session tokens are cryptographically random, expire, can be invalidated on logout, and are not stored/logged in plaintext where avoidable.
- [ ] Cookies use HttpOnly, appropriate SameSite, and Secure behavior under TLS/configured proxy model.
- [ ] State-changing browser requests are CSRF-protected.
- [ ] Public `/` and public navigation do not link to `/admin/`.
- [ ] Auth/security tests cover first use, duplicate setup prevention, login/logout, protected route, expiry/invalid session, CSRF rejection.
- [ ] No account list/create/delete/role UI or API exists.

---

## I-007 — Admin artwork create/edit/upload/visibility UI

### Scope

Build a lightweight responsive server-rendered admin workflow using the auth boundary. Admin can upload artwork, edit required metadata, and control visibility.

### Acceptance criteria

- [ ] Admin can upload an image and enter name/date/tags/surface/medium.
- [ ] Admin can edit existing artwork metadata without re-uploading the image.
- [ ] Admin can hide/publish artwork and public endpoints reflect the state.
- [ ] Validation errors preserve entered non-secret values and are understandable.
- [ ] List view works on narrow mobile and desktop layouts.
- [ ] Forms are keyboard usable with labels, visible focus, and accessible error association.
- [ ] Destructive actions require explicit confirmation and appropriate CSRF protection.
- [ ] No public page links to admin.
- [ ] Tests cover protected access, create/edit/visibility and validation paths.

---

## I-008 — Tags, surfaces, mediums metadata management and filtering

### Scope

Provide the minimum admin affordances and normalization needed to keep tags flexible while surface/medium values remain consistent enough for filtering and museum rules.

### Acceptance criteria

- [ ] Tags support multiple values per artwork with normalized duplicate handling.
- [ ] Surface and medium can be selected/entered through an admin workflow that prevents accidental case/spacing duplicates.
- [ ] Admin can introduce a new valid surface/medium without schema/code changes.
- [ ] Existing values are discoverable/autocompleted/selected rather than repeatedly retyped blindly.
- [ ] Public filtering returns correct visible artworks for tag/surface/medium combinations.
- [ ] Deleting/renaming metadata cannot silently orphan or corrupt artwork references.
- [ ] Tests cover normalization, duplicates, filtering, rename/delete policy.
- [ ] Avoid a generic taxonomy framework beyond these needs.

---

## I-009 — Shared visual tokens, responsive shell, and public home

### Scope

Create the lightweight visual foundation used by public pages and compatible admin components: CSS tokens, typography, spacing, surfaces, focus states, buttons/dialog/chips where actually needed, plus public landing page.

Theme direction is inspired at a high level by the Maren Lowe reference: editorial/image-led, generous whitespace, neutral surfaces, expressive typography/selective letter spacing, minimal chrome, subtle transitions. Use Material interaction/accessibility principles; do not copy the site.

### Acceptance criteria

- [ ] CSS custom properties define core color/type/spacing/radius/elevation/motion tokens.
- [ ] No heavy Material/component framework is required.
- [ ] `/` clearly offers Gallery Lite, Timeline, and Museum entries and does not expose admin.
- [ ] Layout is polished at representative narrow mobile and desktop widths.
- [ ] Interactive states include visible hover/focus/pressed/disabled treatment as relevant.
- [ ] Motion tokens respect reduced-motion preference.
- [ ] Theme supports artwork-first presentation and readable contrast.
- [ ] README documents public routes; no unnecessary design-system abstraction is introduced.

---

## I-010 — Canonical artwork detail pages and deep-link contract

### Scope

Provide semantic HTML artwork routes that remain useful without WebGL and serve as the canonical target from all three public galleries.

### Acceptance criteria

- [ ] `/artwork/{slug}` shows image, name, date, tags, surface, and medium for visible artwork.
- [ ] Hidden/nonexistent artwork does not leak metadata.
- [ ] Page is responsive, keyboard accessible, and uses an appropriate derived image/srcset.
- [ ] Canonical/share metadata is present where practical without adding external services.
- [ ] Tags/surface/medium link to documented public filter URLs where supported.
- [ ] Gallery/timeline/museum can link to the same artwork route.
- [ ] Tests cover visible, hidden, not-found and HTML escaping.

---

## I-011 — Gallery Lite responsive frontend

### Scope

Implement the fastest public gallery experience using semantic server-rendered markup plus minimal JS enhancement.

### Acceptance criteria

- [ ] Responsive image grid works from narrow phones to desktop without horizontal overflow.
- [ ] Uses derived thumbnails/responsive images and native lazy loading where appropriate.
- [ ] Supports tag/surface/medium filters and preserves/shareable URL state.
- [ ] Artwork opens in a lightweight lightbox/dialog or detail flow with next/previous navigation.
- [ ] Keyboard navigation, Escape/close, focus return, touch/swipe where useful, and direct artwork links work.
- [ ] No jQuery/nanogallery2 dependency.
- [ ] Meaningful browsing/detail links still work if enhancement JS fails.
- [ ] Large collection smoke test does not eagerly fetch all large/original images.

---

## I-012 — Horizontal chronological Timeline frontend

### Scope

Build a distinct timeline presentation ordered by artwork date with native horizontal scrolling and restrained enhancement animation.

### Acceptance criteria

- [ ] Artworks are ordered chronologically with clear date context.
- [ ] Desktop wheel/trackpad and direct drag/scroll are intuitive without globally hijacking browser input.
- [ ] Touch scrolling works natively on mobile with sensible snapping/card sizing.
- [ ] Center/near-center artwork may receive subtle scale/fade/parallax emphasis without obscuring art.
- [ ] Keyboard users can traverse artwork and activate canonical details.
- [ ] `prefers-reduced-motion` disables nonessential animation while preserving navigation.
- [ ] Uses appropriate derived image size and lazy loading.
- [ ] Empty/small/large collection states are handled.
- [ ] No heavy animation framework unless justified by measured need.

---

## I-013 — Museum renderer baseline adapted selectively from legacy

### Scope

Port/adapt only the useful low-level 3D concepts from the `legacy` branch into the rewrite-owned frontend structure. Establish a simple scene fed by rewrite data, not the old ARTIC/local API/build pipeline.

### Acceptance criteria

- [ ] Museum page can initialize REGL/WebGL and render a simple floor/walls plus at least one dynamically supplied artwork.
- [ ] Old ARTIC/local API selection and generated static image list are not part of the new data path.
- [ ] Renderer, camera/navigation, mesh drawing, painting drawing, and texture loading have clear rewrite-owned boundaries.
- [ ] License attribution/legacy provenance is preserved where copied code requires it.
- [ ] Museum code is loaded only on `/museum/`, not bundled into other galleries.
- [ ] Unsupported WebGL produces a useful fallback/link to Gallery Lite or artwork pages rather than a blank screen.
- [ ] Legacy code is not merged wholesale and Browserify-era build assumptions are not required unless temporarily documented.

---

## I-014 — Museum rule schema, validation, and grouping contract

### Scope

Define a small renderer-independent validated data model for museum rules and exhibition grouping. Supported conditions should cover tags, surface, medium, and simple date comparisons/ranges. Actions should map to named exhibition groups and a minimal set of presentation/layout hints.

### Acceptance criteria

- [ ] Rule data is serializable/versionable and contains no executable user code.
- [ ] Validation rejects unknown operators/fields, malformed values, duplicate/conflicting identifiers, and unsafe/unbounded inputs.
- [ ] Deterministic evaluation maps artworks to groups with documented priority/conflict behavior.
- [ ] Unmatched artworks are reported as `unclassified` rather than silently lost.
- [ ] Unit tests cover each condition type, multiple conditions, priority/conflict, no-match, invalid config.
- [ ] Rule engine has no WebGL dependency.
- [ ] ARCHITECTURE documents the actual rule schema and semantics.

---

## I-015 — Deterministic dynamic room/layout generator

### Scope

Convert exhibition groups and simple layout settings into a deterministic museum plan: rooms, connections/doorways, wall segments, spawn location, and available placement segments. The renderer should consume this plan; it should not invent semantic grouping.

### Acceptance criteria

- [ ] Same artwork/group/rule-set version and seed yields the same topology/placement IDs across runs.
- [ ] Room size/wall capacity scales from group artwork count and/or artwork aspect/physical dimensions where available, using simple documented rules.
- [ ] Every generated room is reachable from spawn; graph has no accidental isolated groups.
- [ ] Doorways/connections do not make required wall placements unusable without accounting for capacity.
- [ ] Layout validation reports insufficient capacity rather than dropping artworks.
- [ ] Output scene contract is testable as data without WebGL.
- [ ] Tests cover 0, 1, small, large, and uneven group distributions plus deterministic seed behavior.
- [ ] Do not implement arbitrary architectural CAD/general floor-planning features.

---

## I-016 — Explicit metadata-driven artwork placement

### Scope

Assign artworks to stable placement IDs within their exhibition rooms based on the generated plan. Replace the legacy sequential `placement[batch.length]` concept.

### Acceptance criteria

- [ ] Every visible museum artwork receives at most one explicit placement and every assignment references an existing room/wall placement.
- [ ] Group membership constrains placement to the intended exhibition room(s).
- [ ] Placement considers image aspect ratio and available wall width/height with documented spacing rules.
- [ ] Insufficient capacity is a validation error surfaced before publish/public render.
- [ ] Assignment is deterministic for the same inputs/seed.
- [ ] Scene JSON/data includes artwork ID -> placement transform/metadata mapping suitable for renderer consumption.
- [ ] Unit tests cover landscape/portrait/extreme aspect, capacity boundary, deterministic order, and group isolation.

---

## I-017 — Room-aware museum texture loading and culling lifecycle

### Scope

Evolve the legacy distance/index loading into room/spatial-aware dynamic loading appropriate for generated rooms and mobile GPU limits.

### Acceptance criteria

- [ ] High/normal museum textures are loaded for visible/current-near artwork on demand, not all at startup.
- [ ] Adjacent-room preload policy is bounded and documented.
- [ ] Distant textures can be unloaded and later recreated without losing artwork assignment/state.
- [ ] Texture size selection uses museum derivatives/device/network capability conservatively.
- [ ] Frustum/orientation culling continues to avoid needless draw work where practical.
- [ ] Memory/resource cleanup destroys or safely reuses WebGL textures; no unbounded cache growth in a navigation smoke test.
- [ ] Loader handles failed image requests with a stable placeholder/retry policy.
- [ ] Performance instrumentation used during development is not permanently noisy in production.

---

## I-018 — Museum desktop/mobile controls and artwork inspection

### Scope

Provide responsive exploration controls and an accessible bridge from 3D artwork to normal artwork information.

### Acceptance criteria

- [ ] Desktop supports intuitive keyboard movement plus mouse/pointer look with clear instructions/escape behavior.
- [ ] Mobile supports touch-friendly movement/look controls or a simpler guided navigation mode that does not require desktop pointer lock.
- [ ] Controls account for viewport resize/orientation and do not overlap critical artwork/info UI.
- [ ] Selecting/tapping an artwork opens a Material-inspired information surface showing required metadata and a canonical `/artwork/{slug}` link.
- [ ] Info UI is keyboard/focus accessible outside pointer-lock mechanics.
- [ ] Reduced-motion preference removes nonessential camera/UI animation where practical.
- [ ] WebGL/control failure leaves a visible route to Gallery Lite/details.
- [ ] Representative desktop and mobile smoke tests are documented.

---

## I-019 — Admin museum rules draft, preview, and publish workflow

### Scope

Build a focused admin rule editor over the validated rule schema. Maintain draft and published states so editing cannot break the public museum until explicit publish.

### Acceptance criteria

- [ ] Admin can add/edit/delete/reorder supported rules without editing raw executable code.
- [ ] UI exposes supported metadata fields/operators/values and validates errors before save/publish.
- [ ] Preview generates/uses draft grouping/layout and shows unclassified artworks plus capacity/layout validation errors.
- [ ] Public `/api/museum`/museum page uses only the published snapshot.
- [ ] Publishing records a stable version/seed/config so reloads are deterministic.
- [ ] Failed validation cannot replace the last valid published museum.
- [ ] Admin can see whether draft differs from published state.
- [ ] Tests cover draft isolation, validation failure, publish success, published retrieval, and deterministic version behavior.

---

## I-020 — Portable non-root Containerfile and Compose.yml

### Scope

Containerize the single Go application using portable OCI syntax and provide a Compose deployment usable with Docker Compose and rootless Podman-compatible Compose tooling.

### Acceptance criteria

- [ ] File is named `Containerfile`; build uses ordinary OCI/Dockerfile syntax without vendor-only directives.
- [ ] Final image contains only runtime necessities and runs as non-root.
- [ ] `Compose.yml` uses portable syntax, one application service, named persistent data volume, configurable published port, and no privileged/host-network/Docker-socket requirements.
- [ ] Container writes persistent DB/uploads only under documented data mount.
- [ ] Logs go to stdout/stderr and SIGTERM shuts down cleanly.
- [ ] Healthcheck uses documented application endpoint and tools actually present in final image (or Compose-level check is omitted with rationale).
- [ ] Docker build/compose smoke is recorded where available.
- [ ] Rootless Podman build/run/compose smoke is recorded where available; environment limitations are explicitly documented rather than guessed.
- [ ] README and DOCKERHUB match actual ports/env/volume/image commands.

---

## I-021 — CI and repeatable quality gates

### Scope

Add a small CI workflow and local commands that enforce the tests/lint/build guarantees already used by agents. Keep CI provider-specific logic limited to CI orchestration, not application runtime.

### Acceptance criteria

- [ ] CI runs formatting check, `go vet`, `go test`, and production build on pull/push workflow appropriate to repo.
- [ ] Frontend static/build tests are included only if they exist and add value; no Node ecosystem is introduced solely for linting trivial vanilla JS.
- [ ] Museum deterministic rule/layout tests run headlessly without WebGL.
- [ ] Container build check is added when execution cost/tooling is reasonable.
- [ ] Failures are actionable and local equivalent commands are in AGENTS/README.
- [ ] Dependency/cache actions are pinned or reasonably versioned and permissions are minimal.

---

## I-022 — Security, accessibility, and performance hardening

### Scope

Perform a deliberate cross-cutting audit after feature completion; fix concrete defects rather than adding speculative infrastructure.

### Acceptance criteria

- [ ] Authentication/session/CSRF/security headers/upload paths are reviewed against actual deployment model.
- [ ] No public route leaks hidden artwork or admin-only metadata.
- [ ] HTML output escapes untrusted metadata; JSON/content-type handling is correct.
- [ ] Keyboard/focus/labels/dialog behavior is verified for home, Gallery Lite, Timeline, artwork detail, admin forms, and museum info UI.
- [ ] Reduced-motion behavior is verified for timeline/museum/nonessential public motion.
- [ ] Representative mobile/desktop layout has no critical overflow/tap-target defects.
- [ ] Gallery/timeline do not eagerly fetch large/original images; museum texture memory remains bounded during a navigation smoke.
- [ ] Basic HTTP caching for static/derived immutable assets is sensible.
- [ ] Findings/fixes are recorded; unrelated wishlist items become follow-up issues, not scope creep.

---

## I-023 — Backup/restore and operational readiness

### Scope

Define and validate the simplest safe way to back up and restore the single persistent data directory/SQLite state, plus practical operator guidance.

### Acceptance criteria

- [ ] A documented backup procedure produces a restorable copy without recommending unsafe live SQLite copying.
- [ ] Restore to a fresh installation/container is tested with artwork metadata and uploaded images intact.
- [ ] Procedure covers version compatibility/migrations at a practical v1 level.
- [ ] Data directory ownership/permissions work with named volumes and rootless container runtime assumptions.
- [ ] Operator docs include health, logs, graceful stop, upgrade, backup, restore, and reset/fresh-install boundaries.
- [ ] No separate backup service is introduced.

---

## I-024 — End-to-end fresh-install acceptance and v1 release

### Scope

Run the complete product as a user/operator from a clean checkout and fresh volume, resolve acceptance blockers, then finalize v1 docs/release readiness.

### Acceptance criteria

- [ ] Build/run from README succeeds on clean environment assumptions.
- [ ] Fresh `/admin/` setup creates the only admin; logout/login works; no public admin link exists.
- [ ] Upload several artworks spanning different dates/tags/surfaces/mediums; edit metadata and visibility.
- [ ] Verify Gallery Lite filtering/lightbox/keyboard/mobile behavior.
- [ ] Verify Timeline chronology/scroll/touch/reduced-motion behavior.
- [ ] Create draft museum rules, verify preview/unclassified/capacity feedback, publish, and verify deterministic public museum grouping/layout.
- [ ] Navigate museum on desktop and mobile control model; verify dynamic texture lifecycle and artwork detail links.
- [ ] Restart/recreate container and confirm persistent state.
- [ ] Perform backup and restore smoke.
- [ ] Run all quality gates and container smoke tests.
- [ ] Explicitly check all global acceptance criteria in `IMPLEMENTATION_PLAN.md`.
- [ ] Documentation audit confirms README, ARCHITECTURE, AGENTS, DOCKERHUB, IMPLEMENTATION_PLAN, ISSUE_TRACKER match reality.
- [ ] Record final release commit/tag/image references when actually published.

---

## Agent execution log

Append concise entries here as work progresses. Suggested format:

```text
YYYY-MM-DD I-xxx [~|x|!] agent/actor — commit/PR — tests run — note/blocker
```

Do not erase historical blocker/fix entries; append corrections so later agents can understand what happened.
