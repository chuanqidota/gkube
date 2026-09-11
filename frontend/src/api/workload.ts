import request from './request'
import { formatAge } from '@/utils/helpers'
import { createResourceApi } from './factory'

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
      (sum: number, cs: any) => sum + (cs.restartCount || 0),
      0,
    )
    return {
      name: pod.metadata?.name || '',
      namespace: pod.metadata?.namespace || '',
      status: pod.status?.phase || 'Unknown',
      node: pod.spec?.nodeName || '',
      ip: pod.status?.podIP || '',
      hostIP: pod.status?.hostIP || '',
      restarts,
      age: formatAge(pod.metadata?.creationTimestamp, false),
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
    age: formatAge(d.metadata?.creationTimestamp, false),
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
    age: formatAge(d.metadata?.creationTimestamp, false),
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
    age: formatAge(d.metadata?.creationTimestamp, false),
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
    age: formatAge(d.metadata?.creationTimestamp, false),
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
    age: formatAge(d.metadata?.creationTimestamp, false),
  }))
}

// ============ 标准 CRUD（工厂生成） ============

export const podApi = createResourceApi('/k8s/pod', {
  updatePath: '/k8s/pod/update-yaml',
})

export const deploymentApi = createResourceApi('/k8s/deployment', {
  updatePath: '/k8s/deployment/update-yaml',
})

export const statefulSetApi = createResourceApi('/k8s/statefulset')

export const daemonSetApi = createResourceApi('/k8s/daemonset')

export const jobApi = createResourceApi('/k8s/job')

export const cronJobApi = createResourceApi('/k8s/cronjob')

export const replicaSetApi = createResourceApi('/k8s/replicaset', {
  deleteUseParams: true,
})

export const hpaApi = createResourceApi('/k8s/hpa', {
  deleteUseParams: true,
})

// ============ 向后兼容函数别名（deprecated） ============

// Pod
/** @deprecated 使用 podApi.list() */
export const getPodList = podApi.list
/** @deprecated 使用 podApi.detail() */
export const getPodDetail = podApi.detail
/** @deprecated 使用 podApi.getYaml() */
export const getPodYaml = podApi.getYaml
/** @deprecated 使用 podApi.delete() */
export const deletePod = podApi.delete
/** @deprecated 使用 podApi.events() */
export const getPodEvents = podApi.events

// Deployment
/** @deprecated 使用 deploymentApi.list() */
export const getDeploymentList = deploymentApi.list
/** @deprecated 使用 deploymentApi.detail() */
export const getDeploymentDetail = deploymentApi.detail
/** @deprecated 使用 deploymentApi.getYaml() */
export const getDeploymentYaml = deploymentApi.getYaml
/** @deprecated 使用 deploymentApi.create() */
export const createDeployment = deploymentApi.create
/** @deprecated 使用 deploymentApi.updateYaml() */
export const updateDeploymentYaml = deploymentApi.updateYaml
/** @deprecated 使用 deploymentApi.delete() */
export const deleteDeployment = deploymentApi.delete
/** @deprecated 使用 deploymentApi.events() */
export const getDeploymentEvents = deploymentApi.events

// StatefulSet
/** @deprecated 使用 statefulSetApi.list() */
export const getStatefulSetList = statefulSetApi.list
/** @deprecated 使用 statefulSetApi.detail() */
export const getStatefulSetDetail = statefulSetApi.detail
/** @deprecated 使用 statefulSetApi.getYaml() */
export const getStatefulSetYaml = statefulSetApi.getYaml
/** @deprecated 使用 statefulSetApi.create() */
export const createStatefulSet = statefulSetApi.create
/** @deprecated 使用 statefulSetApi.updateYaml() */
export const updateStatefulSetYaml = statefulSetApi.updateYaml
/** @deprecated 使用 statefulSetApi.delete() */
export const deleteStatefulSet = statefulSetApi.delete
/** @deprecated 使用 statefulSetApi.events() */
export const getStatefulSetEvents = statefulSetApi.events

