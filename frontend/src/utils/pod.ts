/**
 * Pod 相关的前端公共工具:状态→标签类型映射、全屏视图 URL 构造。
 * 供 PodList / PodDetail / PodListPanel / TerminalView / LogView 共用,避免重复实现。
 */

/** Element Plus el-tag 的 type 取值。 */
export type TagType = 'success' | 'warning' | 'danger' | 'info'

/**
 * 将 Pod 状态(phase)映射为 el-tag 类型。
 * 统一覆盖 running/succeeded/pending/failed/error/unknown。
 */
export function getPodStatusType(status: string): TagType {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed' || s === 'error') return 'danger'
  return 'info'
}

/**
 * 构造全屏终端/日志视图的 URL(在新标签页打开)。
 * @param kind 'logs' | 'terminal'
 * @param opts.namespace Pod 命名空间
 * @param opts.pod Pod 名称
 * @param opts.cluster 集群名(可选,未传则不带 cluster 参数)
 */
export function buildFullscreenUrl(
  kind: 'logs' | 'terminal',
  opts: { namespace: string; pod: string; cluster?: string },
): string {
  const params = new URLSearchParams({ namespace: opts.namespace, pod: opts.pod })
  if (opts.cluster) params.set('cluster', opts.cluster)
  return `/fullscreen/${kind}?${params.toString()}`
}
