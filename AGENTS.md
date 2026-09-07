# AGENTS.md

This file defines the operating contract for autonomous coding agents. `main` is the canonical product branch; the current v1 corrective cycle is developed on `fix/v1-remediation` and merged to `main` only after its release gates pass.

## Read first

Before changing code, read:

1. `README.md`
2. `ARCHITECTURE.md`
3. `IMPLEMENTATION_PLAN.md`
4. `ISSUE_TRACKER.md`
5. the matching GitHub issue, when one exists

GitHub Issues are enabled for this repository. The linked GitHub issue is authoritative for detailed scope and acceptance criteria; `ISSUE_TRACKER.md` is the compact execution ledger.

The `legacy` branch is reference material only. Do not merge it wholesale. Copy/adapt only code that is justified by the current issue and compatible with the current architecture/licensing.

## Prime directives

- Follow KISS and YAGNI.
- Implement the smallest complete solution that satisfies the issue acceptance criteria.
- Prefer Go standard library, semantic HTML, CSS, and vanilla JavaScript.
- Add a dependency only when it clearly reduces risk/complexity and record the reason in the change/issue.
- Do not introduce a SPA framework, ORM, microservice, queue, cache service, plugin system, generalized rules DSL, or cloud-specific dependency unless an approved issue explicitly requires it.
- Keep public and admin code secure and accessible by default.
- Never expose or log credentials, password hashes, session tokens, or uploaded private filesystem paths.

## Issue execution protocol

Work one issue/tracker item at a time unless it explicitly groups inseparable tasks. Each issue that changes repository content needs one distinct issue-scoped commit; its tests and documentation/tracker updates belong in that same commit.

1. Confirm all declared dependencies are complete.
2. Mark the item `in progress` in `ISSUE_TRACKER.md` with a short note/date.
3. Inspect existing code and tests before editing.
4. Write/adjust tests alongside behavior.
5. Implement only the issue scope.
6. Run the relevant local quality gates.
7. Check every acceptance criterion explicitly.
8. Update all affected docs in the same change.
9. Update `ISSUE_TRACKER.md` with result, validation commands, commit/PR, and follow-up defects.
10. Do not silently expand scope. Create/follow a separate tracker item or GitHub issue for discovered work that is not required to satisfy the current acceptance criteria.

When GitHub Issues are enabled, keep the tracker ID and GitHub issue number cross-linked and keep status synchronized.

## Definition of done for every implementation issue

Unless an issue states otherwise, done means:

- Acceptance criteria are met and demonstrably testable.
- New/changed Go code is formatted and tests pass.
- New/changed JavaScript has no console errors in supported flows.
- Public UI is usable at narrow mobile and desktop widths.
- Keyboard navigation and visible focus are preserved for interactive HTML UI.
- `prefers-reduced-motion` is respected where motion is added.
- No admin link is added to public navigation/home.
- Security-sensitive changes include tests or documented manual verification.
- Container/deployment assumptions remain platform agnostic.
- Relevant docs and `ISSUE_TRACKER.md` are synchronized.
- No unrelated refactors, speculative abstractions, or dead code are included.

## Quality gates

Use the repository-provided commands once they exist. At minimum, the project should converge on equivalents of:

```sh
go fmt ./...
go vet ./...
go test ./...
```

The bootstrap implementation currently uses `go run ./cmd/gallery`, with `GALLERY_LISTEN_ADDR` defaulting to `:8080`, `GALLERY_DATA_DIR` defaulting to `./data`, and `GALLERY_SECURE_COOKIES` defaulting to false.
CI runs on pushes to and pull requests targeting `main`. It checks `test -z "$(gofmt -l cmd internal web)"`, all browser modules with `node --check`, pure browser-module tests, a production build, and an OCI `Containerfile` build. The release workflow is the only CI path that publishes the OCI image, and it runs only for release tags or manual dispatch.

Frontend checks should be dependency-light. Prefer browser/integration tests only where they protect important interactions; do not build a heavyweight testing stack without need.

Container-related work must validate both standard OCI build syntax and rootless operation assumptions. Where the executing environment has Docker or Podman, run the relevant build/smoke test; otherwise document exactly what was not executable.

## Documentation synchronization matrix

Update documentation in the same change when any of these change:

- Routes, setup, configuration, local run commands -> `README.md`
- Component boundaries, data model, security, API, persistence, museum pipeline -> `ARCHITECTURE.md`
- Agent workflow/quality gates -> `AGENTS.md`
- Phase/order/scope/global acceptance -> `IMPLEMENTATION_PLAN.md`
- Container image, Compose, volumes, env vars, registry usage -> `DOCKERHUB.md`
- Issue state, discovered defects, validation notes -> `ISSUE_TRACKER.md`

If none need changes, state that docs were reviewed and remain accurate in the issue/PR notes or tracker execution log.

## Coding guidance

### Go

- Keep packages cohesive and small; avoid `util` dumping grounds.
- Pass dependencies explicitly.
- Use context for request-scoped work and shutdown.
- Return/wrap actionable errors without leaking secrets.
- Keep HTTP handlers thin; domain behavior belongs in focused services/functions where necessary.
- Prefer SQL that can be read and reviewed over an ORM.
- Database migrations are forward-only for v1 unless rollback is explicitly required.

### Frontend

- Start with server-rendered HTML and progressive enhancement.
- Use CSS custom properties for design tokens.
- Keep JS page-specific; do not create a shared framework for a handful of helpers.
- Use native browser features (`dialog`, `IntersectionObserver`, scrolling, responsive images) when adequate.
- Public pages should work meaningfully even if optional animation fails.

### Museum

- Keep semantic rules/layout generation separate from WebGL rendering.
- Generated layouts must be deterministic for a given input/seed.
- Artwork-to-placement assignment must be explicit.
- Do not load full-resolution originals as default GPU textures.
- Mobile controls and fallback access to canonical artwork HTML are required before v1 is considered done.

## Blockers and autonomous decisions

Agents may make ordinary implementation choices consistent with the architecture without asking for approval. Prefer the simplest reversible choice.

Stop and surface a blocker rather than guessing when a decision would:

- contradict a stated product requirement,
- require destructive migration of existing production data,
- materially change authentication/security policy,
- add a paid/proprietary external service,
- change the single-container deployment model,
- or require scope substantially beyond the current issue.

Record unresolved decisions and defects in `ISSUE_TRACKER.md` and the relevant GitHub issue when available.
