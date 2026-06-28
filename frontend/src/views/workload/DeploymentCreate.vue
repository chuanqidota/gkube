<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import WorkloadForm from '@/components/WorkloadForm.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import { createDeployment } from '@/api/resource'

const router = useRouter()
const mode = ref<'form' | 'yaml'>('form')
const yamlContent = ref(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
  labels:
    app: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-app
          image: nginx:latest
          ports:
            - containerPort: 80
`)
const submitting = ref(false)

async function handleYamlSubmit() {
  if (!yamlContent.value.trim()) {
    ElMessage.error('YAML content is required')
    return
  }
  submitting.value = true
  try {
    const parsed = yaml.load(yamlContent.value) as any
    const ns = parsed?.metadata?.namespace || 'default'
    await createDeployment({ namespace: ns, yaml: yamlContent.value })
    ElMessage.success('Deployment created successfully')
    router.push('/workloads/deployments')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/workloads/deployments')
}
</script>

<template>
  <div class="deployment-create">
    <div class="mode-switcher">
      <el-segmented v-model="mode" :options="[{ label: '表单创建', value: 'form' }, { label: 'YAML创建', value: 'yaml' }]" size="small" />
    </div>

    <WorkloadForm v-if="mode === 'form'" kind="Deployment" />

    <div v-else class="yaml-mode">
      <div class="form-header"><h2>Create Deployment (YAML)</h2></div>
      <YamlEditor v-model="yamlContent" height="calc(100vh - 230px)" :read-only="false" editable auto-format />
      <div class="form-actions">
        <el-button @click="handleCancel">Cancel</el-button>
        <el-button type="primary" :loading="submitting" @click="handleYamlSubmit">Create Deployment</el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.deployment-create { max-width: 1100px; margin: 0 auto; padding: 20px 0; }
.mode-switcher { display: flex; justify-content: center; margin-bottom: 12px; }
.yaml-mode { padding: 0 16px; }
.form-header { margin-bottom: 24px; }
.form-header h2 { margin: 0; font-size: 20px; font-weight: 600; }
.form-actions { display: flex; justify-content: flex-end; gap: 12px; padding-top: 24px; border-top: 1px solid var(--gk-color-border); margin-top: 24px; }
</style>
