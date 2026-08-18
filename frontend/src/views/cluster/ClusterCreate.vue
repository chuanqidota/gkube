<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { createCluster, testKubeConfig } from '@/api/cluster'
import { useClusterStore } from '@/stores/cluster'
import type { FormInstance, FormRules, UploadFile } from 'element-plus'

const { t } = useI18n()
const router = useRouter()
const clusterStore = useClusterStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const testing = ref(false)

const form = reactive({
  clusterName: '',
  displayName: '',
  description: '',
  kubeConfig: '',
  labels: [] as Array<{ key: string; value: string }>,
})

const rules: FormRules = {
  clusterName: [{ required: true, message: t('cluster.name'), trigger: 'blur' }],
  kubeConfig: [{ required: true, message: 'KubeConfig', trigger: 'blur' }],
}

function addLabel() {
  form.labels.push({ key: '', value: '' })
}

function removeLabel(index: number) {
  form.labels.splice(index, 1)
}

function handleFileUpload(uploadFile: UploadFile) {
  if (!uploadFile.raw) return
  const reader = new FileReader()
  reader.onload = (e) => {
    const content = e.target?.result as string
    if (content) {
      form.kubeConfig = content
      ElMessage.success(t('cluster.fileUploaded'))
    }
  }
  reader.readAsText(uploadFile.raw)
}

async function handleTestConnection() {
  if (!form.kubeConfig.trim()) {
    ElMessage.warning(t('cluster.kubeConfigRequired'))
    return
  }
  testing.value = true
  try {
    const res: any = await testKubeConfig(form.kubeConfig)
    const info = res.data
    ElMessage.success(
      t('cluster.testSuccess', {
        version: info.clusterVersion,
        nodeCount: info.nodeCount,
        responseTimeMs: info.responseTimeMs,
      })
    )
  } catch (e: any) {
    ElMessage.error(e?.message || t('cluster.testFailed'))
  } finally {
    testing.value = false
  }
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const labels: Record<string, string> = {}
    form.labels.forEach((l) => {
      if (l.key.trim()) labels[l.key.trim()] = l.value
    })

    await createCluster({
      clusterName: form.clusterName,
      displayName: form.displayName,
      description: form.description,
      kubeConfig: form.kubeConfig,
      labels,
    })
    ElMessage.success(t('cluster.clusterCreated'))
    clusterStore.fetchClusters()
    router.push('/clusters')
  } catch (e: any) {
    ElMessage.error(e?.message || t('cluster.createFailed'))
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <h3 style="margin: 0;">{{ t('cluster.addCluster') }}</h3>
          <el-button @click="router.push('/clusters')">{{ t('common.back') }}</el-button>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        style="max-width: 700px;"
      >
        <el-form-item :label="t('cluster.name')" prop="clusterName">
          <el-input v-model="form.clusterName" :placeholder="t('cluster.clusterNamePlaceholder')" />
          <div class="form-tip">{{ t('cluster.clusterNameTip') }}</div>
        </el-form-item>

        <el-form-item :label="t('cluster.displayName')" prop="displayName">
          <el-input v-model="form.displayName" :placeholder="t('cluster.displayNamePlaceholder')" />
        </el-form-item>

        <el-form-item :label="t('cluster.description')" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" :placeholder="t('cluster.descriptionPlaceholder')" />
        </el-form-item>

        <el-form-item label="KubeConfig" prop="kubeConfig">
          <div style="width: 100%; display: flex; align-items: flex-start; gap: 12px;">
            <el-upload
              :auto-upload="false"
              :show-file-list="false"
              :on-change="handleFileUpload"
            >
              <el-button type="primary" plain>{{ t('cluster.uploadKubeConfig') }}</el-button>
            </el-upload>
            <el-button
              type="success"
              plain
              :loading="testing"
              :disabled="!form.kubeConfig.trim()"
              @click="handleTestConnection"
            >{{ t('cluster.testConnection') }}</el-button>
          </div>
          <el-input
            v-model="form.kubeConfig"
            type="textarea"
            :rows="12"
            :placeholder="t('cluster.kubeConfigPlaceholder')"
            style="margin-top: 8px;"
          />
          <div class="form-tip">{{ t('cluster.kubeConfigTip') }}</div>
        </el-form-item>

        <el-form-item :label="t('cluster.labels')">
          <div style="width: 100%;">
            <div
              v-for="(label, index) in form.labels"
              :key="index"
              style="display: flex; gap: 8px; margin-bottom: 8px;"
            >
              <el-input v-model="label.key" :placeholder="t('cluster.keyPlaceholder')" style="flex: 1;" />
              <el-input v-model="label.value" :placeholder="t('cluster.valuePlaceholder')" style="flex: 1;" />
              <el-button type="danger" circle @click="removeLabel(index)">-</el-button>
            </div>
            <el-button @click="addLabel" type="primary" plain>{{ t('cluster.addLabel') }}</el-button>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSubmit">{{ t('cluster.createCluster') }}</el-button>
          <el-button @click="router.push('/clusters')">{{ t('common.cancel') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.form-tip { font-size: 12px; color: var(--gk-color-text-secondary); margin-top: 4px; }
</style>
