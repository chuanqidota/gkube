import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getNamespaceList, extractNamespaceNames } from '@/api/resource'

export const useNamespaceStore = defineStore('namespace', () => {
  const namespaces = ref<string[]>([])
  const loading = ref(false)
  const lastFetched = ref(0)
  const CACHE_TTL = 30_000 // 30 seconds cache

  async function fetchNamespaces(force = false) {
    const now = Date.now()
    if (!force && namespaces.value.length > 0 && now - lastFetched.value < CACHE_TTL) {
      return namespaces.value
    }
    loading.value = true
    try {
      const res: any = await getNamespaceList()
      namespaces.value = extractNamespaceNames(res.data)
      lastFetched.value = now
      return namespaces.value
    } catch {
      return namespaces.value
    } finally {
      loading.value = false
    }
  }

  function clearCache() {
    namespaces.value = []
    lastFetched.value = 0
  }

  return { namespaces, loading, fetchNamespaces, clearCache }
})
