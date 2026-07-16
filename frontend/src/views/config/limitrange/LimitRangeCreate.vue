<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'
import { createLimitRange, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const namespaceList = ref<string[]>([])

const form = ref({
  name: '',
  namespace: 'default',
  labels: [{ key: '', value: '' }] as Array<{ key: string; value: string }>,
  limits: [{
    type: 'Container',
    maxCpu: '', maxMemory: '',
    minCpu: '', minMemory: '',
    defaultCpu: '', defaultMemory: '',
    defaultRequestCpu: '', defaultRequestMemory: '',
    maxLimitRequestRatioCpu: '', maxLimitRequestRatioMemory: '',
  }],
})

const yamlContent = ref('')

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

function buildYaml() {
  const limits = form.value.limits.map(l => {
    const limit: any = { type: l.type }
    const max: any = {}; const min: any = {}; const def: any = {}; const defReq: any = {}
    if (l.maxCpu) max.cpu = l.maxCpu
    if (l.maxMemory) max.memory = l.maxMemory
    if (l.minCpu) min.cpu = l.minCpu
    if (l.minMemory) min.memory = l.minMemory
    if (l.defaultCpu) def.cpu = l.defaultCpu
    if (l.defaultMemory) def.memory = l.defaultMemory
    if (l.defaultRequestCpu) defReq.cpu = l.defaultRequestCpu
    if (l.defaultRequestMemory) defReq.memory = l.defaultRequestMemory
    if (Object.keys(max).length) limit.max = max
    if (Object.keys(min).length) limit.min = min
    if (Object.keys(def).length) limit.default = def
    if (Object.keys(defReq).length) limit.defaultRequest = defReq
    const maxLimitReqRatio: any = {}
    if (l.maxLimitRequestRatioCpu) maxLimitReqRatio.cpu = l.maxLimitRequestRatioCpu
    if (l.maxLimitRequestRatioMemory) maxLimitReqRatio.memory = l.maxLimitRequestRatioMemory
    if (Object.keys(maxLimitReqRatio).length) limit.maxLimitRequestRatio = maxLimitReqRatio
    return limit
  })
  const labels: Record<string, string> = {}
  form.value.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const metadata: Record<string, any> = { name: form.value.name, namespace: form.value.namespace }
  if (Object.keys(labels).length > 0) metadata.labels = labels

  const lr = {
    apiVersion: 'v1',
    kind: 'LimitRange',
    metadata,
    spec: { limits },
  }
  yamlContent.value = JSON.stringify(lr, null, 2)
}

async function handleCreate() {
  if (!form.value.name || !form.value.namespace) {
    ElMessage.warning('Please fill in name and namespace')
    return
  }
  loading.value = true
  try {
    buildYaml()
    await createLimitRange({ namespace: form.value.namespace, yaml: yamlContent.value })
    ElMessage.success('LimitRange created successfully')
    router.push('/config/limitranges')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create LimitRange')
  } finally { loading.value = false }
}

function addLimit() {
  form.value.limits.push({
    type: 'Container', maxCpu: '', maxMemory: '', minCpu: '', minMemory: '',
    defaultCpu: '', defaultMemory: '', defaultRequestCpu: '', defaultRequestMemory: '',
    maxLimitRequestRatioCpu: '', maxLimitRequestRatioMemory: '',
  })
}

function removeLimit(i: number) { form.value.limits.splice(i, 1) }

onMounted(fetchNamespaces)
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 style="margin: 0;">创建 LimitRange</h2>
      <el-button @click="router.push('/config/limitranges')">Back to List</el-button>
    </div>
    <el-card shadow="never">
      <el-form label-width="160px" style="max-width: 700px;">
        <el-form-item label="Name" required><el-input v-model="form.name" placeholder="my-limit-range" /></el-form-item>
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

        <div v-for="(limit, i) in form.limits" :key="i" style="border: 1px solid var(--gk-color-border); border-radius: 8px; padding: 16px; margin-bottom: 16px;">
          <div style="display: flex; justify-content: space-between; margin-bottom: 12px;">
            <el-select v-model="limit.type" style="width: 200px;">
              <el-option label="Container" value="Container" />
              <el-option label="Pod" value="Pod" />
              <el-option label="PersistentVolumeClaim" value="PersistentVolumeClaim" />
            </el-select>
            <el-button v-if="form.limits.length > 1" type="danger" size="small" @click="removeLimit(i)">移除</el-button>
          </div>
          <el-form-item label="Max CPU"><el-input v-model="limit.maxCpu" placeholder="e.g. 4" /></el-form-item>
          <el-form-item label="Max Memory"><el-input v-model="limit.maxMemory" placeholder="e.g. 8Gi" /></el-form-item>
          <el-form-item label="Min CPU"><el-input v-model="limit.minCpu" placeholder="e.g. 100m" /></el-form-item>
          <el-form-item label="Min Memory"><el-input v-model="limit.minMemory" placeholder="e.g. 128Mi" /></el-form-item>
          <el-form-item label="Default CPU"><el-input v-model="limit.defaultCpu" placeholder="e.g. 500m" /></el-form-item>
          <el-form-item label="Default Memory"><el-input v-model="limit.defaultMemory" placeholder="e.g. 512Mi" /></el-form-item>
          <el-form-item label="Default Req CPU"><el-input v-model="limit.defaultRequestCpu" placeholder="e.g. 100m" /></el-form-item>
          <el-form-item label="Default Req Memory"><el-input v-model="limit.defaultRequestMemory" placeholder="e.g. 128Mi" /></el-form-item>
          <el-form-item label="MaxLimit/Request CPU">
            <el-input v-model="limit.maxLimitRequestRatioCpu" placeholder="e.g. 10" />
            <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">限制/请求的最大比率</div>
          </el-form-item>
          <el-form-item label="MaxLimit/Request Memory">
            <el-input v-model="limit.maxLimitRequestRatioMemory" placeholder="e.g. 4" />
          </el-form-item>
        </div>
        <el-button @click="addLimit" style="margin-bottom: 16px;">+ Add Limit</el-button>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleCreate">创建 LimitRange</el-button>
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
