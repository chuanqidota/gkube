import { formatAge } from '@/utils/helpers'
import { createResourceApi } from './factory'

// ============ Transform 函数 ============

export function transformConfigMaps(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((cm: any) => ({
    name: cm.metadata?.name || '',
    namespace: cm.metadata?.namespace || '',
    labels: cm.metadata?.labels || {},
    data_keys_count:
      (cm.data ? Object.keys(cm.data).length : 0) +
      (cm.binaryData ? Object.keys(cm.binaryData).length : 0),
    age: formatAge(cm.metadata?.creationTimestamp, false),
  }))
}

export function transformSecrets(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((s: any) => ({
    name: s.metadata?.name || '',
    namespace: s.metadata?.namespace || '',
    type: s.type || 'Opaque',
    data_keys_count: s.data ? Object.keys(s.data).length : 0,
    age: formatAge(s.metadata?.creationTimestamp, false),
  }))
}

// ============ 标准 CRUD（工厂生成） ============

export const configMapApi = createResourceApi('/k8s/configmap')

export const secretApi = createResourceApi('/k8s/secret')

export const resourceQuotaApi = createResourceApi('/k8s/resourcequota', {
  deleteUseParams: true,
})

export const limitRangeApi = createResourceApi('/k8s/limitrange', {
  deleteUseParams: true,
})

// ============ 向后兼容函数别名（deprecated） ============

// ConfigMap
/** @deprecated 使用 configMapApi.list() */
export const getConfigMapList = configMapApi.list
/** @deprecated 使用 configMapApi.detail() */
export const getConfigMapDetail = configMapApi.detail
/** @deprecated 使用 configMapApi.getYaml() */
export const getConfigMapYaml = configMapApi.getYaml
/** @deprecated 使用 configMapApi.create() */
export const createConfigMap = configMapApi.create
/** @deprecated 使用 configMapApi.updateYaml() */
export const updateConfigMap = configMapApi.updateYaml
/** @deprecated 使用 configMapApi.delete() */
export const deleteConfigMap = configMapApi.delete

// Secret
/** @deprecated 使用 secretApi.list() */
export const getSecretList = secretApi.list
/** @deprecated 使用 secretApi.detail() */
export const getSecretDetail = secretApi.detail
/** @deprecated 使用 secretApi.getYaml() */
export const getSecretYaml = secretApi.getYaml
/** @deprecated 使用 secretApi.create() */
export const createSecret = secretApi.create
/** @deprecated 使用 secretApi.updateYaml() */
export const updateSecret = secretApi.updateYaml
/** @deprecated 使用 secretApi.delete() */
export const deleteSecret = secretApi.delete

// ResourceQuota
/** @deprecated 使用 resourceQuotaApi.list() */
export const getResourceQuotaList = resourceQuotaApi.list
/** @deprecated 使用 resourceQuotaApi.detail() */
export const getResourceQuotaDetail = resourceQuotaApi.detail
/** @deprecated 使用 resourceQuotaApi.getYaml() */
export const getResourceQuotaYaml = resourceQuotaApi.getYaml
/** @deprecated 使用 resourceQuotaApi.create() */
export const createResourceQuota = resourceQuotaApi.create
/** @deprecated 使用 resourceQuotaApi.updateYaml() */
export const updateResourceQuota = resourceQuotaApi.updateYaml
/** @deprecated 使用 resourceQuotaApi.delete() */
export const deleteResourceQuota = resourceQuotaApi.delete

// LimitRange
/** @deprecated 使用 limitRangeApi.list() */
export const getLimitRangeList = limitRangeApi.list
/** @deprecated 使用 limitRangeApi.detail() */
export const getLimitRangeDetail = limitRangeApi.detail
/** @deprecated 使用 limitRangeApi.getYaml() */
export const getLimitRangeYaml = limitRangeApi.getYaml
/** @deprecated 使用 limitRangeApi.create() */
export const createLimitRange = limitRangeApi.create
/** @deprecated 使用 limitRangeApi.updateYaml() */
export const updateLimitRange = limitRangeApi.updateYaml
/** @deprecated 使用 limitRangeApi.delete() */
export const deleteLimitRange = limitRangeApi.delete
