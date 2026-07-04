# GKube UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform GKube's frontend from Element Plus default theme to a Rancher-style K8s management UI with light/dark dual theme support.

**Architecture:** CSS custom properties (design tokens) define all colors, spacing, radii, and shadows. Element Plus CSS variables are overridden to map to these tokens. A `data-theme` attribute on `<html>` switches between light and dark themes. All hardcoded colors in components are replaced with token references.

**Tech Stack:** Vue 3, Element Plus 2.x, CSS Custom Properties, TypeScript

## Global Constraints

- All new CSS files go under `frontend/src/styles/`
- Token prefix: `--gk-` for all custom properties
- Theme attribute: `data-theme="light"` or `data-theme="dark"` on `document.documentElement`
- LocalStorage key: `gk-theme` for persistence
- Element Plus version: ^2.14.1 (CSS variables supported natively)
- No new dependencies required — pure CSS + TypeScript

---

### Task 1: Create Design Token System

**Files:**
- Create: `frontend/src/styles/tokens.css`
- Create: `frontend/src/styles/themes/light.css`
- Create: `frontend/src/styles/themes/dark.css`

**Interfaces:**
- Produces: CSS custom properties consumed by all subsequent tasks
- Consumed by: every `.vue` file and `element-overrides.css`

- [ ] **Step 1: Create tokens.css**

Create the base design token definitions at `frontend/src/styles/tokens.css`:

```css
/* ============================================================
   GKube Design Tokens
   Rancher-inspired design system for K8s multi-cluster management
   ============================================================ */

:root {
  /* === Color System === */
  /* Primary - Rancher Blue */
  --gk-color-primary: #3b82f6;
  --gk-color-primary-light: #60a5fa;
  --gk-color-primary-dark: #2563eb;
  --gk-color-primary-bg: #eff6ff;

  /* Semantic Colors */
  --gk-color-success: #22c55e;
  --gk-color-success-light: #4ade80;
  --gk-color-success-bg: #f0fdf4;
  --gk-color-warning: #f59e0b;
  --gk-color-warning-light: #fbbf24;
  --gk-color-warning-bg: #fffbeb;
  --gk-color-danger: #ef4444;
  --gk-color-danger-light: #f87171;
  --gk-color-danger-bg: #fef2f2;
  --gk-color-info: #6366f1;
  --gk-color-info-light: #818cf8;
  --gk-color-info-bg: #eef2ff;

  /* Neutral - resolved by theme files */
  --gk-color-text-primary: var(--gk-neutral-900);
  --gk-color-text-secondary: var(--gk-neutral-500);
  --gk-color-text-placeholder: var(--gk-neutral-400);
  --gk-color-text-disabled: var(--gk-neutral-300);
  --gk-color-bg-page: var(--gk-neutral-50);
  --gk-color-bg-card: var(--gk-white);
  --gk-color-bg-sidebar: var(--gk-neutral-900);
  --gk-color-bg-header: var(--gk-white);
  --gk-color-border: var(--gk-neutral-200);
  --gk-color-border-light: var(--gk-neutral-100);

  /* === Spacing System (4px base) === */
  --gk-space-1: 4px;
  --gk-space-2: 8px;
  --gk-space-3: 12px;
  --gk-space-4: 16px;
  --gk-space-5: 20px;
  --gk-space-6: 24px;
  --gk-space-8: 32px;
  --gk-space-10: 40px;
  --gk-space-12: 48px;

  /* === Border Radius === */
  --gk-radius-sm: 4px;
  --gk-radius-md: 8px;
  --gk-radius-lg: 12px;
  --gk-radius-xl: 16px;
  --gk-radius-full: 9999px;

  /* === Shadows === */
  --gk-shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --gk-shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  --gk-shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  --gk-shadow-card: 0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.06);

  /* === Sidebar Specific === */
  --gk-sidebar-text: rgba(255, 255, 255, 0.7);
  --gk-sidebar-text-active: #ffffff;
  --gk-sidebar-hover-bg: rgba(255, 255, 255, 0.08);
  --gk-sidebar-active-bg: rgba(59, 130, 246, 0.2);
  --gk-sidebar-active-indicator: #3b82f6;

  /* === Typography === */
  --gk-font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --gk-font-mono: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
  --gk-font-size-xs: 12px;
  --gk-font-size-sm: 13px;
  --gk-font-size-base: 14px;
  --gk-font-size-lg: 16px;
  --gk-font-size-xl: 18px;
  --gk-font-size-2xl: 20px;
  --gk-font-size-3xl: 24px;

  /* === Layout === */
  --gk-sidebar-width: 240px;
  --gk-sidebar-collapsed-width: 64px;
  --gk-header-height: 56px;
  --gk-content-max-width: 1400px;

  /* === Transitions === */
  --gk-transition-fast: 0.15s ease;
  --gk-transition-base: 0.2s ease;
  --gk-transition-slow: 0.3s ease;
}
```

- [ ] **Step 2: Create light.css theme**

Create `frontend/src/styles/themes/light.css`:

```css
/* ============================================================
   Light Theme (Default)
   ============================================================ */

[data-theme="light"],
:root {
  --gk-white: #ffffff;

  /* Neutral Scale */
  --gk-neutral-50: #f8fafc;
  --gk-neutral-100: #f1f5f9;
  --gk-neutral-200: #e2e8f0;
  --gk-neutral-300: #cbd5e1;
  --gk-neutral-400: #94a3b8;
  --gk-neutral-500: #64748b;
  --gk-neutral-600: #475569;
  --gk-neutral-700: #334155;
  --gk-neutral-800: #1e293b;
  --gk-neutral-900: #0f172a;

  /* Semantic Overrides */
  --gk-color-bg-page: #f8fafc;
  --gk-color-bg-card: #ffffff;
  --gk-color-bg-sidebar: #0f172a;
  --gk-color-bg-header: #ffffff;
  --gk-color-text-primary: #0f172a;
  --gk-color-text-secondary: #64748b;
  --gk-color-text-placeholder: #94a3b8;
  --gk-color-border: #e2e8f0;
}
```

- [ ] **Step 3: Create dark.css theme**

Create `frontend/src/styles/themes/dark.css`:

```css
/* ============================================================
   Dark Theme
   ============================================================ */

[data-theme="dark"] {
  --gk-white: #0f172a;

  /* Neutral Scale (inverted) */
  --gk-neutral-50: #1e293b;
  --gk-neutral-100: #1e293b;
  --gk-neutral-200: #334155;
  --gk-neutral-300: #475569;
  --gk-neutral-400: #64748b;
  --gk-neutral-500: #94a3b8;
  --gk-neutral-600: #cbd5e1;
  --gk-neutral-700: #e2e8f0;
  --gk-neutral-800: #f1f5f9;
  --gk-neutral-900: #f8fafc;

  /* Semantic Overrides */
  --gk-color-bg-page: #0f172a;
  --gk-color-bg-card: #1e293b;
  --gk-color-bg-sidebar: #020617;
  --gk-color-bg-header: #1e293b;
  --gk-color-text-primary: #f1f5f9;
  --gk-color-text-secondary: #94a3b8;
  --gk-color-text-placeholder: #64748b;
  --gk-color-border: #334155;

  /* Sidebar specific - darker theme adjustments */
  --gk-sidebar-text: rgba(255, 255, 255, 0.6);
  --gk-sidebar-hover-bg: rgba(255, 255, 255, 0.06);
  --gk-sidebar-active-bg: rgba(59, 130, 246, 0.25);
  --gk-sidebar-active-indicator: #60a5fa;

  /* Shadows - stronger for dark mode visibility */
  --gk-shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.2);
  --gk-shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
  --gk-shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.3);
  --gk-shadow-card: 0 1px 3px rgba(0, 0, 0, 0.25), 0 1px 2px rgba(0, 0, 0, 0.15);

  /* Semantic color backgrounds - muted for dark */
  --gk-color-primary-bg: rgba(59, 130, 246, 0.15);
  --gk-color-success-bg: rgba(34, 197, 94, 0.15);
  --gk-color-warning-bg: rgba(245, 158, 11, 0.15);
  --gk-color-danger-bg: rgba(239, 68, 68, 0.15);
  --gk-color-info-bg: rgba(99, 102, 241, 0.15);
}
```

