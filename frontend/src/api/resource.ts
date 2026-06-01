import request from './request'

export interface Pod {
  name: string
  namespace: string
  status: string
  node: string
  ip: string
  restarts: number
  age: string
}

export interface Deployment {
  name: string
  namespace: string
  ready: string
  up_to_date: number
  available: number
  age: string
}

export interface Service {
  name: string
  namespace: string
  type: string
  cluster_ip: string
  external_ip: string
  ports: string
  age: string
}

export interface Namespace {
  name: string
  status: string
  age: string
}

export interface Ingress {
  name: string
  namespace: string
  hosts: string
  address: string
  age: string
}

export function getPodList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get<Pod[]>('/k8s/pod/list', { params })
}

export function getDeploymentList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get<Deployment[]>('/k8s/deployment/list', { params })
}

export function getServiceList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get<Service[]>('/k8s/service/list', { params })
}

export function getNamespaceList(params?: { cluster_id?: number }) {
  return request.get<Namespace[]>('/k8s/namespace/list', { params })
}

// Service
export function getServiceDetail(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/service/detail', { params })
}

export function getServiceYaml(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/service/get-yaml', { params })
}

export function deleteService(data: { clusterName: string; namespace: string; name: string }) {
  return request.delete('/k8s/service/delete', { data })
}

// Node
export function getNodeList(params?: { cluster_id?: number }) {
  return request.get('/k8s/cluster/nodes', { params })
}

export function getNodeYaml(params: { name: string; cluster_id?: number }) {
  return request.get('/k8s/node/get-yaml', { params })
}

export function cordonNode(data: { name: string; cordon: boolean; cluster_id?: number }) {
  return request.put('/k8s/node/cordon', data)
}

export function taintNode(data: { name: string; taints: any[]; cluster_id?: number }) {
  return request.put('/k8s/node/taint', data)
}

export function getNodePods(params: { name: string; cluster_id?: number }) {
  return request.get('/k8s/node/pods', { params })
}

// Namespace
export function createNamespace(data: { name: string; cluster_id?: number }) {
  return request.post('/k8s/namespace/create', data)
}

// Ingress
export function getIngressList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/ingress/list', { params })
}

export function getIngressDetail(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/ingress/detail', { params })
}

export function getIngressYaml(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/ingress/get-yaml', { params })
}

export function deleteIngress(data: { clusterName: string; namespace: string; name: string }) {
  return request.delete('/k8s/ingress/delete', { data })
}

// Service create
export function createService(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/service/create', data)
}

// Ingress create
export function createIngress(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/ingress/create', data)
}

// ConfigMap create
export function createConfigMap(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/configmap/create', data)
}

// Secret create
export function createSecret(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/secret/create', data)
}

// Pod detail / actions
export function getPodDetail(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/pod/detail', { params })
}

export function getPodYaml(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/pod/get-yaml', { params })
}

export function deletePod(data: { clusterName: string; namespace: string; name: string }) {
  return request.delete('/k8s/pod/delete', { data })
}

export function getPodEvents(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/pod/events', { params })
}

// Deployment create
export function createDeployment(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/deployment/create', data)
}

// Deployment detail / actions
export function getDeploymentDetail(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/deployment/detail', { params })
}

export function getDeploymentYaml(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/deployment/get-yaml', { params })
}

export function scaleDeployment(data: { clusterName: string; namespace: string; name: string; replicas: number }) {
  return request.put('/k8s/deployment/scale', data)
}

export function restartDeployment(data: { clusterName: string; namespace: string; name: string }) {
  return request.post('/k8s/deployment/restart', data)
}

export function rollbackDeployment(data: { clusterName: string; namespace: string; name: string; revision: number }) {
  return request.post('/k8s/deployment/rollback', data)
}

export function deleteDeployment(data: { clusterName: string; namespace: string; name: string }) {
  return request.delete('/k8s/deployment/delete', { data })
}

// Dashboard events
export function getDashboardEvents(params?: { clusterId?: number; type?: string; limit?: number }) {
  return request.get('/dashboard/events', { params })
}

// ConfigMap
export function getConfigMapList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/configmap/list', { params })
}

