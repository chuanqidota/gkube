<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getPvDetail, getPvYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pv = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPvDetail({ clusterName, name })
    pv.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load PV detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getPvYaml({ clusterName, name })
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

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">PersistentVolume: {{ name }}</h2>
      <el-button @click="router.push('/storage/pvs')">Back to List</el-button>
    </div>

    <template v-if="pv">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ pv.name }}</el-descriptions-item>
            <el-descriptions-item label="Capacity">{{ pv.capacity || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Access Modes">{{ pv.accessModes || pv.access_modes || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Storage Class Name">{{ pv.storageClassName || pv.storage_class || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Status">{{ pv.status || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Reclaim Policy">{{ pv.reclaimPolicy || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Volume Mode">{{ pv.volumeMode || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ pv.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Claim Ref -->
          <div v-if="pv.claimRef || pv.claim" style="margin-top: 16px;">
            <h4>Claim Reference</h4>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Namespace">{{ pv.claimRef?.namespace || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Name">{{ pv.claimRef?.name || pv.claim || '-' }}</el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- Labels -->
          <div v-if="pv.labels && Object.keys(pv.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in pv.labels"
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
