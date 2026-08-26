# github-issues-api

A Go REST API that securely brokers create, list, and close operations
against GitHub Issues — built as a technical challenge for a cloud
security engineering role.

---

## Contents

- [Requirements checklist](#requirements-checklist)
- [Architecture](#architecture)
- [Endpoints](#endpoints)
- [Running it](#running-it)
- [CI/CD pipeline](#cicd-pipeline)
- [Engineering & security decisions](#engineering--security-decisions)
- [Testing approach](#testing-approach)
- [Known limitations](#known-limitations)

---

## Requirements checklist

| Requirement | Status |
|---|---|
| REST API: create, list, close issues for a GitHub repo | ✅ |
| Manage authentication | ✅ — two independent layers, see below |
| Pipeline: tests, lint, security check, deploy | ✅ |
| Pipeline runs on GitHub | ✅ — GitHub Actions |
| All done with git | ✅ |
| Language: Go | ✅ |
| Testing prioritized | ✅ — happy paths and failure paths throughout |

## Architecture

```
 Caller  --(API_TOKEN)-->  This API  --(GITHUB_TOKEN)-->  GitHub Issues API
```

This service holds no issue data itself — GitHub remains the single
source of truth. Its job is to validate incoming requests, enforce its
own authentication, and safely mediate access to GitHub on the caller's
behalf.

## Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/repos/{owner}/{repo}/issues` | Create an issue |
| `GET` | `/repos/{owner}/{repo}/issues` | List issues |
| `DELETE` | `/repos/{owner}/{repo}/issues/{number}` | Close an issue |
| `GET` | `/health` | Health check (no auth required) |

> **Note:** `DELETE` closes the issue rather than deleting it — GitHub's
> API has no permanent-delete operation for issues, only close.

## Running it

### Locally

Rename `secret.example.yaml` to secret.yaml and insert your tokens in the corresponding fields

### In Docker

```bash
docker build -t github-issues-api .
docker run -p 8080:8080 \
  -e API_TOKEN="choose-a-secret" \
  -e GITHUB_TOKEN="fine-grained-PAT" \
  github-issues-api
```

### On Kubernetes (minikube)

```bash
./scripts/deploy-local.sh
```

### Example request

```bash
curl -X POST localhost:8080/repos/OWNER/REPO/issues \
  -H "Authorization: Bearer choose-a-secret" \
  -H "Content-Type: application/json" \
  -d '{"title": "Example issue", "body": "Example description"}'
```

## CI/CD pipeline

Defined in `.github/workflows/ci.yml`, five jobs run on every push and
pull request to `main`:

| Job | What it does |
|---|---|
| `build-and-test` | `go vet`, full test suite (`go test -race -cover ./...`) |
| `lint` | `golangci-lint` — style, correctness, unchecked errors |
| `security` | `gosec` (static analysis) + `govulncheck` (dependency CVEs, call-graph-aware) |
| `docker` | Builds the image, scans it with `trivy` |
| `deploy` | Deploys to a `kind` cluster inside the runner |


## Engineering & security decisions

**Two independent authentication layers**
`API_TOKEN` (caller → this API) and `GITHUB_TOKEN` (this API → GitHub)
are fully separate. Callers never see or handle the GitHub token — this
service is a trust boundary, not a credential pass-through.

**Why a bearer token, not OAuth2 or asymmetric keys**
The actual need here is single-tier access control — "is this caller
allowed to use the API at all" — not multi-user identity or delegated
authorization, which is what OAuth2 is for. Asymmetric-key auth (as used
for SSH) earns its complexity when the goal is *never transmitting a
secret at all*; here the goal is verifying possession of a shared secret
over an already-TLS-encrypted connection, which a bearer token handles
directly — the same approach GitHub's own API uses. The token comparison
itself uses `crypto/subtle.ConstantTimeCompare` rather than `==`, to
remove a timing side-channel on the check.

**Least-privilege GitHub token**
Expected to be a fine-grained personal access token scoped to
`Issues: read/write` on a single repository — not a classic token with
full `repo` scope.

**No shelling out to `gh` at runtime**
This service talks to GitHub via `google/go-github` over HTTP directly.
Shelling out to an external CLI from a running service risks command
injection from unsanitized input and would require bundling an extra
binary into the image for no real benefit. `gh` was used only as a local
developer convenience — creating the repository, managing CI secrets —
never invoked by the service itself.

**Distroless base image, non-root, numeric UID**
The final image is built from `gcr.io/distroless/static-debian12` — no
shell, no package manager, minimal attack surface. It runs as UID
`65532` (numeric) rather than the name `nonroot`: Kubernetes'
`runAsNonRoot: true` check needs a numeric UID to verify against, and
fails to start a container running under a named user it can't resolve.

**Multi-stage Docker build**
The build stage — full Go toolchain, source, build cache — never becomes
part of the final image's layers; only the compiled binary is copied
across.

**Explicit server timeouts**
`gosec` (rule G114) flagged the use of `http.ListenAndServe` for
configuring no timeouts, leaving the server exposed to slow-client
(Slowloris-style) denial of service. Replaced with an explicit
`http.Server` specifying read, write, and idle timeouts.

**Interface-first design**
`issues.Service` is implemented by both an in-memory `Fake` (used
throughout the test suite) and `GitHubService` (the real implementation).
Handlers and tests depend only on the interface — switching from fake to
real GitHub integration was a one-line change in `main.go`.

**Bugs caught before shipping**
- An early version of `Fake` shared a single issue-number counter and map
  across all repositories, unlike real GitHub, which numbers issues
  independently per repo. Caught, fixed, and covered by a regression test
  (`TestFakeRepoIsolation`).
- `GITHUB_TOKEN` was initially missing from the Kubernetes Secret and
  Deployment manifests — only `API_TOKEN` was wired in, so the service
  could authenticate callers but not itself to GitHub. Caught via a
  failed real-world minikube deployment test, then fixed in the
  manifests.

## Testing approach

- **Handlers** are tested against the `Fake` using `httptest` — covering
  validation, status codes, and auth middleware, with no network
  dependency.
- **`GitHubService`** is tested against `httptest.NewServer`, simulating
  GitHub's API responses to verify request construction and response
  parsing — without depending on network access or a live token.
  Verifying GitHub's own server-side correctness isn't this test suite's
  responsibility; only whether this code talks to it correctly.
- Failure paths are tested alongside happy paths throughout: invalid
  input, missing or incorrect auth, a nonexistent issue on close, and
  GitHub error responses.

## Known limitations

- `API_TOKEN` is a single, operator-provisioned shared secret. A real
  multi-tenant service would issue and independently revoke tokens per
  client, stored hashed.
- `Close`'s error handling currently maps any error from `Service.Close`
  to `404 Not Found` — correct for the `Fake`, whose only failure mode is
  "not found," but would need refining once other GitHub error types
  (rate limiting, authorization failures) become distinguishable.
- Kubernetes Secrets are used as-is — base64-encoded, not encrypted at
  rest by default. A production deployment would use Sealed Secrets or an
  external secrets manager.