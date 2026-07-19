import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useNamespaceStore } from '@/stores/namespace'

export interface ResourceListOptions {
  /** Resource display name (e.g. 'Deployment', 'Pod') */
  resourceName: string
  /** Fetch the resource list */
  fetchList: (params?: any) => Promise<any>
  /** Transform raw K8s items to display objects (optional) */
  transform?: (items: any[]) => any[]
  /** Get YAML for a resource */
  getYaml: (params: any) => Promise<any>
  /** Update YAML (optional, enables save button) */
  updateYaml?: (data: any) => Promise<any>
  /** Delete a single resource */
  deleteResource: (params: any) => Promise<any>
  /** Route path for detail view (e.g. '/workloads/deployments') */
  detailRoute?: string
  /** Route path for create view */
  createRoute?: string
  /** Custom confirm message for delete */
  deleteConfirm?: (row: any) => string
  /** Enable server-side pagination (default: false) */
  paginated?: boolean
  /** Page size for pagination (default: 50) */
  pageSize?: number
  /** Auto-refresh interval in ms (default: 0 = disabled) */
  autoRefreshInterval?: number
}

export function useResourceList(options: ResourceListOptions) {
  const router = useRouter()
  const namespaceStore = useNamespaceStore()

  const loading = ref(false)
  const list = ref<any[]>([])
  const selectedNamespace = ref('')
  const searchName = ref('')
  const debouncedSearch = ref('')
  const selectedRows = ref<any[]>([])

  // Auto-refresh state
  const autoRefreshEnabled = ref(false)
  let autoRefreshTimer: ReturnType<typeof setInterval> | null = null

  // Pagination state
  const currentPage = ref(1)
  const pageSize = ref(options.pageSize || 50)
  const continueTokens = ref<string[]>([])
  const hasMore = ref(false)
  const totalCount = ref(0)

  // YAML drawer state
  const yamlDialogVisible = ref(false)
  const yamlContent = ref('')
  const yamlLoading = ref(false)
  const yamlTarget = ref<any>(null)
  const yamlEditing = ref(false)
  const yamlSaving = ref(false)

  // Debounce search input
  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  function onSearchInput(value: string) {
    searchName.value = value
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchDebounceTimer = setTimeout(() => {
      debouncedSearch.value = value
    }, 200)
  }

  const filteredList = computed(() => {
    if (!debouncedSearch.value) return list.value
    const keyword = debouncedSearch.value.toLowerCase()
    return list.value.filter((item) => item.name?.toLowerCase().includes(keyword))
  })

  async function fetchNamespaces() {
    await namespaceStore.fetchNamespaces()
  }

  async function fetchResources() {
    loading.value = true
    try {
      const params: any = {}
      if (selectedNamespace.value) params.namespace = selectedNamespace.value

      if (options.paginated) {
        params.limit = pageSize.value
      }

      const res: any = await options.fetchList(params)

      if (options.paginated && res.data?.items) {
        const items = res.data.items || []
        list.value = options.transform ? options.transform(items) : items
        hasMore.value = res.data.hasMore || false
        totalCount.value = res.data.total || items.length

        if (res.data.continue) {
          continueTokens.value = [res.data.continue]
        } else {
          continueTokens.value = []
        }
      } else {
        const items = res.data?.items || res.data || []
        list.value = options.transform ? options.transform(items) : items
        totalCount.value = list.value.length
      }
    } catch (e) {
      // Resource type may legitimately not exist in the cluster; log rather than swallow silently
      console.error(`[useResourceList] Failed to fetch ${options.resourceName} list:`, e)
    } finally {
      loading.value = false
    }
  }

  async function fetchNextPage() {
    if (!hasMore.value || continueTokens.value.length === 0) return
    loading.value = true
    try {
      const params: any = {}
      if (selectedNamespace.value) params.namespace = selectedNamespace.value
      params.limit = pageSize.value
      params.continue = continueTokens.value[continueTokens.value.length - 1]

      const res: any = await options.fetchList(params)
      if (res.data?.items) {
        const items = res.data.items || []
        const transformed = options.transform ? options.transform(items) : items
        list.value = [...list.value, ...transformed]
        hasMore.value = res.data.hasMore || false
        currentPage.value++

        if (res.data.continue) {
          continueTokens.value.push(res.data.continue)
        }
      }
    } catch (e: any) {
      ElMessage.error(e?.message || 'Failed to load more items')
    } finally {
      loading.value = false
    }
  }

  function handleNamespaceChange() {
    currentPage.value = 1
    continueTokens.value = []
    fetchResources()
  }

  function handleSelectionChange(rows: any[]) {
    selectedRows.value = rows
  }

  async function handleViewYaml(row: any) {
    yamlTarget.value = row
    yamlDialogVisible.value = true
    yamlEditing.value = false
    yamlLoading.value = true
    yamlContent.value = ''
    try {
      const res: any = await options.getYaml({
        namespace: row.namespace,
        name: row.name,
      })
      yamlContent.value = res.data?.yaml || res.data || ''
    } catch (e: any) {
      ElMessage.error(e?.message || 'Failed to load YAML')
      yamlDialogVisible.value = false
    } finally {
      yamlLoading.value = false
    }
  }

  async function fetchYaml() {
    if (!yamlTarget.value) return
    yamlLoading.value = true
    try {
      const res: any = await options.getYaml({
        namespace: yamlTarget.value.namespace,
        name: yamlTarget.value.name,
      })
      yamlContent.value = res.data?.yaml || res.data || ''
    } catch (e: any) {
      ElMessage.error(e?.message || 'Failed to load YAML')
    } finally {
      yamlLoading.value = false
    }
  }

  function handleEditYaml() {
    yamlEditing.value = true
  }

  async function handleSaveYaml() {
    if (!yamlTarget.value || !options.updateYaml) return
    yamlSaving.value = true
    try {
      await options.updateYaml({
        namespace: yamlTarget.value.namespace,
        name: yamlTarget.value.name,
        yaml: yamlContent.value,
      })
      ElMessage.success('YAML saved successfully')
      yamlEditing.value = false
      fetchResources()
    } catch (e: any) {
      ElMessage.error(e?.message || 'Failed to save YAML')
    } finally {
      yamlSaving.value = false
    }
  }

  function handleCancelYaml() {
    yamlEditing.value = false
    fetchYaml()
  }

  function handleDetail(row: any) {
    if (options.detailRoute) {
      // Cluster-scoped resources have no namespace segment
      const path = row.namespace
        ? `${options.detailRoute}/${row.namespace}/${row.name}`
        : `${options.detailRoute}/${row.name}`
      router.push(path)
    }
  }

  async function handleDelete(row: any) {
    const msg = options.deleteConfirm
      ? options.deleteConfirm(row)
      : `Delete ${options.resourceName.toLowerCase()} "${row.name}" in namespace "${row.namespace}"?`
    try {
      await ElMessageBox.confirm(msg, 'Confirm', { type: 'warning' })
      await options.deleteResource({ namespace: row.namespace, name: row.name })
      ElMessage.success(`${options.resourceName} deleted`)
      fetchResources()
    } catch {
      // cancelled
    }
  }

  async function handleBatchDelete() {
    if (!selectedRows.value.length) return
    try {
      await ElMessageBox.confirm(
        `Delete ${selectedRows.value.length} selected ${options.resourceName.toLowerCase()}(s)?`,
        'Confirm',
        { type: 'warning' }
      )
      const results = await Promise.allSettled(
        selectedRows.value.map((row) =>
          options.deleteResource({ namespace: row.namespace, name: row.name })
        )
      )
      const successCount = results.filter((r) => r.status === 'fulfilled').length
      const failCount = results.filter((r) => r.status === 'rejected').length
      if (failCount > 0) {
        ElMessage.warning(`Deleted ${successCount}, failed ${failCount}`)
      } else {
        ElMessage.success(`Deleted ${successCount} ${options.resourceName.toLowerCase()}(s)`)
      }
      fetchResources()
    } catch {
      // cancelled
    }
  }

  // Auto-refresh
  function toggleAutoRefresh() {
    autoRefreshEnabled.value = !autoRefreshEnabled.value
    if (autoRefreshEnabled.value) {
      const interval = options.autoRefreshInterval || 30000
      autoRefreshTimer = setInterval(() => {
        fetchResources()
      }, interval)
      ElMessage.info(`Auto-refresh enabled (${interval / 1000}s)`)
    } else {
      if (autoRefreshTimer) {
        clearInterval(autoRefreshTimer)
        autoRefreshTimer = null
      }
      ElMessage.info('Auto-refresh disabled')
    }
  }

  // Keyboard shortcut: R to refresh
  function handleKeyboard(e: KeyboardEvent) {
    if (e.key === 'r' && !e.ctrlKey && !e.metaKey && !e.altKey) {
      const target = e.target as HTMLElement
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return
      e.preventDefault()
      fetchResources()
    }
  }

  onMounted(() => {
    fetchNamespaces()
    fetchResources()
    document.addEventListener('keydown', handleKeyboard)
  })

  onUnmounted(() => {
    if (autoRefreshTimer) clearInterval(autoRefreshTimer)
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    document.removeEventListener('keydown', handleKeyboard)
  })

  return {
    // State
    loading,
    list,
    filteredList,
    selectedNamespace,
    searchName,
    onSearchInput,
    selectedRows,
    // Auto-refresh
    autoRefreshEnabled,
    toggleAutoRefresh,
    // Pagination
    currentPage,
    pageSize,
    hasMore,
    totalCount,
    fetchNextPage,
    // Namespace
    namespaceList: computed(() => namespaceStore.namespaces),
    // YAML drawer
    yamlDialogVisible,
    yamlContent,
    yamlLoading,
    yamlTarget,
    yamlEditing,
    yamlSaving,
    // Methods
    fetchResources,
    handleNamespaceChange,
    handleSelectionChange,
    handleViewYaml,
    handleEditYaml,
    handleSaveYaml,
    handleCancelYaml,
    handleDetail,
    handleDelete,
    handleBatchDelete,
  }
}
