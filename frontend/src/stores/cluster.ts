import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { getClusterList } from '@/api/cluster'
import { useNamespaceStore } from '@/stores/namespace'

export interface Cluster {
  id?: number | string
  clusterName?: string
  cluster_name?: string
  name?: string
  display_name?: string
  [key: string]: unknown
}

export const useClusterStore = defineStore('cluster', () => {
  // Load from localStorage on init
  const savedCluster = (() => {
    try {
      const saved = localStorage.getItem('gkube_cluster')
      return saved ? JSON.parse(saved) : null
    } catch {
      return null
    }
  })()

  const clusterList = ref<Cluster[]>([])
  const currentCluster = ref<Cluster | null>(savedCluster)

  // Persist to localStorage on change
  watch(currentCluster, (val) => {
    if (val) {
      localStorage.setItem('gkube_cluster', JSON.stringify(val))
    } else {
      localStorage.removeItem('gkube_cluster')
    }
  }, { deep: true })

  async function fetchClusters() {
    const res: any = await getClusterList({ page: 1, size: 100 })
    clusterList.value = res.data.items || []

    // 验证当前选中集群是否仍在列表中
    if (currentCluster.value) {
      const stillExists = clusterList.value.some(
        (c: Cluster) => c.id === currentCluster.value?.id
      )
      if (!stillExists) {
        currentCluster.value = null
      }
    }

    // Auto-select first cluster if none selected
    if (!currentCluster.value && clusterList.value.length > 0) {
      setCurrentCluster(clusterList.value[0])
    }
  }

  function setCurrentCluster(cluster: Cluster | null) {
    if (cluster === currentCluster.value) return
    currentCluster.value = cluster
    // 切换集群后清除命名空间缓存，避免跨集群复用旧 ns
    try {
      useNamespaceStore().clearCache()
    } catch {
      // namespace store 尚未注册时忽略（惰性初始化场景）
    }
  }

  return { clusterList, currentCluster, fetchClusters, setCurrentCluster }
})