- [ ] **Step 4: Verify files exist**

Run:
```bash
ls -la frontend/src/styles/
ls -la frontend/src/styles/themes/
```

Expected: Three files created — `tokens.css`, `themes/light.css`, `themes/dark.css`

- [ ] **Step 5: Commit**

```bash
git add frontend/src/styles/
git commit -m "feat: add design token system with light/dark theme variables"
```

---

### Task 2: Create Element Plus Variable Overrides

**Files:**
- Create: `frontend/src/styles/element-overrides.css`

**Interfaces:**
- Consumes: CSS custom properties from `tokens.css` + theme files
- Produces: Element Plus CSS variable overrides consumed by all Element Plus components

- [ ] **Step 1: Create element-overrides.css**

Create `frontend/src/styles/element-overrides.css`:

```css
/* ============================================================
   Element Plus CSS Variable Overrides
   Maps Element Plus design tokens to GKube tokens
   ============================================================ */

:root {
  /* === Primary Color === */
  --el-color-primary: var(--gk-color-primary);
  --el-color-primary-light-3: var(--gk-color-primary-light);
  --el-color-primary-light-5: var(--gk-color-primary-bg);
  --el-color-primary-light-7: var(--gk-color-primary-bg);
  --el-color-primary-light-8: var(--gk-color-primary-bg);
  --el-color-primary-light-9: var(--gk-color-primary-bg);
  --el-color-primary-dark-2: var(--gk-color-primary-dark);

  /* === Success Color === */
  --el-color-success: var(--gk-color-success);
  --el-color-success-light-3: var(--gk-color-success-light);
  --el-color-success-light-5: var(--gk-color-success-bg);
  --el-color-success-light-7: var(--gk-color-success-bg);
  --el-color-success-light-8: var(--gk-color-success-bg);
  --el-color-success-light-9: var(--gk-color-success-bg);
  --el-color-success-dark-2: #16a34a;

  /* === Warning Color === */
  --el-color-warning: var(--gk-color-warning);
  --el-color-warning-light-3: var(--gk-color-warning-light);
  --el-color-warning-light-5: var(--gk-color-warning-bg);
  --el-color-warning-light-7: var(--gk-color-warning-bg);
  --el-color-warning-light-8: var(--gk-color-warning-bg);
  --el-color-warning-light-9: var(--gk-color-warning-bg);
  --el-color-warning-dark-2: #d97706;

  /* === Danger Color === */
  --el-color-danger: var(--gk-color-danger);
  --el-color-danger-light-3: var(--gk-color-danger-light);
  --el-color-danger-light-5: var(--gk-color-danger-bg);
  --el-color-danger-light-7: var(--gk-color-danger-bg);
  --el-color-danger-light-8: var(--gk-color-danger-bg);
  --el-color-danger-light-9: var(--gk-color-danger-bg);
  --el-color-danger-dark-2: #dc2626;

  /* === Info Color === */
  --el-color-info: var(--gk-color-info);
  --el-color-info-light-3: var(--gk-color-info-light);
  --el-color-info-light-5: var(--gk-color-info-bg);
  --el-color-info-light-7: var(--gk-color-info-bg);
  --el-color-info-light-8: var(--gk-color-info-bg);
  --el-color-info-light-9: var(--gk-color-info-bg);
  --el-color-info-dark-2: #4f46e5;

  /* === Background === */
  --el-bg-color: var(--gk-color-bg-card);
  --el-bg-color-page: var(--gk-color-bg-page);
  --el-bg-color-overlay: var(--gk-color-bg-card);

  /* === Fill === */
  --el-fill-color: var(--gk-neutral-100);
  --el-fill-color-light: var(--gk-neutral-100);
  --el-fill-color-lighter: var(--gk-neutral-50);
  --el-fill-color-extra-light: var(--gk-neutral-50);
  --el-fill-color-dark: var(--gk-neutral-200);
  --el-fill-color-blank: var(--gk-color-bg-card);

  /* === Text === */
  --el-text-color-primary: var(--gk-color-text-primary);
  --el-text-color-regular: var(--gk-color-text-primary);
  --el-text-color-secondary: var(--gk-color-text-secondary);
  --el-text-color-placeholder: var(--gk-color-text-placeholder);
  --el-text-color-disabled: var(--gk-color-text-disabled);

  /* === Border === */
  --el-border-color: var(--gk-color-border);
  --el-border-color-light: var(--gk-color-border);
  --el-border-color-lighter: var(--gk-color-border-light);
  --el-border-color-extra-light: var(--gk-color-border-light);
  --el-border-color-dark: var(--gk-neutral-300);

  /* === Border Radius === */
  --el-border-radius-base: var(--gk-radius-md);
  --el-border-radius-small: var(--gk-radius-sm);
  --el-border-radius-round: var(--gk-radius-full);
  --el-border-radius-circle: var(--gk-radius-full);

  /* === Box Shadow === */
  --el-box-shadow: var(--gk-shadow-md);
  --el-box-shadow-light: var(--gk-shadow-sm);
  --el-box-shadow-lighter: var(--gk-shadow-sm);
  --el-box-shadow-dark: var(--gk-shadow-lg);

  /* === Font === */
  --el-font-family: var(--gk-font-sans);
  --el-font-size-base: var(--gk-font-size-base);
  --el-font-size-small: var(--gk-font-size-sm);
  --el-font-size-extra-small: var(--gk-font-size-xs);
  --el-font-size-large: var(--gk-font-size-lg);

  /* === Menu (Sidebar) === */
  --el-menu-bg-color: var(--gk-color-bg-sidebar);
  --el-menu-text-color: var(--gk-sidebar-text);
  --el-menu-active-color: var(--gk-sidebar-text-active);
  --el-menu-hover-bg-color: var(--gk-sidebar-hover-bg);

  /* === Card === */
  --el-card-border-radius: var(--gk-radius-lg);

  /* === Table === */
  --el-table-border-color: var(--gk-color-border-light);
  --el-table-header-bg-color: var(--gk-neutral-50);
  --el-table-row-hover-bg-color: var(--gk-color-primary-bg);

  /* === Input === */
  --el-input-border-radius: var(--gk-radius-md);

  /* === Scrollbar === */
  --el-scrollbar-opacity-hover: 0.6;
}
```

- [ ] **Step 2: Verify file exists**

Run:
```bash
cat frontend/src/styles/element-overrides.css | head -5
```

Expected: First 5 lines of the file with the header comment.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/styles/element-overrides.css
git commit -m "feat: add Element Plus CSS variable overrides for GKube theme"
```

---

### Task 3: Create Theme Switcher

**Files:**
- Create: `frontend/src/styles/theme-switcher.ts`

**Interfaces:**
- Produces: `initTheme()`, `setTheme()`, `getTheme()`, `toggleTheme()`, `useTheme()` composable
- Consumed by: `main.ts` (initTheme), `Header.vue` (useTheme)

- [ ] **Step 1: Create theme-switcher.ts**

Create `frontend/src/styles/theme-switcher.ts`:

```typescript
import { ref, watchEffect } from 'vue'

