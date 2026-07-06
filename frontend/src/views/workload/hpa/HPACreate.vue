<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import HPAForm from './components/HPAForm.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import { createHpa } from '@/api/resource'

const router = useRouter()
const { t } = useI18n()
const mode = ref<'form' | 'yaml'>('form')
const yamlContent = ref(`apiVersion: autoscaling/v2
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
    await createHpa({ namespace: ns, yaml: yamlContent.value })
    ElMessage.success('弹性伸缩创建成功')
    router.push('/workloads/hpa')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/workloads/hpa')
}
</script>

<template>
  <div class="hpa-create">
    <div class="mode-switcher">
      <el-segmented v-model="mode" :options="[{ label: t('common.formCreate'), value: 'form' }, { label: t('common.yamlCreate'), value: 'yaml' }]" size="small" />
    </div>

    <HPAForm v-if="mode === 'form'" />

    <div v-else class="yaml-mode">
      <div class="form-header"><h2>创建弹性伸缩 (YAML)</h2></div>
      <YamlEditor v-model="yamlContent" height="calc(100vh - 230px)" :read-only="false" editable auto-format />
      <div class="form-actions">
        <el-button @click="handleCancel">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleYamlSubmit">创建 HPA</el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.hpa-create { max-width: 1100px; margin: 0 auto; padding: 20px 0; }
.mode-switcher { display: flex; justify-content: center; margin-bottom: 12px; }
.yaml-mode { padding: 0 16px; }
.form-header { margin-bottom: 24px; }
.form-header h2 { margin: 0; font-size: 20px; font-weight: 600; }
.form-actions { display: flex; justify-content: flex-end; gap: 12px; padding-top: 24px; border-top: 1px solid var(--gk-color-border); margin-top: 24px; }
</style>
