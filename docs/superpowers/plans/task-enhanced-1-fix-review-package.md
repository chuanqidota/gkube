# Task 1 Fix Review Package

## Commits
- 0c01361 fix: correct field paths in overview section and related code

## Diff Stats
frontend/src/views/workload/DeploymentDetail.vue | 33 ++++++++++++------------
1 file changed, 16 insertions(+), 17 deletions(-)

## Full Diff
```diff
diff --git a/frontend/src/views/workload/DeploymentDetail.vue b/frontend/src/views/workload/DeploymentDetail.vue
index 31735fd..08034e3 100644
--- a/frontend/src/views/workload/DeploymentDetail.vue
+++ b/frontend/src/views/workload/DeploymentDetail.vue
@@ -88,42 +88,41 @@ async function fetchEvents() {
 }

 async function fetchReplicaSets() {
   replicasetsLoading.value = true
   try {
     const res: any = await getDeploymentReplicaSets({ namespace, name })
     replicasets.value = res.data?.items || res.data || []
     // Auto-select the current revision's ReplicaSet
     if (replicasets.value.length > 0) {
       const currentRevision = deployment.value?.metadata?.annotations?.['deployment.kubernetes.io/revision']
-        || deployment.value?.annotations?.['deployment.kubernetes.io/revision']
       const currentRS = replicasets.value.find(
         (rs: any) => rs.metadata.annotations?.['deployment.kubernetes.io/revision'] === currentRevision
       )
       if (currentRS) {
         handleReplicasetSelect(currentRS)
       }
     }
   } catch (e) {
     console.error('Failed to fetch replicasets:', e)
     ElMessage.error('Failed to load ReplicaSets')
   } finally {
     replicasetsLoading.value = false
   }
 }

 async function fetchReplicasetPods(rsName: string) {
   rsPodsLoading.value = true
   try {
     // The pod-template-hash is the last segment of the ReplicaSet name
     const hash = rsName.split('-').pop()
-    const selector = deployment.value?.selector || {}
+    const selector = deployment.value?.spec?.selector?.matchLabels || {}
     const selectorEntries = Object.entries(selector).map(([k, v]) => `${k}=${v}`).join(',')
     const labelSelector = selectorEntries ? `${selectorEntries},pod-template-hash=${hash}` : `pod-template-hash=${hash}`
     const res: any = await getPodList({ namespace, labelSelector })
     rsPods.value = res.data?.items || res.data || []
   } catch (e) {
     console.error('Failed to fetch pods:', e)
     ElMessage.error('Failed to load pods')
   } finally {
     rsPodsLoading.value = false
   }
@@ -209,28 +208,28 @@ async function handleRestart() {
   try {
     await ElMessageBox.confirm(`Restart deployment "${name}"? This will trigger a rolling update.`, 'Confirm Restart', { type: 'warning' })
     await restartDeployment({ namespace, name })
     ElMessage.success('Deployment restarted')
     fetchDetail()
     fetchReplicaSets()
   } catch { /* cancelled */ }
 }

 function handleRollback() {
-  const annotations = deployment.value?.annotations || {}
+  const annotations = deployment.value?.metadata?.annotations || {}
   const currentRevision = parseInt(annotations['deployment.kubernetes.io/revision'] || '0', 10)
   rollbackRevision.value = Math.max(1, currentRevision - 1)
   rollbackDialogVisible.value = true
 }

 function handleScale() {
-  scaleReplicas.value = deployment.value?.replicas ?? 1
+  scaleReplicas.value = deployment.value?.spec?.replicas ?? 1
   scaleDialogVisible.value = true
 }

 async function handleScaleConfirm() {
   scaleLoading.value = true
   try {
     await scaleDeployment({ namespace, name, replicas: scaleReplicas.value })
     ElMessage.success(`Deployment scaled to ${scaleReplicas.value} replicas`)
     scaleDialogVisible.value = false
     fetchDetail()
@@ -275,55 +274,55 @@ onMounted(() => {
         <el-button type="primary" @click="handleScale">Scale</el-button>
         <el-button type="warning" @click="handleRestart">Restart</el-button>
         <el-button type="danger" @click="handleRollback">Rollback</el-button>
         <el-button @click="handleOpenYaml">YAML</el-button>
         <el-button @click="router.push('/workloads/deployments')">Back to List</el-button>
       </div>
     </div>

     <!-- Overview Section -->
     <div class="overview-section" v-if="deployment">
-      <el-descriptions :column="4" border size="small">
+      <el-descriptions :column="{ xs: 1, sm: 2, md: 3, lg: 4 }" border size="small">
         <el-descriptions-item label="Replicas">
-          {{ deployment.ready ?? 0 }}/{{ deployment.replicas ?? 0 }}
+          {{ deployment.status?.readyReplicas ?? 0 }}/{{ deployment.spec?.replicas ?? 0 }}
         </el-descriptions-item>
         <el-descriptions-item label="Available">
-          {{ deployment.available ?? '-' }}
+          {{ deployment.status?.availableReplicas ?? '-' }}
         </el-descriptions-item>
         <el-descriptions-item label="Updated">
-          {{ deployment.updated ?? '-' }}
+          {{ deployment.status?.updatedReplicas ?? '-' }}
         </el-descriptions-item>
         <el-descriptions-item label="Strategy">
-          {{ deployment.strategy || '-' }}
+          {{ deployment.spec?.strategy?.type || '-' }}
         </el-descriptions-item>
       </el-descriptions>
-      <div class="overview-tags" v-if="deployment.labels && Object.keys(deployment.labels).length > 0">
+      <div class="overview-tags" v-if="deployment.metadata?.labels && Object.keys(deployment.metadata.labels).length > 0">
         <span class="tag-label">Labels:</span>
-        <el-tag v-for="(val, key) in deployment.labels" :key="key" size="small" style="margin-right: 4px;">
+        <el-tag v-for="(val, key) in deployment.metadata.labels" :key="key" size="small">
           {{ key }}={{ val }}
         </el-tag>
       </div>
-      <div class="overview-tags" v-if="deployment.selector && Object.keys(deployment.selector).length > 0">
+      <div class="overview-tags" v-if="deployment.spec?.selector?.matchLabels && Object.keys(deployment.spec.selector.matchLabels).length > 0">
         <span class="tag-label">Selector:</span>
-        <el-tag v-for="(val, key) in deployment.selector" :key="key" size="small" type="info" style="margin-right: 4px;">
+        <el-tag v-for="(val, key) in deployment.spec.selector.matchLabels" :key="key" size="small" type="info">
           {{ key }}={{ val }}
         </el-tag>
       </div>
     </div>

     <template v-if="deployment">
       <div class="main-content">
         <!-- Left Panel: ReplicaSet List -->
         <div class="left-panel">
           <ReplicaSetPanel
             :replicasets="replicasets"
-            :current-revision="parseInt(deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision'] || deployment?.annotations?.['deployment.kubernetes.io/revision'] || '0')"
+            :current-revision="parseInt(deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision'] || '0')"
             :loading="replicasetsLoading"
             :selected-name="selectedReplicaset?.metadata?.name"
             @select="handleReplicasetSelect"
             @rollback="handleReplicasetRollback"
           />
         </div>

         <!-- Right Panel: Events + Pods -->
         <div class="right-panel">
           <!-- Events Section -->
@@ -370,39 +369,39 @@ onMounted(() => {
       </div>
       <div v-loading="yamlLoading">
         <YamlEditor v-model="yamlContent" height="600px" :read-only="!yamlEditing" />
       </div>
     </el-dialog>

     <!-- Rollback Dialog -->
     <el-dialog v-model="rollbackDialogVisible" title="Rollback Deployment" width="480px" destroy-on-close>
       <div>
         <p style="margin-bottom: 16px;">Rollback deployment <strong>{{ name }}</strong> in namespace <strong>{{ namespace }}</strong>.</p>
-        <el-alert v-if="deployment?.annotations?.['deployment.kubernetes.io/revision']" :title="`Current revision: ${deployment.annotations['deployment.kubernetes.io/revision']}`" type="info" :closable="false" style="margin-bottom: 16px;" />
+        <el-alert v-if="deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision']" :title="`Current revision: ${deployment.metadata.annotations['deployment.kubernetes.io/revision']}`" type="info" :closable="false" style="margin-bottom: 16px;" />
         <el-form-item label="Target Revision">
           <el-input-number v-model="rollbackRevision" :min="1" style="width: 200px;" />
         </el-form-item>
         <el-alert title="This will roll back the deployment to the specified revision by restoring the Pod template from that revision's ReplicaSet." type="warning" :closable="false" show-icon />
       </div>
       <template #footer>
         <el-button @click="rollbackDialogVisible = false">Cancel</el-button>
         <el-button type="danger" :loading="rollbackLoading" @click="handleRollbackConfirm">Rollback</el-button>
       </template>
     </el-dialog>

     <!-- Scale Dialog -->
     <el-dialog v-model="scaleDialogVisible" title="Scale Deployment" width="480px" destroy-on-close>
       <div>
         <p style="margin-bottom: 16px;">Scale deployment <strong>{{ name }}</strong> in namespace <strong>{{ namespace }}</strong>.</p>
         <el-descriptions :column="1" border style="margin-bottom: 16px;">
-          <el-descriptions-item label="Current Replicas">{{ deployment?.replicas ?? '-' }}</el-descriptions-item>
-          <el-descriptions-item label="Ready Replicas">{{ deployment?.ready ?? '-' }}</el-descriptions-item>
+          <el-descriptions-item label="Current Replicas">{{ deployment?.spec?.replicas ?? '-' }}</el-descriptions-item>
+          <el-descriptions-item label="Ready Replicas">{{ deployment?.status?.readyReplicas ?? '-' }}</el-descriptions-item>
         </el-descriptions>
         <el-form-item label="Target Replicas">
           <el-input-number v-model="scaleReplicas" :min="0" :max="100" style="width: 200px;" />
         </el-form-item>
         <el-alert v-if="scaleReplicas === 0" title="Setting replicas to 0 will stop all pods." type="warning" :closable="false" show-icon style="margin-top: 8px;" />
       </div>
       <template #footer>
         <el-button @click="scaleDialogVisible = false">Cancel</el-button>
         <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">Scale</el-button>
       </template>
```