export function getConfigMapDetail(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/configmap/detail', { params })
}

export function getConfigMapYaml(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/configmap/get-yaml', { params })
}

export function deleteConfigMap(data: { clusterName: string; namespace: string; name: string }) {
  return request.delete('/k8s/configmap/delete', { data })
}

// Secret
export function getSecretList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/secret/list', { params })
}

export function getSecretDetail(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/secret/detail', { params })
}

export function getSecretYaml(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/secret/get-yaml', { params })
}

export function deleteSecret(data: { clusterName: string; namespace: string; name: string }) {
  return request.delete('/k8s/secret/delete', { data })
}

// PV
export function getPvDetail(params: { clusterName: string; name: string }) {
  return request.get('/k8s/pv/detail', { params })
}

export function getPvList(params?: { cluster_id?: number }) {
  return request.get('/k8s/pv/list', { params })
}

export function getPvYaml(params: { clusterName: string; name: string }) {
  return request.get('/k8s/pv/get-yaml', { params })
}

export function deletePv(data: { clusterName: string; name: string }) {
  return request.delete('/k8s/pv/delete', { data })
}

export function createPv(data: { clusterName: string; yamlContent: string }) {
  return request.post('/k8s/pv/create', data)
}

// PVC
export function getPvcDetail(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/pvc/detail', { params })
}

export function getPvcList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/pvc/list', { params })
}

export function getPvcYaml(params: { clusterName: string; namespace: string; name: string }) {
  return request.get('/k8s/pvc/get-yaml', { params })
}

export function deletePvc(data: { clusterName: string; namespace: string; name: string }) {
  return request.delete('/k8s/pvc/delete', { data })
}

export function createPvc(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/pvc/create', data)
}

// StorageClass
export function getStorageClassDetail(params: { clusterName: string; name: string }) {
  return request.get('/k8s/storageclass/detail', { params })
}

export function getStorageClassList(params?: { cluster_id?: number }) {
  return request.get('/k8s/storageclass/list', { params })
}

export function getStorageClassYaml(params: { clusterName: string; name: string }) {
  return request.get('/k8s/storageclass/get-yaml', { params })
}

// StatefulSet
export function createStatefulSet(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/statefulset/create', data)
}

export function getStatefulSetList(params: any) {
  return request.get('/k8s/statefulset/list', { params })
}
export function getStatefulSetDetail(params: any) {
  return request.get('/k8s/statefulset/detail', { params })
}
export function getStatefulSetYaml(params: any) {
  return request.get('/k8s/statefulset/get-yaml', { params })
}
export function deleteStatefulSet(data: any) {
  return request.delete('/k8s/statefulset/delete', { data })
}

// DaemonSet
export function createDaemonSet(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/daemonset/create', data)
}

export function getDaemonSetList(params: any) {
  return request.get('/k8s/daemonset/list', { params })
}
export function getDaemonSetDetail(params: any) {
  return request.get('/k8s/daemonset/detail', { params })
}
export function getDaemonSetYaml(params: any) {
  return request.get('/k8s/daemonset/get-yaml', { params })
}
export function deleteDaemonSet(data: any) {
  return request.delete('/k8s/daemonset/delete', { data })
}

// Job
export function createJob(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/job/create', data)
}

export function getJobList(params: any) {
  return request.get('/k8s/job/list', { params })
}
export function getJobDetail(params: any) {
  return request.get('/k8s/job/detail', { params })
}
export function getJobYaml(params: any) {
  return request.get('/k8s/job/get-yaml', { params })
}
export function deleteJob(data: any) {
  return request.delete('/k8s/job/delete', { data })
}

// CronJob
export function createCronJob(data: { clusterName: string; namespace: string; yamlContent: string }) {
  return request.post('/k8s/cronjob/create', data)
}

export function getCronJobList(params: any) {
  return request.get('/k8s/cronjob/list', { params })
}
export function getCronJobDetail(params: any) {
  return request.get('/k8s/cronjob/detail', { params })
}
export function getCronJobYaml(params: any) {
  return request.get('/k8s/cronjob/get-yaml', { params })
}
export function deleteCronJob(data: any) {
  return request.delete('/k8s/cronjob/delete', { data })
}
