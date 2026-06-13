# Frontend UX Optimization Design

**Date**: 2026-06-13
**Scope**: 4 incremental improvements to match Kuboard-level UX

---

## 1. Nested Routes with AppLayout Wrapper

### Problem
All routes are flat top-level entries. `AppLayout` exists but is not wired into the router. Each page loads independently without a shared sidebar/header shell.

### Solution
Convert to nested routing: `/login` and `/oidc/callback` use standalone layouts; all other pages are children of `/` which renders `AppLayout`.

### Current Structure
```
/login → LoginView
/dashboard → DashboardView
/workloads/deployments → DeploymentList
...
```

### Target Structure
```
/login → LoginView (standalone)
/oidc/callback → OIDCCallback (standalone)
/ → AppLayout (wrapper with Sidebar + Header)
  /dashboard → DashboardView
  /workloads/deployments → DeploymentList
  /workloads/deployments/:namespace/:name → DeploymentDetail
  ... (all other business pages)
```

### Changes Required
1. **`App.vue`**: Simplify to just `<router-view />` (remove any layout imports)
2. **`router/index.ts`**: Restructure routes:
   - `/login` and `/oidc/callback` remain top-level
   - Add a parent route for `/` with `component: AppLayout` and `redirect: '/dashboard'`
   - Move all other routes as `children` of the `/` route
3. **Remove duplicate imports**: Pages that currently import Sidebar/Header/AppLayout directly should have those imports removed (only if they exist)

### Files Modified
- `frontend/src/App.vue`
- `frontend/src/router/index.ts`

---

## 2. Breadcrumb Enhancement

### Problem
Breadcrumb only shows "Home > <title>". Most routes lack `meta.title`, making the breadcrumb non-functional.

### Solution
- Add `meta.title` to every route (Chinese, matching sidebar labels)
- Add `meta.parent` to detail/create routes for proper breadcrumb hierarchy
- Update `Header.vue` breadcrumb to build from `route.matched`

### Route Meta Examples
```ts
{ path: '/workloads/deployments', meta: { title: 'Deployment' } }
{ path: '/workloads/deployments/:namespace/:name', meta: { title: 'Deployment详情', parent: 'WorkloadDeploymentList' } }
```

### Breadcrumb Logic
```ts
const breadcrumbs = computed(() => {
  const matched = route.matched
  return matched
    .filter(r => r.meta?.title)
    .map(r => ({ title: r.meta.title, path: r.path }))
})
```

### Files Modified
- `frontend/src/router/index.ts` (add meta to all routes)
- `frontend/src/components/Layout/Header.vue` (breadcrumb logic)

---

## 3. Auto-Refresh with Polling

### Problem
Dashboard and list pages require manual refresh. No auto-update mechanism.

### Solution
Create a `useAutoRefresh` composable that polls data at configurable intervals.

### Composable Design
```ts
// src/composables/useAutoRefresh.ts
export function useAutoRefresh(fetchFn: () => Promise<void>, interval = 15000) {
  const isRunning = ref(true)
  const countdown = ref(interval / 1000)
  let timer: ReturnType<typeof setInterval> | null = null
  let countdownTimer: ReturnType<typeof setInterval> | null = null

  const start = () => { /* start polling */ }
  const stop = () => { /* stop polling */ }
  const toggle = () => { /* toggle polling */ }

  // Auto-cleanup on unmount
  onUnmounted(() => stop())

  return { isRunning, countdown, toggle, refresh: fetchFn }
}
```

### Usage in Pages
```ts
// In DeploymentList.vue
const { isRunning, countdown, toggle, refresh } = useAutoRefresh(fetchDeployments, 15000)
```

### UI
- Refresh button shows countdown: `刷新 (15s)`
- Pause button toggles to `已暂停` state
- Manual refresh button calls `refresh()` directly

### Files Created
- `frontend/src/composables/useAutoRefresh.ts`

### Files Modified
- `frontend/src/views/dashboard/DashboardView.vue`
- `frontend/src/views/workload/DeploymentList.vue`
- `frontend/src/views/workload/PodList.vue`
- (All other list pages)

---

## 4. English Translation Fix

### Problem
`en.ts` contains many Chinese strings mixed into English values (e.g., `"deployment": "无Status负载"`).

### Solution
Regenerate `en.ts` with proper English translations for all keys.

### Approach
- Read current `zh-CN.ts` as the source of truth for key structure
- Rewrite `en.ts` with proper English values
- Ensure 1:1 key mapping between zh-CN and en

### Files Modified
- `frontend/src/locales/en.ts`

---

## Implementation Order

1. **Nested routes + AppLayout** (structural foundation)
2. **Breadcrumb enhancement** (depends on route meta from step 1)
3. **Auto-refresh composable** (independent, but benefits from working layout)
4. **English translation fix** (independent, lowest risk)

## Risks

- **Nested routes**: May break existing navigation if route paths change. Mitigation: keep all paths identical, only add parent wrapper.
- **Breadcrumb**: Requires adding meta to ~50 routes. Tedious but low risk.
- **Auto-refresh**: Polling may cause flicker if not debounced. Mitigation: use `v-loading` only on initial load, not on refresh.
- **Translation**: Large file edit. Mitigation: systematic key-by-key approach.
