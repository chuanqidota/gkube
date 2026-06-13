<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getDaemonSetDetail, getDaemonSetYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const daemonSet = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getDaemonSetDetail({ clusterName, namespace, name })
    daemonSet.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load daemonset detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getDaemonSetYaml({ clusterName, namespace, name })
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
      <h2 style="margin: 0;">DaemonSet: {{ name }}</h2>
      <el-button @click="router.push('/workloads/daemonsets')">Back to List</el-button>
    </div>

    <template v-if="daemonSet">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ daemonSet.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ daemonSet.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Desired Scheduled">{{ daemonSet.desiredNumberScheduled ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Number Ready">{{ daemonSet.numberReady ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Update Strategy">{{ daemonSet.updateStrategy || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ daemonSet.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="daemonSet.labels && Object.keys(daemonSet.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in daemonSet.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Selector -->
          <div v-if="daemonSet.selector && Object.keys(daemonSet.selector).length > 0" style="margin-top: 16px;">
            <h4>Selector</h4>
            <el-tag
              v-for="(val, key) in daemonSet.selector"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
              type="info"
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
