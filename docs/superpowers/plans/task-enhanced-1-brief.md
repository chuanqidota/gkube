# Task 1: Add Overview Area to DeploymentDetail.vue

**Files:**
- Modify: `frontend/src/views/workload/DeploymentDetail.vue`

**Interfaces:**
- Consumes: `deployment` ref (already exists)
- Produces: Overview area displaying replicas, strategy, labels, selector

## Step 1: Add overview section to template

Add the following section after the page-header div (around line 281) and before the `template v-if="deployment"` block:

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

## Step 2: Add overview styles

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

## Step 3: Verify TypeScript compiles

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: No TypeScript errors

## Step 4: Commit

```bash
git add frontend/src/views/workload/DeploymentDetail.vue
git commit -m "feat: add overview area to Deployment detail page"
```
