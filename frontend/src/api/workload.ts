import request from './request'
import { calcAge } from './core'

// ============ 类型定义 ============

export interface Pod {
  name: string
  namespace: string
  status: string
  node: string
  ip: string
  hostIP: string
  restarts: number
  age: string
}

export interface Deployment {
  name: string
  namespace: string
  ready: string
  replicas: number
  ready_replicas: number
  up_to_date: number
  available: number
  age: string
}

export interface StatefulSet {
  name: string
  namespace: string
  ready: string
  age: string
  serviceName: string
  updateStrategy: string
}

export interface DaemonSet {
  name: string
  namespace: string
  desired: number
  current: number
  ready: number
  age: string
  updateStrategy: string
}

export interface Job {
  name: string
  namespace: string
  completions: string
  succeeded: number
  active: number
  failed: number
  age: string
}

export interface CronJob {
  name: string
  namespace: string
  schedule: string
  suspend: boolean
  active: number
  lastSchedule: string
  nextScheduleTime: string
  age: string
}

// ============ Transform 函数 ============

export function transformPods(items: any[]): Pod[] {
  if (!Array.isArray(items)) return []
  return items.map((pod: any) => {
    const restarts = (pod.status?.containerStatuses || []).reduce(
      (sum: number, cs: any) => sum + (cs.restartCount || 0), 0
    )
    return {
      name: pod.metadata?.name || '',
      namespace: pod.metadata?.namespace || '',
      status: pod.status?.phase || 'Unknown',
      node: pod.spec?.nodeName || '',
      ip: pod.status?.podIP || '',
      hostIP: pod.status?.hostIP || '',
      restarts,
      age: calcAge(pod.metadata?.creationTimestamp),
    }
  })
}

export function transformDeployments(items: any[]): Deployment[] {
  if (!Array.isArray(items)) return []
  return items.map((d: any) => ({
    name: d.metadata?.name || '',
    namespace: d.metadata?.namespace || '',
    ready: `${d.status?.readyReplicas || 0}/${d.spec?.replicas || 0}`,
    replicas: d.spec?.replicas || 0,
    ready_replicas: d.status?.readyReplicas || 0,
    up_to_date: d.status?.updatedReplicas || 0,
    available: d.status?.availableReplicas || 0,
    age: calcAge(d.metadata?.creationTimestamp),
  }))
}

export function transformStatefulSets(items: any[]): StatefulSet[] {
  if (!Array.isArray(items)) return []
  return items.map((d: any) => ({
    name: d.metadata?.name || '',
    namespace: d.metadata?.namespace || '',
    ready: `${d.status?.readyReplicas || 0}/${d.spec?.replicas || 0}`,
    serviceName: d.spec?.serviceName || '',
    updateStrategy: d.spec?.updateStrategy?.type || 'RollingUpdate',
    age: calcAge(d.metadata?.creationTimestamp),
  }))
}

export function transformDaemonSets(items: any[]): DaemonSet[] {
  if (!Array.isArray(items)) return []
  return items.map((d: any) => ({
    name: d.metadata?.name || '',
    namespace: d.metadata?.namespace || '',
    desired: d.status?.desiredNumberScheduled || 0,
    current: d.status?.currentNumberScheduled || 0,
    ready: d.status?.numberReady || 0,
    updateStrategy: d.spec?.updateStrategy?.type || 'RollingUpdate',
    age: calcAge(d.metadata?.creationTimestamp),
  }))
}

export function transformJobs(items: any[]): Job[] {
  if (!Array.isArray(items)) return []
  return items.map((d: any) => ({
    name: d.metadata?.name || '',
    namespace: d.metadata?.namespace || '',
    completions: `${d.status?.succeeded || 0}/${d.spec?.completions || 1}`,
    succeeded: d.status?.succeeded || 0,
    active: d.status?.active || 0,
    failed: d.status?.failed || 0,
    age: calcAge(d.metadata?.creationTimestamp),
  }))
}

export function transformCronJobs(items: any[]): CronJob[] {
  if (!Array.isArray(items)) return []
  return items.map((d: any) => ({
    name: d.metadata?.name || '',
    namespace: d.metadata?.namespace || '',
    schedule: d.spec?.schedule || '',
    suspend: d.spec?.suspend || false,
    active: d.status?.active?.length || 0,
    lastSchedule: d.status?.lastScheduleTime || '',
    nextScheduleTime: d.nextScheduleTime || '',
    age: calcAge(d.metadata?.creationTimestamp),
  }))
}

// ============ 标准 CRUD（工厂生成） ============

