import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useNamespaceStore } from '@/stores/namespace'
import type { LabelCondition } from '@/components/LabelFilterPopover.vue'

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
  /** Force-delete a single resource (optional; if set, handleDelete accepts force flag) */
  forceDeleteResource?: (params: any) => Promise<any>
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

  // Label selector state
  const labelConditions = ref<LabelCondition[]>([])

  // Auto-refresh state
  const autoRefreshEnabled = ref(false)
  let autoRefreshTimer: ReturnType<typeof setInterval> | null = null

  // Pagination state
  const currentPage = ref(1)
  const pageSize = ref(options.pageSize || 50)
  const continueTokens = ref<string[]>([])
  const hasMore = ref(false)
  const totalCount = ref(0)

  // Pending delete map — prevents auto-refresh from reverting optimistic removal
  // Using Record<string, boolean> instead of Set for Vue 3 reactivity tracking
  const pendingDeleteIds = ref<Record<string, boolean>>({})
  // Track setTimeout IDs for cleanup on unmount
  const pendingDeleteTimers: ReturnType<typeof setTimeout>[] = []

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

  function resourceKey(row: { namespace?: string; name: string }) {
    return `${row.namespace ?? ''}/${row.name}`
  }

  function isPendingDelete(id: string) {
    return !!pendingDeleteIds.value[id]
  }

  function markPendingDelete(ids: string[]) {
    const updated = { ...pendingDeleteIds.value }
    ids.forEach((id) => { updated[id] = true })
    pendingDeleteIds.value = updated
  }

  function clearPendingDelete(ids: string[]) {
    const updated = { ...pendingDeleteIds.value }
    ids.forEach((id) => { delete updated[id] })
    pendingDeleteIds.value = updated
  }

  function scheduleCleanup(ids: string[], delay = 5000) {
    const timer = setTimeout(() => clearPendingDelete(ids), delay)
    pendingDeleteTimers.push(timer)
  }

  function filterPendingItems<T extends { namespace?: string; name: string }>(items: T[]): T[] {
    const keys = Object.keys(pendingDeleteIds.value)
    if (keys.length === 0) return items
    return items.filter((item) => !isPendingDelete(resourceKey(item)))
  }

  async function fetchNamespaces() {
    await namespaceStore.fetchNamespaces()
  }

  async function fetchResources() {
    loading.value = true
    try {
      const body: any = {}
      if (selectedNamespace.value) body.namespace = selectedNamespace.value
      if (labelConditions.value.length > 0) {
        body.labelFilters = labelConditions.value
      }

      if (options.paginated) {
        body.limit = pageSize.value
      }

      const res: any = await options.fetchList(body)

      if (options.paginated && res.data?.items) {
        const items = res.data.items || []
        const transformed = options.transform ? options.transform(items) : items
        list.value = filterPendingItems(transformed)
        hasMore.value = res.data.hasMore || false
        totalCount.value = res.data.total || items.length

        if (res.data.continue) {
          continueTokens.value = [res.data.continue]
        } else {
          continueTokens.value = []
        }
      } else {
        const items = res.data?.items || res.data || []
        const transformed = options.transform ? options.transform(items) : items
        list.value = filterPendingItems(transformed)
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
      const body: any = {}
      if (selectedNamespace.value) body.namespace = selectedNamespace.value
      if (labelConditions.value.length > 0) {
        body.labelFilters = labelConditions.value
      }
      body.limit = pageSize.value
      body.continue = continueTokens.value[continueTokens.value.length - 1]

      const res: any = await options.fetchList(body)
      if (res.data?.items) {
        const items = res.data.items || []
        const transformed = options.transform ? options.transform(items) : items
        const filtered = filterPendingItems(transformed)
        list.value = [...list.value, ...filtered]
        hasMore.value = res.data.hasMore || false
        currentPage.value++

        if (res.data.continue) {
          continueTokens.value.push(res.data.continue)
        }
      }
    } catch (e: any) {
      ElMessage.error(e?.message || '加载更多失败')
    } finally {
      loading.value = false
    }
  }

  function handleNamespaceChange() {
    labelConditions.value = [] // 命名空间切换时清空 label 条件
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
      ElMessage.error(e?.message || '加载 YAML 失败')
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
      ElMessage.error(e?.message || '加载 YAML 失败')
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
      ElMessage.success('YAML 保存成功')
      yamlEditing.value = false
      fetchResources()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存 YAML 失败')
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

  async function handleDelete(row: any, force?: boolean) {
    if (force && options.forceDeleteResource) {
      const msg = `强制删除 ${options.resourceName} "${row.name}" 将跳过优雅终止，控制器管理的 Pod 会被立即重建。确定继续？`
      try {
        await ElMessageBox.confirm(msg, '确认', { type: 'warning' })
      } catch {
        return
      }
      loading.value = true
      try {
        await options.forceDeleteResource({ namespace: row.namespace, name: row.name })
        ElMessage.success(`${options.resourceName} 已强制删除`)
        const id = resourceKey(row)
        markPendingDelete([id])
        list.value = list.value.filter((item) => !isPendingDelete(resourceKey(item)))
        totalCount.value = Math.max(0, totalCount.value - 1)
        // Remove from selection if present
        selectedRows.value = selectedRows.value.filter((r) => resourceKey(r) !== id)
        scheduleCleanup([id])
      } catch (e: any) {
        ElMessage.error(e?.message || `强制删除${options.resourceName}失败`)
      } finally {
        loading.value = false
      }
      return
    }
    const msg = options.deleteConfirm
      ? options.deleteConfirm(row)
      : row.namespace
        ? `删除 ${options.resourceName} "${row.name}"（命名空间: ${row.namespace}）？`
        : `删除 ${options.resourceName} "${row.name}"？`
    try {
      await ElMessageBox.confirm(msg, '确认', { type: 'warning' })
    } catch {
      // 用户取消,不报错
      return
    }
    loading.value = true
    try {
      await options.deleteResource({ namespace: row.namespace, name: row.name })
      ElMessage.success(`${options.resourceName} 已删除`)
      const id = resourceKey(row)
      markPendingDelete([id])
      list.value = list.value.filter((item) => !isPendingDelete(resourceKey(item)))
      totalCount.value = Math.max(0, totalCount.value - 1)
      selectedRows.value = selectedRows.value.filter((r) => resourceKey(r) !== id)
      scheduleCleanup([id])
    } catch (e: any) {
      ElMessage.error(e?.message || `删除${options.resourceName}失败`)
    } finally {
      loading.value = false
    }
  }

  async function handleBatchDelete() {
    if (!selectedRows.value.length) return
    // Capture rows before async gap to avoid race condition
    const rowsToDelete = [...selectedRows.value]
    try {
      await ElMessageBox.confirm(
        `删除选中的 ${rowsToDelete.length} 个 ${options.resourceName}？`,
        '确认',
        { type: 'warning' }
      )
      loading.value = true
      const results = await Promise.allSettled(
        rowsToDelete.map((row) =>
          options.deleteResource({ namespace: row.namespace, name: row.name })
        )
      )
      const successCount = results.filter((r) => r.status === 'fulfilled').length
      const failCount = results.filter((r) => r.status === 'rejected').length
      if (failCount > 0) {
        ElMessage.warning(`成功删除 ${successCount} 个，失败 ${failCount} 个`)
      } else {
        ElMessage.success(`已删除 ${successCount} 个 ${options.resourceName}`)
      }
      // Collect successfully deleted IDs
      const deletedIds: string[] = []
      results.forEach((r, i) => {
        if (r.status === 'fulfilled') {
          deletedIds.push(resourceKey(rowsToDelete[i]))
        }
      })
      if (deletedIds.length > 0) {
        markPendingDelete(deletedIds)
        list.value = list.value.filter((item) => !isPendingDelete(resourceKey(item)))
        totalCount.value = Math.max(0, totalCount.value - deletedIds.length)
        scheduleCleanup(deletedIds)
      }
      // Only clear successfully deleted items from selection; keep failed ones selected
      const failedKeys = new Set(
        results
          .map((r, i) => (r.status === 'rejected' ? resourceKey(rowsToDelete[i]) : null))
          .filter(Boolean) as string[]
      )
      selectedRows.value = selectedRows.value.filter(
        (row) => failedKeys.has(resourceKey(row))
      )
    } catch {
      // cancelled
    } finally {
      loading.value = false
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
      ElMessage.info(`自动刷新已开启（${interval / 1000}s）`)
    } else {
      if (autoRefreshTimer) {
        clearInterval(autoRefreshTimer)
        autoRefreshTimer = null
      }
      ElMessage.info('自动刷新已关闭')
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

  // Label conditions change handler
  function onLabelConditionsChange(conditions: LabelCondition[]) {
    labelConditions.value = conditions
    currentPage.value = 1
    continueTokens.value = []
    fetchResources()
  }

  // URL sync: watch labelConditions and sync to query param `ls`
  // 使用 encodeURIComponent/decodeURIComponent 替代 btoa/atob 以支持 Unicode
  watch(labelConditions, (val) => {
    const query = { ...router.currentRoute.value.query }
    if (val.length > 0) {
      query.ls = encodeURIComponent(JSON.stringify(val))
    } else {
      delete query.ls
    }
    router.replace({ query })
  }, { deep: true })

  onMounted(() => {
    fetchNamespaces()

    // Restore label conditions from URL
    const ls = router.currentRoute.value.query.ls as string
    if (ls && ls.length <= 2000) {
      try {
        const parsed = JSON.parse(decodeURIComponent(ls))
        if (Array.isArray(parsed) && parsed.length > 0) {
          // Validate each condition has required fields
          const validConditions = parsed.filter((c: any) =>
            c && typeof c.key === 'string' && c.key &&
            ['=', '!=', 'in', 'notin'].includes(c.operator) &&
            Array.isArray(c.values) && c.values.length > 0 &&
            c.values.every((v: any) => typeof v === 'string')
          )
          if (validConditions.length > 0) {
            labelConditions.value = validConditions
          }
        }
      } catch {
        // 解析失败，静默清除无效参数
        const query = { ...router.currentRoute.value.query }
        delete query.ls
        router.replace({ query })
      }
    }

    fetchResources()
    document.addEventListener('keydown', handleKeyboard)
  })

  onUnmounted(() => {
    if (autoRefreshTimer) clearInterval(autoRefreshTimer)
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    pendingDeleteTimers.forEach((t) => clearTimeout(t))
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
    // Label selector
    labelConditions,
    onLabelConditionsChange,
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
