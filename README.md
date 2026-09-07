# Virtual Art Gallery — Rewrite

This branch is a clean rewrite of `vm75/virtual-art-gallery` into a self-hosted art platform with one Go backend, three public gallery experiences, and one private admin interface.

The original project is preserved on the `legacy` branch and should be treated as a reference implementation for the WebGL/REGL museum renderer, camera/navigation behavior, procedural geometry ideas, and texture lifecycle. Do not merge the legacy tree wholesale into this branch.

## Product scope

The application serves:

- `/` — public landing page linking only to the three public gallery experiences.
- `/gallery/` — lightweight responsive gallery grid/lightbox.
- `/timeline/` — horizontal chronological gallery with restrained motion and touch-friendly scrolling.
- `/museum/` — 3D museum with dynamically generated rooms, metadata-driven artwork grouping, and dynamic texture loading.
- `/artwork/{slug}` — canonical accessible artwork detail page.
- `/admin/` — private admin UI. It must never be linked from public navigation or the home page.

Each artwork has, at minimum: name, date, tags, surface, medium, source image, derived display images, and visibility state.

The admin can upload images, edit artwork metadata, manage controlled surface/medium values and tags, define museum grouping/layout rules, preview a draft museum, and publish a museum configuration. There is exactly one admin account. Username and password are created on first use; there is no user-management feature.

## Technical direction

- Go backend using the standard library where practical.
- SQLite as the default embedded database; choose a pure-Go driver to avoid CGO/container portability issues.
- Server-rendered HTML plus lightweight vanilla JavaScript/CSS; no SPA framework unless a later issue demonstrates a concrete need.
- Material-inspired component behavior with an editorial, image-first visual theme influenced by the Maren Lowe reference site: generous whitespace, restrained neutral surfaces, expressive typography, subtle motion, and artwork-first layouts. Inspiration only; do not copy assets or proprietary design details.
- Adapt selected ideas/code from `legacy` for the 3D renderer rather than rewriting the entire museum stack without need.
- Platform-agnostic OCI `Containerfile` and `Compose.yml`, tested with Docker and rootless Podman.
- Persistent data stored under a configurable data directory mounted as a volume.

## Principles

KISS and YAGNI are hard requirements. Prefer explicit code, small packages, standard library features, progressive enhancement, and measurable behavior over abstractions or speculative extensibility.

Every change that alters architecture, operation, configuration, deployment, routes, data model, or contributor workflow must update the relevant documentation in the same change.

## Planning documents

- `IMPLEMENTATION_PLAN.md` — phased execution plan and global acceptance criteria.
- `ARCHITECTURE.md` — system boundaries, data flow, security, storage, and museum architecture.
- `AGENTS.md` — autonomous agent execution rules and definition of done.
- `ISSUE_TRACKER.md` — ordered issue ledger, issue-ready specifications, and agent progress log.
- `DOCKERHUB.md` — image publishing and deployment contract.

## Branches

- `legacy` — snapshot of the pre-rewrite application.
- `rewrite` — clean rewrite and active implementation branch.

## Local development

The rewrite is a single Go process. Go 1.26 or newer is required.

```sh
go run ./cmd/gallery
curl http://localhost:8080/healthz
```

The default listen address is `:8080`; configure it with `GALLERY_LISTEN_ADDR`. The persistent data directory is configured with `GALLERY_DATA_DIR` and defaults to `./data`; startup creates `gallery.db` and runs embedded migrations there.

The current public data endpoint is `GET /api/artworks` with optional `tag`, `surface`, `medium`, and `order=asc` filters. Artwork detail JSON is available at `GET /api/artworks/{slug}`.

Canonical semantic artwork pages are available at `/artwork/{slug}` and link taxonomy values back to Gallery Lite filters.

On a fresh data directory, visit `/admin/` to create the single administrator. Subsequent visits use the login flow. Set `GALLERY_SECURE_COOKIES=true` when serving through HTTPS.

Authenticated admin users can add artwork at `/admin/artworks/new` and edit existing records from the `/admin/` list. The upload form accepts the required metadata and controls public visibility.

Authenticated administrators edit museum rules at `/admin/museum`; saving creates a validated draft and publishing is an explicit separate action. The public museum consumes only the published `/api/museum` snapshot.

Surface and medium values are normalized (trimmed and case-folded) and newly entered values are retained in the database for subsequent selection/filtering.

Artwork uploads accept JPEG, PNG, or GIF up to 20 MiB and are stored under the configured data directory with generated names. The image pipeline retains originals and creates thumbnail, medium, museum, and large JPEG derivatives.

Quality gates:

```sh
gofmt -w .
go vet ./...
go test ./...
```

Container development uses the portable Compose file:

```sh
docker compose -f Compose.yml up --build
# or: podman compose -f Compose.yml up --build
```

The container listens on port 8080, stores persistent state under `/data`, runs as a non-root user, and accepts `GALLERY_PORT`, `GALLERY_SECURE_COOKIES`, `GALLERY_LISTEN_ADDR`, and `GALLERY_DATA_DIR` configuration as documented in `DOCKERHUB.md`.

For operations, stop the service before copying persistent data. Use `scripts/backup.sh` and `scripts/restore.sh` as documented in `DOCKERHUB.md`; the procedure backs up SQLite and all image files together and never recommends copying a live database.

## Status

Implementation is proceeding through the ordered issue ledger in `ISSUE_TRACKER.md`, with each item linked to its authoritative GitHub issue.
