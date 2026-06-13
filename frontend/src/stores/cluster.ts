import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { getClusterList } from '@/api/cluster'

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

  const clusterList = ref<any[]>([])
  const currentCluster = ref<any>(savedCluster)

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
    // Auto-select first cluster if none selected
    if (!currentCluster.value && clusterList.value.length > 0) {
      currentCluster.value = clusterList.value[0]
    }
  }

  function setCurrentCluster(cluster: any) {
    currentCluster.value = cluster
  }

  return { clusterList, currentCluster, fetchClusters, setCurrentCluster }
})
