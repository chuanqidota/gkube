/**
 * K8s qualified name 校验（label key、taint key 等均遵循此规则）。
 * 规则：可选 prefix/（DNS 子域名，≤253 字符）+ name（≤63 字符，字母数字开头结尾，中间可含 -_.）。
 * 与 K8s apimachinery IsQualifiedName 对齐。
 */
const QUALIFIED_NAME_RE = /^(?:(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\/)?[A-Za-z0-9](?:[-A-Za-z0-9_.]*[A-Za-z0-9])?$/

/**
 * 校验 K8s qualified name（label/taint key）。返回错误提示，合法时返回空字符串。
 */
export function validateQualifiedName(key: string, maxLen = 63): string {
  if (!key) return 'Key 不能为空'
  // 拆分 prefix/name 分别校验
  const slashIdx = key.indexOf('/')
  if (slashIdx >= 0) {
    const prefix = key.slice(0, slashIdx)
    const name = key.slice(slashIdx + 1)
    if (prefix.length > 253) return 'prefix 部分长度不能超过 253'
    if (name.length > maxLen) return `name 部分长度不能超过 ${maxLen}`
  } else if (key.length > maxLen) {
    return `Key 长度不能超过 ${maxLen}`
  }
  if (!QUALIFIED_NAME_RE.test(key)) return 'Key 格式不合法（需字母数字开头结尾，仅含 -_.，可选 prefix/）'
  return ''
}

/**
 * 校验 K8s label value（可为空，非空时 ≤63 字符，字母数字开头结尾，中间可含 -_.）。
 * 返回错误提示，合法时返回空字符串。
 */
const LABEL_VALUE_RE = /^[A-Za-z0-9](?:[-A-Za-z0-9_.]*[A-Za-z0-9])?$/
export function validateLabelValue(value: string): string {
  if (value === '') return ''
  if (value.length > 63) return 'Value 长度不能超过 63'
  if (!LABEL_VALUE_RE.test(value)) return 'Value 格式不合法（需字母数字开头结尾，仅含 -_.）'
  return ''
}

/**
 * 校验 taint effect 是否为合法枚举值。返回错误提示，合法时返回空字符串。
 */
const VALID_TAINT_EFFECTS = ['NoSchedule', 'PreferNoSchedule', 'NoExecute']
export function validateTaintEffect(effect: string): string {
  if (!VALID_TAINT_EFFECTS.includes(effect)) return 'Effect 必须为 NoSchedule/PreferNoSchedule/NoExecute'
  return ''
}

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
