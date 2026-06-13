<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createPdb, getNamespaceList } from '@/api/resource'
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
})

const yamlContent = ref('')

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = (res.data || []).map((ns: any) => ns.name || ns)
  } catch { /* ignore */ }
}

function buildYaml() {
  const labels: Record<string, string> = {}
  form.value.selectorLabels.forEach(l => { if (l.key) labels[l.key] = l.value })
  const pdb: any = {
    apiVersion: 'policy/v1',
    kind: 'PodDisruptionBudget',
    metadata: { name: form.value.name, namespace: form.value.namespace },
    spec: {
      selector: { matchLabels: labels },
    },
  }
  if (form.value.useMinAvailable && form.value.minAvailable) {
    pdb.spec.minAvailable = form.value.minAvailable
  } else if (!form.value.useMinAvailable && form.value.maxUnavailable) {
    pdb.spec.maxUnavailable = form.value.maxUnavailable
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
    await createPdb({ namespace: form.value.namespace, yamlContent: yamlContent.value })
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
      <h2 style="margin: 0;">Create PodDisruptionBudget</h2>
      <el-button @click="router.push('/workloads/pdb')">Back to List</el-button>
    </div>
    <el-card shadow="never">
      <el-form label-width="160px" style="max-width: 600px;">
        <el-form-item label="Name" required><el-input v-model="form.name" placeholder="my-pdb" /></el-form-item>
        <el-form-item label="Namespace" required>
          <el-select v-model="form.namespace" placeholder="Select namespace" style="width: 100%;">
            <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
          </el-select>
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
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleCreate">Create PDB</el-button>
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
