# Task 1 Fix: Correct Field Paths in Overview Section

## Issues to Fix

### Critical Issue: All field paths are wrong

**File:** `frontend/src/views/workload/DeploymentDetail.vue`

**Problem:** The backend returns the raw `appsv1.Deployment` Kubernetes API object with no transformation. The current field paths reference flat properties that don't exist.

**Fix:** Update the template expressions to use the correct nested paths:

| Current (broken) | Correct |
|---|---|
| `deployment.ready ?? 0` | `deployment.status?.readyReplicas ?? 0` |
| `deployment.replicas ?? 0` | `deployment.spec?.replicas ?? 0` |
| `deployment.available ?? '-'` | `deployment.status?.availableReplicas ?? '-'` |
| `deployment.updated ?? '-'` | `deployment.status?.updatedReplicas ?? '-'` |
| `deployment.strategy \|\| '-'` | `deployment.spec?.strategy?.type \|\| '-'` |
| `deployment.labels` | `deployment.metadata?.labels` |
| `deployment.selector` | `deployment.spec?.selector?.matchLabels` |

### Important Issue: Responsive column count

**Fix:** Update `el-descriptions` to use responsive column count:
```vue
<el-descriptions :column="{ xs: 1, sm: 2, md: 3, lg: 4 }" border size="small">
```

### Minor Issue: Redundant inline styles

**Fix:** Remove `style="margin-right: 4px;"` from el-tag elements since `.overview-tags` already uses `gap: 4px`.

## Implementation Steps

1. Open `frontend/src/views/workload/DeploymentDetail.vue`
2. Find the overview section (around line 275-310)
3. Update the field paths as shown above
4. Update `el-descriptions` to use responsive column count
5. Remove redundant inline styles from el-tag elements
6. Verify TypeScript compiles: `npm run build`
7. Commit: `git commit -m "fix: correct field paths in overview section"`
