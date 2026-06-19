# Stateless Workload Page Refresh Simplification — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove auto-refresh, pause/resume, and countdown from DeploymentList.vue, replacing with a simple manual refresh button with loading state.

**Architecture:** Single-file change to `DeploymentList.vue`. Remove `useAutoRefresh` composable usage, remove pause/resume button and countdown display from template, replace refresh button with a simple click-to-refresh using the existing `loading` ref.

**Tech Stack:** Vue 3, Element Plus, TypeScript

## Global Constraints

- Only `DeploymentList.vue` is modified — no other files change
- The `useAutoRefresh` composable itself is NOT deleted (other pages still use it)
- Existing `fetchDeployments()` function already manages `loading` state — no logic changes needed there
- Follow the established pattern from `ClusterList.vue`, `NamespaceManager.vue`, etc. for simple refresh buttons

---

### Task 1: Simplify DeploymentList.vue Refresh Mechanism

**Files:**
- Modify: `frontend/src/views/workload/DeploymentList.vue`

**Interfaces:**
- No interface changes — purely internal cleanup of one component

- [ ] **Step 1: Remove `useAutoRefresh` import**

In `DeploymentList.vue`, delete line 7:
```ts
import { useAutoRefresh } from '@/composables/useAutoRefresh'
```

- [ ] **Step 2: Remove `useAutoRefresh` usage**

Delete line 157:
```ts
const { isRunning, countdown, toggle, refresh: autoRefresh } = useAutoRefresh(fetchDeployments, 15000)
```

- [ ] **Step 3: Replace refresh button in template**

Replace lines 186-191:
```vue
<el-button type="primary" @click="autoRefresh()">
  <el-icon><Refresh /></el-icon> {{ t('common.refresh') }} ({{ countdown }}s)
</el-button>
<el-button @click="toggle()" :type="isRunning ? 'warning' : 'success'" size="default">
  {{ isRunning ? t('common.paused') : t('common.resume') }}
</el-button>
```

With:
```vue
<el-button @click="fetchDeployments()" :loading="loading">
  <el-icon><Refresh /></el-icon> {{ t('common.refresh') }}
</el-button>
```

- [ ] **Step 4: Verify the build compiles**

Run: `cd frontend && npm run build`
Expected: Build succeeds with no errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/views/workload/DeploymentList.vue
git commit -m "fix: simplify DeploymentList refresh to manual-only button"
```
