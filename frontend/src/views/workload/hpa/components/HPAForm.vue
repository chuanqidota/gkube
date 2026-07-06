<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createHpa, getNamespaceList, extractNamespaceNames } from '@/api/resource'

const router = useRouter()
const loading = ref(false)
const namespaceList = ref<string[]>([])

const form = ref({
  name: '',
  namespace: 'default',
  targetKind: 'Deployment',
  targetName: '',
  minReplicas: 1,
  maxReplicas: 10,
  cpuUtilization: 80,
  memoryUtilization: 0,
  customMetrics: [] as Array<{ type: string; name: string; target: number }>,
})

const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  targetName: [{ required: true, message: '请输入目标名称', trigger: 'blur' }],
  minReplicas: [{ required: true, message: '请输入最小副本数', trigger: 'blur' }],
  maxReplicas: [{ required: true, message: '请输入最大副本数', trigger: 'blur' }],
}

const formRef = ref()

function addCustomMetric() {
  form.value.customMetrics.push({ type: 'Resource', name: 'cpu', target: 80 })
}

function removeCustomMetric(index: number) {
  form.value.customMetrics.splice(index, 1)
}

function buildYaml(): string {
  const metrics: any[] = []

  // CPU metric
  if (form.value.cpuUtilization > 0) {
    metrics.push({
      type: 'Resource',
      resource: {
        name: 'cpu',
        target: {
          type: 'Utilization',
          averageUtilization: form.value.cpuUtilization,
        },
      },
    })
  }

  // Memory metric
  if (form.value.memoryUtilization > 0) {
    metrics.push({
      type: 'Resource',
      resource: {
        name: 'memory',
        target: {
          type: 'Utilization',
          averageUtilization: form.value.memoryUtilization,
        },
      },
    })
  }

  // Custom metrics
  for (const m of form.value.customMetrics) {
    metrics.push({
      type: m.type,
      resource: {
        name: m.name,
        target: {
          type: 'Utilization',
          averageUtilization: m.target,
        },
      },
    })
  }

  const hpa: any = {
    apiVersion: 'autoscaling/v2',
    kind: 'HorizontalPodAutoscaler',
    metadata: {
      name: form.value.name,
      namespace: form.value.namespace,
    },
    spec: {
      scaleTargetRef: {
        apiVersion: 'apps/v1',
        kind: form.value.targetKind,
        name: form.value.targetName,
      },
      minReplicas: form.value.minReplicas,
      maxReplicas: form.value.maxReplicas,
      metrics,
    },
  }

  return JSON.stringify(hpa, null, 2)
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const yaml = buildYaml()
    await createHpa({ namespace: form.value.namespace, yaml })
    ElMessage.success('弹性伸缩创建成功')
    router.push('/workloads/hpa')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  router.push('/workloads/hpa')
}

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

onMounted(fetchNamespaces)
</script>

<template>
  <div class="hpa-form">
    <div class="form-header">
      <h2>创建弹性伸缩 (HPA)</h2>
    </div>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="140px"
      style="max-width: 700px;"
    >
      <el-divider content-position="left">基本信息</el-divider>

      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入HPA名称" />
      </el-form-item>

      <el-form-item label="命名空间" prop="namespace">
        <el-select v-model="form.namespace" placeholder="请选择命名空间" style="width: 100%;">
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
      </el-form-item>

      <el-divider content-position="left">伸缩目标</el-divider>

      <el-form-item label="目标类型" prop="targetKind">
        <el-select v-model="form.targetKind" style="width: 100%;">
          <el-option label="Deployment" value="Deployment" />
          <el-option label="StatefulSet" value="StatefulSet" />
          <el-option label="ReplicaSet" value="ReplicaSet" />
        </el-select>
      </el-form-item>

      <el-form-item label="目标名称" prop="targetName">
        <el-input v-model="form.targetName" placeholder="请输入目标工作负载名称" />
      </el-form-item>

      <el-divider content-position="left">副本范围</el-divider>

      <el-form-item label="最小副本数" prop="minReplicas">
        <el-input-number v-model="form.minReplicas" :min="1" :max="form.maxReplicas" />
      </el-form-item>

      <el-form-item label="最大副本数" prop="maxReplicas">
        <el-input-number v-model="form.maxReplicas" :min="form.minReplicas" :max="1000" />
      </el-form-item>

      <el-divider content-position="left">指标配置</el-divider>

      <el-form-item label="CPU 目标 (%)">
        <el-input-number v-model="form.cpuUtilization" :min="0" :max="100" :step="5" />
        <span class="form-tip">设为 0 表示不使用 CPU 指标</span>
      </el-form-item>

      <el-form-item label="内存目标 (%)">
        <el-input-number v-model="form.memoryUtilization" :min="0" :max="100" :step="5" />
        <span class="form-tip">设为 0 表示不使用内存指标</span>
      </el-form-item>

      <el-form-item label="自定义指标">
        <div v-for="(metric, index) in form.customMetrics" :key="index" class="custom-metric-row">
          <el-select v-model="metric.type" style="width: 120px;">
            <el-option label="Resource" value="Resource" />
            <el-option label="Pods" value="Pods" />
            <el-option label="Object" value="Object" />
            <el-option label="External" value="External" />
          </el-select>
          <el-input v-model="metric.name" placeholder="指标名称" style="width: 150px;" />
          <el-input-number v-model="metric.target" :min="1" :max="10000" style="width: 140px;" />
          <el-button type="danger" link @click="removeCustomMetric(index)">删除</el-button>
        </div>
        <el-button type="primary" link @click="addCustomMetric">+ 添加自定义指标</el-button>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="loading" @click="handleSubmit">创建</el-button>
        <el-button @click="handleCancel">取消</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<style scoped>
.hpa-form {
  padding: 0 16px;
}
.form-header {
  margin-bottom: 24px;
}
.form-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
.form-tip {
  margin-left: 12px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.custom-metric-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
</style>
