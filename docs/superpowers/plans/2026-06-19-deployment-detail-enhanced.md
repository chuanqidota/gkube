# Deployment Detail Enhanced Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Kuboard-style features to the Deployment detail page: overview area, container details, and labels/annotations display.

**Architecture:** Enhance the existing DeploymentDetail.vue with a compact overview area showing deployment status, and enhance PodListPanel.vue with expandable rows showing container details.

**Tech Stack:** Vue 3, TypeScript, Element Plus

## Global Constraints

- Follow existing code patterns: Vue 3 Composition API with `<script setup lang="ts">`
- Use Element Plus components for UI consistency
- Maintain responsive design for tablet/mobile
- Keep the existing left-right split layout

---

## File Structure

### Files to Modify

| File | Action | Purpose |
|------|--------|---------|
| `frontend/src/views/workload/DeploymentDetail.vue` | Modify | Add overview area with deployment status |
| `frontend/src/components/PodListPanel.vue` | Modify | Add expandable rows for container details |

---

## Task 1: Add Overview Area to DeploymentDetail.vue

**Files:**
- Modify: `frontend/src/views/workload/DeploymentDetail.vue`

**Interfaces:**
- Consumes: `deployment` ref (already exists)
- Produces: Overview area displaying replicas, strategy, labels, selector

- [ ] **Step 1: Add overview section to template**

Add the following section after the page-header div and before the main-content div:

```vue
<!-- Overview Section -->
<div class="overview-section" v-if="deployment">
  <el-descriptions :column="4" border size="small">
    <el-descriptions-item label="Replicas">
      {{ deployment.ready ?? 0 }}/{{ deployment.replicas ?? 0 }}
    </el-descriptions-item>
    <el-descriptions-item label="Available">
      {{ deployment.available ?? '-' }}
    </el-descriptions-item>
    <el-descriptions-item label="Updated">
      {{ deployment.updated ?? '-' }}
    </el-descriptions-item>
    <el-descriptions-item label="Strategy">
      {{ deployment.strategy || '-' }}
    </el-descriptions-item>
  </el-descriptions>
  <div class="overview-tags" v-if="deployment.labels && Object.keys(deployment.labels).length > 0">
    <span class="tag-label">Labels:</span>
    <el-tag v-for="(val, key) in deployment.labels" :key="key" size="small" style="margin-right: 4px;">
      {{ key }}={{ val }}
    </el-tag>
  </div>
  <div class="overview-tags" v-if="deployment.selector && Object.keys(deployment.selector).length > 0">
    <span class="tag-label">Selector:</span>
    <el-tag v-for="(val, key) in deployment.selector" :key="key" size="small" type="info" style="margin-right: 4px;">
      {{ key }}={{ val }}
    </el-tag>
  </div>
</div>
```

- [ ] **Step 2: Add overview styles**

Add the following styles to the `<style scoped>` section:

```css
.overview-section {
  padding: 12px 16px;
  background-color: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  margin-bottom: 16px;
}

.overview-tags {
  margin-top: 8px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-right: 8px;
}
```

- [ ] **Step 3: Verify TypeScript compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: No TypeScript errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/workload/DeploymentDetail.vue
git commit -m "feat: add overview area to Deployment detail page"
```

---

## Task 2: Add Container Details to PodListPanel.vue

**Files:**
- Modify: `frontend/src/components/PodListPanel.vue`

**Interfaces:**
- Consumes: Pod object with `spec.containers` array
- Produces: Expandable rows showing container details

- [ ] **Step 1: Update Pod interface to include containers**

Update the Pod interface to include container information:

```typescript
interface Pod {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
  }
  status: {
    phase: string
    containerStatuses?: Array<{
      name: string
      restartCount: number
      ready: boolean
      image: string
    }>
  }
  spec: {
    nodeName?: string
    containers: Array<{
      name: string
      image: string
      ports?: Array<{
        containerPort: number
        protocol?: string
      }>
      resources?: {
        limits?: Record<string, string>
        requests?: Record<string, string>
      }
    }>
  }
}
```

- [ ] **Step 2: Add expandable row to table**

Update the el-table to support expandable rows:

```vue
<el-table v-else :data="pods" style="width: 100%" row-key="metadata.name">
  <el-table-column type="expand">
    <template #default="{ row }">
      <div class="container-details">
        <h4 style="margin: 0 0 12px 0;">Containers</h4>
        <div v-for="container in row.spec.containers" :key="container.name" class="container-item">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="Name">{{ container.name }}</el-descriptions-item>
            <el-descriptions-item label="Image">{{ container.image }}</el-descriptions-item>
            <el-descriptions-item label="Ports" v-if="container.ports && container.ports.length > 0">
              <el-tag v-for="port in container.ports" :key="port.containerPort" size="small" style="margin-right: 4px;">
                {{ port.containerPort }}{{ port.protocol ? `/${port.protocol}` : '' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="Resources" v-if="container.resources">
              <div v-if="container.resources.limits">
                <span class="resource-label">Limits:</span>
                <span v-for="(val, key) in container.resources.limits" :key="key">
                  {{ key }}={{ val }}
                </span>
              </div>
              <div v-if="container.resources.requests">
                <span class="resource-label">Requests:</span>
                <span v-for="(val, key) in container.resources.requests" :key="key">
                  {{ key }}={{ val }}
                </span>
              </div>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </template>
  </el-table-column>
  <!-- Existing columns... -->
</el-table>
```

- [ ] **Step 3: Add container details styles**

Add the following styles to the `<style scoped>` section:

```css
.container-details {
  padding: 16px;
  background-color: var(--el-fill-color-lighter);
}

.container-item {
  margin-bottom: 12px;
}

.container-item:last-child {
  margin-bottom: 0;
}

.resource-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-right: 4px;
}
```

- [ ] **Step 4: Verify TypeScript compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: No TypeScript errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/PodListPanel.vue
git commit -m "feat: add container details expandable rows to PodListPanel"
```

---

## Task 3: Final Verification

- [ ] **Step 1: Verify frontend compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: No TypeScript errors

- [ ] **Step 2: Test the UI**

1. Navigate to Deployment list page
2. Click on a Deployment to view details
3. Verify overview area shows replicas, strategy, labels, selector
4. Click on a Pod row to expand container details
5. Verify container details show name, image, ports, resources

- [ ] **Step 3: Final commit**

```bash
git add -A
git commit -m "feat: complete Deployment detail enhanced features"
```
