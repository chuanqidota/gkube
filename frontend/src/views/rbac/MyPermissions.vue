<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getMyPermissions } from '@/api/rbac'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const isSuperAdmin = ref(false)
const bindings = ref<any[]>([])

interface ClusterEntry {
  clusterId: number
  items: { roleName: string; namespace: string }[]
}

const byCluster = ref<ClusterEntry[]>([])

async function fetchData() {
  loading.value = true
  try {
    const res: any = await getMyPermissions()
    const data = res?.data ?? res
    isSuperAdmin.value = !!data?.isSuperAdmin
    bindings.value = data?.bindings || []
    const map = new Map<number, ClusterEntry>()
    for (const b of bindings.value) {
      let entry = map.get(b.clusterId)
      if (!entry) {
        entry = { clusterId: b.clusterId, items: [] }
        map.set(b.clusterId, entry)
      }
      entry.items.push({
        roleName: b.roleName,
        namespace: b.namespace,
      })
    }
    byCluster.value = [...map.values()]
  } catch (e: any) {
    ElMessage.error(e?.message || t('rbac.loadFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <el-card shadow="never" class="table-card">
      <template #header>
        <span>{{ t('rbac.myPermissions') }} — {{ authStore.user?.username }}</span>
      </template>

      <el-alert v-if="isSuperAdmin" type="success" :closable="false" style="margin-bottom: 16px;">
        {{ t('rbac.superAdminBypass') }}
      </el-alert>

      <div v-for="c in byCluster" :key="c.clusterId" class="cluster-block">
        <div class="cluster-title">{{ t('rbac.clusterLabel') }} #{{ c.clusterId }}</div>
        <div v-for="(item, i) in c.items" :key="i" class="perm-row">
          <el-tag size="small">{{ item.roleName }}</el-tag>
          <el-tag v-if="item.namespace" size="small" type="warning">{{ item.namespace }}</el-tag>
          <el-tag v-else size="small" type="info">{{ t('rbac.clusterScope') }}</el-tag>
        </div>
      </div>
      <el-empty v-if="!isSuperAdmin && byCluster.length === 0" :description="t('rbac.noPermissions')" />
    </el-card>
  </div>
</template>

<style scoped>
.cluster-block { margin-bottom: 16px; }
.cluster-title { font-weight: 600; margin-bottom: 8px; }
.perm-row { display: flex; gap: 8px; margin-bottom: 6px; }
</style>
