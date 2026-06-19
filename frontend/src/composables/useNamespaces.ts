import { ref } from 'vue'
import { getNamespaceList, extractNamespaceNames } from '@/api/resource'

/**
 * Shared composable for fetching and managing namespace lists.
 * Eliminates the 20+ duplicate fetchNamespaces() implementations across views.
 */
export function useNamespaces() {
  const namespaceList = ref<string[]>([])
  const namespaceLoading = ref(false)

  async function fetchNamespaces() {
    namespaceLoading.value = true
    try {
      const res: any = await getNamespaceList()
      namespaceList.value = extractNamespaceNames(res.data)
    } catch (e: any) {
      console.error('Failed to fetch namespaces:', e?.message)
    } finally {
      namespaceLoading.value = false
    }
  }

  return {
    namespaceList,
    namespaceLoading,
    fetchNamespaces,
  }
}
