# Deployment Detail Page Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the Deployment detail page to use a left-right split layout with ReplicaSet list on the left and Pod list on the right.

**Architecture:** Add a new backend API endpoint to fetch ReplicaSet list by Deployment name. Create two new frontend components (ReplicaSetPanel, PodListPanel) and integrate them into DeploymentDetail.vue, replacing the existing Pods tab.

**Tech Stack:** Go (Gin, client-go), Vue 3 (Composition API, TypeScript), Element Plus

## Global Constraints

- Follow existing code patterns: singleton handler struct, 4-step API handler flow, form+json+label param tags
- Frontend API functions use `request.get/post/put/delete` with auto-injected `clusterName`
- Use `response.Success(c, msg, data)` / `response.Fail(c, msg)` for all responses
- All new routes require JWT auth (under `authorized.Group("k8s")`)

---

## File Structure

### Backend Files

| File | Action | Purpose |
|------|--------|---------|
| `backend/pkg/k8s/deployment/api.go` | Modify | Add `GetDeploymentReplicaSets` function |
| `backend/app/k8s/params/deployment.go` | Modify | Add `DeploymentReplicaSetParams` struct |
| `backend/app/k8s/api/deployment.go` | Modify | Add `GetDeploymentReplicaSets` handler |
| `backend/router/router.go` | Modify | Register new route |

### Frontend Files

| File | Action | Purpose |
|------|--------|---------|
| `frontend/src/api/resource.ts` | Modify | Add `getDeploymentReplicaSets` API function |
| `frontend/src/components/ReplicaSetPanel.vue` | Create | Left panel: ReplicaSet list with selection |
| `frontend/src/components/PodListPanel.vue` | Create | Right panel: Pod list for selected RS |
| `frontend/src/views/workload/DeploymentDetail.vue` | Modify | Replace Pods tab with new split layout |

---

## Task 1: Backend - Add GetDeploymentReplicaSets K8s Function

**Files:**
- Modify: `backend/pkg/k8s/deployment/api.go`

**Interfaces:**
- Produces: `GetDeploymentReplicaSets(client *kubernetes.Clientset, namespace, name string) ([]appsv1.ReplicaSet, error)`

- [ ] **Step 1: Add the function to k8s package**

```go
// GetDeploymentReplicaSets returns all ReplicaSets owned by a Deployment
func GetDeploymentReplicaSets(client *kubernetes.Clientset, namespace, name string) ([]appsv1.ReplicaSet, error) {
	// Get the deployment to find its label selector
	deploy, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s/%s: %v", namespace, name, err)
	}

	// Build label selector from deployment's spec.selector
	selector := labels.SelectorFromSet(deploy.Spec.Selector.MatchLabels)
	listOpts := metav1.ListOptions{
		LabelSelector: selector.String(),
	}

	// List ReplicaSets
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list replicasets for deployment %s/%s: %v", namespace, name, err)
	}

	// Sort by revision annotation (descending)
	sort.Slice(rsList.Items, func(i, j int) bool {
		revI := getRevision(&rsList.Items[i])
		revJ := getRevision(&rsList.Items[j])
		return revI > revJ
	})

	return rsList.Items, nil
}

// getRevision extracts the revision number from ReplicaSet annotations
func getRevision(rs *appsv1.ReplicaSet) int64 {
	if rs.Annotations == nil {
		return 0
	}
	revStr := rs.Annotations["deployment.kubernetes.io/revision"]
	rev, _ := strconv.ParseInt(revStr, 10, 64)
	return rev
}
```

