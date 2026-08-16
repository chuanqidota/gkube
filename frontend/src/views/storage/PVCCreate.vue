<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { FullScreen, CopyDocument } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import PVCForm from '@/views/storage/components/PVCForm.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import CloneDialog from '@/components/CloneDialog.vue'
import { useCloneCreate, dumpCloneYaml } from '@/composables/useCloneCreate'
import { createPvc, getPvcList, getPvcYaml } from '@/api/resource'

const router = useRouter()
const { t } = useI18n()
const mode = ref<'form' | 'yaml'>('form')
const yamlEditorRef = ref()
const parsedData = ref<any>(null)
const defaultYaml = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`
const yamlContent = ref(defaultYaml)
const submitting = ref(false)

const {
  cloneMode, cloneNamespace, cloneName, cloneNsOptions, cloneNameOptions,
  cloneNsLoading, cloneNameLoading, cloneLoading, cloneTarget,
  startClone, cancelClone, handleLoadClone,
} = useCloneCreate({
  api: { list: getPvcList, yaml: getPvcYaml },
  onCloneToForm: (parsed) => { parsedData.value = parsed; yamlContent.value = defaultYaml; mode.value = 'form' },
  onCloneToYaml: (parsed) => { yamlContent.value = dumpCloneYaml(parsed); parsedData.value = null; mode.value = 'yaml' },
})

async function handleYamlSubmit() {
  if (!yamlContent.value.trim()) {
    ElMessage.error('YAML内容不能为空')
    return
  }
  submitting.value = true
  try {
    const parsed = yaml.load(yamlContent.value) as any
    const ns = parsed?.metadata?.namespace || 'default'
    await createPvc({ namespace: ns, yaml: yamlContent.value })
    ElMessage.success('PVC创建成功')
    router.push('/storage/pvcs')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/storage/pvcs')
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
  <div class="pvc-create">
    <div class="mode-switcher">
      <el-segmented v-model="mode" :options="[{ label: t('common.formCreate'), value: 'form' }, { label: t('common.yamlCreate'), value: 'yaml' }]" size="small" />
      <el-button size="small" style="margin-left: 12px;" @click="startClone">
        <el-icon><CopyDocument /></el-icon> 从现有资源克隆
      </el-button>
    </div>

    <CloneDialog
      kind-label="PVC"
      v-model="cloneMode"
      v-model:ns-value="cloneNamespace"
      v-model:name-value="cloneName"
      v-model:target="cloneTarget"
      :ns-options="cloneNsOptions"
      :name-options="cloneNameOptions"
      :ns-loading="cloneNsLoading"
      :name-loading="cloneNameLoading"
      :loading="cloneLoading"
      @load="handleLoadClone"
      @cancel="cancelClone"
    />

    <!-- PVCForm 始终挂载（v-show），避免克隆后切换 mode 时组件被销毁重建导致数据丢失 -->
    <PVCForm v-show="mode === 'form'" :initial-data="parsedData" />

    <div v-if="mode !== 'form'" class="yaml-mode">
      <div class="yaml-card">
        <div class="yaml-card-header">
          <div class="yaml-card-left">
            <span class="yaml-card-title">YAML 配置</span>
            <el-button-group>
              <el-button size="small" @click="handleFormat">格式化</el-button>
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
.pvc-create { max-width: 1100px; margin: 0 auto; padding: 20px 0; }
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
