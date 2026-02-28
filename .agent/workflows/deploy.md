---
name: deploy
description: Deployment workflow — pre-deploy checks, deploy, post-deploy verification
trigger: "deploy|release|ship|go live"
---

# Deploy Workflow

## Load Skills

- `deployment-procedures`
- `k8s-argocd-deploy`
- `lint-and-validate`

## Steps

### 1. Pre-Deploy Checks

```bash
# Code quality
make lint
make lint-full
make test
make security

# Swagger up to date
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
git diff --exit-code docs/  # No changes expected

# Dependencies
go mod tidy
git diff --exit-code go.mod go.sum
```

### 2. Build Verification

```bash
# Docker build
docker build -f docker/Dockerfile -t go-boilerplate:test .

# Local smoke test
docker run --rm go-boilerplate:test /app/api --help
```

### 3. Deploy

Deployment is automated via Bitbucket Pipelines:

- Push to `develop` → Deploy to develop/HML
- Push to `main` → Deploy to production

Pipeline: Build → Push ECR → Update Kustomize image → ArgoCD sync

### 4. Post-Deploy Verification

```bash
# Health check
curl -s https://<host>/health | jq .

# Check pods
kubectl get pods -n <namespace> -l app=go-boilerplate

# Watch logs
kubectl logs -f deployment/go-boilerplate -n <namespace> --tail=50

# Verify key endpoint
curl -s https://<host>/api/v1/entities | jq .
```

### 5. Rollback (if needed)

```bash
# Kubernetes
kubectl rollout undo deployment/go-boilerplate -n <namespace>

# Git revert
git revert HEAD
git push origin <branch>
```
