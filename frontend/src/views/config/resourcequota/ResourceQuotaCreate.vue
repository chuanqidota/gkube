<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'
import { createResourceQuota, getNamespaceList } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const namespaceList = ref<string[]>([])

const form = ref({
  name: '',
  namespace: 'default',
  limits: [
    { resource: 'requests.cpu', value: '' },
    { resource: 'requests.memory', value: '' },
    { resource: 'limits.cpu', value: '' },
    { resource: 'limits.memory', value: '' },
    { resource: 'pods', value: '' },
    { resource: 'services', value: '' },
    { resource: 'persistentvolumeclaims', value: '' },
  ],
})

const yamlContent = ref('')

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = (res.data || []).map((ns: any) => ns.name || ns)
  } catch { /* ignore */ }
}

function buildYaml() {
  const hard: Record<string, string> = {}
  form.value.limits.forEach(l => { if (l.value) hard[l.resource] = l.value })
  const rq = {
    apiVersion: 'v1',
    kind: 'ResourceQuota',
    metadata: { name: form.value.name, namespace: form.value.namespace },
    spec: { hard },
  }
  yamlContent.value = JSON.stringify(rq, null, 2)
}

async function handleCreate() {
  if (!form.value.name || !form.value.namespace) {
    ElMessage.warning('Please fill in name and namespace')
    return
  }
  loading.value = true
  try {
    buildYaml()
    await createResourceQuota({ namespace: form.value.namespace, yamlContent: yamlContent.value })
    ElMessage.success('ResourceQuota created successfully')
    router.push('/config/resourcequotas')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create ResourceQuota')
  } finally { loading.value = false }
}

onMounted(fetchNamespaces)
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 style="margin: 0;">Create ResourceQuota</h2>
      <el-button @click="router.push('/config/resourcequotas')">Back to List</el-button>
    </div>
    <el-card shadow="never">
      <el-form label-width="180px" style="max-width: 600px;">
        <el-form-item label="Name" required><el-input v-model="form.name" placeholder="my-resource-quota" /></el-form-item>
        <el-form-item label="Namespace" required>
          <el-select v-model="form.namespace" placeholder="Select namespace" style="width: 100%;">
            <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-divider>Resource Limits</el-divider>
        <el-form-item v-for="(limit, i) in form.limits" :key="i" :label="limit.resource">
          <el-input v-model="limit.value" :placeholder="limit.resource === 'pods' || limit.resource === 'services' || limit.resource === 'persistentvolumeclaims' ? 'e.g. 10' : 'e.g. 4 or 8Gi'" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleCreate">Create ResourceQuota</el-button>
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
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
