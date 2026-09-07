# Container Image and Registry Guide

This document defines the portable container and release contract for the rewrite.

## Image name

Default examples use:

```text
vm75/virtual-art-gallery
```

Git releases use `vMAJOR.MINOR.PATCH` tags. Published image tags omit the leading `v`, so `v1.2.3` publishes `1.2.3`; stable releases also publish `latest`. Pre-release versions such as `v1.2.3-rc.1` never move `latest`.

Images are published to both `vm75/virtual-art-gallery` on Docker Hub and `ghcr.io/vm75/virtual-art-gallery`. The GitHub Actions workflow runs only for pushed `v*` tags or an explicit manual dispatch. Manual dispatch requires a semantic version input and uses the selected workflow ref. Docker Hub requires repository secrets named `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`; GHCR uses the workflow-provided `GITHUB_TOKEN`.

## Platform portability requirements

The project must ship a file named exactly `Containerfile` using ordinary OCI/Dockerfile-compatible instructions and a file named exactly `Compose.yml` using portable Compose syntax.

The runtime must not depend on:

- Docker socket access;
- privileged containers;
- host networking;
- root UID;
- Docker-only `host.docker.internal` behavior;
- Swarm-only settings;
- Kubernetes-specific resources;
- bind mounts that assume a particular host filesystem layout.

It should work with both Docker and rootless Podman when their normal Compose-compatible tooling is available.

## Runtime contract

The final container should:

- run as a non-root user;
- expose one configurable HTTP port;
- use a single persistent application data directory, expected to be mounted at `/data` unless implementation proves another path simpler;
- store SQLite, originals, derived images, and application-managed persistent state under that data directory;
- write logs to stdout/stderr;
- handle SIGTERM cleanly;
- provide a health endpoint suitable for container health checks.

Configuration is environment-based and intentionally small:

```text
GALLERY_LISTEN_ADDR
GALLERY_DATA_DIR
GALLERY_BASE_URL      # only if needed for absolute URL generation
GALLERY_SECURE_COOKIES # only if automatic/proxy detection is insufficient
```

The current defaults are `GALLERY_LISTEN_ADDR=:8080`, `GALLERY_DATA_DIR=./data` for local runs (and `/data` in the image), and `GALLERY_SECURE_COOKIES=false`.

The shipped `Containerfile` builds a static Go binary and runs it as UID 10001 in Alpine. Release builds inject version, commit, and build-date metadata into the binary and OCI labels; local builds default to version `dev`. `Compose.yml` exposes host `${GALLERY_PORT:-8080}` to container port 8080 and persists `/data` in the named `gallery-data` volume. The image healthcheck uses Alpine's bundled `wget` against `/healthz`.

Validation on 2026-09-07: the previously released container baseline passed `podman build -f Containerfile ...` and a temporary rootless-compatible `podman run` health/UID smoke. The versioned release build could not be rerun here because Podman could not set its runtime-directory permissions and Docker could not access its daemon. Docker Compose and Podman Compose were not installed in the execution environment.

## Backup and restore

Stop the application/container first so SQLite is not being written. Then back up the complete persistent directory, including `gallery.db`, originals, and derivatives:

```sh
docker compose -f Compose.yml stop gallery
./scripts/backup.sh /path/to/gallery-data /path/to/backups/gallery-$(date +%Y%m%d).tar.gz
```

Restore into a new or empty data directory before starting a fresh container:

```sh
./scripts/restore.sh /path/to/backups/gallery-20260907.tar.gz /path/to/new-gallery-data
```

For a named volume, stop the service and run the scripts from a temporary helper container or copy the volume contents to a host directory; do not copy a live SQLite file. The restored application reruns forward migrations on startup. Keep backups from the same or older v1 schema and test a restore before upgrading. Rootless runtimes may need the restored directory ownership adjusted to the container’s non-root UID 10001.

To reset a fresh installation, stop the service and remove the application data volume only after taking a backup; this permanently removes the database and uploaded images.

Routine operations use `docker compose -f Compose.yml logs -f gallery` for logs, `... stop gallery` for graceful stop, and `... up -d --build gallery` for an upgrade. The same commands can be run with `podman compose` where that frontend is installed.

Do not add environment variables preemptively. The actual names/defaults must be synchronized here once implemented.

## Target local usage

Docker:

```sh
docker compose -f Compose.yml up --build
```

Rootless Podman, depending on installed Compose frontend:

```sh
podman compose -f Compose.yml up --build
```

or an equivalent documented `podman-compose` invocation.

The Compose file should create/use a named volume so users do not need to understand container UID mappings for the common case.

## Target standalone usage

```sh
docker run --rm \
  -p 8080:8080 \
  -v gallery-data:/data \
  vm75/virtual-art-gallery:latest
```

Equivalent rootless Podman usage should work with the same port and named-volume concept:

```sh
podman run --rm \
  -p 8080:8080 \
  -v gallery-data:/data \
  vm75/virtual-art-gallery:latest
```

Exact ports and environment variables must be updated here when implementation stabilizes.

## Building and multi-architecture images

The Go build should avoid CGO by default where practical so common Linux architectures can be produced without per-platform native SQLite toolchains.

Do not require a specific vendor build extension in the repository contract. Docker Buildx or Podman build tooling may be documented as optional examples for multi-architecture publishing, but the `Containerfile` itself must remain ordinary OCI syntax.

## Publishing checklist

Before publishing a release image:

1. Run Go tests and documented frontend checks.
2. Build the `Containerfile` from a clean checkout.
3. Start with a fresh volume and complete first-run admin setup.
4. Upload an artwork and verify persistence after container recreation.
5. Smoke-test Gallery Lite, Timeline, Museum, and artwork detail route.
6. Confirm `/admin/` is not linked from public pages.
7. Verify container process is non-root.
8. Verify graceful stop.
9. Verify health endpoint.
10. Test Docker Compose and rootless Podman Compose behavior where available.
11. Confirm README, `ARCHITECTURE.md`, this document, and `Compose.yml` use the same ports, paths, env names, and image tags.
12. Publish the immutable version tag; the release workflow adds `latest` only for a stable semantic version.

## GitHub release workflow

Create and push a release tag after the quality workflow is green:

```sh
git tag v1.2.3
git push origin v1.2.3
```

Alternatively, select **Actions → publish container → Run workflow**, enter a semantic version, and choose the ref to build. The workflow validates the version, logs in independently to both registries, builds the `Containerfile` once, and pushes the resulting tags to both registries. It does not run for branch pushes or pull requests.

## Backup contract

Because v1 uses one persistent data directory, a consistent backup should ultimately be documented around that directory. If online SQLite backup semantics are needed, provide a safe application command or documented stop-and-copy procedure rather than encouraging unsafe copying of a busy database.

Backup/restore details must be finalized before v1 acceptance.