- [ ] **Step 2: Verify the code compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/backend && go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add backend/pkg/k8s/deployment/api.go
git commit -m "feat: add GetDeploymentReplicaSets k8s function"
```

---

## Task 2: Backend - Add Params and API Handler

**Files:**
- Modify: `backend/app/k8s/params/deployment.go`
- Modify: `backend/app/k8s/api/deployment.go`

**Interfaces:**
- Consumes: `GetDeploymentReplicaSets(client, namespace, name)` from Task 1
- Produces: `GET /k8s/deployment/replicasets` endpoint

- [ ] **Step 1: Add param struct**

Add to `backend/app/k8s/params/deployment.go`:

```go
// DeploymentReplicaSetParams 获取 Deployment 关联的 ReplicaSet 列表
type DeploymentReplicaSetParams struct {
	ClusterName string `form:"cluster_name" json:"cluster_name" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"Deployment名称"`
}
```

- [ ] **Step 2: Add API handler**

Add to `backend/app/k8s/api/deployment.go`:

```go
// GetDeploymentReplicaSets 获取 Deployment 关联的 ReplicaSet 列表
func (d *deployment) GetDeploymentReplicaSets(c *gin.Context) {
	var query params.DeploymentReplicaSetParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, "获取集群客户端失败: "+err.Error())
		return
	}

	rsList, err := deploymentPkg.GetDeploymentReplicaSets(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, "获取ReplicaSet列表失败: "+err.Error())
		return
	}

	response.Success(c, "获取ReplicaSet列表成功", rsList)
}
```

- [ ] **Step 3: Register route**

Add to `backend/router/router.go` in the deployment routes section (after line 122):

```go
deployment.GET("replicasets", api.Deployment.GetDeploymentReplicaSets)
```

- [ ] **Step 4: Verify the code compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/backend && go build ./...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add backend/app/k8s/params/deployment.go backend/app/k8s/api/deployment.go backend/router/router.go
git commit -m "feat: add deployment replicasets API endpoint"
```

---

## Task 3: Frontend - Add API Function

**Files:**
- Modify: `frontend/src/api/resource.ts`

**Interfaces:**
- Produces: `getDeploymentReplicaSets(params: { namespace, name })` function

- [ ] **Step 1: Add API function**

Add to `frontend/src/api/resource.ts` after the existing deployment functions (around line 370):

```typescript
// 获取 Deployment 关联的 ReplicaSet 列表
export const getDeploymentReplicaSets = (params: { namespace: string; name: string }) => {
  return request.get('/k8s/deployment/replicasets', { params })
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: No TypeScript errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api/resource.ts
git commit -m "feat: add getDeploymentReplicaSets API function"
```

---

## Task 4: Frontend - Create ReplicaSetPanel Component

**Files:**
- Create: `frontend/src/components/ReplicaSetPanel.vue`

**Interfaces:**
- Props: `replicasets: ReplicaSet[]`, `currentRevision: number`, `loading: boolean`
- Emits: `@select(rs: ReplicaSet)`, `@rollback(rs: ReplicaSet)`

- [ ] **Step 1: Create the component**

Create `frontend/src/components/ReplicaSetPanel.vue`:

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { ElTag } from 'element-plus'

interface ReplicaSet {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
    annotations?: Record<string, string>
    ownerReferences?: Array<{ name: string; kind: string }>
  }
  spec: {
    replicas?: number
    template: {
      spec: {
        containers: Array<{ name: string; image: string }>
      }
    }
  }
  status: {
    readyReplicas?: number
    availableReplicas?: number
  }
}

const props = defineProps<{
  replicasets: ReplicaSet[]
  currentRevision: number
  loading: boolean
  selectedName?: string
}>()

const emit = defineEmits<{
  select: [rs: ReplicaSet]
  rollback: [rs: ReplicaSet]
}>()

const getRevision = (rs: ReplicaSet): number => {
  const revStr = rs.metadata.annotations?.['deployment.kubernetes.io/revision']
  return revStr ? parseInt(revStr, 10) : 0
}

const getImage = (rs: ReplicaSet): string => {
  const image = rs.spec.template.spec.containers[0]?.image || ''
  return image.length > 30 ? image.substring(0, 30) + '...' : image
}

const getReplicas = (rs: ReplicaSet): string => {
  const ready = rs.status.readyReplicas || 0
  const desired = rs.spec.replicas || 0
  return `${ready}/${desired}`
}

const getStatus = (rs: ReplicaSet): { text: string; type: 'success' | 'primary' | 'info' } => {
  const revision = getRevision(rs)
  if (revision === props.currentRevision) {
    return { text: 'Current', type: 'success' }
  }
  const ready = rs.status.readyReplicas || 0
  if (ready > 0) {
    return { text: 'Active', type: 'primary' }
  }
  return { text: 'Inactive', type: 'info' }
}

const formatAge = (timestamp: string): string => {
  const now = new Date()
  const created = new Date(timestamp)
  const diffMs = now.getTime() - created.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffDays > 0) return `${diffDays}d ago`
  if (diffHours > 0) return `${diffHours}h ago`
  return `${diffMins}m ago`
}

