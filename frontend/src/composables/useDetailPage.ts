import { ref, onMounted, type Ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAutoRefresh } from './useAutoRefresh'
import { useClusterNameRef } from './useClusterName'

export interface DetailPageOptions {
  /** 资源显示名，如 'Deployment' */
  resourceName: string
  /** 获取详情 */
  fetchDetail: (params: any) => Promise<any>
  /** 获取事件（可选） */
  fetchEvents?: (params: any) => Promise<any>
  /** 删除资源 */
  deleteResource: (data: any) => Promise<any>
  /** 强制删除（可选，Pod 用） */
  forceDeleteResource?: (data: any) => Promise<any>
  /** 列表页路由（返回按钮用） */
  listRoute: string
  /** 从 route params 构造 API 请求参数 */
  buildParams: () => Record<string, any>
  /** 自定义删除确认消息（可选） */
  deleteConfirm?: (detail: any, name: string) => string
  /** 自动刷新时的额外回调（如 fetchPods/fetchReplicaSets） */
  onRefresh?: () => Promise<void>
  /** 自动启动刷新（默认 false） */
  autoStart?: boolean
}

export function useDetailPage(options: DetailPageOptions) {
  const route = useRoute()
  const router = useRouter()
  const clusterName = useClusterNameRef()

  const loading = ref(false)
  const detail: Ref<any> = ref(null)
  const events = ref<any[]>([])
  const eventsLoading = ref(false)
  const yamlDialogVisible = ref(false)
  const deleteLoading = ref(false)

  const namespace = route.params.namespace as string
  const name = route.params.name as string

  // ---- 数据获取 ----

  async function fetchDetail() {
    loading.value = true
    try {
      const res: any = await options.fetchDetail(options.buildParams())
      detail.value = res?.data ?? res
    } catch (e: any) {
      ElMessage.error(e?.message || `获取${options.resourceName}详情失败`)
    } finally {
      loading.value = false
    }
  }

  async function fetchEvents() {
    if (!options.fetchEvents) return
    eventsLoading.value = true
    try {
      const res: any = await options.fetchEvents(options.buildParams())
      events.value = res?.data ?? res ?? []
    } catch {
      // 事件加载失败不阻塞页面
    } finally {
      eventsLoading.value = false
    }
  }

  // ---- 删除 ----

  async function handleDelete(force = false) {
    const msg = options.deleteConfirm?.(detail.value, name)
      || `确定要删除 ${options.resourceName} "${name}" 吗？此操作不可恢复。`
    try {
      await ElMessageBox.confirm(msg, '确认删除', { type: 'error', confirmButtonText: '确定删除', cancelButtonText: '取消' })
    } catch { return }

    deleteLoading.value = true
    try {
      const params = options.buildParams()
      if (force && options.forceDeleteResource) {
        await options.forceDeleteResource({ ...params, force: true })
      } else {
        await options.deleteResource(params)
      }
      ElMessage.success('已删除')
      router.push(options.listRoute)
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    } finally {
      deleteLoading.value = false
    }
  }

  // ---- YAML ----

  function handleOpenYaml() {
    yamlDialogVisible.value = true
  }

  // ---- 自动刷新 ----

  const { isRunning, countdown, currentInterval, availableIntervals,
    toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
    await fetchDetail()
    if (options.onRefresh) await options.onRefresh()
    await fetchEvents()
  }, { autoStart: options.autoStart ?? false })

  // ---- 初始化 ----

  onMounted(async () => {
    await fetchDetail()
    if (options.onRefresh) await options.onRefresh()
    fetchEvents()
  })

  return {
    // 路由参数
    namespace, name,
    // 状态
    loading, detail, events, eventsLoading, yamlDialogVisible, deleteLoading,
    // 集群
    clusterName,
    // 自动刷新
    isRunning, countdown, currentInterval, availableIntervals,
    toggle, manualRefresh, setIntervalOption,
    // 操作
    fetchDetail, fetchEvents, handleDelete, handleOpenYaml,
    // 路由
    router,
  }
}
