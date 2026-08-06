/**
 * K8s resource quantity 格式化工具。
 *
 * K8s 的 CPU/内存数量有多种后缀表示（m 毫核、Ki/Mi/Gi/Ti 二进制后缀、裸数值），
 * 这里提供面向"显示"的格式化：CPU 统一成 "X.XX Core"，内存/存储统一成人类可读的 Gi/Mi/KiB。
 * 注意：与把数量解析成数值（如 ResourceQuotaDetail 的 parseResourceValue）是不同关注点，不在此处处理。
 */

/**
 * 格式化 CPU 核数（数值，非 K8s quantity 字符串）为定长字符串。
 * 用于进度条/卡片显示已请求/可分配的核数，如 1.50。
 */
export function formatCpuCores(n: number): string {
  return (n || 0).toFixed(2)
}

/**
 * 格式化内存 GiB 数值（已请求/可分配）为定长字符串，如 3.2。
 */
export function formatMemGiB(n: number): string {
  return (n || 0).toFixed(1)
}

/**
 * 格式化 K8s CPU 数量字符串。
 * - "1500m" → "1.50 Core"（毫核转核，保留两位小数）
 * - "2" → "2.00 Core"
 * - 其他无法解析的原样返回
 */
export function formatK8sCPU(val: string | undefined): string {
  if (!val) return '-'
  const s = String(val)
  if (s.endsWith('m')) {
    const milli = parseFloat(s)
    if (!isNaN(milli)) return `${(milli / 1000).toFixed(2)} Core`
    return s
  }
  const num = parseFloat(s)
  if (!isNaN(num)) return `${num.toFixed(2)} Core`
  return s
}

/**
 * 格式化 K8s 内存/存储数量字符串为人类可读单位。
 * - "1048576Ki" → "1.0 Gi"
 * - "512Mi" → "512 Mi"
 * - "2Gi" → "2Gi"（原样）
 * - 裸数值按字节向上换算
 */
export function formatK8sMemory(val: string | undefined): string {
  if (!val) return '-'
  const s = String(val)
  if (s.endsWith('Ki')) {
    const ki = parseInt(s)
    if (ki >= 1048576) return `${(ki / 1048576).toFixed(1)} Gi`
    else if (ki >= 1024) return `${(ki / 1024).toFixed(0)} Mi`
    return `${ki} KiB`
  }
  if (s.endsWith('Mi')) {
    const mi = parseInt(s)
    if (mi >= 1024) return `${(mi / 1024).toFixed(1)} Gi`
    return `${mi} Mi`
  }
  if (s.endsWith('Gi')) return s
  if (s.endsWith('Ti')) return s
  const num = parseInt(s)
  if (!isNaN(num)) {
    if (num >= 1073741824) return `${(num / 1073741824).toFixed(1)} Gi`
    else if (num >= 1048576) return `${(num / 1048576).toFixed(0)} Mi`
    else if (num >= 1024) return `${(num / 1024).toFixed(0)} KiB`
    return `${num} B`
  }
  return s
}
