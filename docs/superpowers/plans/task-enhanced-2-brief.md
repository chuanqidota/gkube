# Task 2: Add Container Details to PodListPanel.vue

**Files:**
- Modify: `frontend/src/components/PodListPanel.vue`

**Interfaces:**
- Consumes: Pod object with `spec.containers` array
- Produces: Expandable rows showing container details

## Step 1: Update Pod interface to include containers

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

## Step 2: Add expandable row to table

Update the el-table to support expandable rows. Add the following column after the opening `<el-table>` tag:

```vue
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
```

Also add `row-key="metadata.name"` to the `<el-table>` tag for proper expand functionality.

## Step 3: Add container details styles

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

## Step 4: Verify TypeScript compiles

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: No TypeScript errors

## Step 5: Commit

```bash
git add frontend/src/components/PodListPanel.vue
git commit -m "feat: add container details expandable rows to PodListPanel"
```