export type Theme = 'light' | 'dark'

const THEME_KEY = 'gk-theme'

/**
 * Get the stored theme or default to 'light'
 */
export function getTheme(): Theme {
  const stored = localStorage.getItem(THEME_KEY)
  if (stored === 'light' || stored === 'dark') return stored
  // Respect system preference
  if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark'
  return 'light'
}

/**
 * Apply theme to document and persist to localStorage
 */
export function setTheme(theme: Theme): void {
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem(THEME_KEY, theme)
}

/**
 * Initialize theme on app startup (call once in main.ts)
 */
export function initTheme(): void {
  setTheme(getTheme())
}

/**
 * Composable for reactive theme state
 */
export function useTheme() {
  const currentTheme = ref<Theme>(getTheme())

  function toggle() {
    currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
  }

  function set(theme: Theme) {
    currentTheme.value = theme
  }

  // Sync to DOM and localStorage whenever it changes
  watchEffect(() => {
    setTheme(currentTheme.value)
  })

  const isDark = ref(currentTheme.value === 'dark')

  watchEffect(() => {
    isDark.value = currentTheme.value === 'dark'
  })

  return {
    currentTheme,
    isDark,
    toggle,
    set,
  }
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run:
```bash
cd frontend && npx vue-tsc --noEmit --pretty 2>&1 | grep "theme-switcher" || echo "No errors in theme-switcher.ts"
```

Expected: No errors related to theme-switcher.ts.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/styles/theme-switcher.ts
git commit -m "feat: add theme switcher composable with light/dark toggle"
```

---

### Task 4: Create Style Entry Point and Update main.ts

**Files:**
- Create: `frontend/src/styles/index.css`
- Modify: `frontend/src/main.ts`
- Modify: `frontend/src/App.vue`

**Interfaces:**
- Consumes: all CSS files from `styles/` directory
- Produces: single import point for all styles

- [ ] **Step 1: Create styles/index.css**

Create `frontend/src/styles/index.css`:

```css
/* ============================================================
   GKube Styles Entry Point
   Import order matters: tokens first, then themes, then overrides
   ============================================================ */

/* Design Tokens */
@import './tokens.css';

/* Theme Definitions */
@import './themes/light.css';
@import './themes/dark.css';

/* Element Plus Overrides */
@import './element-overrides.css';
```

- [ ] **Step 2: Update main.ts**

Replace the contents of `frontend/src/main.ts`:

```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { install as MonacoVueEditor } from '@guolao/vue-monaco-editor'
import router from './router'
import i18n from './locales'
import App from './App.vue'
import { initTheme } from './styles/theme-switcher'
import './styles/index.css'
import './style.css'

// Initialize theme before mounting
initTheme()

const app = createApp(App)

// Register all Element Plus icons globally
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(ElementPlus)
app.use(MonacoVueEditor)
app.use(i18n)
app.use(router)
app.mount('#app')
```

- [ ] **Step 3: Clean up App.vue**

Replace the contents of `frontend/src/App.vue` — remove duplicate global reset (already in style.css):

```vue
<template>
  <router-view />
</template>
```

- [ ] **Step 4: Verify dev server starts**

Run:
```bash
cd frontend && npm run dev 2>&1 &
sleep 5
curl -s http://localhost:5173 | head -20
kill %1
```

Expected: HTML response with the app shell.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/main.ts frontend/src/App.vue frontend/src/styles/index.css
git commit -m "feat: integrate theme system into app bootstrap"
```

---

### Task 5: Redesign AppLayout Component

**Files:**
- Modify: `frontend/src/components/Layout/AppLayout.vue`

**Interfaces:**
- Consumes: CSS custom properties from tokens
- Produces: layout structure consumed by Sidebar and Header

- [ ] **Step 1: Rewrite AppLayout.vue**

Replace the contents of `frontend/src/components/Layout/AppLayout.vue`:

```vue
<template>
  <el-container class="app-layout">
    <el-aside
      :width="isCollapse ? 'var(--gk-sidebar-collapsed-width)' : 'var(--gk-sidebar-width)'"
      class="app-aside"
    >
      <Sidebar :is-collapse="isCollapse" />
    </el-aside>
    <el-container class="app-content-wrapper">
      <el-header class="app-header">
        <Header @toggle-collapse="isCollapse = !isCollapse" />
      </el-header>
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'

const isCollapse = ref(false)
</script>

<style scoped>
.app-layout {
  height: 100vh;
  overflow: hidden;
}

.app-aside {
  background: var(--gk-color-bg-sidebar);
  transition: width var(--gk-transition-slow);
  overflow-x: hidden;
  overflow-y: auto;
  border-right: 1px solid var(--gk-color-border);
}

.app-header {
  padding: 0;
  height: var(--gk-header-height);
  border-bottom: 1px solid var(--gk-color-border);
  box-shadow: var(--gk-shadow-sm);
  background: var(--gk-color-bg-header);
}

.app-main {
  background: var(--gk-color-bg-page);
  padding: 0;
  overflow-y: auto;
  min-height: 0;
}

.app-content-wrapper {
  overflow: hidden;
}

/* Responsive: auto-collapse sidebar on small screens */
@media (max-width: 768px) {
  .app-aside {
    position: fixed;
    z-index: 1000;
    height: 100vh;
    box-shadow: var(--gk-shadow-lg);
  }
}
</style>
```

- [ ] **Step 2: Verify layout renders**

Run:
```bash
cd frontend && npm run dev 2>&1 &
sleep 5
curl -s http://localhost:5173 | grep -o "app-layout" | head -1
kill %1
```

Expected: `app-layout` found in HTML.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/Layout/AppLayout.vue
git commit -m "refactor: redesign AppLayout with CSS custom properties"
```

---

### Task 6: Redesign Sidebar Component

**Files:**
- Modify: `frontend/src/components/Layout/Sidebar.vue`

**Interfaces:**
- Consumes: `isCollapse` prop, CSS custom properties
- Produces: Rancher-style sidebar navigation

- [ ] **Step 1: Rewrite Sidebar.vue styles**

Replace the contents of `frontend/src/components/Layout/Sidebar.vue`:

```vue
<template>
  <div class="sidebar-container">
    <div class="sidebar-logo">
      <div class="logo-icon">
        <svg viewBox="0 0 32 32" width="28" height="28" fill="none">
          <rect width="32" height="32" rx="8" fill="#3b82f6"/>
          <path d="M8 16L14 22L24 10" stroke="white" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
      <transition name="fade">
        <span v-show="!isCollapse" class="logo-text">GKube</span>
      </transition>
    </div>
    <el-menu
      :default-active="activeMenu"
      :collapse="isCollapse"
      :collapse-transition="false"
      router
      class="sidebar-menu"
    >
      <el-menu-item index="/dashboard">
        <el-icon><Odometer /></el-icon>
        <template #title>{{ t('sidebar.dashboard') }}</template>
      </el-menu-item>
      <el-menu-item index="/system/overview">
        <el-icon><Monitor /></el-icon>
        <template #title>{{ t('sidebar.systemOverview') }}</template>
      </el-menu-item>
      <el-menu-item index="/clusters">
        <el-icon><Connection /></el-icon>
        <template #title>{{ t('sidebar.clusters') }}</template>
      </el-menu-item>
      <el-sub-menu index="workloads">
        <template #title>
          <el-icon><Box /></el-icon>
          <span>{{ t('sidebar.workloads') }}</span>
        </template>
        <el-menu-item index="/workloads/pods">
          <el-icon><Coin /></el-icon>
          <template #title>{{ t('sidebar.pods') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/deployments">
          <el-icon><DocumentCopy /></el-icon>
          <template #title>{{ t('sidebar.deployments') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/statefulsets">
          <el-icon><List /></el-icon>
          <template #title>{{ t('sidebar.statefulsets') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/daemonsets">
          <el-icon><SetUp /></el-icon>
          <template #title>{{ t('sidebar.daemonsets') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/jobs">
          <el-icon><Finished /></el-icon>
          <template #title>{{ t('sidebar.jobs') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/cronjobs">
          <el-icon><Timer /></el-icon>
          <template #title>{{ t('sidebar.cronjobs') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/hpa">
          <el-icon><DataLine /></el-icon>
          <template #title>{{ t('sidebar.hpa') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/pdb">
          <el-icon><Warning /></el-icon>
          <template #title>{{ t('sidebar.pdb') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="config">
        <template #title>
          <el-icon><Tickets /></el-icon>
          <span>{{ t('sidebar.config') }}</span>
        </template>
        <el-menu-item index="/config/configmaps">
          <el-icon><Tickets /></el-icon>
          <template #title>{{ t('sidebar.configmaps') }}</template>
        </el-menu-item>
        <el-menu-item index="/config/secrets">
          <el-icon><Key /></el-icon>
          <template #title>{{ t('sidebar.secrets') }}</template>
        </el-menu-item>
        <el-menu-item index="/config/resourcequotas">
          <el-icon><Coin /></el-icon>
          <template #title>{{ t('sidebar.resourcequotas') }}</template>
        </el-menu-item>
        <el-menu-item index="/config/limitranges">
          <el-icon><ScaleToOriginal /></el-icon>
          <template #title>{{ t('sidebar.limitranges') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="storage">
        <template #title>
          <el-icon><Coin /></el-icon>
          <span>{{ t('sidebar.storage') }}</span>
        </template>
        <el-menu-item index="/storage/pvs">
          <el-icon><Coin /></el-icon>
          <template #title>{{ t('sidebar.pvs') }}</template>
        </el-menu-item>
        <el-menu-item index="/storage/pvcs">
          <el-icon><Box /></el-icon>
          <template #title>{{ t('sidebar.pvcs') }}</template>
        </el-menu-item>
        <el-menu-item index="/storage/storageclasses">
          <el-icon><Files /></el-icon>
          <template #title>{{ t('sidebar.storageclasses') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="network">
        <template #title>
          <el-icon><Share /></el-icon>
          <span>{{ t('sidebar.network') }}</span>
        </template>
        <el-menu-item index="/services">
          <el-icon><Connection /></el-icon>
          <template #title>{{ t('sidebar.services') }}</template>
        </el-menu-item>
        <el-menu-item index="/ingresses">
          <el-icon><Link /></el-icon>
          <template #title>{{ t('sidebar.ingresses') }}</template>
        </el-menu-item>
        <el-menu-item index="/network/networkpolicies">
          <el-icon><Lock /></el-icon>
          <template #title>{{ t('sidebar.networkpolicies') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <el-menu-item index="/nodes">
        <el-icon><Cpu /></el-icon>
        <template #title>{{ t('sidebar.nodes') }}</template>
      </el-menu-item>
      <el-menu-item index="/namespaces">
        <el-icon><FolderOpened /></el-icon>
        <template #title>{{ t('sidebar.namespaces') }}</template>
      </el-menu-item>
      <el-menu-item index="/events">
        <el-icon><Bell /></el-icon>
        <template #title>{{ t('sidebar.events') }}</template>
      </el-menu-item>
      <el-menu-item index="/crd">
        <el-icon><Grid /></el-icon>
        <template #title>{{ t('sidebar.crd') }}</template>
      </el-menu-item>
      <el-sub-menu index="system">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>{{ t('sidebar.system') }}</span>
        </template>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <template #title>{{ t('sidebar.users') }}</template>
        </el-menu-item>
        <el-menu-item index="/roles">
          <el-icon><UserFilled /></el-icon>
          <template #title>{{ t('sidebar.roles') }}</template>
        </el-menu-item>
        <el-menu-item index="/settings/auth">
          <el-icon><Setting /></el-icon>
          <template #title>{{ t('sidebar.authSettings') }}</template>
        </el-menu-item>
        <el-menu-item index="/audit">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('sidebar.audit') }}</template>
        </el-menu-item>
      </el-sub-menu>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Odometer,
  Connection,
  Setting,
  User,
  UserFilled,
  Monitor,
  Document,
  Box,
  Coin,
  Files,
  Share,
  Link,
  Cpu,
  FolderOpened,
  Tickets,
  Key,
  Bell,
  DataLine,
  Lock,
  Warning,
  Grid,
  ScaleToOriginal,
  DocumentCopy,
  List,
  SetUp,
  Finished,
  Timer,
} from '@element-plus/icons-vue'

