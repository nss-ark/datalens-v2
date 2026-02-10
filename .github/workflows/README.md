# GitHub Actions Workflows

This directory contains CI/CD automation workflows for DataLens 2.0.

## Workflows

### CI Pipeline (`ci.yml`)

Runs on every push to `main`/`develop` and on all pull requests.

**Jobs:**

1. **Backend (Go)**
   - Lints code with `go vet`
   - Builds all packages
   - Runs tests with race detector
   - Generates coverage report
   - Uses service containers: PostgreSQL 16, Redis 7, NATS 2

2. **Frontend (React)**
   - Lints code with ESLint
   - Builds production bundle with Vite
   - Uploads build artifacts

3. **Docker Build & Push** (main branch only)
   - Builds multi-stage Docker images
   - Pushes to GitHub Container Registry (`ghcr.io`)
   - Tags: `latest` and `sha-<commit>`

---

## Required GitHub Secrets

| Secret | Description | Required For |
|--------|-------------|--------------|
| `GITHUB_TOKEN` | Auto-provided by GitHub Actions | Docker push to GHCR |
| `CODECOV_TOKEN` | Codecov.io API token | Coverage upload (optional) |

---

## Running Workflows Manually

Navigate to **Actions** tab → Select workflow → **Run workflow** → Choose branch

---

## Troubleshooting

### Backend tests failing in CI but passing locally

**Cause:** Service containers not ready
**Solution:** Workflow uses health checks to wait for services. Check service logs in workflow output.

### Docker push fails with permission denied

**Cause:** Package write permission not configured
**Solution:** Ensure workflow has `permissions: packages: write` (already configured)

### Frontend build fails in CI

**Cause:** Node modules cache mismatch
**Solution:** Clear cache by updating `cache-dependency-path` or deleting workflow cache

---

## Badge URLs

Add to `README.md`:

```markdown
[![CI](https://github.com/{owner}/{repo}/actions/workflows/ci.yml/badge.svg)](https://github.com/{owner}/{repo}/actions/workflows/ci.yml)
```

---

## Local Testing

Test the workflow locally using [act](https://github.com/nektos/act):

```bash
# Install act
winget install nektos.act

# Run workflow
act -j backend
act -j frontend
```
