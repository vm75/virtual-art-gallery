# Container Image and Registry Guide

This document defines the intended portable container contract for the rewrite. Until the containerization issues are complete, commands here are a target specification rather than a release claim.

## Image name

Default examples use:

```text
vm75/virtual-art-gallery
```

Tags should be immutable for releases (`v1.2.3`) with optional moving convenience tags (`1`, `1.2`, `latest`) only after a release policy is established.

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

Configuration should be environment-based and intentionally small. Expected categories include:

```text
GALLERY_LISTEN_ADDR
GALLERY_DATA_DIR
GALLERY_BASE_URL      # only if needed for absolute URL generation
GALLERY_SECURE_COOKIES # only if automatic/proxy detection is insufficient
```

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
12. Publish immutable version tag; only then update moving tags if policy permits.

## Backup contract

Because v1 uses one persistent data directory, a consistent backup should ultimately be documented around that directory. If online SQLite backup semantics are needed, provide a safe application command or documented stop-and-copy procedure rather than encouraging unsafe copying of a busy database.

Backup/restore details must be finalized before v1 acceptance.
