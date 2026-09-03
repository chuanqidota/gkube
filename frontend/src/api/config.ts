import request from './request'
import { calcAge } from './core'

// ============ Transform 函数 ============

export function transformConfigMaps(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((cm: any) => ({
    name: cm.metadata?.name || '',
    namespace: cm.metadata?.namespace || '',
    labels: cm.metadata?.labels || {},
    data_keys_count: (cm.data ? Object.keys(cm.data).length : 0) + (cm.binaryData ? Object.keys(cm.binaryData).length : 0),
    age: calcAge(cm.metadata?.creationTimestamp),
  }))
}

export function transformSecrets(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((s: any) => ({
    name: s.metadata?.name || '',
    namespace: s.metadata?.namespace || '',
    type: s.type || 'Opaque',
    data_keys_count: s.data ? Object.keys(s.data).length : 0,
    age: calcAge(s.metadata?.creationTimestamp),
  }))
}

// ============ ConfigMap API ============

export function getConfigMapList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/configmap/list', { params })
}

export function getConfigMapDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/configmap/detail', { params })
}

export function getConfigMapYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/configmap/get-yaml', { params })
}

export function createConfigMap(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/configmap/create', data, { timeout: 30000 })
}

export function updateConfigMap(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/configmap/update', data)
}

export function deleteConfigMap(data: { namespace: string; name: string }) {
  return request.delete('/k8s/configmap/delete', { data })
}

// ============ Secret API ============

export function getSecretList(params?: { namespace?: string; cluster_id?: number }) {
  return request.get('/k8s/secret/list', { params })
}

export function getSecretDetail(params: { namespace: string; name: string }) {
  return request.get('/k8s/secret/detail', { params })
}

export function getSecretYaml(params: { namespace: string; name: string }) {
  return request.get('/k8s/secret/get-yaml', { params })
}

export function createSecret(data: { namespace: string; yaml: string }) {
  return request.post('/k8s/secret/create', data, { timeout: 30000 })
}

export function updateSecret(data: { namespace: string; name: string; yaml: string }) {
  return request.put('/k8s/secret/update', data)
}

export function deleteSecret(data: { namespace: string; name: string }) {
  return request.delete('/k8s/secret/delete', { data })
}

// ============ ResourceQuota API ============

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

// ============ LimitRange API ============

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
