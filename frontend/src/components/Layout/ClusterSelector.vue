<template>
  <el-select
    v-model="clusterStore.currentCluster"
    value-key="id"
    :placeholder="t('common.selectCluster')"
    :loading="loading"
    size="small"
    class="cluster-select"
    clearable
    @change="handleChange"
  >
    <template #prefix>
      <el-icon><Connection /></el-icon>
    </template>
    <el-option
      v-for="c in clusterStore.clusterList"
      :key="c.id"
      :label="c.clusterName"
      :value="c"
    />
  </el-select>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Connection } from '@element-plus/icons-vue'
import { useClusterStore } from '@/stores/cluster'

const { t } = useI18n()
const clusterStore = useClusterStore()
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    await clusterStore.fetchClusters()
  } finally {
    loading.value = false
  }
})

function handleChange(val: any) {
  clusterStore.setCurrentCluster(val || null)
}
</script>

<style scoped>
.cluster-select {
  min-width: 160px;
  max-width: 240px;
}

.cluster-select :deep(.el-input__wrapper) {
  border-radius: var(--gk-radius-md);
}
</style>
