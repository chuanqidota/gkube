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

export function getPodList(params?: { namespace?: string; cluster_id?: number; labelSelector?: string }) {
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

/**
 * Extract namespace names from API response.
 * Handles both formats:
 * 1. Simple string array: ["default", "kube-system", ...]
 * 2. Full K8s object: { namespaces: { items: [{metadata: {name: "default"}}, ...] } }
 */
export function extractNamespaceNames(data: any): string[] {
  if (Array.isArray(data)) {
    return data
  }
  if (data?.namespaces?.items) {
    return data.namespaces.items
      .map((ns: any) => ns.metadata?.name)
      .filter(Boolean)
  }
  return []
}

/**
 * Calculate age string from a creation timestamp.
 */
function calcAge(creationTimestamp: string): string {
  if (!creationTimestamp) return ''
  const created = new Date(creationTimestamp).getTime()
  const now = Date.now()
  const diff = Math.floor((now - created) / 1000)
  if (diff < 60) return `${diff}s`
  if (diff < 3600) return `${Math.floor(diff / 60)}m`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h`
  const days = Math.floor(diff / 86400)
  if (days < 365) return `${days}d`
  return `${Math.floor(days / 365)}y`
}

/**
 * Transform raw K8s Pod objects into simplified display format.
 */
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
      restarts,
      age: calcAge(pod.metadata?.creationTimestamp),
    }
  })
}

/**
 * Transform raw K8s Deployment objects into simplified display format.
 */
export function transformDeployments(items: any[]): Deployment[] {
  if (!Array.isArray(items)) return []
  return items.map((d: any) => ({
    name: d.metadata?.name || '',
    namespace: d.metadata?.namespace || '',
    ready: `${d.status?.readyReplicas || 0}/${d.spec?.replicas || 0}`,
    up_to_date: d.status?.updatedReplicas || 0,
    available: d.status?.availableReplicas || 0,
    age: calcAge(d.metadata?.creationTimestamp),
  }))
}

/**
 * Transform raw K8s Service objects into simplified display format.
 */
export function transformServices(items: any[]): Service[] {
  if (!Array.isArray(items)) return []
  return items.map((svc: any) => {
    const ports = (svc.spec?.ports || [])
      .map((p: any) => {
        let s = `${p.port}`
        if (p.targetPort && p.targetPort !== p.port) s += `:${p.targetPort}`
        if (p.nodePort) s += `/${p.nodePort}`
        if (p.protocol && p.protocol !== 'TCP') s += `/${p.protocol}`
        return s
      })
      .join(', ')
    const externalIps = svc.spec?.externalIPs?.join(', ') || svc.status?.loadBalancer?.ingress?.map((i: any) => i.ip || i.hostname).join(', ') || ''
    return {
      name: svc.metadata?.name || '',
      namespace: svc.metadata?.namespace || '',
      type: svc.spec?.type || 'ClusterIP',
      cluster_ip: svc.spec?.clusterIP || '',
      external_ip: externalIps,
      ports,
      age: calcAge(svc.metadata?.creationTimestamp),
    }
  })
}

/**
 * Transform raw K8s Ingress objects into simplified display format.
 */
export function transformIngresses(items: any[]): Ingress[] {
  if (!Array.isArray(items)) return []
  return items.map((ing: any) => {
    const hosts = (ing.spec?.rules || []).map((r: any) => r.host || '*').join(', ')
    const address = ing.status?.loadBalancer?.ingress?.map((i: any) => i.ip || i.hostname).join(', ') || ''
    return {
      name: ing.metadata?.name || '',
      namespace: ing.metadata?.namespace || '',
      hosts,
      address,
      age: calcAge(ing.metadata?.creationTimestamp),
    }
  })
}

/**
 * Transform raw K8s ConfigMap objects into simplified display format.
 */
export function transformConfigMaps(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((cm: any) => ({
    name: cm.metadata?.name || '',
    namespace: cm.metadata?.namespace || '',
    data: cm.data ? Object.keys(cm.data).length : 0,
    age: calcAge(cm.metadata?.creationTimestamp),
  }))
}

/**
 * Transform raw K8s Secret objects into simplified display format.
 */
export function transformSecrets(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((s: any) => ({
    name: s.metadata?.name || '',
    namespace: s.metadata?.namespace || '',
    type: s.type || 'Opaque',
    data: s.data ? Object.keys(s.data).length : 0,
    age: calcAge(s.metadata?.creationTimestamp),
  }))
}

export function getNamespaceDetail(params: { name: string }) {
  return request.get('/k8s/namespace/detail', { params })
}

export function getNamespaceYaml(params: { name: string }) {
  return request.get('/k8s/namespace/get-yaml', { params })
}

export function updateNamespace(data: { yamlContent: string }) {
  return request.put('/k8s/namespace/update', data)
}

export function deleteNamespace(params: { name: string }) {
  return request.delete('/k8s/namespace/delete', { params })
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

export function getPodLogs(params: { namespace: string; name: string; container?: string; tailLines?: number }) {
  return request.get('/k8s/log', { params })
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

// 获取 Deployment 关联的 ReplicaSet 列表
export const getDeploymentReplicaSets = (params: { namespace: string; name: string }) => {
  return request.get('/k8s/deployment/replicasets', { params })
}

// 获取 Deployment 关联的 Pod 列表
export const getDeploymentPodList = (params: { namespace: string; name: string }) => {
  return request.get('/k8s/deployment/pods', { params })
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

export function createStorageClass(data: any) {
  return request.post('/k8s/storageclass/create', data)
}

export function deleteStorageClass(params: { name: string }) {
  return request.delete('/k8s/storageclass/delete', { params })
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

// PDB
export function getPdbList(params?: { namespace?: string }) {
  return request.get('/k8s/pdb/list', { params })
}

export function getPdbDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/pdb/detail', { params })
}

export function getPdbYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/pdb/yaml', { params })
}

export function createPdb(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/pdb/create', data)
}

export function updatePdb(data: { namespace: string; yamlContent: string }) {
  return request.put('/k8s/pdb/update', data)
}

export function deletePdb(params: { namespace: string; name: string }) {
  return request.delete('/k8s/pdb/delete', { params })
}

// ResourceQuota
export function getResourceQuotaList(params?: { namespace?: string }) {
  return request.get('/k8s/resourcequota/list', { params })
}

export function getResourceQuotaDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/resourcequota/detail', { params })
}

export function getResourceQuotaYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/resourcequota/get-yaml', { params })
}

export function createResourceQuota(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/resourcequota/create', data)
}

export function updateResourceQuota(data: { namespace: string; yamlContent: string }) {
  return request.put('/k8s/resourcequota/update', data)
}

export function deleteResourceQuota(params: { namespace: string; name: string }) {
  return request.delete('/k8s/resourcequota/delete', { params })
}

// LimitRange
export function getLimitRangeList(params?: { namespace?: string }) {
  return request.get('/k8s/limitrange/list', { params })
}

export function getLimitRangeDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/limitrange/detail', { params })
}

export function getLimitRangeYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/limitrange/get-yaml', { params })
}

export function createLimitRange(data: { namespace: string; yamlContent: string }) {
  return request.post('/k8s/limitrange/create', data)
}

export function updateLimitRange(data: { namespace: string; yamlContent: string }) {
  return request.put('/k8s/limitrange/update', data)
}

export function deleteLimitRange(params: { namespace: string; name: string }) {
  return request.delete('/k8s/limitrange/delete', { params })
}

// CRD
export function getCrdList() {
  return request.get('/k8s/crd/list')
}

export function getCrdDetail(params: { name: string }) {
  return request.get('/k8s/crd/detail', { params })
}

export function getCrdYaml(params: { name: string }) {
  return request.get('/k8s/crd/get-yaml', { params })
}

export function createCrd(data: { yamlContent: string }) {
  return request.post('/k8s/crd/create', data)
}

export function updateCrd(data: { yamlContent: string }) {
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

export function createCustomResource(data: { group: string; version: string; resource: string; namespace?: string; yamlContent: string }) {
  return request.post('/k8s/crd/resource/create', data)
}

export function deleteCustomResource(params: { group: string; version: string; resource: string; namespace?: string; name: string }) {
  return request.delete('/k8s/crd/resource', { params })
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