const handleSelect = (rs: ReplicaSet) => {
  emit('select', rs)
}

const handleRollback = (rs: ReplicaSet) => {
  emit('rollback', rs)
}
</script>

<template>
  <div class="replicaset-panel" v-loading="loading">
    <div v-if="replicasets.length === 0" class="empty-state">
      No ReplicaSets found
    </div>
    <div
      v-for="rs in replicasets"
      :key="rs.metadata.name"
      class="rs-item"
      :class="{ selected: rs.metadata.name === selectedName }"
      @click="handleSelect(rs)"
    >
      <div class="rs-header">
        <span class="rs-name">{{ rs.metadata.name }}</span>
        <el-tag :type="getStatus(rs).type" size="small">
          {{ getStatus(rs).text }}
        </el-tag>
      </div>
      <div class="rs-details">
        <div class="rs-detail">
          <span class="label">Image:</span>
          <span class="value">{{ getImage(rs) }}</span>
        </div>
        <div class="rs-detail">
          <span class="label">Created:</span>
          <span class="value">{{ formatAge(rs.metadata.creationTimestamp) }}</span>
        </div>
        <div class="rs-detail">
          <span class="label">Replicas:</span>
          <span class="value">{{ getReplicas(rs) }}</span>
        </div>
      </div>
      <div class="rs-actions" v-if="getRevision(rs) !== currentRevision">
        <el-button size="small" @click.stop="handleRollback(rs)">Rollback</el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.replicaset-panel {
  height: 100%;
  overflow-y: auto;
  border-right: 1px solid var(--el-border-color-lighter);
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: var(--el-text-color-secondary);
}

.rs-item {
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--el-border-color-lighter);
  transition: background-color 0.2s;
}

.rs-item:hover {
  background-color: var(--el-fill-color-light);
}

.rs-item.selected {
  background-color: var(--el-color-primary-light-9);
  border-left: 3px solid var(--el-color-primary);
}

.rs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.rs-name {
  font-weight: 500;
  font-size: 14px;
}

.rs-details {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.rs-detail {
  margin-bottom: 4px;
}

.rs-detail .label {
  margin-right: 4px;
}

.rs-actions {
  margin-top: 8px;
}
</style>
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: No TypeScript errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ReplicaSetPanel.vue
git commit -m "feat: add ReplicaSetPanel component"
```

---

## Task 5: Frontend - Create PodListPanel Component

**Files:**
- Create: `frontend/src/components/PodListPanel.vue`

**Interfaces:**
- Props: `pods: Pod[]`, `loading: boolean`, `replicasetName: string`
- Emits: `@logs(pod: Pod)`, `@exec(pod: Pod)`, `@delete(pod: Pod)`

- [ ] **Step 1: Create the component**

Create `frontend/src/components/PodListPanel.vue`:

```vue
<script setup lang="ts">
import { ElTag, ElButton } from 'element-plus'

interface Pod {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
  }
  status: {
    phase: string
    containerStatuses?: Array<{
      restartCount: number
    }>
  }
  spec: {
    nodeName?: string
  }
}

const props = defineProps<{
  pods: Pod[]
  loading: boolean
  replicasetName: string
}>()

const emit = defineEmits<{
  logs: [pod: Pod]
  exec: [pod: Pod]
  delete: [pod: Pod]
}>()

const getStatusType = (phase: string): 'success' | 'warning' | 'danger' | 'info' => {
  switch (phase) {
    case 'Running':
      return 'success'
    case 'Pending':
      return 'warning'
    case 'Failed':
      return 'danger'
    case 'Succeeded':
      return 'info'
    default:
      return 'info'
  }
}

const getRestarts = (pod: Pod): number => {
  return pod.status.containerStatuses?.reduce((sum, cs) => sum + cs.restartCount, 0) || 0
}

const formatAge = (timestamp: string): string => {
  const now = new Date()
  const created = new Date(timestamp)
  const diffMs = now.getTime() - created.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffDays > 0) return `${diffDays}d`
  if (diffHours > 0) return `${diffHours}h`
  return `${diffMins}m`
}

const handleLogs = (pod: Pod) => {
  emit('logs', pod)
}

const handleExec = (pod: Pod) => {
  emit('exec', pod)
}

const handleDelete = (pod: Pod) => {
  emit('delete', pod)
}
</script>

