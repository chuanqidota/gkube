<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
import { ElMessage } from 'element-plus'
import { createStorageClass } from '@/api/resource'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)

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

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const parameters: Record<string, string> = {}
    form.parameters.forEach((p) => { if (p.key.trim()) parameters[p.key.trim()] = p.value })
    const labels: Record<string, string> = {}
    form.labels.forEach((l) => { if (l.key.trim()) labels[l.key.trim()] = l.value })

    await createStorageClass({
      name: form.name,
      provisioner: form.provisioner,
      reclaimPolicy: form.reclaimPolicy,
      volumeBindingMode: form.volumeBindingMode,
      parameters,
      labels,
    })
    ElMessage.success('StorageClass 创建成功')
    router.push('/storage/storageclasses')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally { loading.value = false }
}
</script>

<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <h3 style="margin: 0;">创建 StorageClass</h3>
          <el-button @click="router.push('/storage/storageclasses')">返回</el-button>
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
          <el-button type="primary" :loading="loading" @click="handleSubmit">创建</el-button>
          <el-button @click="router.push('/storage/storageclasses')">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