defineProps<{
  isCollapse: boolean
}>()

const route = useRoute()
const { t } = useI18n()
const activeMenu = computed(() => route.path)
</script>

<style scoped>
.sidebar-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--gk-color-bg-sidebar);
}

.sidebar-logo {
  height: var(--gk-header-height);
  display: flex;
  align-items: center;
  padding: 0 var(--gk-space-4);
  gap: var(--gk-space-3);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  flex-shrink: 0;
}

.logo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-text {
  color: #ffffff;
  font-size: var(--gk-font-size-xl);
  font-weight: 700;
  white-space: nowrap;
  letter-spacing: -0.02em;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
  overflow-y: auto;
  overflow-x: hidden;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: var(--gk-sidebar-width);
}

/* Override Element Plus menu item styles */
.sidebar-menu .el-menu-item,
.sidebar-menu :deep(.el-sub-menu__title) {
  height: 44px;
  line-height: 44px;
  margin: 2px var(--gk-space-2);
  border-radius: var(--gk-radius-md);
  transition: all var(--gk-transition-fast);
}

.sidebar-menu .el-menu-item:hover,
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  background-color: var(--gk-sidebar-hover-bg);
}

.sidebar-menu .el-menu-item.is-active {
  background-color: var(--gk-sidebar-active-bg);
  color: var(--gk-sidebar-text-active);
  position: relative;
}

.sidebar-menu .el-menu-item.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 0 2px 2px 0;
  background-color: var(--gk-sidebar-active-indicator);
}

/* Sub-menu items - more indentation */
.sidebar-menu .el-sub-menu .el-menu-item {
  padding-left: 52px !important;
}

/* Scrollbar styling for sidebar */
.sidebar-menu::-webkit-scrollbar {
  width: 4px;
}

.sidebar-menu::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 2px;
}

.sidebar-menu::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.25);
}

.sidebar-menu::-webkit-scrollbar-track {
  background: transparent;
}

/* Fade transition for logo text */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--gk-transition-fast);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
```

- [ ] **Step 2: Verify sidebar renders**

Run:
```bash
cd frontend && npm run dev 2>&1 &
sleep 5
curl -s http://localhost:5173 | grep -o "sidebar-container" | head -1
kill %1
```

Expected: `sidebar-container` found.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/Layout/Sidebar.vue
git commit -m "feat: redesign Sidebar with Rancher-style navigation"
```

---

### Task 7: Redesign Header Component with Theme Toggle

**Files:**
- Modify: `frontend/src/components/Layout/Header.vue`

**Interfaces:**
- Consumes: `useTheme()` from `theme-switcher.ts`, CSS custom properties
- Produces: Header with theme toggle button

- [ ] **Step 1: Rewrite Header.vue**

Replace the contents of `frontend/src/components/Layout/Header.vue`:

