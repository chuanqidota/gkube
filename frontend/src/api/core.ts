import request from './request'
import type { AxiosResponse } from 'axios'
import { formatAge } from '@/utils/helpers'

// ============ 类型定义 ============

export interface Namespace {
  name: string
  status: string
  labels: Record<string, string>
  annotations: Record<string, string>
  age: string
}

export interface NodeInfo {
  name: string
  status: string
  roles: string
  version: string
  internal_ip: string
  external_ip: string
  architecture: string
  unschedulable: boolean
  pod_count: number
  labels: Record<string, string>
  taints: { key: string; value: string; effect: string }[]
  is_ready: boolean
  os_image: string
  kernel_version: string
  container_runtime: string
  creationTimestamp: string
  cpu_used: number
  cpu_total: number
  mem_used: number
  mem_total: number
  pod_total: number
}

export interface NodeDetail {
  name: string
  status: string
  roles: string
  version: string
  os: string
  kernel: string
  container_runtime: string
  architecture: string
  internal_ip: string
  external_ip: string
  hostname: string
  unschedulable: boolean
  labels: Record<string, string>
  taints: { key: string; value: string; effect: string }[]
  conditions: { type: string; status: string; reason: string; message: string; lastTransitionTime: string }[]
  capacity: Record<string, string>
  allocatable: Record<string, string>
  creationTimestamp: string
}

export interface NodeEvent {
  type: string
  reason: string
  message: string
  last_seen: string
}

export interface DrainResult {
  evicted: string[]
  skipped: string[]
  failed: string[]
}

export interface K8sPod {
  name: string
  namespace: string
  creationTimestamp: string
  status: {
    phase: string
    podIP: string
    conditions: { type: string; status: string }[]
    containerStatuses: { restartCount: number }[]
  }
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

/**
 * Calculate age string from a creation timestamp.
 * 委托到统一的 @/utils/helpers formatAge（无 " ago" 后缀，与原语义一致）。
 */
export function calcAge(creationTimestamp: string): string {
  if (!creationTimestamp) return ''
  return formatAge(creationTimestamp, false)
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

export function getNamespaceList(params?: { cluster_id?: number }) {
  return request.get<Namespace[]>('/k8s/namespace/list', { params })
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

// ============ Node API ============

export function getNodeList(params?: { clusterName?: string }): Promise<AxiosResponse<NodeInfo[]>> {
  return request.get('/k8s/cluster/nodes', { params })
}

export function getNodeDetail(params: { name: string }): Promise<AxiosResponse<NodeDetail>> {
  return request.get('/k8s/node/detail', { params })
}

export function getNodeYaml(params: { name: string }): Promise<AxiosResponse<{ yaml: string }>> {
  return request.get('/k8s/node/get-yaml', { params })
}

export function updateNodeYaml(data: { name: string; yaml: string }): Promise<AxiosResponse<null>> {
  return request.put('/k8s/node/update-yaml', data)
}

export function cordonNode(data: { name: string; cordon: boolean }): Promise<AxiosResponse<{ isCordon: boolean }>> {
  return request.put('/k8s/node/cordon', data)
}

export function updateNodeTaints(data: { name: string; taints: { key: string; value: string; effect: string }[] }): Promise<AxiosResponse<null>> {
  return request.put('/k8s/node/taints', data)
}

export function updateNodeLabels(data: { name: string; labels: Record<string, string> }): Promise<AxiosResponse<null>> {
  return request.put('/k8s/node/labels', data)
}

export function drainNode(data: {
  name: string
  ignoreDaemonSets?: boolean
  deleteLocalData?: boolean
  gracePeriod?: number
  force?: boolean
}): Promise<AxiosResponse<DrainResult>> {
  return request.put('/k8s/node/drain', data)
}

export function deleteNode(data: { name: string }): Promise<AxiosResponse<null>> {
  return request.delete('/k8s/node/delete', { data })
}

export function getNodePods(params: { name: string }): Promise<AxiosResponse<K8sPod[]>> {
  return request.get('/k8s/node/pods', { params })
}

export function getNodeEvents(params: { name: string }): Promise<AxiosResponse<NodeEvent[]>> {
  return request.get('/k8s/node/events', { params })
}

// ============ Event API ============

export function getEventList(params: {
  namespace?: string
  fieldSelector?: string
  limit?: number
  continue?: string
}) {
  return request.get('/k8s/event/list', { params })
}

// ============ Dashboard Events API ============

export function getDashboardEvents(params?: {
  clusterId?: number
  type?: string
  namespace?: string
  limit?: number
  continue?: string
  fieldSelector?: string
}) {
  return request.get('/dashboard/events', { params })
}

// ============ CRD API ============

export function getCrdList() {
  return request.get('/k8s/crd/list')
}

export function getCrdDetail(params: { name: string }) {
  return request.get('/k8s/crd/detail', { params })
}

export function getCrdYaml(params: { name: string }) {
  return request.get('/k8s/crd/get-yaml', { params })
}

export function createCrd(data: { yaml: string }) {
  return request.post('/k8s/crd/create', data)
}

export function updateCrd(data: { yaml: string }) {
  return request.put('/k8s/crd/update', data)
}

export function deleteCrd(params: { name: string }) {
  return request.delete('/k8s/crd/delete', { params })
}

export function getCustomResourceList(params: { group: string; version: string; resource: string; namespace?: string }) {
  return request.get('/k8s/crd/resources', { params })
}

export function getCustomResourceYaml(params: { group: string; version: string; resource: string; namespace?: string; name: string }) {
  return request.get('/k8s/crd/resource/yaml', { params })
}

export function getCustomResourceDetail(params: { group: string; version: string; resource: string; namespace?: string; name: string }) {
  return request.get('/k8s/crd/resource/detail', { params })
}

export function createCustomResource(data: { group: string; version: string; resource: string; namespace?: string; yaml: string }) {
  return request.post('/k8s/crd/resource/create', data)
}

export function deleteCustomResource(params: { group: string; version: string; resource: string; namespace?: string; name: string }) {
  return request.delete('/k8s/crd/resource', { params })
}

export function updateCustomResource(data: {
  group: string
  version: string
  resource: string
  namespace?: string
  yaml: string
}) {
  return request.put('/k8s/crd/resource/update', data)
}

export function patchCustomResource(data: {
  group: string
  version: string
  resource: string
  namespace?: string
  name: string
  patch: string
  patchType?: 'strategic' | 'merge' | 'json'
}) {
  return request.patch('/k8s/crd/resource/patch', data)
}
