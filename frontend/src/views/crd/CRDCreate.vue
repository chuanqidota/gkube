<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { FullScreen } from '@element-plus/icons-vue'
import { createCrd } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()
const router = useRouter()
const yamlEditorRef = ref()
const submitting = ref(false)
const yamlContent = ref(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: myresources.example.com
spec:
  group: example.com
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                name:
                  type: string
  scope: Namespaced
  names:
    plural: myresources
    singular: myresource
    kind: MyResource
    shortNames:
      - mr
`)

async function handleSubmit() {
  if (!yamlContent.value.trim()) {
    ElMessage.warning(t('crd.crdYamlRequired'))
    return
  }
  submitting.value = true
  try {
    await createCrd({ yaml: yamlContent.value })
    ElMessage.success(t('crd.crdCreated'))
    router.push('/crd')
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.createFailed'))
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/crd')
}

function handleFormat() {
  yamlEditorRef.value?.handleFormat()
}

function handleCopy() {
  yamlEditorRef.value?.handleCopy()
}

function handleMaximize() {
  yamlEditorRef.value?.toggleFullscreen()
}
</script>

<template>
  <div class="crd-create">
    <div class="yaml-card">
      <div class="yaml-card-header">
        <div class="yaml-card-left">
          <span class="yaml-card-title">{{ t('crd.yamlConfig') }}</span>
          <el-button-group>
            <el-button size="small" @click="handleFormat">Format</el-button>
            <el-button size="small" @click="handleCopy">{{ t('common.copy') }}</el-button>
          </el-button-group>
          <el-tooltip :content="t('crd.maximize')" placement="top">
            <el-icon class="maximize-btn" @click="handleMaximize"><FullScreen /></el-icon>
          </el-tooltip>
        </div>
        <div class="yaml-card-actions">
          <el-button size="small" @click="handleCancel">{{ t('common.cancel') }}</el-button>
          <el-button size="small" type="primary" :loading="submitting" @click="handleSubmit">{{
            t('common.create')
          }}</el-button>
        </div>
      </div>
      <div class="yaml-card-body">
        <YamlEditor
          ref="yamlEditorRef"
          v-model="yamlContent"
          height="calc(100dvh - 180px)"
          :read-only="false"
          editable
          auto-format
          :show-toolbar="false"
          :title="t('crd.yamlConfig')"
        >
          <template #fullscreen-actions>
            <el-button size="small" @click="handleCancel">{{ t('common.cancel') }}</el-button>
            <el-button size="small" type="primary" :loading="submitting" @click="handleSubmit">{{
              t('common.create')
            }}</el-button>
          </template>
        </YamlEditor>
      </div>
    </div>
  </div>
</template>

<style scoped>
.crd-create {
  max-width: 1100px;
  margin: 0 auto;
  padding: var(--gk-space-5) var(--gk-space-4);
}

.yaml-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--gk-radius-md);
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
  gap: var(--gk-space-3);
}

.yaml-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.yaml-card-actions {
  display: flex;
  gap: var(--gk-space-2);
}

.yaml-card-body {
  padding: 0;
}

.maximize-btn {
  cursor: pointer;
  font-size: 16px;
  color: var(--el-text-color-secondary);
  margin-left: 4px;
  transition: color 0.2s;
}
.maximize-btn:hover {
  color: var(--el-color-primary);
}
</style>