```vue
<template>
  <div class="header">
    <div class="header-left">
      <el-icon
        class="collapse-btn"
        role="button"
        tabindex="0"
        :aria-label="t('common.toggleSidebar')"
        @click="$emit('toggleCollapse')"
        @keyup.enter="$emit('toggleCollapse')"
      >
        <Fold />
      </el-icon>
      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/dashboard' }">{{ t('common.home') }}</el-breadcrumb-item>
        <el-breadcrumb-item
          v-for="item in breadcrumbs"
          :key="item.path || item.title"
          :to="item.to"
        >
          {{ item.title }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>
    <div class="header-right">
      <!-- Cluster Selector -->
      <el-select
        v-model="clusterStore.currentCluster"
        value-key="id"
        :placeholder="t('common.selectCluster')"
        :loading="clusterLoading"
        size="small"
        class="cluster-select"
        clearable
        @change="handleClusterChange"
      >
        <template #prefix>
          <el-icon><Connection /></el-icon>
        </template>
        <el-option
          v-for="c in clusterStore.clusterList"
          :key="c.id"
          :label="c.clusterName"
          :value="c"
        />
      </el-select>

      <!-- Language Switcher -->
      <el-dropdown @command="handleLangChange">
        <el-button size="small" text class="header-action-btn">
          <el-icon><Switch /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="zh-CN">中文</el-dropdown-item>
            <el-dropdown-item command="en">English</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <!-- Theme Toggle -->
      <el-tooltip :content="isDark ? t('common.lightMode') : t('common.darkMode')" placement="bottom">
        <el-button size="small" text class="header-action-btn" @click="toggle()">
          <el-icon :size="18">
            <Sunny v-if="isDark" />
            <Moon v-else />
          </el-icon>
        </el-button>
      </el-tooltip>

      <!-- User Menu -->
      <el-dropdown @command="handleCommand">
        <div class="user-info">
          <el-avatar :size="32" class="user-avatar">
            {{ (authStore.user?.username || '?')[0].toUpperCase() }}
          </el-avatar>
          <span class="username">{{ authStore.user?.displayName || authStore.user?.username || '-' }}</span>
          <el-icon class="user-arrow"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>
              <el-icon><User /></el-icon>
              {{ authStore.user?.username }}
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon>
              {{ t('common.logout') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useClusterStore } from '@/stores/cluster'
import { useTheme } from '@/styles/theme-switcher'
import {
  Fold,
  Switch,
  ArrowDown,
  User,
  SwitchButton,
  Connection,
  Sunny,
  Moon,
} from '@element-plus/icons-vue'

defineEmits(['toggleCollapse'])
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const clusterStore = useClusterStore()
const { locale, t } = useI18n()
const clusterLoading = ref(false)
const { isDark, toggle } = useTheme()

const currentLang = computed(() => locale.value === 'zh-CN' ? '中文' : 'English')

const breadcrumbs = computed(() => {
  const items: Array<{ title: string; path?: string; to?: { path: string } }> = []

  if (route.meta?.parent) {
    const parentRoute = router.getRoutes().find(r => r.name === route.meta.parent)
    if (parentRoute?.meta?.title) {
      items.push({
        title: parentRoute.meta.title as string,
        path: parentRoute.path,
        to: { path: parentRoute.path },
      })
    }
  }

  if (route.meta?.title) {
    items.push({ title: route.meta.title as string })
  }

  return items
})

onMounted(async () => {
  clusterLoading.value = true
  try {
    await clusterStore.fetchClusters()
  } finally {
    clusterLoading.value = false
  }
})

function handleLangChange(lang: string) {
  locale.value = lang
  localStorage.setItem('gkube_locale', lang)
}

function handleClusterChange(val: any) {
  clusterStore.setCurrentCluster(val || null)
}

function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.header {
  height: var(--gk-header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--gk-space-5);
  background: var(--gk-color-bg-header);
  border-bottom: 1px solid var(--gk-color-border);
  box-shadow: var(--gk-shadow-sm);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--gk-space-4);
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: var(--gk-color-text-secondary);
  transition: color var(--gk-transition-fast);
  padding: var(--gk-space-1);
  border-radius: var(--gk-radius-sm);
}

.collapse-btn:hover,
.collapse-btn:focus-visible {
  color: var(--gk-color-primary);
  background: var(--gk-color-primary-bg);
  outline: none;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
}

.cluster-select {
  width: 200px;
}

.cluster-select :deep(.el-input__wrapper) {
  border-radius: var(--gk-radius-md);
}

.header-action-btn {
  color: var(--gk-color-text-secondary);
  border-radius: var(--gk-radius-md);
  padding: var(--gk-space-2);
}

.header-action-btn:hover {
  color: var(--gk-color-primary);
  background: var(--gk-color-primary-bg);
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
  cursor: pointer;
  padding: var(--gk-space-1) var(--gk-space-2);
  border-radius: var(--gk-radius-md);
  transition: background-color var(--gk-transition-fast);
}

.user-info:hover {
  background: var(--gk-neutral-100);
}

.user-avatar {
  background: var(--gk-color-primary);
  color: #ffffff;
  font-weight: 600;
  font-size: var(--gk-font-size-sm);
}

.username {
  font-size: var(--gk-font-size-base);
  color: var(--gk-color-text-primary);
  font-weight: 500;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-arrow {
  color: var(--gk-color-text-secondary);
  font-size: 12px;
}
</style>
```

- [ ] **Step 2: Verify theme toggle works**

Run:
```bash
cd frontend && npm run dev 2>&1 &
sleep 5
curl -s http://localhost:5173 | grep -o "theme-switcher" | head -1
kill %1
```

Expected: The page should load without errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/Layout/Header.vue
git commit -m "feat: redesign Header with theme toggle and Rancher-style layout"
```

---

### Task 8: Refactor Global Styles (style.css)

**Files:**
- Modify: `frontend/src/style.css`

**Interfaces:**
- Consumes: CSS custom properties from tokens
- Produces: global utility classes using tokens

- [ ] **Step 1: Rewrite style.css**

Replace the contents of `frontend/src/style.css`:

```css
/* ============================================================
   GKube Global Styles
   Uses design tokens from styles/tokens.css
   ============================================================ */

/* === Global Reset === */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  font-family: var(--gk-font-sans);
  font-size: var(--gk-font-size-base);
  color: var(--gk-color-text-primary);
  background: var(--gk-color-bg-page);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* === Scrollbar === */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-thumb {
  background: var(--gk-neutral-300);
  border-radius: var(--gk-radius-sm);
}

::-webkit-scrollbar-thumb:hover {
  background: var(--gk-neutral-400);
}

::-webkit-scrollbar-track {
  background: transparent;
}

/* Dark theme scrollbar */
[data-theme="dark"] ::-webkit-scrollbar-thumb {
  background: var(--gk-neutral-600);
}

[data-theme="dark"] ::-webkit-scrollbar-thumb:hover {
  background: var(--gk-neutral-500);
}

/* === Focus Visible === */
:focus-visible {
  outline: 2px solid var(--gk-color-primary);
  outline-offset: 2px;
}

button:focus-visible,
.el-button:focus-visible {
  outline: 2px solid var(--gk-color-primary);
  outline-offset: 2px;
  border-radius: var(--gk-radius-sm);
}

/* === Reduced Motion === */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

/* === Page Container === */
.page-container {
  padding: var(--gk-space-6);
  max-width: var(--gk-content-max-width);
  margin: 0 auto;
}

/* === Card Styles === */
.gk-card {
  background: var(--gk-color-bg-card);
  border-radius: var(--gk-radius-lg);
  box-shadow: var(--gk-shadow-card);
  padding: var(--gk-space-6);
  border: 1px solid var(--gk-color-border);
  transition: box-shadow var(--gk-transition-base);
}

.gk-card:hover {
  box-shadow: var(--gk-shadow-md);
}

