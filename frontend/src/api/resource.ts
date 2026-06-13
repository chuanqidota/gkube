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
export function getServiceDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/service/detail', { params })
}

export function getServiceYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/service/get-yaml', { params })
}

export function deleteService(data: { namespace: string; name: string }) {
  return request.delete('/k8s/service/delete', { data })
}

// Node
export function getNodeList(params?: { cluster_id?: number }) {
  return request.get('/k8s/cluster/nodes', { params })
}

export function getNodeDetail(params: { name: string }) {
  return request.get('/k8s/node/detail', { params })
}

export function getNodeYaml(params: { name: string }) {
  return request.get('/k8s/node/get-yaml', { params })
}

export function updateNodeYaml(data: { name: string; yaml: string }) {
  return request.put('/k8s/node/update-yaml', data)
}

export function cordonNode(data: { name: string; cordon: boolean }) {
  return request.put('/k8s/node/cordon', data)
}

export function taintNode(data: { name: string; taints: any[] }) {
  return request.put('/k8s/node/taint', data)
}

export function getNodePods(params: { name: string }) {
  return request.get('/k8s/node/pods', { params })
}

export function getNodeEvents(params: { name: string }) {
  return request.get('/k8s/node/events', { params })
}

// Namespace
export function createNamespace(data: { name: string; cluster_id?: number }) {
  return request.post('/k8s/namespace/create', data)
}

// Ingress
export function getIngressList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/ingress/list', { params })
}

export function getIngressDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/ingress/detail', { params })
}

export function getIngressYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/ingress/get-yaml', { params })
}

export function deleteIngress(data: { namespace: string; name: string }) {
  return request.delete('/k8s/ingress/delete', { data })
}

// Service create
export function createService(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/service/create', data)
}

// Ingress create
export function createIngress(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/ingress/create', data)
}

// ConfigMap create
export function createConfigMap(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/configmap/create', data)
}

// Secret create
export function createSecret(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/secret/create', data)
}

// Pod detail / actions
export function getPodDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/pod/detail', { params })
}

export function getPodYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/pod/get-yaml', { params })
}

export function updatePodYaml(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/pod/update-yaml', data)
}

export function deletePod(data: { namespace: string; name: string }) {
  return request.delete('/k8s/pod/delete', { data })
}

export function getPodEvents(params: { namespace: string; name: string }) {
  return request.get('/k8s/pod/events', { params })
}

// Deployment create
export function createDeployment(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/deployment/create', data)
}

// Deployment detail / actions
export function getDeploymentDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/deployment/detail', { params })
}

export function getDeploymentYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/deployment/get-yaml', { params })
}

export function updateDeploymentYaml(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/deployment/update-yaml', data)
}

export function getDeploymentEvents(params: { namespace: string; name: string }) {
  return request.get('/k8s/deployment/events', { params })
}

export function scaleDeployment(data: { namespace: string; name: string; replicas: number }) {
  return request.put('/k8s/deployment/scale', data)
}

export function restartDeployment(data: { namespace: string; name: string }) {
  return request.post('/k8s/deployment/restart', data)
}

export function rollbackDeployment(data: { namespace: string; name: string; revision: number }) {
  return request.post('/k8s/deployment/rollback', data)
}

export function deleteDeployment(data: { namespace: string; name: string }) {
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

export function getConfigMapDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/configmap/detail', { params })
}

export function getConfigMapYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/configmap/get-yaml', { params })
}

export function deleteConfigMap(data: { namespace: string; name: string }) {
  return request.delete('/k8s/configmap/delete', { data })
}

// Secret
export function getSecretList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/secret/list', { params })
}

export function getSecretDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/secret/detail', { params })
}

export function getSecretYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/secret/get-yaml', { params })
}

export function deleteSecret(data: { namespace: string; name: string }) {
  return request.delete('/k8s/secret/delete', { data })
}

// PV
export function getPvDetail(params: { name: string }) {
  return request.get('/k8s/pv/detail', { params })
}

export function getPvList(params?: { cluster_id?: number }) {
  return request.get('/k8s/pv/list', { params })
}

export function getPvYaml(params: { name: string }) {
  return request.get('/k8s/pv/get-yaml', { params })
}

export function deletePv(data: { name: string }) {
  return request.delete('/k8s/pv/delete', { data })
}

