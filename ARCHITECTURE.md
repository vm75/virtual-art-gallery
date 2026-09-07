# Architecture

## Goals

Build one self-hosted application that owns artwork metadata, image storage, admin authentication, museum rules, and three independent public gallery presentations. Keep the system simple enough to run as a single container with one persistent volume.

## Non-goals for v1

- Multi-user or role-based administration.
- External identity providers.
- Distributed services, message queues, Kubernetes, object-storage requirements, or microservices.
- A general-purpose CMS.
- Real-time collaborative editing.
- Arbitrary scripting in museum rules.
- Runtime streaming of newly generated room geometry while walking; v1 generates the museum plan at load/publish time and streams artwork textures as needed.

## Process model

A single Go process serves HTML, JSON, static assets, uploaded/derived media, admin endpoints, and health endpoints.

```text
Browser
  |-- /, /gallery/, /timeline/, /artwork/*  -> public HTML + assets
  |-- /museum/                              -> museum HTML + WebGL assets
  |-- /api/*                                -> public read JSON
  |-- /admin/*                              -> authenticated HTML/JSON
  v
Go application
  |-- HTTP/router
  |-- artwork service
  |-- image service
  |-- museum rules/layout service
  |-- admin auth/session service
  v
SQLite + data directory
```

The current bootstrap process is `cmd/gallery`: it uses `net/http` with standard-library routing, reads `GALLERY_LISTEN_ADDR` (default `:8080`) and `GALLERY_DATA_DIR` (default `./data`), opens `gallery.db` in that data directory, exposes `GET /healthz`, and shuts down on SIGINT/SIGTERM with a ten-second deadline. Feature routes are added incrementally behind this single process.

The admin boundary currently supports first-use setup at `/admin/setup`, login at `/admin/login`, authenticated `/admin/`, and POST logout. Passwords use bcrypt; sessions store only SHA-256 token digests with a 12-hour expiry, and a separate CSRF token is required for logout. Login failures are throttled in-process by the direct peer IP (not forwarded headers), with expired entries removed and the key map bounded. `GALLERY_SECURE_COOKIES=true` enables the Secure cookie flag for TLS deployments.

The current artwork admin UI uses `/admin/artworks/new` for multipart creation and `/admin/artworks/edit?slug=...` for metadata/visibility updates. Upload processing completes before image references are persisted; failed creates delete the draft record.

Museum rules are persisted as one draft row and one published row. `/admin/museum` validates and saves draft JSON; `/admin/museum/publish` explicitly promotes it. `GET /api/museum` evaluates only the published snapshot against visible artworks and returns its version plus deterministic layout; an absent published snapshot is a public 404/fallback.

Surfaces and mediums are small normalized value tables, seeded with common values and extended on valid artwork creation/update. Tags remain flexible normalized strings joined to artworks; filtering uses the same public API for every frontend.

Public artwork JSON is served from `GET /api/artworks` and `GET /api/artworks/{slug}`. List filters are optional `tag`, `surface`, and `medium` query parameters; `order=asc` sorts oldest first, while the default is newest first. Only visible records are returned publicly. Each response includes an `image` object reserved for derivative URLs added by the image pipeline.

JPEG derivatives are served below `/media/` with immutable cache headers. The media handler rejects traversal and original-image paths; original files remain application storage inputs rather than public display resources.

All routes pass through security headers: restrictive same-origin CSP, `nosniff`, strict cross-origin referrer policy, and disabled unnecessary browser permissions. Stable static asset URLs require revalidation on every use so deployments update reliably; generated derivatives use long-lived immutable caching. Original uploads are never a public media response.

## Proposed repository layout

```text
cmd/gallery/              application entrypoint
internal/app/             wiring and lifecycle
internal/config/          env/config parsing
internal/httpx/           shared HTTP helpers/middleware
internal/artwork/         artwork domain and persistence
internal/images/          upload validation, storage, derivatives
internal/auth/            first-run setup, password/session handling
internal/museum/          rules, grouping, deterministic layout contracts
internal/store/           SQLite connection and migrations
web/templates/            server-rendered templates
web/static/               shared CSS/JS/assets
web/gallery/              gallery-lite JS/CSS
web/timeline/             timeline JS/CSS
web/museum/               adapted REGL/WebGL renderer and controls
web/admin/                admin-specific CSS/JS/templates
migrations/               ordered SQL migrations
```

This is a direction, not a mandate to create packages before they are needed.

## Persistence