.gk-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--gk-space-4);
  padding-left: var(--gk-space-3);
  border-left: 3px solid var(--gk-color-primary);
}

.gk-card__title {
  font-size: var(--gk-font-size-lg);
  font-weight: 600;
  color: var(--gk-color-text-primary);
}

/* === Element Plus Card Overrides === */
.el-card {
  border-radius: var(--gk-radius-lg);
  border-color: var(--gk-color-border);
  transition: box-shadow var(--gk-transition-base);
}

.el-card:hover {
  box-shadow: var(--gk-shadow-md);
}

/* === Filter Card === */
.filter-card {
  margin-bottom: var(--gk-space-4);
}

.filter-card .el-card__body {
  padding: var(--gk-space-4) var(--gk-space-5);
}

/* === Table Card === */
.table-card {
  margin-bottom: var(--gk-space-4);
}

.table-card .el-card__header {
  padding: var(--gk-space-3) var(--gk-space-5);
  border-bottom: 1px solid var(--gk-color-border);
}

/* === Filter Bar === */
.filter-bar {
  display: flex;
  align-items: center;
  gap: var(--gk-space-3);
  flex-wrap: wrap;
}

/* === Card Header === */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* === Detail Header === */
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--gk-space-6);
}

.detail-title {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
  font-size: var(--gk-font-size-2xl);
  font-weight: 600;
  color: var(--gk-color-text-primary);
}

.detail-actions {
  display: flex;
  gap: var(--gk-space-2);
}

/* === Table Overrides === */
.el-table {
  font-size: var(--gk-font-size-sm);
  border-radius: var(--gk-radius-lg);
  overflow: hidden;
}

.el-table th {
  background-color: var(--gk-neutral-50) !important;
  color: var(--gk-color-text-primary);
  font-weight: 600;
  font-size: var(--gk-font-size-sm);
}

.el-table td {
  padding: var(--gk-space-3) var(--gk-space-4);
}

.el-table--border .el-table__inner-wrapper::after,
.el-table--border::after,
.el-table--border::before,
.el-table__inner-wrapper::before {
  display: none;
}

/* === Workload Grid === */
.workload-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: var(--gk-space-4);
}

.workload-item {
  text-align: center;
  padding: var(--gk-space-4);
  background: var(--gk-neutral-50);
  border-radius: var(--gk-radius-lg);
  border: 1px solid var(--gk-color-border);
  transition: all var(--gk-transition-base);
}

.workload-item:hover {
  background: var(--gk-color-primary-bg);
  border-color: var(--gk-color-primary-light);
}

.workload-count {
  font-size: var(--gk-font-size-3xl);
  font-weight: 700;
  color: var(--gk-color-primary);
}

.workload-label {
  font-size: var(--gk-font-size-sm);
  color: var(--gk-color-text-secondary);
  margin-top: var(--gk-space-1);
}

/* === Status Colors (legacy utility classes - prefer .gk-status) === */
.status-running { color: var(--gk-color-success); }
.status-pending { color: var(--gk-color-warning); }
.status-failed { color: var(--gk-color-danger); }
.status-succeeded { color: var(--gk-color-primary); }

/* === Total Count Badge === */
.total-count {
  color: var(--gk-color-text-secondary);
  font-size: var(--gk-font-size-sm);
  margin-left: auto;
  padding: var(--gk-space-1) var(--gk-space-3);
  background: var(--gk-neutral-100);
  border-radius: var(--gk-radius-full);
  font-weight: 500;
}

/* === Load More === */
.load-more {
  display: flex;
  justify-content: center;
  padding: var(--gk-space-3) 0;
  border-top: 1px solid var(--gk-color-border-light);
}

/* === Table Actions === */
.table-actions {
  display: flex;
  gap: var(--gk-space-1);
}

/* === Empty State === */
.empty-state {
  padding: var(--gk-space-12) var(--gk-space-5);
  text-align: center;
  color: var(--gk-color-text-secondary);
}

.empty-state .el-icon {
  font-size: 48px;
  margin-bottom: var(--gk-space-3);
  color: var(--gk-color-text-placeholder);
}

/* === Transitions === */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--gk-transition-base);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* === Status Badge Component === */
.gk-status {
  display: inline-flex;
  align-items: center;
  gap: var(--gk-space-2);
  font-size: var(--gk-font-size-sm);
  font-weight: 500;
}

.gk-status__dot {
  width: 8px;
  height: 8px;
  border-radius: var(--gk-radius-full);
  flex-shrink: 0;
}

.gk-status--running .gk-status__dot { background: var(--gk-color-success); }
.gk-status--pending .gk-status__dot { background: var(--gk-color-warning); }
.gk-status--failed .gk-status__dot { background: var(--gk-color-danger); }
.gk-status--succeeded .gk-status__dot { background: var(--gk-color-primary); }
.gk-status--unknown .gk-status__dot { background: var(--gk-neutral-400); }
```

- [ ] **Step 2: Verify styles load**

Run:
```bash
cd frontend && npm run dev 2>&1 &
sleep 5
curl -s http://localhost:5173 | grep -o "gk-card" | head -1
kill %1
```

Expected: Page loads without errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/style.css
git commit -m "refactor: replace hardcoded colors with design token references"
```

---

### Task 9: Create Color Replacement Script and Migrate Workload Views

**Files:**
- Modify: All `.vue` files in `frontend/src/views/workload/`

**Interfaces:**
- Consumes: global CSS classes from `style.css`, CSS custom properties
- Produces: consistent list page styling across all workload views

- [ ] **Step 1: Run automated color replacement on workload views**

The following sed commands replace all hardcoded colors with token references. Run from the project root:

```bash
cd /Users/zqqzqq/05_github/gkube/frontend

# Find all .vue files in views/workload/
find src/views/workload -name "*.vue" -type f | while read file; do
  # Replace Element Plus default blue
  sed -i '' 's/#409eff/var(--gk-color-primary)/g' "$file"
  sed -i '' 's/#409EFF/var(--gk-color-primary)/g' "$file"
  # Replace success green
  sed -i '' 's/#67c23a/var(--gk-color-success)/g' "$file"
  sed -i '' 's/#67C23A/var(--gk-color-success)/g' "$file"
  # Replace warning orange
  sed -i '' 's/#e6a23c/var(--gk-color-warning)/g' "$file"
  sed -i '' 's/#E6A23C/var(--gk-color-warning)/g' "$file"
  # Replace danger red
  sed -i '' 's/#f56c6c/var(--gk-color-danger)/g' "$file"
  sed -i '' 's/#F56C6C/var(--gk-color-danger)/g' "$file"
  # Replace text colors
  sed -i '' 's/#606266/var(--gk-color-text-primary)/g' "$file"
  sed -i '' 's/#909399/var(--gk-color-text-secondary)/g' "$file"
  # Replace backgrounds
  sed -i '' 's/#f5f7fa/var(--gk-neutral-100)/g' "$file"
  sed -i '' 's/#F5F7FA/var(--gk-neutral-100)/g' "$file"
  sed -i '' 's/#f0f2f5/var(--gk-color-bg-page)/g' "$file"
  sed -i '' 's/#F0F2F5/var(--gk-color-bg-page)/g' "$file"
  # Replace borders
  sed -i '' 's/#e4e7ed/var(--gk-color-border)/g' "$file"
  sed -i '' 's/#E4E7ED/var(--gk-color-border)/g' "$file"
  sed -i '' 's/#ebeef5/var(--gk-color-border-light)/g' "$file"
  sed -i '' 's/#EBEEF5/var(--gk-color-border-light)/g' "$file"
  # Replace white backgrounds (only in style blocks, not SVGs)
  sed -i '' 's/background: #fff;/background: var(--gk-color-bg-card);/g' "$file"
  sed -i '' 's/background: #ffffff;/background: var(--gk-color-bg-card);/g' "$file"
  sed -i '' 's/background:#fff;/background: var(--gk-color-bg-card);/g' "$file"
  sed -i '' 's/background:#ffffff;/background: var(--gk-color-bg-card);/g' "$file"
  # Replace box-shadow hardcoded values
  sed -i '' 's/box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1)/box-shadow: var(--gk-shadow-md)/g' "$file"
  echo "Processed: $file"
done
```

