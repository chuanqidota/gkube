# Feature Simplification Design

**Date:** 2026-06-14
**Status:** Approved
**Goal:** Reduce infrastructure dependencies, lower maintenance burden, simplify UI, remove incomplete features.

## Background

gkube currently has 26 feature domains with heavy infrastructure requirements (MySQL, Redis, ES, S3, Prometheus, Helm CLI, ArgoCD CLI). As an internal platform product, many enterprise features and external integrations are unnecessary. This design removes non-core features to focus the platform on its core value: multi-cluster Kubernetes resource management.

## Removal Inventory

### 1. External CLI Dependencies

| Feature | Frontend | Backend |
|---------|----------|---------|
| Catalog (Helm App Store) | `src/views/catalog/` | `app/k8s/api/catalog.go` |
| GitOps (ArgoCD Integration) | `src/views/gitops/` | `app/k8s/api/gitops.go` |

**Rationale:** Both require external CLI tools on the backend server and fall back to mock/sample data when unavailable. High maintenance cost for low actual usage.

### 2. Incomplete / Stub Features

| Feature | Frontend |
|---------|----------|
| Notification Center | `src/views/notification/` |
| Resource Watcher | `src/views/watcher/` |
| Batch Operations | `src/views/tools/BatchOperations.vue` |

**Rationale:** These pages have frontend routes but lack complete backend implementation or only display placeholder content.

### 3. Enterprise Features

| Feature | Frontend | Backend |
|---------|----------|---------|
| Approval Workflows | `src/views/approval/` | `app/k8s/api/approval.go`, model |
| Multi-Tenancy | `src/views/tenancy/` | `app/k8s/api/tenancy.go` |

**Rationale:** Approval workflows and namespace-based tenancy add significant complexity without being needed for the internal platform's use case.

### 4. Monitoring

| Feature | Frontend | Backend |
|---------|----------|---------|
| Monitoring pages | `src/views/monitoring/` | `app/k8s/api/metrics.go`, `prometheus.go` |
| Prometheus config | - | `config/prometheus.json` |

**Rationale:** Requires a separate Prometheus server. Monitoring is better handled by dedicated tools (Grafana, etc.).

### 5. Developer Tools

| Feature | Frontend |
|---------|----------|
| Resource Diff | `src/views/tools/ResourceDiff.vue` |
| Standalone YAML Editor | `src/views/tools/YAMLEditor.vue` |
| Topology view | `src/views/topology/` |

**Rationale:** Resource detail pages already have inline YAML editing. Standalone tools are rarely used. Topology view is a nice-to-have with limited practical value.

### 6. K8s Native RBAC Pages

| Feature | Frontend |
|---------|----------|
| ServiceAccount list/detail | `src/views/rbac/` (K8s RBAC sub-pages) |
| ClusterRole/Role list/detail | `src/views/rbac/` |
| ClusterRoleBinding/RoleBinding list/detail | `src/views/rbac/` |

**Rationale:** gkube has its own RBAC permission matrix for application-level access control. K8s native RBAC resource viewing adds complexity without clear value for the platform's audience.

### 7. Redis Dependency

| Component | Path |
|-----------|------|
| Redis client wrapper | `pkg/redis/` |
| Redis config | `config/config.yaml` `redis` section |
| Docker Compose Redis | `backend/docker-compose.yaml` |

**Rationale:** Redis is used only as a generic key-value store with no deep integration into core features. Removing it eliminates one infrastructure dependency.

## Retained Feature Set

### Core K8s Resource Management

- **Workloads:** Deployment (Scale/Restart/Rollback), StatefulSet, DaemonSet, Job, CronJob, Pod, HPA, PDB
- **Configuration:** ConfigMap, Secret, ResourceQuota, LimitRange
- **Storage:** PV, PVC, StorageClass
- **Networking:** Service, Ingress, NetworkPolicy
- **Cluster Resources:** Node, Namespace, Event, CRD

### Platform Features

