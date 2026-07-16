<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'
import { createResourceQuota, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const namespaceList = ref<string[]>([])

const form = ref({
  name: '',
  namespace: 'default',
  labels: [{ key: '', value: '' }] as Array<{ key: string; value: string }>,
  scopes: [] as string[],
  scopeSelector: '',
  limits: [
    { resource: 'requests.cpu', value: '' },
    { resource: 'requests.memory', value: '' },
    { resource: 'limits.cpu', value: '' },
    { resource: 'limits.memory', value: '' },
    { resource: 'pods', value: '' },
    { resource: 'services', value: '' },
    { resource: 'persistentvolumeclaims', value: '' },
    { resource: 'configmaps', value: '' },
    { resource: 'secrets', value: '' },
    { resource: 'replicationcontrollers', value: '' },
  ],
})

const yamlContent = ref('')

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

function buildYaml() {
  const hard: Record<string, string> = {}
  form.value.limits.forEach(l => { if (l.value) hard[l.resource] = l.value })

  const labels: Record<string, string> = {}
  form.value.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const metadata: Record<string, any> = { name: form.value.name, namespace: form.value.namespace }
  if (Object.keys(labels).length > 0) metadata.labels = labels

  const spec: any = { hard }
  if (form.value.scopes.length > 0) spec.scopes = form.value.scopes
  if (form.value.scopeSelector) {
    spec.scopeSelector = {
      matchExpressions: [{
        scopeName: form.value.scopeSelector,
        operator: 'In',
        values: [''],
      }],
    }
  }

  const rq = {
    apiVersion: 'v1',
    kind: 'ResourceQuota',
    metadata,
    spec,
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
    await createResourceQuota({ namespace: form.value.namespace, yaml: yamlContent.value })
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
      <h2 style="margin: 0;">创建 ResourceQuota</h2>
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
        <el-form-item label="作用域 (Scopes)">
          <el-select v-model="form.scopes" multiple clearable placeholder="选择作用域（可选）" style="width: 100%;">
            <el-option label="Terminating" value="Terminating" />
            <el-option label="NotTerminating" value="NotTerminating" />
            <el-option label="BestEffort" value="BestEffort" />
            <el-option label="NotBestEffort" value="NotBestEffort" />
            <el-option label="PriorityClass" value="PriorityClass" />
            <el-option label="CrossNamespacePodAffinity" value="CrossNamespacePodAffinity" />
          </el-select>
        </el-form-item>
        <el-divider>Resource Limits</el-divider>
        <el-form-item v-for="(limit, i) in form.limits" :key="i" :label="limit.resource">
          <el-input v-model="limit.value" :placeholder="limit.resource === 'pods' || limit.resource === 'services' || limit.resource === 'persistentvolumeclaims' ? 'e.g. 10' : 'e.g. 4 or 8Gi'" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleCreate">创建 ResourceQuota</el-button>
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
