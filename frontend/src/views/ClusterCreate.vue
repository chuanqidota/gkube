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
  clusterName: [{ required: true, message: '请输入集群名称', trigger: 'blur' }],
  kubeConfig: [{ required: true, message: '请输入 KubeConfig', trigger: 'blur' }],
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
    ElMessage.success('集群创建成功')
    router.push('/clusters')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
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
          <h3 style="margin: 0;">添加集群</h3>
          <el-button @click="router.push('/clusters')">返回</el-button>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        style="max-width: 700px;"
      >
        <el-form-item label="集群名称" prop="clusterName">
          <el-input v-model="form.clusterName" placeholder="例如: prod-cluster-01" />
          <div class="form-tip">集群唯一标识，创建后不可修改</div>
        </el-form-item>

        <el-form-item label="显示名称" prop="displayName">
          <el-input v-model="form.displayName" placeholder="例如: 生产集群" />
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="可选描述信息" />
        </el-form-item>

        <el-form-item label="KubeConfig" prop="kubeConfig">
          <el-input
            v-model="form.kubeConfig"
            type="textarea"
            :rows="12"
            placeholder="粘贴您的 kubeconfig YAML 内容"
          />
          <div class="form-tip">系统将验证 kubeconfig 的连通性</div>
        </el-form-item>

        <el-form-item label="标签">
          <div style="width: 100%;">
            <div
              v-for="(label, index) in form.labels"
              :key="index"
              style="display: flex; gap: 8px; margin-bottom: 8px;"
            >
              <el-input v-model="label.key" placeholder="键" style="flex: 1;" />
              <el-input v-model="label.value" placeholder="值" style="flex: 1;" />
              <el-button type="danger" circle @click="removeLabel(index)">-</el-button>
            </div>
            <el-button @click="addLabel" type="primary" plain>添加标签</el-button>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSubmit">创建集群</el-button>
          <el-button @click="router.push('/clusters')">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.form-tip { font-size: 12px; color: #909399; margin-top: 4px; }
</style>