export function createPv(data: { yamlContent: string }) {
  return request.post('/k8s/pv/create', data)
}

// PVC
export function getPvcDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/pvc/detail', { params })
}

export function getPvcList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/pvc/list', { params })
}

export function getPvcYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/pvc/get-yaml', { params })
}

export function deletePvc(data: { namespace: string; name: string }) {
  return request.delete('/k8s/pvc/delete', { data })
}

export function createPvc(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/pvc/create', data)
}

// StorageClass
export function getStorageClassDetail(params: { name: string }) {
  return request.get('/k8s/storageclass/detail', { params })
}

export function getStorageClassList(params?: { cluster_id?: number }) {
  return request.get('/k8s/storageclass/list', { params })
}

export function getStorageClassYaml(params: { name: string }) {
  return request.get('/k8s/storageclass/get-yaml', { params })
}

// HPA
export function getHpaList(params?: { namespace?: string }) {
  return request.get('/k8s/hpa/list', { params })
}

export function getHpaDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/hpa/detail', { params })
}

export function getHpaYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/hpa/yaml', { params })
}

export function createHpa(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/hpa/create', data)
}

export function updateHpa(data: { namespace: string; yamlContent: string }) {
  return request.put('/k8s/hpa/update', data)
}

export function deleteHpa(params: { namespace: string; name: string }) {
  return request.delete('/k8s/hpa/delete', { params })
}

// NetworkPolicy
export function getNetworkPolicyList(params?: { namespace?: string }) {
  return request.get('/k8s/networkpolicy/list', { params })
}

export function getNetworkPolicyDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/networkpolicy/detail', { params })
}

export function getNetworkPolicyYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/networkpolicy/yaml', { params })
}

export function createNetworkPolicy(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/networkpolicy/create', data)
}

export function updateNetworkPolicy(data: { namespace: string; yamlContent: string }) {
  return request.put('/k8s/networkpolicy/update', data)
}

export function deleteNetworkPolicy(params: { namespace: string; name: string }) {
  return request.delete('/k8s/networkpolicy/delete', { params })
}

// RBAC - ServiceAccount
export function getServiceAccountList(params?: { namespace?: string }) {
  return request.get('/k8s/serviceaccount/list', { params })
}

export function getServiceAccountYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/serviceaccount/yaml', { params })
}

export function deleteServiceAccount(params: { namespace: string; name: string }) {
  return request.delete('/k8s/serviceaccount/delete', { params })
}

// RBAC - ClusterRole
export function getClusterRoleList() {
  return request.get('/k8s/clusterrole/list')
}

export function getClusterRoleYaml(params: { name: string }) {
  return request.get('/k8s/clusterrole/yaml', { params })
}

export function deleteClusterRole(params: { name: string }) {
  return request.delete('/k8s/clusterrole/delete', { params })
}

// RBAC - Role
export function getRoleList(params?: { namespace?: string }) {
  return request.get('/k8s/role/list', { params })
}

export function getRoleYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/role/yaml', { params })
}

export function deleteRole(params: { namespace: string; name: string }) {
  return request.delete('/k8s/role/delete', { params })
}

// RBAC - ClusterRoleBinding
export function getClusterRoleBindingList() {
  return request.get('/k8s/clusterrolebinding/list')
}

export function getClusterRoleBindingYaml(params: { name: string }) {
  return request.get('/k8s/clusterrolebinding/yaml', { params })
}

export function deleteClusterRoleBinding(params: { name: string }) {
  return request.delete('/k8s/clusterrolebinding/delete', { params })
}

// RBAC - RoleBinding
export function getRoleBindingList(params?: { namespace?: string }) {
  return request.get('/k8s/rolebinding/list', { params })
}

export function getRoleBindingYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/rolebinding/yaml', { params })
}

export function deleteRoleBinding(params: { namespace: string; name: string }) {
  return request.delete('/k8s/rolebinding/delete', { params })
}

// StatefulSet
export function createStatefulSet(data: { namespace: string; yamlContent: string }) {
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
export function createDaemonSet(data: { namespace: string; yamlContent: string }) {
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
export function createJob(data: { namespace: string; yamlContent: string }) {
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
export function createCronJob(data: { namespace: string; yamlContent: string }) {
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