// DaemonSet
/** @deprecated 使用 daemonSetApi.list() */
export const getDaemonSetList = daemonSetApi.list
/** @deprecated 使用 daemonSetApi.detail() */
export const getDaemonSetDetail = daemonSetApi.detail
/** @deprecated 使用 daemonSetApi.getYaml() */
export const getDaemonSetYaml = daemonSetApi.getYaml
/** @deprecated 使用 daemonSetApi.create() */
export const createDaemonSet = daemonSetApi.create
/** @deprecated 使用 daemonSetApi.updateYaml() */
export const updateDaemonSetYaml = daemonSetApi.updateYaml
/** @deprecated 使用 daemonSetApi.delete() */
export const deleteDaemonSet = daemonSetApi.delete
/** @deprecated 使用 daemonSetApi.events() */
export const getDaemonSetEvents = daemonSetApi.events

// Job
/** @deprecated 使用 jobApi.list() */
export const getJobList = jobApi.list
/** @deprecated 使用 jobApi.detail() */
export const getJobDetail = jobApi.detail
/** @deprecated 使用 jobApi.getYaml() */
export const getJobYaml = jobApi.getYaml
/** @deprecated 使用 jobApi.create() */
export const createJob = jobApi.create
/** @deprecated 使用 jobApi.updateYaml() */
export const updateJobYaml = jobApi.updateYaml
/** @deprecated 使用 jobApi.delete() */
export const deleteJob = jobApi.delete
/** @deprecated 使用 jobApi.events() */
export const getJobEvents = jobApi.events

// CronJob
/** @deprecated 使用 cronJobApi.list() */
export const getCronJobList = cronJobApi.list
/** @deprecated 使用 cronJobApi.detail() */
export const getCronJobDetail = cronJobApi.detail
/** @deprecated 使用 cronJobApi.getYaml() */
export const getCronJobYaml = cronJobApi.getYaml
/** @deprecated 使用 cronJobApi.create() */
export const createCronJob = cronJobApi.create
/** @deprecated 使用 cronJobApi.updateYaml() */
export const updateCronJobYaml = cronJobApi.updateYaml
/** @deprecated 使用 cronJobApi.delete() */
export const deleteCronJob = cronJobApi.delete
/** @deprecated 使用 cronJobApi.events() */
export const getCronJobEvents = cronJobApi.events

// ReplicaSet
/** @deprecated 使用 replicaSetApi.list() */
export const getReplicaSetList = replicaSetApi.list
/** @deprecated 使用 replicaSetApi.detail() */
export const getReplicaSetDetail = replicaSetApi.detail
/** @deprecated 使用 replicaSetApi.getYaml() */
export const getReplicaSetYaml = replicaSetApi.getYaml
/** @deprecated 使用 replicaSetApi.delete() */
export const deleteReplicaSet = replicaSetApi.delete
/** @deprecated 使用 replicaSetApi.events() */
export const getReplicaSetEvents = replicaSetApi.events

// HPA
/** @deprecated 使用 hpaApi.list() */
export const getHpaList = hpaApi.list
/** @deprecated 使用 hpaApi.detail() */
export const getHpaDetail = hpaApi.detail
/** @deprecated 使用 hpaApi.getYaml() */
export const getHpaYaml = hpaApi.getYaml
/** @deprecated 使用 hpaApi.create() */
export const createHpa = hpaApi.create
/** @deprecated 使用 hpaApi.updateYaml() */
export const updateHpa = hpaApi.updateYaml
/** @deprecated 使用 hpaApi.delete() */
export const deleteHpa = hpaApi.delete
/** @deprecated 使用 hpaApi.events() */
export const getHpaEvents = hpaApi.events

// ============ 资源特有操作 ============

// Pod
export const getPodLogs = (params: {
  namespace: string
  name: string
  container?: string
  tailLines?: number
}) => request.get('/k8s/pod/logs', { params })

