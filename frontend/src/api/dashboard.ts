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
  ingresses: number
}

export interface NamespaceUsage {
  name: string
  pod_count: number
  running_pods: number
  cpu_used: number
  mem_used: number
}

export interface NamespaceResources {
  total_cpu: number
  total_mem: number
  namespaces: NamespaceUsage[]
}

export interface AbnormalPod {
  name: string
  namespace: string
  phase: string
  reason: string
  node: string
}

export interface RestartingPod {
  name: string
  namespace: string
  restart_count: number
  node: string
}

export interface PressureNode {
  name: string
  pressures: string[]
}

export interface AbnormalPVC {
  name: string
  namespace: string
  phase: string
}

export interface HealthSummary {
  healthy_pods: number
  abnormal_pods: number
  ready_nodes: number
  not_ready_nodes: number
  bound_pvcs: number
  abnormal_pvcs: number
}

export interface ClusterHealth {
  summary: HealthSummary
  abnormal_pods: AbnormalPod[] | null
  restarting_pods: RestartingPod[] | null
  not_ready_nodes: string[] | null
  pressure_nodes: PressureNode[] | null
  abnormal_pvcs: AbnormalPVC[] | null
}

export interface K8sEvent {
  type: string
  reason: string
  message: string
  namespace: string
  involved_object: string
  involved_object_kind: string
  involved_object_name: string
  first_seen: string
  last_seen: string
  count: number
  reporting_component: string
  reporting_instance: string
  action: string
  cluster_name: string
}

export function getOverview(params?: { clusterId?: number }) {
  return request.get<Overview>('/dashboard/overview', { params })
}

export function getResources(params?: { clusterId?: number }) {
  return request.get<ResourceMetrics>('/dashboard/resources', { params })
}

export function getWorkloads(params?: { clusterId?: number }) {
  return request.get<WorkloadSummary>('/dashboard/workloads', { params })
}

export function getNamespaceResources(params?: { clusterId?: number }) {
  return request.get<NamespaceResources>('/dashboard/namespaces', { params })
}

export function getHealth(params?: { clusterId?: number }) {
  return request.get<ClusterHealth>('/dashboard/health', { params })
}

export interface EventsResponse {
  items: K8sEvent[]
  total: number
  continue: string
  has_more: boolean
}

export function getEvents(params?: {
  clusterId?: number
  namespace?: string
  type?: string
  fieldSelector?: string
  limit?: number
  continue?: string
}) {
  return request.get<EventsResponse>('/dashboard/events', { params })
}
