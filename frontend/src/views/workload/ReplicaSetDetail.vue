<template>
  <div class="page-container">
    <el-page-header @back="$router.back()" :content="'ReplicaSet: ' + ($route.query.name || '')" />

    <el-card v-if="detail" class="info-card" shadow="never">
      <template #header>基本信息</template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="名称">{{ detail.rs?.metadata?.name }}</el-descriptions-item>
        <el-descriptions-item label="命名空间">{{ detail.rs?.metadata?.namespace }}</el-descriptions-item>
        <el-descriptions-item label="期望副本数">{{ detail.rs?.spec?.replicas }}</el-descriptions-item>
        <el-descriptions-item label="当前副本数">{{ detail.rs?.status?.replicas }}</el-descriptions-item>
        <el-descriptions-item label="就绪副本数">{{ detail.rs?.status?.readyReplicas }}</el-descriptions-item>
        <el-descriptions-item label="可用副本数">{{ detail.rs?.status?.availableReplicas }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card v-if="detail?.controllerOf" class="info-card" shadow="never">
      <template #header>由以下控制器管理</template>
      <el-tag type="primary">
        {{ detail.controllerOf.kind }}: {{ detail.controllerOf.name }}
      </el-tag>
    </el-card>

    <el-card v-if="detail?.pods?.length" class="info-card" shadow="never">
      <template #header>关联 Pod ({{ detail.pods.length }})</template>
      <el-table :data="detail.pods" stripe>
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="status" label="状态" />
        <el-table-column prop="ready" label="就绪" />
        <el-table-column prop="node" label="节点" />
        <el-table-column prop="restarts" label="重启次数" />
      </el-table>
    </el-card>

    <el-card v-else-if="detail" class="info-card" shadow="never">
      <el-empty description="没有关联的 Pod" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getReplicaSetDetail } from '@/api/resource'
import { ElMessage } from 'element-plus'

const route = useRoute()
const detail = ref<any>(null)

onMounted(async () => {
  try {
    const namespace = route.query.namespace as string
    const name = route.query.name as string
    if (!namespace || !name) {
      ElMessage.error('缺少参数')
      return
    }
    detail.value = await getReplicaSetDetail({ namespace, name })
  } catch (e: any) {
    ElMessage.error(e?.message || '获取详情失败')
  }
})
</script>

<style scoped>
.page-container { padding: 16px; }
.info-card { margin-top: 16px; }
.info-card:first-of-type { margin-top: 20px; }
</style>