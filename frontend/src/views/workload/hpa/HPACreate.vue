<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ElMessage } from 'element-plus'
import { createHpa, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const namespaceList = ref<string[]>([])

const form = ref({
  name: '',
  namespace: '',
  targetKind: 'Deployment',
  targetName: '',
  minReplicas: 1,
  maxReplicas: 10,
  cpuUtilization: 80,
})

const defaultYaml = `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-hpa
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-deployment
  minReplicas: 1
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 80
`

const yamlContent = ref(defaultYaml)

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

function buildYaml() {
  const hpa = {
    apiVersion: 'autoscaling/v2',
    kind: 'HorizontalPodAutoscaler',
    metadata: {
      name: form.value.name,
      namespace: form.value.namespace,
    },
    spec: {
      scaleTargetRef: {
        apiVersion: 'apps/v1',
        kind: form.value.targetKind,
        name: form.value.targetName,
      },
      minReplicas: form.value.minReplicas,
      maxReplicas: form.value.maxReplicas,
      metrics: [
        {
          type: 'Resource',
          resource: {
            name: 'cpu',
            target: {
              type: 'Utilization',
              averageUtilization: form.value.cpuUtilization,
            },
          },
        },
      ],
    },
  }
  yamlContent.value = JSON.stringify(hpa, null, 2)
}

async function handleCreate() {
  if (!form.value.name || !form.value.namespace || !form.value.targetName) {
    ElMessage.warning('Please fill in all required fields')
    return
  }
  loading.value = true
  try {
    buildYaml()
    await createHpa({ namespace: form.value.namespace, yaml: yamlContent.value })
    ElMessage.success('HPA created successfully')
    router.push('/workloads/hpa')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create HPA')
  } finally { loading.value = false }
}

onMounted(fetchNamespaces)
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 style="margin: 0;">创建 HPA</h2>
      <el-button @click="router.push('/workloads/hpa')">Back to List</el-button>
    </div>

    <el-card shadow="never">
      <el-form label-width="140px" style="max-width: 600px;">
        <el-form-item label="Name" required>
          <el-input v-model="form.name" placeholder="my-hpa" />
        </el-form-item>
        <el-form-item label="Namespace" required>
          <el-select v-model="form.namespace" placeholder="Select namespace" style="width: 100%;">
            <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="Target Kind">
          <el-select v-model="form.targetKind" style="width: 100%;">
            <el-option label="Deployment" value="Deployment" />
            <el-option label="StatefulSet" value="StatefulSet" />
            <el-option label="ReplicaSet" value="ReplicaSet" />
          </el-select>
        </el-form-item>
        <el-form-item label="Target Name" required>
          <el-input v-model="form.targetName" placeholder="my-deployment" />
        </el-form-item>
        <el-form-item label="Min Replicas">
          <el-input-number v-model="form.minReplicas" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="Max Replicas">
          <el-input-number v-model="form.maxReplicas" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="CPU Target (%)">
          <el-input-number v-model="form.cpuUtilization" :min="1" :max="100" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleCreate">创建 HPA</el-button>
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
