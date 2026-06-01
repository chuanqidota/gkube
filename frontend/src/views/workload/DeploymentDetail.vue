<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDeploymentDetail,
  getDeploymentYaml,
  restartDeployment,
  rollbackDeployment,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const deployment = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

// Rollback dialog
const rollbackDialogVisible = ref(false)
const rollbackRevision = ref<number>(1)
const rollbackLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getDeploymentDetail({ clusterName, namespace, name })
    deployment.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load deployment detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getDeploymentYaml({ clusterName, namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) {
    fetchYaml()
  }
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(
      `Restart deployment "${name}"? This will trigger a rolling update.`,
      'Confirm Restart',
      { type: 'warning' }
    )
    await restartDeployment({ clusterName, namespace, name })
    ElMessage.success('Deployment restarted')
    fetchDetail()
  } catch {
    // cancelled
  }
}

function handleRollback() {
  // Try to parse current revision from annotations
  const annotations = deployment.value?.annotations || {}
  const currentRevision = parseInt(annotations['deployment.kubernetes.io/revision'] || '0', 10)
  rollbackRevision.value = Math.max(1, currentRevision - 1)
  rollbackDialogVisible.value = true
}

async function handleRollbackConfirm() {
  if (!rollbackRevision.value || rollbackRevision.value < 1) {
    ElMessage.warning('Please enter a valid revision number')
    return
  }
  rollbackLoading.value = true
  try {
    await rollbackDeployment({
      clusterName,
      namespace,
      name,
      revision: rollbackRevision.value,
    })
    ElMessage.success(`Deployment rolled back to revision ${rollbackRevision.value}`)
    rollbackDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to rollback deployment')
  } finally {
    rollbackLoading.value = false
  }
}

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Deployment: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button type="warning" @click="handleRestart">Restart</el-button>
        <el-button type="danger" @click="handleRollback">Rollback</el-button>
        <el-button @click="router.push('/workloads/deployments')">Back to List</el-button>
      </div>
    </div>

    <template v-if="deployment">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ deployment.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ deployment.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Replicas">{{ deployment.replicas ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Ready">{{ deployment.ready ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Updated">{{ deployment.updated ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Available">{{ deployment.available ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Strategy">{{ deployment.strategy || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ deployment.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="deployment.labels && Object.keys(deployment.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in deployment.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Selector -->
          <div v-if="deployment.selector && Object.keys(deployment.selector).length > 0" style="margin-top: 16px;">
            <h4>Selector</h4>
            <el-tag
              v-for="(val, key) in deployment.selector"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
              type="info"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Conditions -->
          <div v-if="deployment.conditions && deployment.conditions.length > 0" style="margin-top: 16px;">
            <h4>Conditions</h4>
            <el-table :data="deployment.conditions" border stripe>
              <el-table-column prop="type" label="Type" width="160" />
              <el-table-column label="Status" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">
                    {{ row.status }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="Reason" width="160" />
              <el-table-column prop="message" label="Message" min-width="250" show-overflow-tooltip />
              <el-table-column prop="lastUpdateTime" label="Last Update" width="180" />
            </el-table>
          </div>
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <div v-loading="yamlLoading">
            <YamlEditor v-model="yamlContent" height="600px" read-only />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>

    <!-- Rollback Dialog -->
    <el-dialog v-model="rollbackDialogVisible" title="Rollback Deployment" width="480px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">
          Rollback deployment <strong>{{ name }}</strong> in namespace <strong>{{ namespace }}</strong>.
        </p>
        <el-alert
          v-if="deployment?.annotations?.['deployment.kubernetes.io/revision']"
          :title="`Current revision: ${deployment.annotations['deployment.kubernetes.io/revision']}`"
          type="info"
          :closable="false"
          style="margin-bottom: 16px;"
        />
        <el-form-item label="Target Revision">
          <el-input-number v-model="rollbackRevision" :min="1" style="width: 200px;" />
        </el-form-item>
        <el-alert
          title="This will roll back the deployment to the specified revision by restoring the Pod template from that revision's ReplicaSet."
          type="warning"
          :closable="false"
          show-icon
        />
      </div>
      <template #footer>
        <el-button @click="rollbackDialogVisible = false">Cancel</el-button>
        <el-button type="danger" :loading="rollbackLoading" @click="handleRollbackConfirm">Rollback</el-button>
      </template>
    </el-dialog>
  </div>
</template>
