<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'
import { createPdb, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const namespaceList = ref<string[]>([])

const form = ref({
  name: '',
  namespace: 'default',
  minAvailable: '1',
  maxUnavailable: '',
  selectorLabels: [{ key: 'app', value: '' }],
  useMinAvailable: true,
  labels: [{ key: '', value: '' }] as Array<{ key: string; value: string }>,
  unhealthyPodEvictionPolicy: '',
})

const yamlContent = ref('')

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

function buildYaml() {
  const matchLabels: Record<string, string> = {}
  form.value.selectorLabels.forEach(l => { if (l.key) matchLabels[l.key] = l.value })

  const labels: Record<string, string> = {}
  form.value.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const metadata: Record<string, any> = { name: form.value.name, namespace: form.value.namespace }
  if (Object.keys(labels).length > 0) metadata.labels = labels

  const pdb: any = {
    apiVersion: 'policy/v1',
    kind: 'PodDisruptionBudget',
    metadata,
    spec: {
      selector: { matchLabels },
    },
  }
  if (form.value.useMinAvailable && form.value.minAvailable) {
    pdb.spec.minAvailable = form.value.minAvailable
  } else if (!form.value.useMinAvailable && form.value.maxUnavailable) {
    pdb.spec.maxUnavailable = form.value.maxUnavailable
  }
  if (form.value.unhealthyPodEvictionPolicy) {
    pdb.spec.unhealthyPodEvictionPolicy = form.value.unhealthyPodEvictionPolicy
  }
  yamlContent.value = JSON.stringify(pdb, null, 2)
}

async function handleCreate() {
  if (!form.value.name || !form.value.namespace) {
    ElMessage.warning('Please fill in name and namespace')
    return
  }
  loading.value = true
  try {
    buildYaml()
    await createPdb({ namespace: form.value.namespace, yaml: yamlContent.value })
    ElMessage.success('PDB created successfully')
    router.push('/workloads/pdb')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create PDB')
  } finally { loading.value = false }
}

function addLabel() { form.value.selectorLabels.push({ key: '', value: '' }) }
function removeLabel(i: number) { form.value.selectorLabels.splice(i, 1) }

onMounted(fetchNamespaces)
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 style="margin: 0;">创建 PodDisruptionBudget</h2>
      <el-button @click="router.push('/workloads/pdb')">Back to List</el-button>
    </div>
    <el-card shadow="never">
      <el-form label-width="160px">
        <el-form-item label="Name" required><el-input v-model="form.name" placeholder="my-pdb" /></el-form-item>
        <el-form-item label="Namespace" required>
          <el-select v-model="form.namespace" placeholder="Select namespace" style="width: 100%;">
            <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <div style="width: 100%;">
            <div v-for="(label, i) in form.labels" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="label.key" placeholder="Key" style="flex: 1;" />
              <el-input v-model="label.value" placeholder="Value" style="flex: 1;" />
              <el-button type="danger" circle size="small" @click="form.labels.splice(i, 1)">X</el-button>
            </div>
            <el-button size="small" @click="form.labels.push({ key: '', value: '' })">+ 添加标签</el-button>
          </div>
        </el-form-item>
        <el-form-item label="Pod Selector">
          <div style="width: 100%;">
            <div v-for="(label, i) in form.selectorLabels" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="label.key" placeholder="Key" style="flex: 1;" />
              <el-input v-model="label.value" placeholder="Value" style="flex: 1;" />
              <el-button type="danger" circle size="small" @click="removeLabel(i)">X</el-button>
            </div>
            <el-button size="small" @click="addLabel">+ Add Label</el-button>
          </div>
        </el-form-item>
        <el-form-item label="Policy">
          <el-radio-group v-model="form.useMinAvailable">
            <el-radio :value="true">Min Available</el-radio>
            <el-radio :value="false">Max Unavailable</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.useMinAvailable" label="Min Available">
          <el-input v-model="form.minAvailable" placeholder="e.g. 1 or 25%" />
        </el-form-item>
        <el-form-item v-else label="Max Unavailable">
          <el-input v-model="form.maxUnavailable" placeholder="e.g. 1 or 25%" />
        </el-form-item>
        <el-form-item label="不健康 Pod 驱逐策略">
          <el-select v-model="form.unhealthyPodEvictionPolicy" clearable placeholder="默认" style="width: 100%;">
            <el-option label="Default (默认)" value="" />
            <el-option label="AlwaysAllow (始终允许)" value="AlwaysAllow" />
            <el-option label="IfHealthy (健康时)" value="IfHealthy" />
          </el-select>
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">K8s 1.27+ 支持</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleCreate">创建 PDB</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card shadow="never" style="margin-top: 16px;">
      <template #header><span>YAML Preview</span></template>
      <YamlEditor :model-value="yamlContent" height="300px" read-only />
    </el-card>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
  max-width: 900px;
  margin: 0 auto;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
</style>
