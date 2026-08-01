/**
 * Shared utility functions used across the application
 */

/**
 * Returns the Element Plus tag type for a given status string
 */
export function statusType(status: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  const s = (status || '').toLowerCase()
  if (['running', 'ready', 'active', 'bound', 'available', 'true', 'ok', 'healthy'].includes(s)) return 'success'
  if (['pending', 'waiting', 'containercreating', 'terminating', 'released'].includes(s)) return 'warning'
  if (['failed', 'error', 'crashloopbackoff', 'imagepullbackoff', 'errimagepull', 'OOMKilled', 'unknown', 'notready', 'false'].includes(s)) return 'danger'
  if (['succeeded', 'completed'].includes(s)) return 'info'
  return ''
}

/**
 * Returns a human-readable label for a status string
 */
export function statusLabel(status: string): string {
  const map: Record<string, string> = {
    'Running': '运行中',
    'Pending': '等待中',
    'Succeeded': '成功',
    'Failed': '失败',
    'Unknown': '未知',
    'Terminating': '终止中',
    'ContainerCreating': '创建中',
    'ImagePullBackOff': '镜像拉取失败',
    'ErrImagePull': '镜像拉取错误',
    'CrashLoopBackOff': '重启循环',
    'OOMKilled': '内存溢出',
    'Ready': '就绪',
    'NotReady': '未就绪',
    'Active': '活跃',
    'Bound': '已绑定',
    'Available': '可用',
    'Released': '已释放',
    'Lost': '丢失',
    'True': '正常',
    'False': '异常',
  }
  return map[status] || status
}

/**
 * Base64 encode a UTF-8 string
 */
export function base64Encode(str: string): string {
  // 使用 TextEncoder 获取 UTF-8 字节，再编码为 base64（避免废弃的 escape/unescape）
  const bytes = new TextEncoder().encode(str)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i])
  }
  return btoa(binary)
}

/**
 * Base64 decode a string to UTF-8（使用 TextDecoder + Uint8Array，避免废弃的 escape/unescape）
 */
export function base64Decode(str: string): string {
  const binary = atob(str)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return new TextDecoder().decode(bytes)
}

/**
 * Format bytes to human readable string
 */
export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['B', 'Ki', 'Mi', 'Gi', 'Ti', 'Pi', 'Ei', 'Zi', 'Yi']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

/**
 * Format a Kubernetes resource age from creation timestamp.
 * 最完整的年龄格式化实现（含秒/分/时/天/年），作为 utils/time.ts 与 api/resource.ts calcAge 的统一来源。
 * @param suffix 为 true 时追加 " ago" 后缀（兼容旧 time.formatAge 调用方）
 */
export function formatAge(creationTimestamp: string, suffix: boolean = true): string {
  if (!creationTimestamp) return '-'
  const created = new Date(creationTimestamp).getTime()
  const now = Date.now()
  const diff = Math.floor((now - created) / 1000)
  const ago = suffix ? ' ago' : ''
  if (diff < 60) return `${diff}s${ago}`
  if (diff < 3600) return `${Math.floor(diff / 60)}m${ago}`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h${ago}`
  const days = Math.floor(diff / 86400)
  if (days < 365) return `${days}d${ago}`
  return `${Math.floor(days / 365)}y${ago}`
}

/**
 * Truncate a string to a maximum length with ellipsis
 */
export function truncate(str: string, maxLen: number): string {
  if (!str || str.length <= maxLen) return str
  return str.substring(0, maxLen) + '...'
}

/**
 * Deep clone an object.
 * 限制：基于 JSON 序列化，无法保留 undefined/函数/Symbol/Date 对象/循环引用等，
 * 仅适用于可安全 JSON 序列化的 K8s 资源数据。
 */
export function deepClone<T>(obj: T): T {
  return JSON.parse(JSON.stringify(obj))
}

/**
 * 计算使用率百分比（0-100），total 非正时返回 0
 */
export function usagePercent(used: number, total: number): number {
  if (!total || total <= 0) return 0
  return Math.min(100, Math.round((used / total) * 100))
}

/**
 * 根据使用率百分比返回进度条颜色（>=90 危险 / >=70 警告 / 其余主色）
 */
export function progressColor(percent: number): string {
  if (percent >= 90) return 'var(--gk-color-danger)'
  if (percent >= 70) return 'var(--gk-color-warning)'
  return 'var(--gk-color-primary)'
}
