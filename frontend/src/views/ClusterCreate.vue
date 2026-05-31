<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createCluster } from '@/api/cluster'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  clusterName: '',
  displayName: '',
  description: '',
  kubeConfig: '',
  labels: [] as Array<{ key: string; value: string }>,
})

const rules: FormRules = {
  clusterName: [{ required: true, message: 'Cluster name is required', trigger: 'blur' }],
  kubeConfig: [{ required: true, message: 'KubeConfig is required', trigger: 'blur' }],
}

function addLabel() {
  form.labels.push({ key: '', value: '' })
}

function removeLabel(index: number) {
  form.labels.splice(index, 1)
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
    ElMessage.success('Cluster created')
    router.push('/clusters')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h2>Add Cluster</h2>
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="140px"
      style="max-width: 700px;"
    >
      <el-form-item label="Cluster Name" prop="clusterName">
        <el-input v-model="form.clusterName" placeholder="e.g. prod-cluster-01" />
      </el-form-item>

      <el-form-item label="Display Name" prop="displayName">
        <el-input v-model="form.displayName" placeholder="e.g. Production Cluster" />
      </el-form-item>

      <el-form-item label="Description" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="3" placeholder="Optional description" />
      </el-form-item>

      <el-form-item label="KubeConfig" prop="kubeConfig">
        <el-input
          v-model="form.kubeConfig"
          type="textarea"
          :rows="10"
          placeholder="Paste your kubeconfig YAML here"
        />
      </el-form-item>

      <el-form-item label="Labels">
        <div style="width: 100%;">
          <div
            v-for="(label, index) in form.labels"
            :key="index"
            style="display: flex; gap: 8px; margin-bottom: 8px;"
          >
            <el-input v-model="label.key" placeholder="Key" style="flex: 1;" />
            <el-input v-model="label.value" placeholder="Value" style="flex: 1;" />
            <el-button type="danger" :icon="'Delete'" circle @click="removeLabel(index)" />
          </div>
          <el-button @click="addLabel">Add Label</el-button>
        </div>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="loading" @click="handleSubmit">Create</el-button>
        <el-button @click="router.push('/clusters')">Cancel</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>
