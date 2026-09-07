# Implementation Plan

## Objective

Deliver a self-hosted virtual art gallery platform with one Go backend, three responsive public gallery frontends, and one private single-admin interface. The project must remain simple to operate, portable across Docker/rootless Podman, and maintainable by autonomous coding agents.

## Global scope

### In scope

- Go HTTP server and configuration.
- SQLite persistence and migrations.
- Artwork records with required metadata: name, date, tags, surface, medium.
- Image upload, validation, original retention, and derived display sizes.
- Public landing page with links to Gallery Lite, Timeline, and Museum only.
- Gallery Lite responsive grid/lightbox/filtering.
- Timeline horizontal chronological experience with responsive/touch/reduced-motion behavior.
- 3D museum adapted selectively from the `legacy` branch, with runtime artwork loading, deterministic dynamically generated rooms, metadata-driven grouping, explicit artwork placement, and desktop/mobile navigation.
- Canonical accessible artwork detail pages.
- Admin setup/login/logout; exactly one admin credential initialized on first use.
- Admin artwork upload/edit/visibility and metadata management.
- Admin museum rules editor with draft preview and publish flow.
- Platform-agnostic OCI `Containerfile` and `Compose.yml` compatible with Docker and rootless Podman.
- Health endpoints, graceful shutdown, persistent-volume contract, and operational documentation.
- Automated tests focused on domain behavior, persistence, HTTP/security boundaries, museum rule/layout determinism, and critical UI flows.
- Documentation kept synchronized with code.

### Explicitly out of scope for v1

- Multiple admins, roles, invitations, OAuth/SSO, password reset email.
- Comments, likes, social network features, ecommerce, payments, print ordering.
- Generic CMS/page builder.
- Cloud storage as a requirement.
- Kubernetes manifests, Helm charts, Terraform, or cloud-specific deployment.
- Arbitrary scripting/plugins for museum rules.
- Multiplayer/shared museum sessions.
- WebXR/VR.
- Realtime room streaming/generation while walking.
- Complex 3D sculpture import pipeline unless separately approved after v1.

## Global acceptance criteria

V1 is accepted only when all of the following are true:

1. A clean checkout can be built and run locally with documented Go commands.
2. The application can be built from `Containerfile` and launched with `Compose.yml` using Docker Compose and rootless Podman/Podman Compose semantics without privileged mode, Docker socket, or Docker-specific extensions.
3. Application state persists in one documented data volume across container replacement.
4. First visit to `/admin/` on a fresh data directory permits creation of exactly one admin username/password. After initialization, a second admin cannot be created through normal or direct HTTP flows.
5. Public home/navigation contains no admin link.
6. Admin authentication uses modern password hashing, secure session handling, CSRF protection for browser state changes, upload limits, and content validation.
7. Admin can upload an artwork and edit at least name, date, tags, surface, medium, and visibility.
8. Uploaded artwork is available through responsive derived images; public grid/timeline/museum do not require downloading full originals for normal display.
9. `/gallery/` is usable with keyboard, touch, and desktop pointer; supports lazy loading and artwork detail/lightbox behavior.
10. `/timeline/` is chronological, horizontally navigable on desktop/mobile, does not trap ordinary browser interaction, and respects `prefers-reduced-motion`.
11. `/museum/` loads artwork dynamically, generates a deterministic room/layout plan from published metadata/rules, explicitly assigns artworks to placements, and supports desktop plus touch-friendly mobile navigation.
12. Museum texture lifecycle avoids retaining all high-resolution textures indefinitely; distant work can be unloaded/preloaded according to a documented policy.
13. Museum rules can group by supported metadata (tags/surface/medium/date conditions), be edited as draft, previewed, and published. Published public museum state is not changed merely by editing a draft.
14. Every public artwork has a canonical semantic HTML detail route showing required metadata and usable without WebGL.
15. Public UI is responsive at narrow mobile and modern desktop sizes; core interactions have visible focus and meaningful accessible names.
16. Theme is modern, restrained, image-first, and Material-inspired, taking high-level cues from the Maren Lowe reference (generous whitespace, editorial typography, neutral surfaces, subtle motion) without copying assets/layout wholesale.
17. Core Go tests and museum rules/layout determinism tests pass in CI/local documented commands.
18. Application shuts down cleanly on SIGTERM and exposes documented health behavior.
19. `README.md`, `ARCHITECTURE.md`, `AGENTS.md`, `DOCKERHUB.md`, `IMPLEMENTATION_PLAN.md`, and `ISSUE_TRACKER.md` accurately reflect the delivered system.
20. No known critical/high security defect or acceptance-blocking issue remains open for the v1 milestone/tracker.

