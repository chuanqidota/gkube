<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { FullScreen } from '@element-plus/icons-vue'
import { createCrd } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

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
    ElMessage.warning('请输入 CRD YAML')
    return
  }
  submitting.value = true
  try {
    await createCrd({ yaml: yamlContent.value })
    ElMessage.success('CRD 创建成功')
    router.push('/crd')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
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
          <span class="yaml-card-title">YAML 配置</span>
          <el-button-group>
            <el-button size="small" @click="handleFormat">Format</el-button>
            <el-button size="small" @click="handleCopy">复制</el-button>
          </el-button-group>
          <el-tooltip content="最大化" placement="top">
            <el-icon class="maximize-btn" @click="handleMaximize"><FullScreen /></el-icon>
          </el-tooltip>
        </div>
        <div class="yaml-card-actions">
          <el-button size="small" @click="handleCancel">取消</el-button>
          <el-button size="small" type="primary" :loading="submitting" @click="handleSubmit">创建</el-button>
        </div>
      </div>
      <div class="yaml-card-body">
        <YamlEditor
          ref="yamlEditorRef"
          v-model="yamlContent"
          height="calc(100vh - 180px)"
          :read-only="false"
          editable
          auto-format
          :show-toolbar="false"
          title="YAML 配置"
        >
          <template #fullscreen-actions>
            <el-button size="small" @click="handleCancel">取消</el-button>
            <el-button size="small" type="primary" :loading="submitting" @click="handleSubmit">创建</el-button>
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
  padding: 20px 16px;
}

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
