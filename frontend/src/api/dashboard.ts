import request from './request'

export interface Overview {
  cluster_count: number
  node_count: number
  pod_count: number
  namespace_count: number
}

export interface ResourceMetrics {
  cpu: { used: number; total: number }
  memory: { used: number; total: number }
  storage: { used: number; total: number }
}

export interface WorkloadSummary {
  deployments: number
  statefulsets: number
  daemonsets: number
  jobs: number
  cronjobs: number
}

export interface K8sEvent {
  type: string
  reason: string
  message: string
  namespace: string
  involved_object: string
  last_seen: string
}

export function getOverview() {
  return request.get<Overview>('/dashboard/overview')
}

export function getResources() {
  return request.get<ResourceMetrics>('/dashboard/resources')
}

export function getWorkloads() {
  return request.get<WorkloadSummary>('/dashboard/workloads')
}

export function getEvents() {
  return request.get<K8sEvent[]>('/dashboard/events')
}
