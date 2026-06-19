# Deployment List Button Simplification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove Scale and Restart row actions from the Deployment list page, reducing the Actions column from 320px to 200px to match other list pages.

**Architecture:** Single-file change to `DeploymentList.vue`. The detail page already has Scale, Restart, and Rollback buttons, so no additions needed elsewhere. This is purely a removal of code from the list page.

**Tech Stack:** Vue 3, Element Plus, TypeScript

## Global Constraints

- Actions column width: 200px (matches StatefulSet, DaemonSet, Service list pages)
- Detail page already has Scale, Restart, Rollback — no changes needed there
- No test files exist in this project — skip TDD steps

---

### Task 1: Remove Scale and Restart from DeploymentList.vue

**Files:**
- Modify: `frontend/src/views/workload/DeploymentList.vue`

**Interfaces:**
- No interface changes — this is a UI-only removal
- The `scaleDeployment` and `restartDeployment` API functions remain available for the detail page

- [ ] **Step 1: Remove Scale and Restart buttons from the Actions column**

In the template, change the Actions column (line 239) from 320px to 200px, and remove the Scale and Restart buttons:

```vue
<!-- Before (lines 239-245) -->
<el-table-column label="Actions" width="320" fixed="right">
  <template #default="{ row }">
    <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
    <el-button size="small" type="warning" @click="handleScale(row)">Scale</el-button>
    <el-button size="small" type="success" @click="handleRestart(row)">Restart</el-button>
    <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
  </template>
</el-table-column>

<!-- After -->
<el-table-column label="Actions" width="200" fixed="right">
  <template #default="{ row }">
    <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
    <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
  </template>
</el-table-column>
```

- [ ] **Step 2: Remove the Scale dialog template**

Delete the Scale dialog (lines 258-269):

```vue
<!-- DELETE THIS ENTIRE BLOCK -->
<!-- Scale Dialog -->
<el-dialog v-model="scaleDialogVisible" title="Scale Deployment" width="400px" destroy-on-close>
  <div v-if="scaleTarget">
    <p style="margin-bottom: 12px;">Deployment: <strong>{{ scaleTarget.name }}</strong></p>
    <el-form-item label="Replicas">
      <el-input-number v-model="scaleReplicas" :min="0" :max="100" />
    </el-form-item>
  </div>
  <template #footer>
    <el-button @click="scaleDialogVisible = false">Cancel</el-button>
    <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">Scale</el-button>
  </template>
</el-dialog>
```

- [ ] **Step 3: Remove Scale dialog state variables**

Delete these refs (lines 34-38):

```typescript
// Scale dialog
const scaleDialogVisible = ref(false)
const scaleTarget = ref<any>(null)
const scaleReplicas = ref(1)
const scaleLoading = ref(false)
```

- [ ] **Step 4: Remove Scale and Restart handler functions**

Delete `handleScale` (lines 96-102), `handleScaleConfirm` (lines 104-121), and `handleRestart` (lines 123-132):

```typescript
// DELETE: handleScale function
function handleScale(row: any) {
  scaleTarget.value = row
  const readyStr = row.ready || '0'
  const parts = readyStr.split('/')
  scaleReplicas.value = parseInt(parts[1] || parts[0]) || 1
  scaleDialogVisible.value = true
}

// DELETE: handleScaleConfirm function
async function handleScaleConfirm() {
  if (!scaleTarget.value) return
  scaleLoading.value = true
  try {
    await scaleDeployment({
      namespace: scaleTarget.value.namespace,
      name: scaleTarget.value.name,
      replicas: scaleReplicas.value,
    })
    ElMessage.success(`Scaled to ${scaleReplicas.value} replicas`)
    scaleDialogVisible.value = false
    fetchDeployments()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to scale')
  } finally {
    scaleLoading.value = false
  }
}

// DELETE: handleRestart function
async function handleRestart(row: any) {
  try {
    await ElMessageBox.confirm(`Restart deployment "${row.name}"?`, 'Confirm', { type: 'warning' })
    await restartDeployment({ namespace: row.namespace, name: row.name })
    ElMessage.success('Deployment restarted')
    fetchDeployments()
  } catch {
    // cancelled
  }
}
```

- [ ] **Step 5: Remove unused imports**

Remove `scaleDeployment` and `restartDeployment` from the import statement (line 11-12):

```typescript
// Before (lines 8-17)
import {
  getDeploymentList,
  getDeploymentYaml,
  scaleDeployment,       // DELETE
  restartDeployment,     // DELETE
  deleteDeployment,
  getNamespaceList,
  extractNamespaceNames,
  transformDeployments,
} from '@/api/resource'

// After
import {
  getDeploymentList,
  getDeploymentYaml,
  deleteDeployment,
  getNamespaceList,
  extractNamespaceNames,
  transformDeployments,
} from '@/api/resource'
```

- [ ] **Step 6: Verify the result**

Run the dev server and confirm:
- Deployment list loads without errors
- Actions column shows only YAML and Delete buttons
- Actions column is 200px wide
- YAML dialog still works
- Delete still works
- Detail page still has Scale, Restart, Rollback buttons

```bash
cd frontend && npm run dev
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/views/workload/DeploymentList.vue
git commit -m "feat: simplify Deployment list actions to YAML and Delete only

Remove Scale and Restart row actions from the Deployment list page.
These operations remain available in the Deployment detail page.
Actions column width reduced from 320px to 200px for consistency
with other list pages (StatefulSet, DaemonSet, Service)."
```
