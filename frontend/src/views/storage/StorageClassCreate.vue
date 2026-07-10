<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { FullScreen } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import YamlEditor from '@/components/YamlEditor.vue'
import { createStorageClass } from '@/api/resource'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const mode = ref<'form' | 'yaml'>('form')
const yamlEditorRef = ref()
const submitting = ref(false)

// Form mode state
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  provisioner: '',
  reclaimPolicy: 'Delete',
  volumeBindingMode: 'Immediate',
  parameters: [] as Array<{ key: string; value: string }>,
  labels: [] as Array<{ key: string; value: string }>,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  provisioner: [{ required: true, message: '请输入 Provisioner', trigger: 'blur' }],
}

function addParam() { form.parameters.push({ key: '', value: '' }) }
function removeParam(i: number) { form.parameters.splice(i, 1) }
function addLabel() { form.labels.push({ key: '', value: '' }) }
function removeLabel(i: number) { form.labels.splice(i, 1) }

// YAML mode state
const yamlContent = ref(`apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: my-storage-class
provisioner: kubernetes.io/aws-ebs
reclaimPolicy: Delete
volumeBindingMode: Immediate
parameters:
  type: gp3
`)

function buildYamlFromForm(): string {
  const parameters: Record<string, string> = {}
  form.parameters.forEach((p) => { if (p.key.trim()) parameters[p.key.trim()] = p.value })
  const labels: Record<string, string> = {}
  form.labels.forEach((l) => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const obj: any = {
    apiVersion: 'storage.k8s.io/v1',
    kind: 'StorageClass',
    metadata: {
      name: form.name,
    },
    provisioner: form.provisioner,
    reclaimPolicy: form.reclaimPolicy,
    volumeBindingMode: form.volumeBindingMode,
  }
  if (Object.keys(parameters).length > 0) obj.parameters = parameters
  if (Object.keys(labels).length > 0) obj.metadata.labels = labels
  return yaml.dump(obj, { indent: 2 })
}

async function handleFormSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const yamlStr = buildYamlFromForm()
    await createStorageClass({ yaml: yamlStr })
    ElMessage.success('StorageClass 创建成功')
    router.push('/storage/storageclasses')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    submitting.value = false
  }
}

async function handleYamlSubmit() {
  if (!yamlContent.value.trim()) {
    ElMessage.error('YAML 内容不能为空')
    return
  }
  submitting.value = true
  try {
    await createStorageClass({ yaml: yamlContent.value })
    ElMessage.success('StorageClass 创建成功')
    router.push('/storage/storageclasses')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/storage/storageclasses')
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
  <div class="sc-create">
    <div class="mode-switcher">
      <el-segmented v-model="mode" :options="[{ label: '表单创建', value: 'form' }, { label: 'YAML 创建', value: 'yaml' }]" size="small" />
    </div>

    <!-- Form Mode -->
    <div v-if="mode === 'form'" class="form-mode">
      <el-card shadow="never">
        <template #header>
          <div class="card-header">
            <h3 style="margin: 0;">创建 StorageClass</h3>
            <el-button @click="handleCancel">返回</el-button>
          </div>
        </template>

        <el-form ref="formRef" :model="form" :rules="rules" label-width="160px" style="max-width: 700px;">
          <el-form-item label="名称" prop="name">
            <el-input v-model="form.name" placeholder="例如: fast-ssd" />
          </el-form-item>

          <el-form-item label="Provisioner" prop="provisioner">
            <el-input v-model="form.provisioner" placeholder="例如: kubernetes.io/aws-ebs, kubernetes.io/gce-pd" />
          </el-form-item>

          <el-form-item label="回收策略">
            <el-radio-group v-model="form.reclaimPolicy">
              <el-radio value="Delete">Delete</el-radio>
              <el-radio value="Retain">Retain</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="卷绑定模式">
            <el-radio-group v-model="form.volumeBindingMode">
              <el-radio value="Immediate">Immediate</el-radio>
              <el-radio value="WaitForFirstConsumer">WaitForFirstConsumer</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="参数">
            <div style="width: 100%;">
              <div v-for="(p, i) in form.parameters" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
                <el-input v-model="p.key" placeholder="键" style="flex: 1;" />
                <el-input v-model="p.value" placeholder="值" style="flex: 1;" />
                <el-button type="danger" circle @click="removeParam(i)">-</el-button>
              </div>
              <el-button @click="addParam" type="primary" plain>添加参数</el-button>
            </div>
          </el-form-item>

          <el-form-item label="标签">
            <div style="width: 100%;">
              <div v-for="(l, i) in form.labels" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
                <el-input v-model="l.key" placeholder="键" style="flex: 1;" />
                <el-input v-model="l.value" placeholder="值" style="flex: 1;" />
                <el-button type="danger" circle @click="removeLabel(i)">-</el-button>
              </div>
              <el-button @click="addLabel" type="primary" plain>添加标签</el-button>
            </div>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" :loading="submitting" @click="handleFormSubmit">创建</el-button>
            <el-button @click="handleCancel">取消</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>

    <!-- YAML Mode -->
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
.sc-create { max-width: 1100px; margin: 0 auto; padding: 20px 0; }
.mode-switcher { display: flex; justify-content: center; margin-bottom: 12px; }
.form-mode { padding: 0 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
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
