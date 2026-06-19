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
 * Base64 encode a string
 */
export function base64Encode(str: string): string {
  return btoa(unescape(encodeURIComponent(str)))
}

/**
 * Base64 decode a string
 */
export function base64Decode(str: string): string {
  return decodeURIComponent(escape(atob(str)))
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
 * Format a Kubernetes resource age from creation timestamp
 */
export function formatAge(creationTimestamp: string): string {
  if (!creationTimestamp) return '-'
  const created = new Date(creationTimestamp)
  const now = new Date()
  const diff = now.getTime() - created.getTime()

  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}d`
  if (hours > 0) return `${hours}h`
  if (minutes > 0) return `${minutes}m`
  return `${seconds}s`
}

/**
 * Truncate a string to a maximum length with ellipsis
 */
export function truncate(str: string, maxLen: number): string {
  if (!str || str.length <= maxLen) return str
  return str.substring(0, maxLen) + '...'
}

/**
 * Deep clone an object
 */
export function deepClone<T>(obj: T): T {
  return JSON.parse(JSON.stringify(obj))
}
