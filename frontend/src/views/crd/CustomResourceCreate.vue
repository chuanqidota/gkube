<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createCustomResource } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const submitting = ref(false)

const group = route.query.group as string
const version = route.query.version as string
const resource = route.query.resource as string
const scope = route.query.scope as string
const kind = route.query.kind as string || ''

const isNamespaced = scope === 'Namespaced'

const yamlContent = ref(
`apiVersion: ${group}/${version}
kind: ${kind}
metadata:
  name: example${isNamespaced ? '\n  namespace: default' : ''}
spec:
  # Add your spec here
`)

const backRoute = computed(() =>
  `/crd/resources?group=${group}&version=${version}&resource=${resource}&scope=${scope}`
)

async function handleSubmit() {
  if (!yamlContent.value.trim()) {
    ElMessage.warning('请输入 YAML')
    return
  }
  submitting.value = true
  try {
    const data: any = { group, version, resource, yaml: yamlContent.value }
    await createCustomResource(data)
    ElMessage.success('自定义资源创建成功')
    router.push(backRoute.value)
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push(backRoute.value)
}
</script>

<template>
  <div class="cr-create">
    <div class="yaml-card">
      <div class="yaml-card-header">
        <div class="yaml-card-left">
          <span class="yaml-card-title">创建 {{ kind || resource }}</span>
          <el-tag size="small" type="info">{{ group }}/{{ version }}</el-tag>
        </div>
        <div class="yaml-card-actions">
          <el-button size="small" @click="handleCancel">取消</el-button>
          <el-button size="small" type="primary" :loading="submitting" @click="handleSubmit">创建</el-button>
        </div>
      </div>
      <div class="yaml-card-body">
        <YamlEditor
          v-model="yamlContent"
          height="calc(100vh - 180px)"
          editable
          auto-format
          :show-toolbar="true"
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
.cr-create {
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
</style>
