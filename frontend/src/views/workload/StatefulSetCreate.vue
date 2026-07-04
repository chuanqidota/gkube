<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import WorkloadForm from '@/views/workload/components/WorkloadForm.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import { createStatefulSet } from '@/api/resource'

const router = useRouter()
const { t } = useI18n()
const mode = ref<'form' | 'yaml'>('form')
const yamlContent = ref(`apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-statefulset
  namespace: default
  labels:
    app: my-statefulset
spec:
  replicas: 1
  serviceName: my-statefulset
  selector:
    matchLabels:
      app: my-statefulset
  template:
    metadata:
      labels:
        app: my-statefulset
    spec:
      containers:
        - name: my-statefulset
          image: nginx:latest
          ports:
            - containerPort: 80
          volumeMounts:
            - name: data
              mountPath: /data
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
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
    await createStatefulSet({ namespace: ns, yaml: yamlContent.value })
    ElMessage.success('StatefulSet created successfully')
    router.push('/workloads/statefulsets')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/workloads/statefulsets')
}
</script>

<template>
  <div class="statefulset-create">
    <div class="mode-switcher">
      <el-segmented v-model="mode" :options="[{ label: t('common.formCreate'), value: 'form' }, { label: t('common.yamlCreate'), value: 'yaml' }]" size="small" />
    </div>

    <WorkloadForm v-if="mode === 'form'" kind="StatefulSet" />

    <div v-else class="yaml-mode">
      <div class="form-header"><h2>Create StatefulSet (YAML)</h2></div>
      <YamlEditor v-model="yamlContent" height="calc(100vh - 230px)" :read-only="false" editable auto-format />
      <div class="form-actions">
        <el-button @click="handleCancel">Cancel</el-button>
        <el-button type="primary" :loading="submitting" @click="handleYamlSubmit">Create StatefulSet</el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.statefulset-create { max-width: 1100px; margin: 0 auto; padding: 20px 0; }
.mode-switcher { display: flex; justify-content: center; margin-bottom: 12px; }
.yaml-mode { padding: 0 16px; }
.form-header { margin-bottom: 24px; }
.form-header h2 { margin: 0; font-size: 20px; font-weight: 600; }
.form-actions { display: flex; justify-content: flex-end; gap: 12px; padding-top: 24px; border-top: 1px solid var(--gk-color-border); margin-top: 24px; }
</style>
