import request from './request'
import type { AxiosResponse } from 'axios'

// ============ 类型定义 ============

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
  conditions: {
    type: string
    status: string
    reason: string
    message: string
    lastTransitionTime: string
  }[]
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

export function cordonNode(data: {
  name: string
  cordon: boolean
}): Promise<AxiosResponse<{ isCordon: boolean }>> {
  return request.put('/k8s/node/cordon', data)
}

export function updateNodeTaints(data: {
  name: string
  taints: { key: string; value: string; effect: string }[]
}): Promise<AxiosResponse<null>> {
  return request.put('/k8s/node/taints', data)
}

export function updateNodeLabels(data: {
  name: string
  labels: Record<string, string>
}): Promise<AxiosResponse<null>> {
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
