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
export function getServiceDetail(params: { name: string; namespace: string; cluster_id?: number }) {
  return request.get('/k8s/service/detail', { params })
}

export function getServiceYaml(params: { name: string; namespace: string; cluster_id?: number }) {
  return request.get('/k8s/service/get-yaml', { params })
}

export function deleteService(data: { name: string; namespace: string; cluster_id?: number }) {
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

export function getIngressYaml(params: { name: string; namespace: string; cluster_id?: number }) {
  return request.get('/k8s/ingress/get-yaml', { params })
}

export function deleteIngress(data: { name: string; namespace: string; cluster_id?: number }) {
  return request.delete('/k8s/ingress/delete', { data })
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

export function deleteDeployment(data: { clusterName: string; namespace: string; name: string }) {
  return request.delete('/k8s/deployment/delete', { data })
}

// ConfigMap
export function getConfigMapList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/configmap/list', { params })
}

export function getConfigMapDetail(params: { name: string; namespace: string; cluster_id?: number }) {
  return request.get('/k8s/configmap/detail', { params })
}

export function getConfigMapYaml(params: { name: string; namespace: string; cluster_id?: number }) {
  return request.get('/k8s/configmap/get-yaml', { params })
}

export function deleteConfigMap(data: { name: string; namespace: string; cluster_id?: number }) {
  return request.delete('/k8s/configmap/delete', { data })
}

// Secret
export function getSecretList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/secret/list', { params })
}

export function getSecretDetail(params: { name: string; namespace: string; cluster_id?: number }) {
  return request.get('/k8s/secret/detail', { params })
}

export function getSecretYaml(params: { name: string; namespace: string; cluster_id?: number }) {
  return request.get('/k8s/secret/get-yaml', { params })
}

export function deleteSecret(data: { name: string; namespace: string; cluster_id?: number }) {
  return request.delete('/k8s/secret/delete', { data })
}

// PV
export function getPvList(params?: { cluster_id?: number }) {
  return request.get('/k8s/pv/list', { params })
}

export function getPvYaml(params: { name: string; cluster_id?: number }) {
  return request.get('/k8s/pv/get-yaml', { params })
}

export function deletePv(data: { name: string; cluster_id?: number }) {
  return request.delete('/k8s/pv/delete', { data })
}

// PVC
export function getPvcList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/pvc/list', { params })
}

export function getPvcYaml(params: { name: string; namespace: string; cluster_id?: number }) {
  return request.get('/k8s/pvc/get-yaml', { params })
}

export function deletePvc(data: { name: string; namespace: string; cluster_id?: number }) {
  return request.delete('/k8s/pvc/delete', { data })
}

// StorageClass
export function getStorageClassList(params?: { cluster_id?: number }) {
  return request.get('/k8s/storageclass/list', { params })
}

export function getStorageClassYaml(params: { name: string; cluster_id?: number }) {
  return request.get('/k8s/storageclass/get-yaml', { params })
}