Use SQLite in the configured data directory through the pure-Go `modernc.org/sqlite` driver, so the default build does not require CGO. The application creates `<data-dir>/gallery.db`; every pooled connection receives the SQLite `foreign_keys=1` and `busy_timeout=5000` pragmas through the driver's connection DSN. Embedded numbered SQL migrations are claimed and recorded in `schema_migrations` transactionally at startup, and concurrent/repeated starts are safe.

Core records:

- `artworks`: id, slug, name, date, surface, medium, editable bounded alt text, visibility, image metadata, timestamps. Existing slugs are stable; new names that cannot form an ASCII slug receive a deterministic URL-safe digest fallback.
- `tags` and artwork-tag join table.
- controlled `surfaces` and `mediums`, unless a simpler normalized-string approach meets the issue acceptance criteria.
- `admin`: exactly one credential record.
- `sessions`: opaque admin sessions with expiration/revocation.
- `museum_rule_sets`: draft/published rule configuration and deterministic seed/version.

Do not add generalized entity systems, plugin schemas, or event sourcing.

## Artwork image pipeline

Original uploads are retained. Derived images are generated for distinct use cases: thumbnail/grid, medium/timeline, museum texture, and large/detail. Exact dimensions/formats should be chosen in the implementation issue based on browser support and library simplicity.

Rules:

- Validate MIME/content, dimensions, and upload size before persistence.
- Generate stable, non-user-controlled storage names.
- Store image width/height/aspect ratio.
- Do not serve original file paths supplied by users.
- Public pages use responsive derived images rather than downloading originals unnecessarily.
- 3D museum textures are sized for GPU use and loaded/unloaded according to proximity/visibility.

The image pipeline accepts decodable JPEG, PNG, and GIF uploads up to 20 MiB and 16,000 pixels per dimension by default. It retains the original and writes application-generated immutable JPEG derivatives, preserving aspect ratio without upscaling: thumbnail up to 480×480, medium up to 1200×1200, museum up to 2048×2048, and large up to 2400×2400. Bilinear resampling improves display quality while bounding portrait and landscape output. Processing happens in a temporary directory and failed processing removes it; public consumers use derivatives rather than originals.

## Public frontend architecture

Favor server-rendered semantic HTML with vanilla ES modules for interaction. Shared design tokens and tiny reusable CSS components are preferred over a large component framework.

Theme direction: editorial and image-led, inspired by the referenced Maren Lowe site: generous negative space, neutral surfaces, strong typography with selective tracking, minimal chrome, large imagery, quiet transitions, and clear hierarchy. Material principles should govern interaction states, accessibility, elevation/surfaces, form controls, dialogs, focus, and responsive behavior. Do not clone the reference site.

### Gallery Lite

Fastest and most accessible experience. Responsive grid, lazy loading, filters, lightbox/detail interaction, keyboard support, touch support, and shareable URLs.

### Timeline

Chronological horizontal experience. Native scroll remains the primary mechanic. Enhance with snapping, center emphasis, subtle parallax/scale/fade, touch dragging, keyboard navigation, and `prefers-reduced-motion` behavior. Avoid scroll-jacking that prevents normal browser interaction.

The current timeline is server-rendered at `/timeline/`, orders visible artworks oldest-first, uses medium derivatives with lazy loading, and applies CSS scroll snapping plus an optional IntersectionObserver emphasis class. Native horizontal scrolling remains available when JavaScript is absent.

### Museum

The `legacy` branch is reference material. Preserve useful low-level concepts such as REGL rendering, generated wall meshes, first-person camera, culling, and texture lifecycle where they remain appropriate.

Refactor the museum into clear layers:

```text
artwork metadata
   -> published museum rules
   -> exhibition groups
   -> deterministic room/layout plan
   -> wall/placement scene description
   -> WebGL/REGL renderer
```

The semantic rule/layout layer must not depend on WebGL so it can be unit tested.

Rules operate on supported metadata (`tags`, `surface`, `medium`, date ranges) and map artworks into named exhibition groups plus simple layout/presentation parameters. No arbitrary code execution or user-authored script language.

The rule contract is versioned JSON: `{version, seed, groups:[{id,name}], rules:[{id,priority,all:[{field,op,value/from/to}],group}]}`. Fields are `tag`, `surface`, `medium`, and `date`; text operators are `equals`/`contains`, and date operators are `before`/`after`/`between`. Rules are evaluated by descending priority and declaration order; the first matching rule wins. Invalid identifiers, unknown operators/fields, empty or oversized values, duplicate IDs, reversed date ranges, and unknown groups are rejected. Works matching no rule appear in `unclassified`.

The generated museum is deterministic for a given published rule-set version/seed and artwork set.