export const podApi = {
  list:    (data?: any) => request.post('/k8s/pod/list', data),
  detail:  (params: { namespace: string; name: string }) => request.get('/k8s/pod/detail', { params }),
  getYaml: (params: { namespace: string; name: string }) => request.get('/k8s/pod/get-yaml', { params }),
  delete:  (data: { namespace: string; name: string; force?: boolean }) => request.delete('/k8s/pod/delete', { data }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/pod/events', { params }),
}

export const deploymentApi = {
  list:    (data?: any) => request.post('/k8s/deployment/list', data),
  detail:  (params: { namespace: string; name: string }) => request.get('/k8s/deployment/detail', { params }),
  getYaml: (params: { namespace: string; name: string }) => request.get('/k8s/deployment/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/deployment/create', data),
  updateYaml: (data: { namespace: string; name: string; yaml: string }) => request.put('/k8s/deployment/update-yaml', data),
  delete:  (data: { namespace: string; name: string }) => request.delete('/k8s/deployment/delete', { data }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/deployment/events', { params }),
}

export const statefulSetApi = {
  list:    (data?: any) => request.post('/k8s/statefulset/list', data),
  detail:  (params: any) => request.get('/k8s/statefulset/detail', { params }),
  getYaml: (params: any) => request.get('/k8s/statefulset/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/statefulset/create', data),
  updateYaml: (data: { namespace: string; name: string; yaml: string }) => request.put('/k8s/statefulset/update', data),
  delete:  (data: any) => request.delete('/k8s/statefulset/delete', { data }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/statefulset/events', { params }),
}

export const daemonSetApi = {
  list:    (data?: any) => request.post('/k8s/daemonset/list', data),
  detail:  (params: any) => request.get('/k8s/daemonset/detail', { params }),
  getYaml: (params: any) => request.get('/k8s/daemonset/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/daemonset/create', data),
  updateYaml: (data: { namespace: string; name: string; yaml: string }) => request.put('/k8s/daemonset/update', data),
  delete:  (data: { namespace: string; name: string }) => request.delete('/k8s/daemonset/delete', { data }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/daemonset/events', { params }),
}

export const jobApi = {
  list:    (data?: any) => request.post('/k8s/job/list', data),
  detail:  (params: any) => request.get('/k8s/job/detail', { params }),
  getYaml: (params: any) => request.get('/k8s/job/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/job/create', data),
  updateYaml: (data: { namespace: string; name: string; yaml: string }) => request.put('/k8s/job/update', data),
  delete:  (data: any) => request.delete('/k8s/job/delete', { data }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/job/events', { params }),
}

export const cronJobApi = {
  list:    (data?: any) => request.post('/k8s/cronjob/list', data),
  detail:  (params: any) => request.get('/k8s/cronjob/detail', { params }),
  getYaml: (params: any) => request.get('/k8s/cronjob/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/cronjob/create', data),
  updateYaml: (data: { namespace: string; name: string; yaml: string }) => request.put('/k8s/cronjob/update', data),
  delete:  (data: any) => request.delete('/k8s/cronjob/delete', { data }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/cronjob/events', { params }),
}

export const replicaSetApi = {
  list:    (data?: any) => request.post('/k8s/replicaset/list', data),
  detail:  (params: { namespace: string; name: string }) => request.get('/k8s/replicaset/detail', { params }),
  getYaml: (params: { namespace: string; name: string }) => request.get('/k8s/replicaset/get-yaml', { params }),
  delete:  (params: { namespace: string; name: string }) => request.delete('/k8s/replicaset/delete', { params }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/replicaset/events', { params }),
}

export const hpaApi = {
  list:    (data?: any) => request.post('/k8s/hpa/list', data),
  detail:  (params: { namespace: string; name: string }) => request.get('/k8s/hpa/detail', { params }),
  getYaml: (params: { namespace: string; name: string }) => request.get('/k8s/hpa/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/hpa/create', data),
  updateYaml: (data: { namespace: string; yaml: string }) => request.put('/k8s/hpa/update', data),
  delete:  (params: { namespace: string; name: string }) => request.delete('/k8s/hpa/delete', { params }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/hpa/events', { params }),
}

// ============ 资源特有操作 ============

// Pod
export const getPodList = podApi.list
export const getPodDetail = podApi.detail
export const getPodYaml = podApi.getYaml
export const deletePod = podApi.delete
export const getPodEvents = podApi.events

// Deployment
export const getDeploymentList = deploymentApi.list
export const getDeploymentDetail = deploymentApi.detail
export const getDeploymentYaml = deploymentApi.getYaml
export const createDeployment = deploymentApi.create
export const updateDeploymentYaml = deploymentApi.updateYaml
export const deleteDeployment = deploymentApi.delete
export const getDeploymentEvents = deploymentApi.events
export const scaleDeployment = (data: { namespace: string; name: string; replicas: number }) => request.put('/k8s/deployment/scale', data)
export const restartDeployment = (data: { namespace: string; name: string }) => request.post('/k8s/deployment/restart', data)
export const rollbackDeployment = (data: { namespace: string; name: string; revision: number }) => request.post('/k8s/deployment/rollback', data)
export const updateDeploymentImage = (data: { namespace: string; name: string; containerName: string; image: string }) => request.put('/k8s/deployment/update-image', data)
export const getDeploymentReplicaSets = (params: { namespace: string; name: string }) => request.get('/k8s/deployment/replicasets', { params })
export const getDeploymentPodList = (params: { namespace: string; name: string }) => request.get('/k8s/deployment/pods', { params })

// StatefulSet
export const getStatefulSetList = statefulSetApi.list
export const getStatefulSetDetail = statefulSetApi.detail
export const getStatefulSetYaml = statefulSetApi.getYaml
export const createStatefulSet = statefulSetApi.create
export const updateStatefulSetYaml = statefulSetApi.updateYaml
export const deleteStatefulSet = statefulSetApi.delete
export const getStatefulSetEvents = statefulSetApi.events
export const scaleStatefulSet = (data: { namespace: string; name: string; replicas: number }) => request.put('/k8s/statefulset/scale', data)
export const restartStatefulSet = (data: { namespace: string; name: string }) => request.post('/k8s/statefulset/restart', data)
export const rollbackStatefulSet = (data: { namespace: string; name: string; revision: number }) => request.post('/k8s/statefulset/rollback', data)
export const updateStatefulSetImage = (data: { namespace: string; name: string; containerName: string; image: string }) => request.put('/k8s/statefulset/update-image', data)
export const getStatefulSetRollbacks = (params: { namespace: string; name: string }) => request.get('/k8s/statefulset/rollbacks', { params })
export const getStatefulSetPVCs = (params: { namespace: string; name: string }) => request.get('/k8s/statefulset/pvcs', { params })
export const getStatefulSetPods = (params: { namespace: string; name: string }) => request.get('/k8s/statefulset/pods', { params })

// DaemonSet
export const getDaemonSetList = daemonSetApi.list
export const getDaemonSetDetail = daemonSetApi.detail
export const getDaemonSetYaml = daemonSetApi.getYaml
export const createDaemonSet = daemonSetApi.create
export const updateDaemonSetYaml = daemonSetApi.updateYaml
export const deleteDaemonSet = daemonSetApi.delete
export const getDaemonSetEvents = daemonSetApi.events
export const getDaemonSetPods = (params: { namespace: string; name: string }) => request.get('/k8s/daemonset/pods', { params })
export const restartDaemonSet = (data: { namespace: string; name: string }) => request.post('/k8s/daemonset/restart', data)
export const updateDaemonSetImage = (data: { namespace: string; name: string; containerName: string; image: string }) => request.put('/k8s/daemonset/update-image', data)
export const rollbackDaemonSet = (data: { namespace: string; name: string; revision: number }) => request.post('/k8s/daemonset/rollback', data)
export const getDaemonSetRollbacks = (params: { namespace: string; name: string }) => request.get('/k8s/daemonset/rollbacks', { params })

// Job
export const getJobList = jobApi.list
export const getJobDetail = jobApi.detail
export const getJobYaml = jobApi.getYaml
export const createJob = jobApi.create
export const updateJobYaml = jobApi.updateYaml
export const deleteJob = jobApi.delete
export const getJobEvents = jobApi.events
export const getJobPods = (params: { namespace: string; name: string }) => request.get('/k8s/job/pods', { params })
export const rerunJob = (params: { namespace: string; name: string }) => request.post('/k8s/job/rerun', undefined, { params })

// CronJob
export const getCronJobList = cronJobApi.list
export const getCronJobDetail = cronJobApi.detail
export const getCronJobYaml = cronJobApi.getYaml
export const createCronJob = cronJobApi.create
export const updateCronJobYaml = cronJobApi.updateYaml
export const deleteCronJob = cronJobApi.delete
export const getCronJobEvents = cronJobApi.events
export const getCronJobExecutionHistory = (params: { namespace: string; name: string }) => request.get('/k8s/cronjob/jobs', { params })
export const suspendCronJob = (params: { namespace: string; name: string }) => request.put('/k8s/cronjob/suspend', undefined, { params })
export const resumeCronJob = (params: { namespace: string; name: string }) => request.put('/k8s/cronjob/resume', undefined, { params })
export const triggerCronJob = (params: { namespace: string; name: string }) => request.post('/k8s/cronjob/trigger', undefined, { params })

// ReplicaSet
export const getReplicaSetList = replicaSetApi.list
export const getReplicaSetDetail = replicaSetApi.detail
export const getReplicaSetYaml = replicaSetApi.getYaml
export const deleteReplicaSet = replicaSetApi.delete
export const getReplicaSetEvents = replicaSetApi.events
export const getReplicaSetPodList = (params: { namespace: string; name: string }) => request.get('/k8s/replicaset/pods', { params })

// HPA
export const getHpaList = hpaApi.list
export const getHpaDetail = hpaApi.detail
export const getHpaYaml = hpaApi.getYaml
export const createHpa = hpaApi.create
export const updateHpa = hpaApi.updateYaml
export const deleteHpa = hpaApi.delete
export const getHpaEvents = hpaApi.events
export const pauseHpa = (params: { namespace: string; name: string }) => request.post('/k8s/hpa/pause', null, { params })
export const resumeHpa = (params: { namespace: string; name: string }) => request.post('/k8s/hpa/resume', null, { params })
