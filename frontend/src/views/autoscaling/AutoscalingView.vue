<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const activeTab = computed(() => (route.path.startsWith('/autoscaling/vpa') ? 'vpa' : 'hpa'))

function handleTabChange(tab: string | number) {
  router.push(tab === 'vpa' ? '/autoscaling/vpa' : '/autoscaling/hpa')
}
</script>

<template>
  <div class="autoscaling-page">
    <el-card shadow="never" class="autoscaling-tabs-card">
      <el-tabs :model-value="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="HPA" name="hpa" />
        <el-tab-pane label="VPA" name="vpa" />
      </el-tabs>
    </el-card>

    <router-view />
  </div>
</template>

<style scoped>
.autoscaling-page {
  min-height: 100%;
}

.autoscaling-tabs-card {
  margin: 20px 20px 0;
  border-radius: 8px;
}

.autoscaling-tabs-card :deep(.el-card__body) {
  padding: 0 16px;
}

.autoscaling-tabs-card :deep(.el-tabs__header) {
  margin: 0;
}
</style>