Expected: All `.vue` files in `views/workload/` updated with token references.

- [ ] **Step 2: Verify replacements were applied**

Run:
```bash
grep -rn "#409eff\|#67c23a\|#e6a23c\|#f56c6c\|#606266\|#909399" frontend/src/views/workload/ --include="*.vue" | head -10
```

Expected: No matches (all hardcoded colors replaced).

- [ ] **Step 3: Verify build**

Run:
```bash
cd frontend && npm run build 2>&1 | tail -10
```

Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/workload/
git commit -m "refactor: migrate workload views to use design tokens"
```

---

### Task 10: Migrate Network, Storage, Config Views to Use Tokens

**Files:**
- Modify: All `.vue` files in `frontend/src/views/network/`
- Modify: All `.vue` files in `frontend/src/views/storage/`
- Modify: All `.vue` files in `frontend/src/views/config/`

**Interfaces:**
- Consumes: global CSS classes, CSS custom properties
- Produces: consistent styling across network, storage, config views

- [ ] **Step 1: Run automated color replacement on network, storage, config views**

Run the same sed replacement script from Task 9, targeting these directories:

```bash
cd /Users/zqqzqq/05_github/gkube/frontend

for dir in src/views/network src/views/storage src/views/config; do
  find "$dir" -name "*.vue" -type f | while read file; do
    sed -i '' 's/#409eff/var(--gk-color-primary)/g' "$file"
    sed -i '' 's/#409EFF/var(--gk-color-primary)/g' "$file"
    sed -i '' 's/#67c23a/var(--gk-color-success)/g' "$file"
    sed -i '' 's/#67C23A/var(--gk-color-success)/g' "$file"
    sed -i '' 's/#e6a23c/var(--gk-color-warning)/g' "$file"
    sed -i '' 's/#E6A23C/var(--gk-color-warning)/g' "$file"
    sed -i '' 's/#f56c6c/var(--gk-color-danger)/g' "$file"
    sed -i '' 's/#F56C6C/var(--gk-color-danger)/g' "$file"
    sed -i '' 's/#606266/var(--gk-color-text-primary)/g' "$file"
    sed -i '' 's/#909399/var(--gk-color-text-secondary)/g' "$file"
    sed -i '' 's/#f5f7fa/var(--gk-neutral-100)/g' "$file"
    sed -i '' 's/#F5F7FA/var(--gk-neutral-100)/g' "$file"
    sed -i '' 's/#f0f2f5/var(--gk-color-bg-page)/g' "$file"
    sed -i '' 's/#F0F2F5/var(--gk-color-bg-page)/g' "$file"
    sed -i '' 's/#e4e7ed/var(--gk-color-border)/g' "$file"
    sed -i '' 's/#E4E7ED/var(--gk-color-border)/g' "$file"
    sed -i '' 's/#ebeef5/var(--gk-color-border-light)/g' "$file"
    sed -i '' 's/#EBEEF5/var(--gk-color-border-light)/g' "$file"
    sed -i '' 's/background: #fff;/background: var(--gk-color-bg-card);/g' "$file"
    sed -i '' 's/background: #ffffff;/background: var(--gk-color-bg-card);/g' "$file"
    sed -i '' 's/background:#fff;/background: var(--gk-color-bg-card);/g' "$file"
    sed -i '' 's/background:#ffffff;/background: var(--gk-color-bg-card);/g' "$file"
    sed -i '' 's/box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1)/box-shadow: var(--gk-shadow-md)/g' "$file"
    echo "Processed: $file"
  done
done
```

Expected: All `.vue` files in the three directories updated.

- [ ] **Step 2: Verify build**

Run:
```bash
cd frontend && npm run build 2>&1 | tail -10
```

Expected: Build succeeds.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/network/ frontend/src/views/storage/ frontend/src/views/config/
git commit -m "refactor: migrate network, storage, config views to use design tokens"
```

---

### Task 11: Migrate All Remaining Views to Use Tokens

**Files:**
- Modify: All `.vue` files in `frontend/src/views/node/`
- Modify: All `.vue` files in `frontend/src/views/namespace/`
- Modify: All `.vue` files in `frontend/src/views/event/`
- Modify: All `.vue` files in `frontend/src/views/dashboard/`
- Modify: All `.vue` files in `frontend/src/views/crd/`
- Modify: All `.vue` files in `frontend/src/views/settings/`
- Modify: All `.vue` files in `frontend/src/views/login/`
- Modify: All `.vue` files in `frontend/src/views/audit/`
- Modify: All `.vue` files in `frontend/src/views/terminal/`
- Modify: All `.vue` files in `frontend/src/views/logviewer/`
- Modify: `frontend/src/views/ClusterCreate.vue`
- Modify: `frontend/src/views/ClusterDetail.vue`
- Modify: `frontend/src/views/ClusterList.vue`
- Modify: `frontend/src/views/RoleList.vue`
- Modify: `frontend/src/views/UserList.vue`

**Interfaces:**
- Consumes: global CSS classes, CSS custom properties
- Produces: consistent styling across all remaining views

- [ ] **Step 1: Run automated color replacement on all remaining subdirectory views**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend

for dir in src/views/node src/views/namespace src/views/event src/views/dashboard \
           src/views/crd src/views/settings src/views/login src/views/audit \
           src/views/terminal src/views/logviewer; do
  find "$dir" -name "*.vue" -type f 2>/dev/null | while read file; do
    sed -i '' 's/#409eff/var(--gk-color-primary)/g' "$file"
    sed -i '' 's/#409EFF/var(--gk-color-primary)/g' "$file"
    sed -i '' 's/#67c23a/var(--gk-color-success)/g' "$file"
    sed -i '' 's/#67C23A/var(--gk-color-success)/g' "$file"
    sed -i '' 's/#e6a23c/var(--gk-color-warning)/g' "$file"
    sed -i '' 's/#E6A23C/var(--gk-color-warning)/g' "$file"
    sed -i '' 's/#f56c6c/var(--gk-color-danger)/g' "$file"
    sed -i '' 's/#F56C6C/var(--gk-color-danger)/g' "$file"
    sed -i '' 's/#606266/var(--gk-color-text-primary)/g' "$file"
    sed -i '' 's/#909399/var(--gk-color-text-secondary)/g' "$file"
    sed -i '' 's/#f5f7fa/var(--gk-neutral-100)/g' "$file"
    sed -i '' 's/#F5F7FA/var(--gk-neutral-100)/g' "$file"
    sed -i '' 's/#f0f2f5/var(--gk-color-bg-page)/g' "$file"
    sed -i '' 's/#F0F2F5/var(--gk-color-bg-page)/g' "$file"
    sed -i '' 's/#e4e7ed/var(--gk-color-border)/g' "$file"
    sed -i '' 's/#E4E7ED/var(--gk-color-border)/g' "$file"
    sed -i '' 's/#ebeef5/var(--gk-color-border-light)/g' "$file"
    sed -i '' 's/#EBEEF5/var(--gk-color-border-light)/g' "$file"
    sed -i '' 's/background: #fff;/background: var(--gk-color-bg-card);/g' "$file"
    sed -i '' 's/background: #ffffff;/background: var(--gk-color-bg-card);/g' "$file"
    sed -i '' 's/background:#fff;/background: var(--gk-color-bg-card);/g' "$file"
    sed -i '' 's/background:#ffffff;/background: var(--gk-color-bg-card);/g' "$file"
    sed -i '' 's/box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1)/box-shadow: var(--gk-shadow-md)/g' "$file"
    echo "Processed: $file"
  done
