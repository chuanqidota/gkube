<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { FullScreen } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import IngressForm from '@/views/network/components/IngressForm.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import { createIngress } from '@/api/resource'

const router = useRouter()
const { t } = useI18n()
const mode = ref<'form' | 'yaml'>('form')
const yamlEditorRef = ref()
const yamlContent = ref(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  namespace: default
  labels:
    app: my-ingress
spec:
  ingressClassName: nginx
  rules:
    - host: example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-service
                port:
                  number: 80
`)
const submitting = ref(false)

async function handleYamlSubmit() {
  if (!yamlContent.value.trim()) {
    ElMessage.error('YAML 内容不能为空')
    return
  }
  submitting.value = true
  try {
    const parsed = yaml.load(yamlContent.value) as any
    const ns = parsed?.metadata?.namespace || 'default'
    await createIngress({ namespace: ns, yaml: yamlContent.value })
    ElMessage.success('Ingress 创建成功')
    router.push('/network/ingresses')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/network/ingresses')
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
  <div class="ingress-create">
    <div class="mode-switcher">
      <el-segmented v-model="mode" :options="[{ label: t('common.formCreate'), value: 'form' }, { label: t('common.yamlCreate'), value: 'yaml' }]" size="small" />
    </div>

    <IngressForm v-if="mode === 'form'" />

    <div v-else class="yaml-mode">
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
            <el-button size="small" type="primary" :loading="submitting" @click="handleYamlSubmit">创建</el-button>
          </div>
        </div>
        <div class="yaml-card-body">
          <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="calc(100vh - 180px)" :read-only="false" editable auto-format :show-toolbar="false" title="YAML 配置">
            <template #fullscreen-actions>
              <el-button size="small" @click="handleCancel">取消</el-button>
              <el-button size="small" type="primary" :loading="submitting" @click="handleYamlSubmit">创建</el-button>
            </template>
          </YamlEditor>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ingress-create { max-width: 1100px; margin: 0 auto; padding: 20px 0; }
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
