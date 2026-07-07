<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import CronJobForm from '@/views/workload/components/CronJobForm.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import { createCronJob } from '@/api/resource'

const router = useRouter()
const { t } = useI18n()
const mode = ref<'form' | 'yaml'>('form')
const yamlEditorRef = ref()
const yamlContent = ref(`apiVersion: batch/v1
kind: CronJob
metadata:
  name: my-cronjob
  namespace: default
  labels:
    app: my-cronjob
spec:
  schedule: "*/5 * * * *"
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app: my-cronjob
        spec:
          containers:
            - name: my-cronjob
              image: busybox:latest
              command: ["echo", "Hello from CronJob"]
          restartPolicy: Never
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
    await createCronJob({ namespace: ns, yaml: yamlContent.value })
    ElMessage.success('CronJob created successfully')
    router.push('/workloads/cronjobs')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/workloads/cronjobs')
}

function handleFormat() {
  yamlEditorRef.value?.handleFormat()
}

function handleCopy() {
  yamlEditorRef.value?.handleCopy()
}
</script>

<template>
  <div class="cronjob-create">
    <div class="mode-switcher">
      <el-segmented v-model="mode" :options="[{ label: t('common.formCreate'), value: 'form' }, { label: t('common.yamlCreate'), value: 'yaml' }]" size="small" />
    </div>

    <CronJobForm v-if="mode === 'form'" />

    <div v-else class="yaml-mode">
      <div class="yaml-card">
        <div class="yaml-card-header">
          <div class="yaml-card-left">
            <span class="yaml-card-title">YAML 配置</span>
            <el-button-group>
              <el-button size="small" @click="handleFormat">Format</el-button>
              <el-button size="small" @click="handleCopy">复制</el-button>
            </el-button-group>
          </div>
          <div class="yaml-card-actions">
            <el-button size="small" @click="handleCancel">取消</el-button>
            <el-button size="small" type="primary" :loading="submitting" @click="handleYamlSubmit">创建</el-button>
          </div>
        </div>
        <div class="yaml-card-body">
          <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="calc(100vh - 180px)" :read-only="false" editable auto-format :show-toolbar="false" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cronjob-create { max-width: 1100px; margin: 0 auto; padding: 20px 0; }
.mode-switcher { display: flex; justify-content: center; margin-bottom: 12px; }
.yaml-mode { padding: 0 16px; }

.yaml-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  overflow: hidden;
  background: var(--el-bg-color);
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.yaml-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--el-border-color-light);
}

.yaml-card-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.yaml-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.yaml-card-actions {
  display: flex;
  gap: 8px;
}

.yaml-card-body {
  padding: 0;
}
</style>
