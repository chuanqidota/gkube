<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'
import { createNetworkPolicy, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const namespaceList = ref<string[]>([])

const form = ref({
  name: '',
  namespace: '',
  policyTypes: ['Ingress', 'Egress'],
  podSelectorLabels: [{ key: 'app', value: '' }],
})

const defaultYaml = `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: my-network-policy
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: my-app
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: allowed-app
      ports:
        - protocol: TCP
          port: 80
  egress:
    - to:
        - ipBlock:
            cidr: 10.0.0.0/8
      ports:
        - protocol: TCP
          port: 443
`

const yamlContent = ref(defaultYaml)

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

function buildYaml() {
  const labels: Record<string, string> = {}
  form.value.podSelectorLabels.forEach(l => { if (l.key) labels[l.key] = l.value })
  const np = {
    apiVersion: 'networking.k8s.io/v1',
    kind: 'NetworkPolicy',
    metadata: { name: form.value.name, namespace: form.value.namespace },
    spec: {
      podSelector: { matchLabels: labels },
      policyTypes: form.value.policyTypes,
      ingress: [{ from: [{ podSelector: { matchLabels: {} } }] }],
      egress: [{ to: [{ ipBlock: { cidr: '0.0.0.0/0' } }] }],
    },
  }
  yamlContent.value = JSON.stringify(np, null, 2)
}

async function handleCreate() {
  if (!form.value.name || !form.value.namespace) {
    ElMessage.warning('Please fill in name and namespace')
    return
  }
  loading.value = true
  try {
    buildYaml()
    await createNetworkPolicy({ namespace: form.value.namespace, yaml: yamlContent.value })
    ElMessage.success('NetworkPolicy created successfully')
    router.push('/network/networkpolicies')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create NetworkPolicy')
  } finally { loading.value = false }
}

function addLabel() { form.value.podSelectorLabels.push({ key: '', value: '' }) }
function removeLabel(i: number) { form.value.podSelectorLabels.splice(i, 1) }

onMounted(fetchNamespaces)
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 style="margin: 0;">Create NetworkPolicy</h2>
      <el-button @click="router.push('/network/networkpolicies')">Back to List</el-button>
    </div>
    <el-card shadow="never">
      <el-form label-width="140px" style="max-width: 600px;">
        <el-form-item label="Name" required>
          <el-input v-model="form.name" placeholder="my-network-policy" />
        </el-form-item>
        <el-form-item label="Namespace" required>
          <el-select v-model="form.namespace" placeholder="Select namespace" style="width: 100%;">
            <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="Policy Types">
          <el-checkbox-group v-model="form.policyTypes">
            <el-checkbox label="Ingress" value="Ingress" />
            <el-checkbox label="Egress" value="Egress" />
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="Pod Selector">
          <div v-for="(label, i) in form.podSelectorLabels" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
            <el-input v-model="label.key" placeholder="Key" style="flex: 1;" />
            <el-input v-model="label.value" placeholder="Value" style="flex: 1;" />
            <el-button type="danger" circle size="small" @click="removeLabel(i)">X</el-button>
          </div>
          <el-button size="small" @click="addLabel">Add Label</el-button>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleCreate">Create NetworkPolicy</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card shadow="never" style="margin-top: 16px;">
      <template #header><span>YAML Preview</span></template>
      <YamlEditor v-model="yamlContent" height="400px" read-only />
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
