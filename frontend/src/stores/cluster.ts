import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getClusterList } from '@/api/cluster'

export const useClusterStore = defineStore('cluster', () => {
  const clusterList = ref<any[]>([])
  const currentCluster = ref<any>(null)

  async function fetchClusters() {
    const res: any = await getClusterList({ page: 1, size: 100 })
    clusterList.value = res.data.items || []
  }

  function setCurrentCluster(cluster: any) {
    currentCluster.value = cluster
  }

  return { clusterList, currentCluster, fetchClusters, setCurrentCluster }
})