<template>
  <div class="pod-list-panel" v-loading="loading">
    <div class="panel-header">
      <span class="title">Pods ({{ pods.length }})</span>
    </div>
    <div v-if="pods.length === 0" class="empty-state">
      No pods found
    </div>
    <el-table v-else :data="pods" style="width: 100%">
      <el-table-column label="Name" min-width="200">
        <template #default="{ row }">
          <router-link
            :to="{ name: 'PodDetail', params: { namespace: row.metadata.namespace, name: row.metadata.name } }"
            class="pod-link"
          >
            {{ row.metadata.name }}
          </router-link>
        </template>
      </el-table-column>
      <el-table-column label="Status" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status.phase)" size="small">
            {{ row.status.phase }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Restarts" width="80">
        <template #default="{ row }">
          <span :class="{ warning: getRestarts(row) > 0 }">
            {{ getRestarts(row) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="Age" width="80">
        <template #default="{ row }">
          {{ formatAge(row.metadata.creationTimestamp) }}
        </template>
      </el-table-column>
      <el-table-column label="Node" width="120">
        <template #default="{ row }">
          {{ row.spec.nodeName || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="Actions" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleLogs(row)">Logs</el-button>
          <el-button size="small" @click="handleExec(row)">Exec</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.pod-list-panel {
  height: 100%;
  overflow-y: auto;
}

.panel-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.title {
  font-weight: 500;
  font-size: 14px;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: var(--el-text-color-secondary);
}

.pod-link {
  color: var(--el-color-primary);
  text-decoration: none;
}

.pod-link:hover {
  text-decoration: underline;
}

.warning {
  color: var(--el-color-warning);
  font-weight: 500;
}
</style>
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: No TypeScript errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/PodListPanel.vue
git commit -m "feat: add PodListPanel component"
```

---

## Task 6: Frontend - Integrate Components into DeploymentDetail

**Files:**
- Modify: `frontend/src/views/workload/DeploymentDetail.vue`

**Interfaces:**
- Consumes: `getDeploymentReplicaSets` from Task 3
- Consumes: `ReplicaSetPanel` from Task 4
- Consumes: `PodListPanel` from Task 5

- [ ] **Step 1: Add imports and state**

Add to the script section:

```typescript
import ReplicaSetPanel from '@/components/ReplicaSetPanel.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import { getDeploymentReplicaSets } from '@/api/resource'

// Add new state
const replicasets = ref<any[]>([])
const replicasetsLoading = ref(false)
const selectedReplicaset = ref<any>(null)
const rsPods = ref<any[]>([])
const rsPodsLoading = ref(false)
```

- [ ] **Step 2: Add fetch functions**

```typescript
const fetchReplicaSets = async () => {
  replicasetsLoading.value = true
  try {
    const res = await getDeploymentReplicaSets({ namespace, name })
    replicasets.value = res.data?.items || res.data || []
    // Auto-select the current RS
    if (replicasets.value.length > 0) {
      const currentRevision = deployment.value?.metadata?.annotations?.['deployment.kubernetes.io/revision']
      const currentRS = replicasets.value.find(rs => 
        rs.metadata.annotations?.['deployment.kubernetes.io/revision'] === currentRevision
      )
      if (currentRS) {
        handleReplicasetSelect(currentRS)
      }
    }
  } catch (error) {
    console.error('Failed to fetch replicasets:', error)
  } finally {
    replicasetsLoading.value = false
  }
}

const fetchReplicasetPods = async (rsName: string) => {
  rsPodsLoading.value = true
  try {
    // Reuse existing getPodList with label selector
    const res = await getPodList({ 
      namespace, 
      labelSelector: `pod-template-hash=${rsName.split('-').pop()}`
    })
    rsPods.value = res.data?.items || res.data || []
  } catch (error) {
    console.error('Failed to fetch pods:', error)
  } finally {
    rsPodsLoading.value = false
  }
}
```

- [ ] **Step 3: Add event handlers**

```typescript
const handleReplicasetSelect = (rs: any) => {
  selectedReplicaset.value = rs
  fetchReplicasetPods(rs.metadata.name)
}

const handleReplicasetRollback = async (rs: any) => {
  const revision = rs.metadata.annotations?.['deployment.kubernetes.io/revision']
  if (!revision) return
  
  try {
    await ElMessageBox.confirm(
      `Are you sure you want to rollback to revision ${revision}?`,
      'Confirm Rollback',
      { type: 'warning' }
    )
    await rollbackDeployment({ namespace, name, revision: parseInt(revision, 10) })
    ElMessage.success('Rollback successful')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('Rollback failed')
    }
  }
}

const handlePodLogs = (pod: any) => {
  router.push({ name: 'LogViewer', params: { namespace: pod.metadata.namespace, podName: pod.metadata.name } })
}

const handlePodExec = (pod: any) => {
  router.push({ name: 'Terminal', params: { namespace: pod.metadata.namespace, podName: pod.metadata.name } })
}

const handlePodDelete = async (pod: any) => {
  try {
    await ElMessageBox.confirm(
      `Are you sure you want to delete pod ${pod.metadata.name}?`,
      'Confirm Delete',
      { type: 'warning' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod deleted')
    fetchReplicasetPods(selectedReplicaset.value?.metadata.name)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('Delete failed')
    }
  }
}
```

- [ ] **Step 4: Update handleTabChange**

```typescript
const handleTabChange = (tab: string) => {
  if (tab === 'info' && !deployment.value) {
    fetchDeploymentDetail()
  } else if (tab === 'replicasets' && replicasets.value.length === 0) {
    fetchReplicaSets()
  } else if (tab === 'events' && events.value.length === 0) {
    fetchEvents()
  } else if (tab === 'yaml' && !yamlContent.value) {
    fetchYaml()
  }
}
```

- [ ] **Step 5: Update template**

Replace the existing `pods` tab with:

```vue
<el-tab-pane label="Replicasets & Pods" name="replicasets">
  <div class="rs-pods-container">
    <div class="rs-panel">
      <ReplicaSetPanel
        :replicasets="replicasets"
        :current-revision="parseInt(deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision'] || '0')"
        :loading="replicasetsLoading"
        :selected-name="selectedReplicaset?.metadata?.name"
        @select="handleReplicasetSelect"
        @rollback="handleReplicasetRollback"
      />
    </div>
    <div class="pods-panel">
      <PodListPanel
        :pods="rsPods"
        :loading="rsPodsLoading"
        :replicaset-name="selectedReplicaset?.metadata?.name || ''"
        @logs="handlePodLogs"
        @exec="handlePodExec"
        @delete="handlePodDelete"
      />
    </div>
  </div>
</el-tab-pane>
```

- [ ] **Step 6: Add styles**

```vue
<style scoped>
.rs-pods-container {
  display: flex;
  height: 600px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
}

.rs-panel {
  width: 320px;
  min-width: 320px;
  border-right: 1px solid var(--el-border-color-lighter);
}

.pods-panel {
  flex: 1;
  overflow: hidden;
}

@media (max-width: 768px) {
  .rs-pods-container {
    flex-direction: column;
    height: auto;
  }
  
  .rs-panel {
    width: 100%;
    min-width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--el-border-color-lighter);
    max-height: 300px;
  }
}
</style>
```

- [ ] **Step 7: Verify TypeScript compiles**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
Expected: No TypeScript errors

- [ ] **Step 8: Commit**

```bash
git add frontend/src/views/workload/DeploymentDetail.vue
git commit -m "feat: integrate ReplicaSet and Pod panels into DeploymentDetail"
```

---

## Task 7: Testing and Verification

- [ ] **Step 1: Start backend server**

Run: `cd /Users/zqqzqq/05_github/gkube/backend && go run main.go`
Expected: Server starts on port 8080

- [ ] **Step 2: Start frontend dev server**

Run: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run dev`
Expected: Vite dev server starts

- [ ] **Step 3: Test the API endpoint**

Run: `curl -H "Authorization: Bearer <token>" "http://localhost:8080/v1/k8s/deployment/replicasets?cluster_name=<cluster>&namespace=<ns>&name=<name>"`
Expected: JSON response with ReplicaSet list

- [ ] **Step 4: Test the UI**

1. Navigate to Deployment list page
2. Click on a Deployment to view details
3. Click "Replicasets & Pods" tab
4. Verify left panel shows ReplicaSets
5. Click on a ReplicaSet
6. Verify right panel shows Pods for that RS
7. Click Rollback on a historical RS
8. Verify rollback confirmation dialog appears

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat: complete Deployment detail page redesign with RS/Pod split view"
```
