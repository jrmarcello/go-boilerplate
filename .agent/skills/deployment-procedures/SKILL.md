---
name: deployment-procedures
description: Deploy procedures — pre-deploy checklist, validation, rollback, zero-downtime, post-deploy verification
---

# Deployment Procedures

## Pre-Deploy Checklist

```text
Code Quality
- [ ] make lint
- [ ] make lint-full
- [ ] make test

Security
- [ ] No hardcoded secrets
- [ ] Environment variables documented
- [ ] Dependencies without known CVEs

Database
- [ ] Migrations tested locally
- [ ] Migrations reversible

Documentation
- [ ] ADR created if architectural change
- [ ] Swagger updated: swag init -g cmd/api/main.go -o docs
```

## Deploy Flow

```text
1. PREPARE → Tests pass? Build OK? Env vars configured?
2. DEPLOY  → Push to branch → CI Pipeline → ArgoCD sync
3. VERIFY  → Health check? Logs clean? Key flows working?
4. CONFIRM → All OK → Done. Problems → Rollback immediately
```

## Post-Deploy Verification

```bash
# Health check
curl -s http://<host>/health

# Check pods
kubectl get pods -n <namespace>

# Check logs
kubectl logs -f deployment/<app> -n <namespace> --tail=50
```

## Rollback Triggers

| Symptom | Action |
| --- | --- |
| Service down | Rollback immediately |
| Critical errors in logs | Rollback |
| Performance degraded >50% | Consider rollback |
| Minor issues | Fix forward if quick |
