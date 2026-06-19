# Task 2 Review Package

## Commits
- 03c505e feat: add container details expandable rows to PodListPanel

## Diff Stats
frontend/src/components/PodListPanel.vue | 68 +++++++++++++++++++++++++++++++-
1 file changed, 67 insertions(+), 1 deletion(-)

## Full Diff
```diff
diff --git a/frontend/src/components/PodListPanel.vue b/frontend/src/components/PodListPanel.vue
index 6ae199b..4427b9f 100644
--- a/frontend/src/components/PodListPanel.vue
+++ b/frontend/src/components/PodListPanel.vue
@@ -4,25 +4,40 @@ import { formatAge } from '@/utils/time'

 interface Pod {
   metadata: {
     name: string
     namespace: string
     creationTimestamp: string
   }
   status: {
     phase: string
     containerStatuses?: Array<{
+      name: string
       restartCount: number
+      ready: boolean
+      image: string
     }>
   }
   spec: {
     nodeName?: string
+    containers: Array<{
+      name: string
+      image: string
+      ports?: Array<{
+        containerPort: number
+        protocol?: string
+      }>
+      resources?: {
+        limits?: Record<string, string>
+        requests?: Record<string, string>
+      }
+    }>
   }
 }

 const props = defineProps<{
   pods: Pod[]
   loading: boolean
   replicasetName: string
 }>()

 const emit = defineEmits<{
@@ -64,21 +79,53 @@ const handleDelete = (pod: Pod) => {
 </script>

 <template>
   <div class="pod-list-panel" v-loading="loading">
     <div class="panel-header">
       <span class="title">{{ replicasetName ? `Pods for ${replicasetName}` : 'Pods' }} ({{ pods.length }})</span>
     </div>
     <div v-if="pods.length === 0" class="empty-state">
       No pods found
     </div>
-    <el-table v-else :data="pods" style="width: 100%">
+    <el-table v-else :data="pods" style="width: 100%" row-key="metadata.name">
+      <el-table-column type="expand">
+        <template #default="{ row }">
+          <div class="container-details">
+            <h4 style="margin: 0 0 12px 0;">Containers</h4>
+            <div v-for="container in row.spec.containers" :key="container.name" class="container-item">
+              <el-descriptions :column="2" border size="small">
+                <el-descriptions-item label="Name">{{ container.name }}</el-descriptions-item>
+                <el-descriptions-item label="Image">{{ container.image }}</el-descriptions-item>
+                <el-descriptions-item label="Ports" v-if="container.ports && container.ports.length > 0">
+                  <el-tag v-for="port in container.ports" :key="port.containerPort" size="small" style="margin-right: 4px;">
+                    {{ port.containerPort }}{{ port.protocol ? `/${port.protocol}` : '' }}
+                  </el-tag>
+                </el-descriptions-item>
+                <el-descriptions-item label="Resources" v-if="container.resources">
+                  <div v-if="container.resources.limits">
+                    <span class="resource-label">Limits:</span>
+                    <span v-for="(val, key) in container.resources.limits" :key="key">
+                      {{ key }}={{ val }}
+                    </span>
+                  </div>
+                  <div v-if="container.resources.requests">
+                    <span class="resource-label">Requests:</span>
+                    <span v-for="(val, key) in container.resources.requests" :key="key">
+                      {{ key }}={{ val }}
+                    </span>
+                  </div>
+                </el-descriptions-item>
+              </el-descriptions>
+            </div>
+          </div>
+        </template>
+      </el-table-column>
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
@@ -145,11 +192,30 @@ const handleDelete = (pod: Pod) => {
 }

 .pod-link:hover {
   text-decoration: underline;
 }

 .warning {
   color: var(--el-color-warning);
   font-weight: 500;
 }
+
+.container-details {
+  padding: 16px;
+  background-color: var(--el-fill-color-lighter);
+}
+
+.container-item {
+  margin-bottom: 12px;
+}
+
+.container-item:last-child {
+  margin-bottom: 0;
+}
+
+.resource-label {
+  font-size: 12px;
+  color: var(--el-text-color-secondary);
+  margin-right: 4px;
+}
 </style>
```
