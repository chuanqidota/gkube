import request from './request'
import { calcAge } from './core'
import { createResourceApi } from './factory'

// ============ 类型定义 ============

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

// ============ Transform 函数 ============

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

export function transformStorageClasses(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((sc: any) => {
    const annotations = sc.metadata?.annotations || {}
    const isDefault = annotations['storageclass.kubernetes.io/is-default-class'] === 'true' ||
                      annotations['storageclass.beta.kubernetes.io/is-default-class'] === 'true'
    return {
      name: sc.metadata?.name || '',
      provisioner: sc.provisioner || '-',
      reclaim_policy: sc.reclaimPolicy || '-',
      volume_binding_mode: sc.volumeBindingMode || '-',
      default: isDefault,
      age: calcAge(sc.metadata?.creationTimestamp),
    }
  })
}

// ============ 标准 CRUD（工厂生成） ============

export const pvApi = createResourceApi('/k8s/pv', {
  deleteUseParams: true,
})

export const pvcApi = createResourceApi('/k8s/pvc')

export const storageClassApi = createResourceApi('/k8s/storageclass')

export const volumeSnapshotApi = createResourceApi('/k8s/volumesnapshot')

export const volumeSnapshotClassApi = createResourceApi('/k8s/volumesnapshotclass', {
  deleteUseParams: true,
})

// ============ 向后兼容函数别名（deprecated） ============

// PV
/** @deprecated 使用 pvApi.list() */
export const getPvList = pvApi.list
/** @deprecated 使用 pvApi.detail() */
export const getPvDetail = pvApi.detail
/** @deprecated 使用 pvApi.getYaml() */
export const getPvYaml = pvApi.getYaml
/** @deprecated 使用 pvApi.create() */
export const createPv = pvApi.create
/** @deprecated 使用 pvApi.updateYaml() */
export const updatePvYaml = pvApi.updateYaml
/** @deprecated 使用 pvApi.delete() */
export const deletePv = pvApi.delete

// PVC
/** @deprecated 使用 pvcApi.list() */
export const getPvcList = pvcApi.list
/** @deprecated 使用 pvcApi.detail() */
export const getPvcDetail = pvcApi.detail
/** @deprecated 使用 pvcApi.getYaml() */
export const getPvcYaml = pvcApi.getYaml
/** @deprecated 使用 pvcApi.create() */
export const createPvc = pvcApi.create
/** @deprecated 使用 pvcApi.updateYaml() */
export const updatePvcYaml = pvcApi.updateYaml
/** @deprecated 使用 pvcApi.delete() */
export const deletePvc = pvcApi.delete

// StorageClass
/** @deprecated 使用 storageClassApi.list() */
export const getStorageClassList = storageClassApi.list
/** @deprecated 使用 storageClassApi.detail() */
export const getStorageClassDetail = storageClassApi.detail
/** @deprecated 使用 storageClassApi.getYaml() */
export const getStorageClassYaml = storageClassApi.getYaml
/** @deprecated 使用 storageClassApi.create() */
export const createStorageClass = storageClassApi.create
/** @deprecated 使用 storageClassApi.updateYaml() */
export const updateStorageClass = storageClassApi.updateYaml
/** @deprecated 使用 storageClassApi.delete() */
export const deleteStorageClass = storageClassApi.delete
/** @deprecated 使用 storageClassApi.events() */
export const getStorageClassEvents = storageClassApi.events

// VolumeSnapshot
/** @deprecated 使用 volumeSnapshotApi.list() */
export const getVolumeSnapshotList = volumeSnapshotApi.list
/** @deprecated 使用 volumeSnapshotApi.detail() */
export const getVolumeSnapshotDetail = volumeSnapshotApi.detail
/** @deprecated 使用 volumeSnapshotApi.getYaml() */
export const getVolumeSnapshotYaml = volumeSnapshotApi.getYaml
/** @deprecated 使用 volumeSnapshotApi.create() */
export const createVolumeSnapshot = volumeSnapshotApi.create
/** @deprecated 使用 volumeSnapshotApi.updateYaml() */
export const updateVolumeSnapshot = volumeSnapshotApi.updateYaml
/** @deprecated 使用 volumeSnapshotApi.delete() */
export const deleteVolumeSnapshot = volumeSnapshotApi.delete

// VolumeSnapshotClass
/** @deprecated 使用 volumeSnapshotClassApi.list() */
export const getVolumeSnapshotClassList = volumeSnapshotClassApi.list
/** @deprecated 使用 volumeSnapshotClassApi.detail() */
export const getVolumeSnapshotClassDetail = volumeSnapshotClassApi.detail
/** @deprecated 使用 volumeSnapshotClassApi.getYaml() */
export const getVolumeSnapshotClassYaml = volumeSnapshotClassApi.getYaml
/** @deprecated 使用 volumeSnapshotClassApi.create() */
export const createVolumeSnapshotClass = volumeSnapshotClassApi.create
/** @deprecated 使用 volumeSnapshotClassApi.updateYaml() */
export const updateVolumeSnapshotClass = volumeSnapshotClassApi.updateYaml
/** @deprecated 使用 volumeSnapshotClassApi.delete() */
export const deleteVolumeSnapshotClass = volumeSnapshotClassApi.delete

// ============ 资源特有操作 ============

export const getPvcListByStorageClass = (params: { storageClassName: string }) => request.get('/k8s/pvc/list-by-storageclass', { params })