The `/museum/` page loads only its page module and its rewrite-owned ES modules: `museum.js` wires public APIs and accessible controls, `museum-camera.js` owns world-coordinate view movement, `museum-renderer.js` turns room/doorway/placement data into WebGL floor, wall, doorway, and artwork-plane geometry, and `texture-loader.js` owns image lifetime. The renderer consumes explicit layout transforms rather than artwork list order. Browsers without WebGL, or without a published scene, receive an explicit Gallery Lite fallback. This direct browser-WebGL implementation has no legacy build pipeline or ARTIC data source.

The renderer-independent layout generator sorts group IDs and artwork slugs, creates one connected linear room per group plus an optional unclassified room, and emits stable IDs and world-space room centers, doorways, wall normals, and placement transforms (position, normal, and Euler rotation). Each room has two usable artwork walls with four three-metre slots per wall segment; doorways occupy the east/west connection faces, so capacity exactly matches the emitted slots. Rooms scale to ten segments (80 works); excess work is reported in `errors` rather than dropped. Publish and public-scene generation reject capacity, assignment, group-isolation, or physical-placement validation errors.

Placement sizing uses a 2.4-unit target height and a 2.8-unit maximum width, preserving recorded image aspect ratio. `ValidatePlacements` rejects unknown rooms, duplicate artwork assignments, missing assignments, and references outside the evaluated artwork set before a scene is rendered or published.

Museum texture loading is bounded by the rewrite-owned `MuseumTextureLifecycle` (six entries). It loads only the current room's nearest artworks plus visible works in directly connected rooms; no farther rooms are preloaded. It selects the museum derivative normally and the smaller medium derivative for Save-Data or devices reporting 2 GiB memory or less. Leaving that target set releases both cached image references and renderer-owned WebGL textures; artwork placement metadata remains intact, so re-entering a room recreates textures on demand. Renderer draw work also applies distance/facing culling. Failed requests retry once and then use a stable placeholder; page teardown clears image and WebGL resources.

Museum controls expose keyboard/WASD movement, pointer drag look, and visible directional buttons suitable for touch. Escape releases canvas focus; resize/orientation redraws retain the world-coordinate view. A click/tap without a drag selects the nearest facing rendered artwork, while the accessible artwork list provides an equivalent keyboard path. Both open a native dialog with required metadata and a canonical detail link; the WebGL-unavailable state keeps Gallery Lite links visible.

Artwork assignment must be explicit; do not rely on API order mapping artwork N to placement N.

V1 generates the complete room topology before exploration. Texture loading remains dynamic: current room gets appropriate resolution, adjacent rooms may preload, distant textures unload.

## Admin architecture

`/admin/` is intentionally undiscoverable from the public UI but must still be fully secured because obscurity is not authentication.

First-run flow:

1. If no admin credential exists, `/admin/` shows setup.
2. User chooses username and password.
3. Password is stored only as a modern password hash with per-password salt and sane parameters.
4. Once initialized, setup cannot create a second account.

No user list, invitations, roles, password sharing workflows, or account creation API.

Admin functions: login/logout, artwork upload/edit, metadata/taxonomy editing, visibility control, museum rule editing, draft preview, and publish.

Use secure cookie sessions, CSRF protection for state-changing browser requests, upload limits, content validation, and conservative defaults. Deployment behind TLS reverse proxy must be supported.

## HTTP/API boundaries

Keep public API small and resource-oriented. Exact endpoints can evolve, but likely shapes are:

- `GET /api/artworks`
- `GET /api/artworks/{slug}`
- `GET /api/tags`, `/api/surfaces`, `/api/mediums`
- `GET /api/museum` for published museum scene/configuration
- authenticated `/admin/api/...` only where HTML forms are insufficient
- `GET /healthz` for liveness and `GET /readyz` if readiness is materially distinct

Do not create separate duplicated artwork APIs for each frontend.

## Container/deployment contract

Ship a platform-agnostic `Containerfile` using standard OCI syntax and a `Compose.yml` that works with both Docker Compose and Podman Compose/rootless Podman without vendor-only extensions.

Requirements:

- Multi-stage build where useful.
- Final container runs as a non-root user.
- Listen address/port are configurable.
- Persistent application data lives under one documented mount point (for example `/data`).
- No requirement for privileged mode, host networking, fixed UID 0, or Docker socket access.
- Clean SIGTERM shutdown.
- Healthcheck behavior documented.
- Image is buildable for common Linux architectures supported by Go.

## Documentation invariant

A change is incomplete if it changes an external contract but leaves documentation stale. The same pull request/commit must update README/architecture/agent/deployment/tracker docs as relevant.
