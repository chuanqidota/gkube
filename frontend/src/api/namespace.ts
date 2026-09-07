import request from './request'

// ============ 类型定义 ============

export interface Namespace {
  name: string
  status: string
  labels: Record<string, string>
  annotations: Record<string, string>
  age: string
}

// ============ 工具函数 ============

/**
 * Extract namespace names from API response.
 * Handles both formats:
 * 1. Object array: [{name: "default", ...}, ...]
 * 2. Simple string array: ["default", "kube-system", ...]
 */
export function extractNamespaceNames(data: any): string[] {
  if (!Array.isArray(data)) return []
  return data
    .map((item: any) => (typeof item === 'string' ? item : item.name))
    .filter(Boolean)
}

// ============ Transform 函数 ============

export function transformNamespaces(items: any[]): Namespace[] {
  if (!Array.isArray(items)) return []
  return items.map((ns: any) => ({
    name: ns.name || '',
    status: ns.status || 'Unknown',
    labels: ns.labels || {},
    annotations: ns.annotations || {},
    age: ns.age || '',
  }))
}

// ============ Namespace API ============

export function getNamespaceList(data?: any) {
  return request.post<Namespace[]>('/k8s/namespace/list', data)
}

export function getNamespaceDetail(params: { name: string }) {
  return request.get('/k8s/namespace/detail', { params })
}

export function getNamespaceYaml(params: { name: string }) {
  return request.get('/k8s/namespace/get-yaml', { params })
}

export function createNamespace(data: {
  namespace: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
}) {
  return request.post('/k8s/namespace/create', data)
}

export function updateNamespace(data: { yaml: string }) {
  return request.put('/k8s/namespace/update', data)
}

export function updateNamespaceLabels(data: {
  namespace: string
  labels: Record<string, string>
}) {
  return request.put('/k8s/namespace/labels', data)
}

export function deleteNamespace(params: { name: string }) {
  return request.delete('/k8s/namespace/delete', { params })
}