- **Auth:** Local login + OIDC SSO, User/Role management, RBAC permission matrix
- **Cluster Management:** Multi-cluster registration, health checks
- **Dashboard:** Overview, resource stats, events feed
- **Web Terminal:** Container exec + session recording/playback (ES + S3)
- **Log Viewer:** Real-time log streaming via SSE
- **Audit Logs:** Operation audit records (ES)
- **Resource Search:** Global resource search

### Retained Infrastructure Dependencies

| Dependency | Purpose |
|------------|---------|
| MySQL | Core database (users, roles, clusters, audit) |
| Elasticsearch | Terminal recording storage + audit logs |
| S3/MinIO | Terminal recording file hosting |

## Simplified Sidebar Structure

```
1. Dashboard
2. Resource Search
3. Clusters
4. Workloads (Pod, Deployment, StatefulSet, DaemonSet, Job, CronJob, HPA, PDB)
5. Configuration (ConfigMap, Secret, ResourceQuota, LimitRange)
6. Storage (PV, PVC, StorageClass)
7. Network (Service, Ingress, NetworkPolicy)
8. Nodes
9. Namespaces
10. Events
11. CRD
12. Tools (Terminal, Logs)
13. System (Users, Roles, Auth Settings, Audit)
```

Reduction: 19 top-level menu items → 13 top-level menu items. Six independent feature domains removed.

## Implementation Order

Single-pass cleanup, frontend-first approach:

### Step 1: Frontend Cleanup

1. Delete removed page component directories:
   - `src/views/catalog/`, `src/views/gitops/`, `src/views/notification/`
   - `src/views/watcher/`, `src/views/approval/`, `src/views/tenancy/`
   - `src/views/monitoring/`, `src/views/topology/`
2. Delete individual tool files: `BatchOperations.vue`, `ResourceDiff.vue`, `YAMLEditor.vue`
3. Delete K8s native RBAC sub-pages under `src/views/rbac/`
4. Update `src/router/index.ts` — remove related routes
5. Update `src/components/Layout/Sidebar.vue` — remove menu items
6. Clean up `src/api/` — remove standalone API files for removed features
7. Reorganize `src/views/tools/` if needed (keep only terminal + logs entry points)

### Step 2: Backend API Cleanup

1. Delete handler files: `catalog.go`, `gitops.go`, `approval.go`, `tenancy.go`, `metrics.go`, `prometheus.go`, `topology.go`
2. Delete model files (e.g., `model/approval.go`)
3. Update `backend/router/router.go` — remove route registrations
4. Clean up cross-references in remaining API files

### Step 3: Redis Dependency Removal

1. Delete `backend/pkg/redis/` directory
2. Search and remove all `import` references to `pkg/redis`
3. Remove `redis` config section from `config/config.yaml`
4. Remove Redis service from `backend/docker-compose.yaml`
5. Remove `redis.Init()` call from `backend/cmd/root.go`

### Step 4: Configuration and Documentation

1. Delete `backend/config/prometheus.json`
2. Update `docker-compose.yaml` — remove Redis service
3. Update `CLAUDE.md` — reflect new feature scope and dependencies
4. Update `config/config.yaml` — remove Redis section, keep ES and S3

### Step 5: Verification

1. `cd frontend && npm run build` — ensure frontend compiles
2. `cd backend && go build -o gkube .` — ensure backend compiles
3. Manual check: sidebar renders correctly, routes work, no broken imports

## Risk Assessment

- **Low risk:** Removing stub/incomplete features (no real functionality lost)
- **Low risk:** Removing Catalog/GitOps (already falling back to mock data)
- **Low risk:** Removing monitoring (delegated to external tools)
- **Medium risk:** Redis removal — need to verify no hidden usages in middleware or session management
- **Medium risk:** K8s RBAC removal — some users may rely on viewing cluster RBAC config

## Out of Scope

- Refactoring the remaining code (e.g., splitting the large `k8s/api/` directory)
- Adding tests
- Performance optimization
- UI redesign of remaining features