## Delivery phases

### Phase 0 — Planning and guardrails

Establish clean rewrite branch, preserve legacy branch, architecture, agent workflow, issue tracker, and detailed GitHub backlog. No product code.

### Phase 1 — Backend foundation

Create minimal Go module/application lifecycle, configuration, HTTP routing, structured logging, graceful shutdown, health route, SQLite connection, and migrations. Establish test conventions before feature growth.

### Phase 2 — Artwork and image core

Implement artwork persistence/model/API, tags/surface/medium semantics, image storage, upload validation, stable filenames, image metadata, and derived sizes. Add canonical artwork detail route.

### Phase 3 — Single-admin control plane

Implement first-run admin setup, authentication/session/CSRF boundaries, then responsive admin UI for upload/edit/visibility and metadata. No user-management concepts.

### Phase 4 — Shared visual foundation + Gallery Lite

Create design tokens, page shell, responsive typography/components, public landing page, and Gallery Lite. Use Gallery Lite to validate the complete data/image/API pipeline and accessibility baseline.

### Phase 5 — Timeline

Build the chronological horizontal experience using native scrolling and progressive animation, then test keyboard/touch/reduced motion and large collections.

### Phase 6 — Museum renderer baseline

Study/copy only needed pieces from `legacy`; isolate REGL/WebGL rendering, camera, mesh and texture loading behind rewrite-owned interfaces. First milestone renders a simple generated scene with dynamic artwork data, not hard-coded ARTIC/local lists.

### Phase 7 — Rules-driven dynamic museum

Define a small validated museum rule model, exhibition grouping, deterministic layout/room generation, explicit placement mapping, collision/spatial queries, room-aware texture loading, responsive controls, and artwork inspection UI.

### Phase 8 — Admin museum workflow

Add rule editing, validation, draft preview, published snapshot/version/seed, and safe publish. Show unclassified artworks and layout validation errors before publish.

### Phase 9 — Containerization and operations

Add non-root `Containerfile`, portable `Compose.yml`, persistent volume, environment contract, healthcheck, image metadata/labels where useful, Docker Hub instructions, and rootless Podman verification.

### Phase 10 — Hardening and release

Cross-cutting accessibility, security, performance, backup/restore documentation, smoke/e2e tests, CI, docs audit, fresh-install acceptance run, and v1 release checklist.

## Sequencing rules for autonomous agents

- Foundation/persistence issues precede feature issues that store data.
- Authentication must precede admin mutation UI.
- Artwork/image core must precede public galleries.
- Shared visual foundation precedes UI polish, but page-specific behavior should not wait for a large design-system project.
- Museum renderer baseline precedes rules-driven layout.
- Rule model/layout must be testable without WebGL before admin editor work begins.
- Containerization can begin once the Go process/data-directory contract exists, then be finalized during release hardening.
- If an issue reveals a required prerequisite, create a focused issue and update `ISSUE_TRACKER.md`; do not silently implement a second project.

## Architectural constraints

- Single Go process and one persistent data directory for v1.
- Server-rendered HTML + vanilla JS/CSS by default.
- No ORM unless raw SQL demonstrably becomes the larger complexity.
- No SPA framework unless an approved issue records why native/server-rendered UI is insufficient.
- No arbitrary museum rule scripts; rules are validated data.
- Museum semantic layout code must be renderer-independent and deterministic.
- `legacy` is a source/reference branch, not the base tree for the rewrite.

## Documentation completion rule

Every issue's acceptance criteria implicitly include documentation synchronization according to `AGENTS.md`. A feature is not complete if users/operators/agents would receive stale instructions after the change.
