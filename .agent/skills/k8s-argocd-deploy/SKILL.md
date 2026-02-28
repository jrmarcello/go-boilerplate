---
name: k8s-argocd-deploy
description: Kubernetes deployment with ArgoCD, Kustomize overlays, Bitbucket Pipelines, and Kind local cluster
---

# Kubernetes + ArgoCD Deploy

## Architecture

```text
Git Push → Bitbucket Pipelines → Docker Build → ECR Push → Kustomize Update → ArgoCD Sync
```

## Environments

| Environment | Cluster | Overlay |
| --- | --- | --- |
| Local | Kind | `deploy/overlays/develop/` |
| Homologação | AWS EKS | `deploy/overlays/homologacao/` |
| Produção | AWS EKS | `deploy/overlays/producao/` |

## Kustomize Structure

```text
deploy/
├── base/                 # Shared resources
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── hpa.yaml
│   └── kustomization.yaml
└── overlays/
    ├── develop/          # Kind local
    ├── homologacao/      # AWS EKS HML
    └── producao/         # AWS EKS PRD
```

## Commands

```bash
make kind-setup     # Full Kind setup (cluster + db + migrate + deploy)
make kind-up        # Create Kind cluster
make kind-deploy    # Deploy to Kind

# Validate manifests
kubectl kustomize deploy/overlays/develop/

# Apply dry-run
kubectl apply -k deploy/overlays/develop/ --dry-run=client
```

## Migrations

ArgoCD PreSync Job runs `cmd/migrate` binary before app deployment. See `docs/adr/006-migration-strategy.md`.

## Rollback

```bash
# kubectl
kubectl rollout undo deployment/go-boilerplate -n <namespace>

# ArgoCD (revert git)
git revert HEAD && git push origin develop
```
