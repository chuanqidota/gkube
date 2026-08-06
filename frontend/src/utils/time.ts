// formatAge 的实现已统一到 @/utils/helpers（含秒/分/时/天/年 + 可选 " ago" 后缀）。
// 此处仅做转发，保持既有 `import { formatAge } from '@/utils/time'` 调用方可用。
export { formatAge } from '@/utils/helpers'

/**
 * 将 ISO 时间戳格式化为 "YYYY-MM-DD HH:mm:ss" 本地展示字符串。
 * 用于"创建时间"/"最后变更"等需要绝对时间的场景（区别于 formatAge 的相对年龄）。
 * 解析失败时原样返回，避免吞错。
 */
export function formatDateTime(iso: string): string {
  if (!iso) return '-'
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

