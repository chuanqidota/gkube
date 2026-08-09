import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import {
  extractK8sItems,
  extractResourceNames,
  fetchAndPrepareClone,
  loadNamespaceList,
} from './useCloneResource'
import { getNamespaceList, extractNamespaceNames } from '@/api/resource'

export interface CloneApi {
  /** List resources. For namespaced resources receives `{ namespace }`; for cluster-scoped `{}`. */
  list: (params: any) => Promise<any>
  /** Fetch a resource's YAML. Receives `{ namespace, name }` or `{ name }`. */
  yaml: (params: any) => Promise<any>
}

export interface UseCloneCreateOptions {
  api: CloneApi
  /** true (default) = namespaced resource; false = cluster-scoped (PV/StorageClass/VolumeSnapshotClass). */
  namespaceScoped?: boolean
  /** true (default) = page has a form; false = YAML-only page (target forced to 'yaml'). */
  hasForm?: boolean
  /** Invoked when user clones into the form. Optional (YAML-only pages omit it). */
  onCloneToForm?: (parsed: any) => void
  /** Invoked when user clones into the YAML editor. */
  onCloneToYaml: (parsed: any) => void
}

/**
 * Encapsulates the full clone-from-existing flow for a resource create page.
 * Returns reactive state + handlers ready to bind to <CloneDialog>.
 *
 * Pages destructure the return so the refs are top-level bindings (auto-unwrapped in template).
 */
export function useCloneCreate(opts: UseCloneCreateOptions) {
  const namespaceScoped = opts.namespaceScoped !== false
  const hasForm = opts.hasForm !== false

  const cloneMode = ref(false)
  const cloneNamespace = ref('default')
  const cloneName = ref('')
  const cloneNsOptions = ref<string[]>([])
  const cloneNameOptions = ref<string[]>([])
  const cloneNsLoading = ref(false)
  const cloneNameLoading = ref(false)
  const cloneLoading = ref(false)
  const cloneTarget = ref<'form' | 'yaml'>(hasForm ? 'form' : 'yaml')
  const parsedData = ref<any>(null)

  function resetCloneState() {
    cloneName.value = ''
    parsedData.value = null
  }

  async function fetchCloneNamespaces() {
    cloneNsLoading.value = true
    try {
      cloneNsOptions.value = await loadNamespaceList(async () => {
        const res: any = await getNamespaceList()
        return extractNamespaceNames(res.data)
      })
    } finally {
      cloneNsLoading.value = false
    }
  }

  async function fetchCloneNames(namespace: string) {
    cloneNameOptions.value = []
    cloneName.value = ''
    if (namespaceScoped && !namespace) return
    cloneNameLoading.value = true
    try {
      const res: any = await opts.api.list(namespaceScoped ? { namespace } : {})
      cloneNameOptions.value = extractResourceNames(extractK8sItems(res))
    } finally {
      cloneNameLoading.value = false
    }
  }

  async function handleLoadClone() {
    cloneLoading.value = true
    try {
      const result = await fetchAndPrepareClone(
        opts.api.yaml,
        namespaceScoped ? cloneNamespace.value : undefined,
        cloneName.value,
      )
      if (!result) return
      parsedData.value = result.parsed
      cloneMode.value = false
      if (cloneTarget.value === 'yaml' || !hasForm) {
        opts.onCloneToYaml(result.parsed)
      } else {
        opts.onCloneToForm?.(result.parsed)
      }
      ElMessage.success(`已成功克隆 "${result.originalName}"，请确认后点击创建`)
    } finally {
      cloneLoading.value = false
    }
  }

  function startClone() {
    cloneMode.value = true
    // 记录重置前的值：若已是 'default'，watch 不会触发，需显式拉取；
    // 若不是，watch 会触发拉取，此处不再重复调用，避免双重请求。
    const wasDefault = cloneNamespace.value === 'default'
    cloneNamespace.value = 'default'
    resetCloneState()
    cloneNameOptions.value = []
    if (namespaceScoped) {
      fetchCloneNamespaces()
      if (wasDefault) fetchCloneNames('default')
    } else {
      fetchCloneNames('')
    }
  }

  function cancelClone() {
    cloneMode.value = false
    resetCloneState()
  }

  if (namespaceScoped) {
    // 单一机制：命名空间变化时自动加载资源列表，避免重复请求
    watch(cloneNamespace, (newNs) => {
      if (newNs) fetchCloneNames(newNs)
    })
  }

  return {
    cloneMode,
    cloneNamespace,
    cloneName,
    cloneNsOptions,
    cloneNameOptions,
    cloneNsLoading,
    cloneNameLoading,
    cloneLoading,
    cloneTarget,
    parsedData,
    namespaceScoped,
    hasForm,
    startClone,
    cancelClone,
    handleLoadClone,
  }
}

/** Shared YAML dump options for clone-to-yaml. */
export const CLONE_YAML_DUMP_OPTS = { indent: 2, lineWidth: -1, noRefs: true } as const

/** Convenience: dump a parsed K8s object to YAML using the shared opts. */
export function dumpCloneYaml(parsed: any): string {
  return yaml.dump(parsed, CLONE_YAML_DUMP_OPTS)
}
