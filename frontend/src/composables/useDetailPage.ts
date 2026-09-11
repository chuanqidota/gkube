import { ref, onMounted, type Ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useAutoRefresh } from './useAutoRefresh'
import { useClusterNameRef } from './useClusterName'

export interface DetailError {
  type: 'not-found' | 'network' | 'forbidden' | 'unknown'
  message: string
}

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
  const { t } = useI18n()
  const clusterName = useClusterNameRef()

  const loading = ref(false)
  const detail: Ref<any> = ref(null)
  const error: Ref<DetailError | null> = ref(null)
  const events = ref<any[]>([])
  const eventsLoading = ref(false)
  const yamlDialogVisible = ref(false)
  const deleteLoading = ref(false)

  const namespace = route.params.namespace as string
  const name = route.params.name as string

  // ---- 数据获取 ----

  async function fetchDetail() {
    loading.value = true
    error.value = null
    try {
      const res: any = await options.fetchDetail(options.buildParams())
      detail.value = res?.data ?? res
    } catch (e: any) {
      const status = e?.response?.status
      if (status === 404) {
        error.value = { type: 'not-found', message: '该资源不存在或已被删除' }
      } else if (status === 403) {
        error.value = { type: 'forbidden', message: '没有权限访问该资源' }
      } else if (!e?.response) {
        error.value = { type: 'network', message: '网络错误，请检查连接后重试' }
      } else {
        error.value = { type: 'unknown', message: e?.message || '加载失败' }
      }
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
    const msg =
      options.deleteConfirm?.(detail.value, name) ||
      t('common.deleteResourceConfirm', { type: options.resourceName, name })
    try {
      await ElMessageBox.confirm(msg, t('common.confirmDelete'), {
        type: 'error',
        confirmButtonText: t('common.confirmDelete'),
        cancelButtonText: t('common.cancel'),
      })
    } catch {
      return
    }

    deleteLoading.value = true
    try {
      const params = options.buildParams()
      if (force && options.forceDeleteResource) {
        await options.forceDeleteResource({ ...params, force: true })
      } else {
        await options.deleteResource(params)
      }
      ElMessage.success(t('common.deleteSuccess', { type: options.resourceName }))
      router.push(options.listRoute)
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.deleteFailed', { type: options.resourceName }))
    } finally {
      deleteLoading.value = false
    }
  }

  // ---- YAML ----

  function handleOpenYaml() {
    yamlDialogVisible.value = true
  }

  // ---- 自动刷新 ----

  const {
    isRunning,
    countdown,
    currentInterval,
    availableIntervals,
    toggle,
    refresh: manualRefresh,
    setIntervalOption,
  } = useAutoRefresh(
    async () => {
      await fetchDetail()
      if (options.onRefresh) await options.onRefresh()
      await fetchEvents()
    },
    { autoStart: options.autoStart ?? false },
  )

  // ---- 初始化 ----

  onMounted(async () => {
    await fetchDetail()
    if (options.onRefresh) await options.onRefresh()
    fetchEvents()
  })

  return {
    // 路由参数
    namespace,
    name,
    // 状态
    loading,
    detail,
    error,
    events,
    eventsLoading,
    yamlDialogVisible,
    deleteLoading,
    // 集群
    clusterName,
    // 自动刷新
    isRunning,
    countdown,
    currentInterval,
    availableIntervals,
    toggle,
    manualRefresh,
    setIntervalOption,
    // 操作
    fetchDetail,
    fetchEvents,
    handleDelete,
    handleOpenYaml,
    // 路由
    router,
  }
}
