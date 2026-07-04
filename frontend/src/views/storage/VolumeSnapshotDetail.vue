<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getVolumeSnapshotDetail, getVolumeSnapshotYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const snapshot = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getVolumeSnapshotDetail({ namespace, name })
    snapshot.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load VolumeSnapshot detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getVolumeSnapshotYaml({ namespace, name })
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

function getStatus(): string {
  if (!snapshot.value) return '-'
  if (snapshot.value.status?.readyToUse) return 'Ready'
  if (snapshot.value.status?.error?.message) return 'Error'
  if (snapshot.value.status?.boundVolumeSnapshotContentName) return 'Bound'
  return 'Pending'
}

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'ready' || s === 'bound') return 'success'
  if (s === 'pending') return 'warning'
  if (s === 'error') return 'danger'
  return 'info'
}

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">VolumeSnapshot: {{ name }}</h2>
      <el-button @click="router.push('/storage/volumesnapshots')">Back to List</el-button>
    </div>

    <template v-if="snapshot">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ snapshot.metadata?.name || name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ snapshot.metadata?.namespace || namespace }}</el-descriptions-item>
            <el-descriptions-item label="Status">
              <el-tag :type="statusType(getStatus())" size="small">{{ getStatus() }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="Snapshot Class">{{ snapshot.spec?.volumeSnapshotClassName || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Source PVC">{{ snapshot.spec?.source?.persistentVolumeClaimName || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Restore Size">{{ snapshot.status?.restoreSize || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Creation Time">{{ snapshot.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Bound VSC">{{ snapshot.status?.boundVolumeSnapshotContentName || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Error -->
          <div v-if="snapshot.status?.error?.message" style="margin-top: 16px;">
            <h4>Error</h4>
            <el-alert :title="snapshot.status.error.message" type="error" :closable="false" show-icon />
          </div>

          <!-- Labels -->
          <div v-if="snapshot.metadata?.labels && Object.keys(snapshot.metadata.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in snapshot.metadata.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Annotations -->
          <div v-if="snapshot.metadata?.annotations && Object.keys(snapshot.metadata.annotations).length > 0" style="margin-top: 16px;">
            <h4>Annotations</h4>
            <el-tag
              v-for="(val, key) in snapshot.metadata.annotations"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <div v-loading="yamlLoading">
            <YamlEditor v-model="yamlContent" height="600px" read-only />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>