done
```

- [ ] **Step 2: Run automated color replacement on root-level view files**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend

for file in src/views/ClusterCreate.vue src/views/ClusterDetail.vue \
            src/views/ClusterList.vue src/views/RoleList.vue src/views/UserList.vue; do
  if [ -f "$file" ]; then
    sed -i '' 's/#409eff/var(--gk-color-primary)/g' "$file"
    sed -i '' 's/#409EFF/var(--gk-color-primary)/g' "$file"
    sed -i '' 's/#67c23a/var(--gk-color-success)/g' "$file"
    sed -i '' 's/#67C23A/var(--gk-color-success)/g' "$file"
    sed -i '' 's/#e6a23c/var(--gk-color-warning)/g' "$file"
    sed -i '' 's/#E6A23C/var(--gk-color-warning)/g' "$file"
    sed -i '' 's/#f56c6c/var(--gk-color-danger)/g' "$file"
    sed -i '' 's/#F56C6C/var(--gk-color-danger)/g' "$file"
    sed -i '' 's/#606266/var(--gk-color-text-primary)/g' "$file"
    sed -i '' 's/#909399/var(--gk-color-text-secondary)/g' "$file"
    sed -i '' 's/#f5f7fa/var(--gk-neutral-100)/g' "$file"
    sed -i '' 's/#F5F7FA/var(--gk-neutral-100)/g' "$file"
    sed -i '' 's/#f0f2f5/var(--gk-color-bg-page)/g' "$file"
    sed -i '' 's/#F0F2F5/var(--gk-color-bg-page)/g' "$file"
    sed -i '' 's/#e4e7ed/var(--gk-color-border)/g' "$file"
    sed -i '' 's/#E4E7ED/var(--gk-color-border)/g' "$file"
    sed -i '' 's/#ebeef5/var(--gk-color-border-light)/g' "$file"
    sed -i '' 's/#EBEEF5/var(--gk-color-border-light)/g' "$file"
    sed -i '' 's/background: #fff;/background: var(--gk-color-bg-card);/g' "$file"
    sed -i '' 's/background: #ffffff;/background: var(--gk-color-bg-card);/g' "$file"
    echo "Processed: $file"
  fi
done
```

- [ ] **Step 3: Handle dashboard progressColor() function**

The dashboard's `progressColor()` function uses hardcoded hex values for echarts. These MUST remain as hex because echarts does not support CSS custom properties. Leave them as-is. The function already uses `#F56C6C`, `#E6A23C`, `#409EFF` which are standard echarts colors.

- [ ] **Step 4: Verify build**

Run:
```bash
cd frontend && npm run build 2>&1 | tail -10
```

Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/views/
git commit -m "refactor: migrate all remaining views to use design tokens"
```

---

### Task 12: Migrate Shared Components and FullscreenLayout to Use Tokens

**Files:**
- Modify: All `.vue` files in `frontend/src/components/`
- Modify: `frontend/src/components/Layout/FullscreenLayout.vue`

**Interfaces:**
- Consumes: CSS custom properties
- Produces: consistent component styling

- [ ] **Step 1: Run automated color replacement on all components**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend

find src/components -name "*.vue" -type f | while read file; do
  # Skip files that are already using tokens (Layout files done in Tasks 5-7)
  if echo "$file" | grep -q "Layout/AppLayout\|Layout/Sidebar\|Layout/Header"; then
    echo "Skipping (already migrated): $file"
    continue
  fi
  sed -i '' 's/#409eff/var(--gk-color-primary)/g' "$file"
  sed -i '' 's/#409EFF/var(--gk-color-primary)/g' "$file"
  sed -i '' 's/#67c23a/var(--gk-color-success)/g' "$file"
  sed -i '' 's/#67C23A/var(--gk-color-success)/g' "$file"
  sed -i '' 's/#e6a23c/var(--gk-color-warning)/g' "$file"
  sed -i '' 's/#E6A23C/var(--gk-color-warning)/g' "$file"
  sed -i '' 's/#f56c6c/var(--gk-color-danger)/g' "$file"
  sed -i '' 's/#F56C6C/var(--gk-color-danger)/g' "$file"
  sed -i '' 's/#606266/var(--gk-color-text-primary)/g' "$file"
  sed -i '' 's/#909399/var(--gk-color-text-secondary)/g' "$file"
  sed -i '' 's/#f5f7fa/var(--gk-neutral-100)/g' "$file"
  sed -i '' 's/#F5F7FA/var(--gk-neutral-100)/g' "$file"
  sed -i '' 's/#f0f2f5/var(--gk-color-bg-page)/g' "$file"
  sed -i '' 's/#F0F2F5/var(--gk-color-bg-page)/g' "$file"
  sed -i '' 's/#e4e7ed/var(--gk-color-border)/g' "$file"
  sed -i '' 's/#E4E7ED/var(--gk-color-border)/g' "$file"
  sed -i '' 's/#ebeef5/var(--gk-color-border-light)/g' "$file"
  sed -i '' 's/#EBEEF5/var(--gk-color-border-light)/g' "$file"
  sed -i '' 's/background: #fff;/background: var(--gk-color-bg-card);/g' "$file"
  sed -i '' 's/background: #ffffff;/background: var(--gk-color-bg-card);/g' "$file"
  sed -i '' 's/background:#fff;/background: var(--gk-color-bg-card);/g' "$file"
  sed -i '' 's/background:#ffffff;/background: var(--gk-color-bg-card);/g' "$file"
  echo "Processed: $file"
done
```

- [ ] **Step 2: Verify Monaco Editor colors are not broken**

Monaco Editor uses its own theme system (not CSS variables). Check `YamlEditor.vue` — if it sets `theme` prop on the editor, leave it as-is. Monaco's default dark/light theme will work correctly. Do NOT replace Monaco-specific hex colors.

Run:
```bash
grep -n "theme" frontend/src/components/YamlEditor.vue | head -5
```

Expected: Monaco theme configuration left intact.

- [ ] **Step 3: Verify build**

Run:
```bash
cd frontend && npm run build 2>&1 | tail -10
```

Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/
git commit -m "refactor: migrate shared components to use design tokens"
```

---

### Task 13: Final Verification and Cleanup

**Files:**
- Verify: all `.vue` and `.css` files

**Interfaces:**
- Final integration test of the complete theme system

- [ ] **Step 1: Run full build**

Run:
```bash
cd frontend && npm run build 2>&1
```

Expected: Build succeeds with no errors.

- [ ] **Step 2: Verify no remaining hardcoded colors**

Run:
```bash
grep -rn "#[0-9a-fA-F]\{3,8\}" frontend/src/ --include="*.vue" --include="*.css" | grep -v "var(--" | grep -v "node_modules" | grep -v ".svg" | grep -v "tokens.css" | grep -v "themes/" | grep -v "element-overrides" | wc -l
```

Expected: A small number (some hex colors are acceptable in specific contexts like echarts configs).

- [ ] **Step 3: Test theme switching**

Start the dev server and verify:
1. Page loads with light theme by default
2. Clicking the moon icon switches to dark theme
3. Refreshing the page persists the theme choice
4. All components render correctly in both themes

Run:
```bash
cd frontend && npm run dev
```

Manual verification required — check sidebar, header, tables, cards, forms in both themes.

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete UI redesign with Rancher-style theme and light/dark support"
```
