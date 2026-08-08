import { useClusterStore } from '@/stores/cluster'
import { computed } from 'vue'

/**
 * Returns a reactive computed for the current cluster name,
 * compatible with different field shapes (clusterName / cluster_name / name).
 * Use this in detail pages where the cluster may change while the page is open.
 */
export function useClusterNameRef() {
  const clusterStore = useClusterStore()
  return computed(() => clusterStore.clusterName)
}
