export function formatAge(timestamp: string, suffix: boolean = true): string {
  const now = new Date()
  const created = new Date(timestamp)
  const diffMs = now.getTime() - created.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  const ago = suffix ? ' ago' : ''
  if (diffDays > 0) return `${diffDays}d${ago}`
  if (diffHours > 0) return `${diffHours}h${ago}`
  return `${diffMins}m${ago}`
}
