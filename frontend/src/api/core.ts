// Re-export 门面 — 保持向后兼容
// 所有 import 从 core.ts 来的页面不需要改任何代码

export * from './namespace'
export * from './node'
export * from './event'
export * from './crd'

import { formatAge } from '@/utils/helpers'

/**
 * Calculate age string from a creation timestamp.
 * @deprecated 使用 formatAge from @/utils/helpers
 */
export function calcAge(creationTimestamp: string): string {
  if (!creationTimestamp) return ''
  return formatAge(creationTimestamp, false)
}
