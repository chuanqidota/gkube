import request from './request'

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
  labels: Record<string, string>
  annotations: Record<string, string>
  age: string
}

export interface Ingress {
  name: string
  namespace: string
  hosts: string
  address: string
  age: string
}

export interface Pv {
  name: string
  capacity: string
  access_modes: string
  status: string
  claim: string
  storage_class: string
  reclaim_policy: string
  volume_mode: string
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
 */
export function calcAge(creationTimestamp: string): string {
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
      hostIP: pod.status?.hostIP || '',
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

/**
 * Transform namespace list response (already flat from backend) into display format.
 */
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

/**
 * Transform raw K8s PersistentVolume objects into simplified display format.
 */
export function transformPvs(items: any[]): Pv[] {
  if (!Array.isArray(items)) return []
  return items.map((pv: any) => {
    const capacity = pv.spec?.capacity?.storage || '-'
    const accessModes = (pv.spec?.accessModes || []).join(', ')
    const status = pv.status?.phase || 'Unknown'
    const claimRef = pv.spec?.claimRef
    const claim = claimRef ? `${claimRef.namespace}/${claimRef.name}` : '-'
    const storageClass = pv.spec?.storageClassName || '-'
    const reclaimPolicy = pv.spec?.persistentVolumeReclaimPolicy || '-'
    const volumeMode = pv.spec?.volumeMode || '-'
    return {
      name: pv.metadata?.name || '',
      capacity,
      access_modes: accessModes,
      status,
      claim,
      storage_class: storageClass,
      reclaim_policy: reclaimPolicy,
      volume_mode: volumeMode,
      age: calcAge(pv.metadata?.creationTimestamp),
    }
  })
}

export function getNamespaceDetail(params: { name: string }) {
  return request.get('/k8s/namespace/detail', { params })
}

export function getNamespaceYaml(params: { name: string }) {
  return request.get('/k8s/namespace/get-yaml', { params })
}

export function updateNamespace(data: { yaml: string }) {
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

export function updateService(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/service/update', data)
}

export function updateServiceYaml(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/service/update', data)
}

export function getServiceEvents(params: { namespace: string; name: string }) {
  return request.get('/k8s/service/events', { params })
}

export function getServicePods(params: { namespace: string; name: string }) {
  return request.get('/k8s/service/pods', { params })
}

// Node
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
  capacity_cpu: string
  capacity_memory: string
  allocatable_cpu: string
  allocatable_mem: string
  os_image: string
  kernel_version: string
  container_runtime: string
  age: string
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
  age: string
}

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

export function updateNodeTaints(data: { name: string; taints: { key: string; value: string; effect: string }[] }) {
  return request.put('/k8s/node/taints', data)
}

export function updateNodeLabels(data: { name: string; labels: Record<string, string> }) {
  return request.put('/k8s/node/labels', data)
}

export function drainNode(data: {
  name: string
  ignoreDaemonSets?: boolean
  deleteLocalData?: boolean
  gracePeriod?: number
  force?: boolean
}) {
  return request.put('/k8s/node/drain', data)
}

export function deleteNode(data: { name: string }) {
  return request.delete('/k8s/node/delete', { data })
}

export function getNodePods(params: { name: string }) {
  return request.get('/k8s/node/pods', { params })
}

export function getNodeEvents(params: { name: string }) {
  return request.get('/k8s/node/events', { params })
}

// Namespace
export function createNamespace(data: {
  namespace: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
}) {
  return request.post('/k8s/namespace/create', data)
}

export function updateNamespaceLabels(data: {
  namespace: string
  labels: Record<string, string>
}) {
  return request.put('/k8s/namespace/labels', data)
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

export function updateIngress(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/ingress/update', data)
}

export function updateIngressYaml(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/ingress/update', data)
}

export function getIngressEvents(params: { namespace: string; name: string }) {
  return request.get('/k8s/ingress/events', { params })
}

// Service create
export function createService(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/service/create', data)
}

// Ingress create
export function createIngress(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/ingress/create', data)
}

// ConfigMap create
export function createConfigMap(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/configmap/create', data)
}

// ConfigMap update
export function updateConfigMap(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/configmap/update', data)
}

// Secret create
export function createSecret(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/secret/create', data)
}

// Secret update
export function updateSecret(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/secret/update', data)
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

export function getPodLogs(params: { namespace: string; podName: string; container?: string; tailLines?: number }) {
  return request.get('/k8s/log', { params })
}

// Deployment create
export function createDeployment(data: { namespace: string; yaml: string }) {
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

export function updateDeploymentImage(data: { namespace: string; name: string; containerName: string; image: string }) {
  return request.put('/k8s/deployment/update-image', data)
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

export function createPv(data: { yaml: string }) {
  return request.post('/k8s/pv/create', data)
}

export function updatePvYaml(data: { name: string; yaml: string }) {
  return request.put('/k8s/pv/update', data)
}

/**
 * Transform raw K8s PersistentVolumeClaim objects into simplified display format.
 */
export function transformPvcs(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((pvc: any) => ({
    name: pvc.metadata?.name || '',
    namespace: pvc.metadata?.namespace || '',
    status: pvc.status?.phase || 'Unknown',
    volume: pvc.spec?.volumeName || '-',
    capacity: pvc.status?.capacity?.storage || '-',
    storage_class: pvc.spec?.storageClassName || '-',
    access_modes: (pvc.spec?.accessModes || []).join(', '),
    age: calcAge(pvc.metadata?.creationTimestamp),
  }))
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

export function createPvc(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/pvc/create', data)
}

export function updatePvcYaml(data: { namespace: string; yaml: string }) {
  return request.put('/k8s/pvc/update', data)
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

export function updateStorageClass(data: { name: string; yaml: string }) {
  return request.put('/k8s/storageclass/update', data)
}

export const updateStorageClassYaml = updateStorageClass

export function deleteStorageClass(data: { name: string }) {
  return request.delete('/k8s/storageclass/delete', { data })
}

export function getStorageClassEvents(params: { name: string }) {
  return request.get('/k8s/storageclass/events', { params })
}

// VolumeSnapshot
export function getVolumeSnapshotList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/volumesnapshot/list', { params })
}

export function getVolumeSnapshotDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/volumesnapshot/detail', { params })
}

export function getVolumeSnapshotYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/volumesnapshot/get-yaml', { params })
}

export function createVolumeSnapshot(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/volumesnapshot/create', data)
}

export function updateVolumeSnapshot(data: { namespace: string; yaml: string }) {
  return request.put('/k8s/volumesnapshot/update', data)
}

export function deleteVolumeSnapshot(data: { namespace: string; name: string }) {
  return request.delete('/k8s/volumesnapshot/delete', { data })
}

// VolumeSnapshotClass
export function getVolumeSnapshotClassList(params?: { cluster_id?: number }) {
  return request.get('/k8s/volumesnapshotclass/list', { params })
}

export function getVolumeSnapshotClassDetail(params: { name: string }) {
  return request.get('/k8s/volumesnapshotclass/detail', { params })
}

export function getVolumeSnapshotClassYaml(params: { name: string }) {
  return request.get('/k8s/volumesnapshotclass/get-yaml', { params })
}

export function createVolumeSnapshotClass(data: { yaml: string }) {
  return request.post('/k8s/volumesnapshotclass/create', data)
}

export function updateVolumeSnapshotClass(data: { yaml: string }) {
  return request.put('/k8s/volumesnapshotclass/update', data)
}

export function deleteVolumeSnapshotClass(params: { name: string }) {
  return request.delete('/k8s/volumesnapshotclass/delete', { params })
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

export function createHpa(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/hpa/create', data)
}

export function updateHpa(data: { namespace: string; yaml: string }) {
  return request.put('/k8s/hpa/update', data)
}

export function deleteHpa(params: { namespace: string; name: string }) {
  return request.delete('/k8s/hpa/delete', { params })
}

// NetworkPolicy
export function getNetworkPolicyList(params?: { namespace?: string }) {
  return request.get('/k8s/networkpolicy/list', { params })
}

/**
 * Transform raw K8s NetworkPolicy objects into simplified display format.
 */
export function transformNetworkPolicies(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((np: any) => ({
    name: np.metadata?.name || '',
    namespace: np.metadata?.namespace || '',
    pod_selector: Object.entries(np.spec?.podSelector?.matchLabels || {}).map(([k, v]) => `${k}=${v}`).join(', ') || 'All',
    policy_types: np.spec?.policyTypes || [],
    ingress_rules: np.spec?.ingress?.length || 0,
    egress_rules: np.spec?.egress?.length || 0,
    age: calcAge(np.metadata?.creationTimestamp),
  }))
}

export function getNetworkPolicyDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/networkpolicy/detail', { params })
}

export function getNetworkPolicyYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/networkpolicy/yaml', { params })
}

export function createNetworkPolicy(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/networkpolicy/create', data)
}

export function updateNetworkPolicy(data: { namespace: string; yaml: string }) {
  return request.put('/k8s/networkpolicy/update', data)
}

export function updateNetworkPolicyYaml(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/networkpolicy/update', data)
}

export function deleteNetworkPolicy(params: { namespace: string; name: string }) {
  return request.delete('/k8s/networkpolicy/delete', { params })
}

export function getNetworkPolicyEvents(params: { namespace: string; name: string }) {
  return request.get('/k8s/networkpolicy/events', { params })
}

export function getNetworkPolicyPods(params: { namespace: string; name: string }) {
  return request.get('/k8s/networkpolicy/pods', { params })
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

export function createClusterRole(data: { yaml: string }) {
  return request.post('/k8s/clusterrole/create', data)
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

export function createRole(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/role/create', data)
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

export function createClusterRoleBinding(data: { yaml: string }) {
  return request.post('/k8s/clusterrolebinding/create', data)
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

export function createRoleBinding(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/rolebinding/create', data)
}

export function deleteRoleBinding(params: { namespace: string; name: string }) {
  return request.delete('/k8s/rolebinding/delete', { params })
}

// ReplicaSet
export function getReplicaSetList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/replicaset/list', { params })
}

export function getReplicaSetYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/replicaset/yaml', { params })
}

export function deleteReplicaSet(params: { namespace: string; name: string }) {
  return request.delete('/k8s/replicaset/delete', { params })
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

export function createPdb(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/pdb/create', data)
}

export function updatePdb(data: { namespace: string; yaml: string }) {
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

export function createResourceQuota(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/resourcequota/create', data)
}

export function updateResourceQuota(data: { namespace: string; yaml: string }) {
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

export function createLimitRange(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/limitrange/create', data)
}

export function updateLimitRange(data: { namespace: string; yaml: string }) {
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

export function createCustomResource(data: { group: string; version: string; resource: string; namespace?: string; yaml: string }) {
  return request.post('/k8s/crd/resource/create', data)
}

export function deleteCustomResource(params: { group: string; version: string; resource: string; namespace?: string; name: string }) {
  return request.delete('/k8s/crd/resource', { params })
}

// StatefulSet
export interface StatefulSet {
  name: string
  namespace: string
  ready: string
  age: string
  serviceName: string
  updateStrategy: string
}

/**
 * Transform raw K8s StatefulSet objects into simplified display format.
 */
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

export function createStatefulSet(data: { namespace: string; yaml: string }) {
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
export function updateStatefulSetYaml(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/statefulset/update', data)
}
export function deleteStatefulSet(data: any) {
  return request.delete('/k8s/statefulset/delete', { data })
}

export function scaleStatefulSet(data: { namespace: string; name: string; replicas: number }) {
  return request.put('/k8s/statefulset/scale', data)
}

export function restartStatefulSet(data: { namespace: string; name: string }) {
  return request.put('/k8s/statefulset/restart', data)
}

export function getStatefulSetEvents(params: { namespace: string; name: string }) {
  return request.get('/k8s/statefulset/events', { params })
}
export function getStatefulSetPods(params: { namespace: string; name: string }) {
  return request.get('/k8s/statefulset/pods', { params })
}

// DaemonSet
export interface DaemonSet {
  name: string
  namespace: string
  desired: number
  current: number
  ready: number
  age: string
  updateStrategy: string
}

/**
 * Transform raw K8s DaemonSet objects into simplified display format.
 */
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

export function createDaemonSet(data: { namespace: string; yaml: string }) {
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
export function updateDaemonSetYaml(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/daemonset/update', data)
}
export function deleteDaemonSet(data: any) {
  return request.delete('/k8s/daemonset/delete', { data })
}
export function getDaemonSetEvents(params: { namespace: string; name: string }) {
  return request.get('/k8s/daemonset/events', { params })
}
export function getDaemonSetPods(params: { namespace: string; name: string }) {
  return request.get('/k8s/daemonset/pods', { params })
}
export function restartDaemonSet(data: { namespace: string; name: string }) {
  return request.post('/k8s/daemonset/restart', data)
}

// Job
export interface Job {
  name: string
  namespace: string
  completions: string
  succeeded: number
  age: string
}

/**
 * Transform raw K8s Job objects into simplified display format.
 */
export function transformJobs(items: any[]): Job[] {
  if (!Array.isArray(items)) return []
  return items.map((d: any) => ({
    name: d.metadata?.name || '',
    namespace: d.metadata?.namespace || '',
    completions: `${d.status?.succeeded || 0}/${d.spec?.completions || 1}`,
    succeeded: d.status?.succeeded || 0,
    age: calcAge(d.metadata?.creationTimestamp),
  }))
}

export function createJob(data: { namespace: string; yaml: string }) {
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
export function updateJobYaml(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/job/update', data)
}
export function deleteJob(data: any) {
  return request.delete('/k8s/job/delete', { data })
}
export function getJobEvents(params: { namespace: string; name: string }) {
  return request.get('/k8s/job/events', { params })
}
export function getJobPods(params: { namespace: string; name: string }) {
  return request.get('/k8s/job/pods', { params })
}

// CronJob
export interface CronJob {
  name: string
  namespace: string
  schedule: string
  suspend: boolean
  active: number
  lastSchedule: string
  age: string
}

/**
 * Transform raw K8s CronJob objects into simplified display format.
 */
export function transformCronJobs(items: any[]): CronJob[] {
  if (!Array.isArray(items)) return []
  return items.map((d: any) => ({
    name: d.metadata?.name || '',
    namespace: d.metadata?.namespace || '',
    schedule: d.spec?.schedule || '',
    suspend: d.spec?.suspend || false,
    active: d.status?.active?.length || 0,
    lastSchedule: d.status?.lastScheduleTime || '',
    age: calcAge(d.metadata?.creationTimestamp),
  }))
}

export function createCronJob(data: { namespace: string; yaml: string }) {
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
export function updateCronJobYaml(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/cronjob/update', data)
}
export function deleteCronJob(data: any) {
  return request.delete('/k8s/cronjob/delete', { data })
}
export function getCronJobEvents(params: { namespace: string; name: string }) {
  return request.get('/k8s/cronjob/events', { params })
}
export function getCronJobJobs(params: { namespace: string; name: string }) {
  return request.get('/k8s/cronjob/jobs', { params })
}

// ReplicaSet detail
export function getReplicaSetDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/replicaset/detail', { params })
}

// CRD custom resource update/patch
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

// Event API
export function getEventList(params: {
  namespace?: string
  fieldSelector?: string
  limit?: number
  continue?: string
}) {
  return request.get('/k8s/event/list', { params })
}

// CronJob operations
export function suspendCronJob(params: { namespace: string; name: string }) {
  return request.put('/k8s/cronjob/suspend', undefined, { params })
}

export function resumeCronJob(params: { namespace: string; name: string }) {
  return request.put('/k8s/cronjob/resume', undefined, { params })
}

export function triggerCronJob(params: { namespace: string; name: string }) {
  return request.post('/k8s/cronjob/trigger', undefined, { params })
}

// StatefulSet rollback + update-image
export function rollbackStatefulSet(data: { namespace: string; name: string; revision: number }) {
  return request.post('/k8s/statefulset/rollback', data)
}

export function updateStatefulSetImage(data: { namespace: string; name: string; containerName: string; image: string }) {
  return request.put('/k8s/statefulset/update-image', data)
}

// DaemonSet rollback + update-image
export function rollbackDaemonSet(data: { namespace: string; name: string; revision: number }) {
  return request.post('/k8s/daemonset/rollback', data)
}

export function updateDaemonSetImage(data: { namespace: string; name: string; containerName: string; image: string }) {
  return request.put('/k8s/daemonset/update-image', data)
}

// RBAC detail methods
export function getServiceAccountDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/serviceaccount/detail', { params })
}

export function createServiceAccount(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/serviceaccount/create', data)
}

export function updateServiceAccount(data: { yaml: string }) {
  return request.put('/k8s/serviceaccount/update', data)
}

export function getRoleDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/role/detail', { params })
}

export function updateRole(data: { yaml: string }) {
  return request.put('/k8s/role/update', data)
}

export function getClusterRoleDetail(params: { name: string }) {
  return request.get('/k8s/clusterrole/detail', { params })
}

export function updateClusterRole(data: { yaml: string }) {
  return request.put('/k8s/clusterrole/update', data)
}

export function getRoleBindingDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/rolebinding/detail', { params })
}

export function updateRoleBinding(data: { yaml: string }) {
  return request.put('/k8s/rolebinding/update', data)
}

export function getClusterRoleBindingDetail(params: { name: string }) {
  return request.get('/k8s/clusterrolebinding/detail', { params })
}

export function updateClusterRoleBinding(data: { yaml: string }) {
  return request.put('/k8s/clusterrolebinding/update', data)
}
