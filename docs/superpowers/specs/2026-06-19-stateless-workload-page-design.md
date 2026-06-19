# Stateless Workload (Deployment) List Page Design

**Date:** 2026-06-19
**Status:** Approved
**Scope:** Simplify the Deployment list page refresh mechanism

## Problem

The current DeploymentList.vue has an auto-refresh mechanism (15-second polling) with pause/resume toggle and countdown display. This adds unnecessary complexity — users don't need automatic polling on a list page, and the two extra buttons (refresh with countdown + pause/resume) clutter the filter bar.

## Design Decisions

| Element | Decision | Rationale |
|---------|----------|-----------|
| Auto-refresh | **Remove** | Not needed for a list page; users refresh manually when needed |
| Pause/Resume button | **Remove** | Goes away with auto-refresh |
| Countdown display | **Remove** | Goes away with auto-refresh |
| Manual refresh button | **Keep** (simplified) | Users still need a way to reload data without F5 |
| Search by name | **Keep** | Client-side filter, useful for quick lookup |
| Namespace selector | **Keep** | Server-side filter, essential for multi-namespace clusters |
| Create button | **Keep** | Primary action |
| Batch delete | **Keep** | Useful for cleaning up multiple deployments |
| Row actions (YAML + Delete) | **Keep** | Sufficient for list-level operations |
| Table columns | **Keep as-is** | Name, Namespace, Ready, Up-to-date, Available, Age |

## Changes

### Filter Bar (Before → After)

**Before:**
```
[Search] [Namespace ▾] [Refresh (12s)] [Pause/Resume] [Create] [Delete (0)]
```

**After:**
```
[Search] [Namespace ▾] [Refresh] [Create] [Delete (0)]
```

### Code Changes

1. **Remove `useAutoRefresh` import and usage**
   - Remove `import { useAutoRefresh } from '@/composables/useAutoRefresh'`
   - Remove `const { isRunning, countdown, toggle, refresh: autoRefresh } = useAutoRefresh(fetchDeployments, 15000)`

2. **Simplify refresh button**
   - Replace the countdown-based button with a simple refresh button
   - Use `loading` ref to show loading state during fetch
   - Button text: just the icon + "Refresh" (no countdown)

3. **Remove pause/resume button entirely**

4. **`fetchDeployments()` already handles loading state** — no changes needed to the fetch function itself

### Template Changes

```vue
<!-- Before -->
<el-button type="primary" @click="autoRefresh()">
  <el-icon><Refresh /></el-icon> {{ t('common.refresh') }} ({{ countdown }}s)
</el-button>
<el-button @click="toggle()" :type="isRunning ? 'warning' : 'success'" size="default">
  {{ isRunning ? t('common.paused') : t('common.resume') }}
</el-button>

<!-- After -->
<el-button @click="fetchDeployments()" :loading="loading">
  <el-icon><Refresh /></el-icon> {{ t('common.refresh') }}
</el-button>
```

## Files to Modify

- `frontend/src/views/workload/DeploymentList.vue` — the only file that changes

## Out of Scope

- Other list pages (StatefulSet, DaemonSet, etc.) — each page is independent; this change is Deployment-only
- Generic list component abstraction — not in scope for this change
- YAML dialog or table column changes — no changes needed
