import yaml from 'js-yaml'
import { ElMessage } from 'element-plus'

/**
 * Reusable helpers for clone-from-existing flows.
 * Pure functions, no internal state.
 */

/**
 * Normalize K8s list API response (handles both paginated `{ data: { items: [...] } }`
 * and non-paginated `{ data: [...] }` shapes) into a flat array of items.
 */
export function extractK8sItems(res: any): any[] {
  if (Array.isArray(res?.data)) return res.data
  if (Array.isArray(res?.data?.items)) return res.data.items
  return []
}

/**
 * Extract resource names from a list of K8s objects.
 */
export function extractResourceNames(items: any[]): string[] {
  return items.map((item: any) => item?.name || item?.metadata?.name).filter(Boolean) as string[]
}

/**
 * Extract the YAML string from various API response shapes.
 */
export function extractYamlString(res: any): string {
  if (res?.data && typeof res.data === 'object' && (res.data as any).yaml) {
    return (res.data as any).yaml
  }
  if (typeof res?.data === 'string') return res.data
  return ''
}

/**
 * Fetch the source resource's YAML, parse it, strip cluster-managed immutable fields,
 * and prepare a cloned K8s object ready for create-submit.
 *
 * `namespace`:
 *   - a string  → namespaced resource (validated non-empty, written into metadata.namespace)
 *   - undefined → cluster-scoped resource (no namespace; metadata.namespace stripped if present)
 *
 * Returns null on any failure (and surfaces an ElMessage error to the user).
 */
export async function fetchAndPrepareClone(
  yamlFetcher: (params: { namespace?: string; name: string }) => Promise<any>,
  namespace: string | undefined,
  name: string,
): Promise<{ parsed: any; originalName: string } | null> {
  if (!name) {
    ElMessage.warning('请选择资源名称')
    return null
  }
  if (namespace !== undefined && !namespace) {
    ElMessage.warning('请选择命名空间和资源名称')
    return null
  }
  try {
    const res: any = await yamlFetcher(namespace !== undefined ? { namespace, name } : { name })
    const rawYaml = extractYamlString(res)
    if (!rawYaml) {
      const errMsg = res?.msg || res?.message || `响应为空 (code=${res?.code || 'unknown'})`
      ElMessage.error(`获取 YAML 失败: ${errMsg}`)
      return null
    }
    const parsed = yaml.load(rawYaml) as any
    if (!parsed || typeof parsed !== 'object' || !parsed.metadata) {
      ElMessage.error('解析克隆源失败：YAML 内容无效或缺失 metadata')
      return null
    }
    const originalName = parsed.metadata.name
    parsed.metadata.name = `${originalName}-copy`
    if (namespace !== undefined) {
      parsed.metadata.namespace = namespace
    } else {
      delete parsed.metadata.namespace
    }
    for (const key of ['uid', 'resourceVersion', 'selfLink', 'creationTimestamp', 'generation', 'managedFields']) {
      delete parsed.metadata[key]
    }
    return { parsed, originalName }
  } catch (e: any) {
    ElMessage.error(e?.message || '加载克隆源失败')
    return null
  }
}

/**
 * Load namespace list with default fallback to ['default'] on failure.
 */
export async function loadNamespaceList(fetchFn: () => Promise<string[]>): Promise<string[]> {
  try {
    return await fetchFn()
  } catch {
    return ['default']
  }
}