// Deployment
export const scaleDeployment = (data: { namespace: string; name: string; replicas: number }) =>
  request.put('/k8s/deployment/scale', data)
export const restartDeployment = (data: { namespace: string; name: string }) =>
  request.post('/k8s/deployment/restart', data)
export const rollbackDeployment = (data: { namespace: string; name: string; revision: number }) =>
  request.post('/k8s/deployment/rollback', data)
export const updateDeploymentImage = (data: {
  namespace: string
  name: string
  containerName: string
  image: string
}) => request.put('/k8s/deployment/update-image', data)
export const getDeploymentReplicaSets = (params: { namespace: string; name: string }) =>
  request.get('/k8s/deployment/replicasets', { params })
export const getDeploymentPodList = (params: { namespace: string; name: string }) =>
  request.get('/k8s/deployment/pods', { params })

// StatefulSet
export const scaleStatefulSet = (data: { namespace: string; name: string; replicas: number }) =>
  request.put('/k8s/statefulset/scale', data)
export const restartStatefulSet = (data: { namespace: string; name: string }) =>
  request.post('/k8s/statefulset/restart', data)
export const rollbackStatefulSet = (data: { namespace: string; name: string; revision: number }) =>
  request.post('/k8s/statefulset/rollback', data)
export const updateStatefulSetImage = (data: {
  namespace: string
  name: string
  containerName: string
  image: string
}) => request.put('/k8s/statefulset/update-image', data)
export const getStatefulSetRollbacks = (params: { namespace: string; name: string }) =>
  request.get('/k8s/statefulset/rollbacks', { params })
export const getStatefulSetPVCs = (params: { namespace: string; name: string }) =>
  request.get('/k8s/statefulset/pvcs', { params })
export const getStatefulSetPods = (params: { namespace: string; name: string }) =>
  request.get('/k8s/statefulset/pods', { params })

// DaemonSet
export const getDaemonSetPods = (params: { namespace: string; name: string }) =>
  request.get('/k8s/daemonset/pods', { params })
export const restartDaemonSet = (data: { namespace: string; name: string }) =>
  request.post('/k8s/daemonset/restart', data)
export const updateDaemonSetImage = (data: {
  namespace: string
  name: string
  containerName: string
  image: string
}) => request.put('/k8s/daemonset/update-image', data)
export const rollbackDaemonSet = (data: { namespace: string; name: string; revision: number }) =>
  request.post('/k8s/daemonset/rollback', data)
export const getDaemonSetRollbacks = (params: { namespace: string; name: string }) =>
  request.get('/k8s/daemonset/rollbacks', { params })

// Job
export const getJobPods = (params: { namespace: string; name: string }) =>
  request.get('/k8s/job/pods', { params })
export const rerunJob = (params: { namespace: string; name: string }) =>
  request.post('/k8s/job/rerun', undefined, { params })

// CronJob
export const getCronJobExecutionHistory = (params: { namespace: string; name: string }) =>
  request.get('/k8s/cronjob/jobs', { params })
export const suspendCronJob = (params: { namespace: string; name: string }) =>
  request.put('/k8s/cronjob/suspend', undefined, { params })
export const resumeCronJob = (params: { namespace: string; name: string }) =>
  request.put('/k8s/cronjob/resume', undefined, { params })
export const triggerCronJob = (params: { namespace: string; name: string }) =>
  request.post('/k8s/cronjob/trigger', undefined, { params })

// ReplicaSet
export const getReplicaSetPodList = (params: { namespace: string; name: string }) =>
  request.get('/k8s/replicaset/pods', { params })

// HPA
export const pauseHpa = (params: { namespace: string; name: string }) =>
  request.post('/k8s/hpa/pause', null, { params })
export const resumeHpa = (params: { namespace: string; name: string }) =>
  request.post('/k8s/hpa/resume', null, { params })
