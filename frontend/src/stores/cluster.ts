import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getClusterList } from '@/api/cluster'

export const useClusterStore = defineStore('cluster', () => {
  const clusterList = ref<any[]>([])
  const currentCluster = ref<any>(null)

  async function fetchClusters() {
    const res: any = await getClusterList()
    clusterList.value = res.data || []
  }

  function setCurrentCluster(cluster: any) {
    currentCluster.value = cluster
  }

  return { clusterList, currentCluster, fetchClusters, setCurrentCluster }
})
