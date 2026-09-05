import request from './request'
import { calcAge } from './core'

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

// ============ PV API ============

export const pvApi = {
  list:    (data?: any) => request.post('/k8s/pv/list', data),
  detail:  (params: { name: string }) => request.get('/k8s/pv/detail', { params }),
  getYaml: (params: { name: string }) => request.get('/k8s/pv/get-yaml', { params }),
  create:  (data: { yaml: string }) => request.post('/k8s/pv/create', data),
  updateYaml: (data: { name: string; yaml: string }) => request.put('/k8s/pv/update', data),
  delete:  (data: { name: string }) => request.delete('/k8s/pv/delete', { data }),
}

export const getPvList = pvApi.list
export const getPvDetail = pvApi.detail
export const getPvYaml = pvApi.getYaml
export const createPv = pvApi.create
export const updatePvYaml = pvApi.updateYaml
export const deletePv = pvApi.delete

// ============ PVC API ============

export const pvcApi = {
  list:    (data?: any) => request.post('/k8s/pvc/list', data),
  detail:  (params: { namespace: string; name: string }) => request.get('/k8s/pvc/detail', { params }),
  getYaml: (params: { namespace: string; name: string }) => request.get('/k8s/pvc/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/pvc/create', data),
  updateYaml: (data: { namespace: string; name?: string; yaml: string }) => request.put('/k8s/pvc/update', data),
  delete:  (data: { namespace: string; name: string }) => request.delete('/k8s/pvc/delete', { data }),
}

export const getPvcList = pvcApi.list
export const getPvcDetail = pvcApi.detail
export const getPvcYaml = pvcApi.getYaml
export const createPvc = pvcApi.create
export const updatePvcYaml = pvcApi.updateYaml
export const deletePvc = pvcApi.delete
export const getPvcListByStorageClass = (params: { storageClassName: string }) => request.get('/k8s/pvc/list-by-storageclass', { params })

// ============ StorageClass API ============

export const storageClassApi = {
  list:    (data?: any) => request.post('/k8s/storageclass/list', data),
  detail:  (params: { name: string }) => request.get('/k8s/storageclass/detail', { params }),
  getYaml: (params: { name: string }) => request.get('/k8s/storageclass/get-yaml', { params }),
  create:  (data: any) => request.post('/k8s/storageclass/create', data),
  updateYaml: (data: { name: string; yaml: string }) => request.put('/k8s/storageclass/update', data),
  delete:  (data: { name: string }) => request.delete('/k8s/storageclass/delete', { data }),
  events:  (params: { name: string }) => request.get('/k8s/storageclass/events', { params }),
}

export const getStorageClassList = storageClassApi.list
export const getStorageClassDetail = storageClassApi.detail
export const getStorageClassYaml = storageClassApi.getYaml
export const createStorageClass = storageClassApi.create
export const updateStorageClass = storageClassApi.updateYaml
export const deleteStorageClass = storageClassApi.delete
export const getStorageClassEvents = storageClassApi.events

// ============ VolumeSnapshot API ============

export const volumeSnapshotApi = {
  list:    (data?: any) => request.post('/k8s/volumesnapshot/list', data),
  detail:  (params: { namespace: string; name: string }) => request.get('/k8s/volumesnapshot/detail', { params }),
  getYaml: (params: { namespace: string; name: string }) => request.get('/k8s/volumesnapshot/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/volumesnapshot/create', data),
  updateYaml: (data: { namespace: string; yaml: string }) => request.put('/k8s/volumesnapshot/update', data),
  delete:  (data: { namespace: string; name: string }) => request.delete('/k8s/volumesnapshot/delete', { data }),
}

export const getVolumeSnapshotList = volumeSnapshotApi.list
export const getVolumeSnapshotDetail = volumeSnapshotApi.detail
export const getVolumeSnapshotYaml = volumeSnapshotApi.getYaml
export const createVolumeSnapshot = volumeSnapshotApi.create
export const updateVolumeSnapshot = volumeSnapshotApi.updateYaml
export const deleteVolumeSnapshot = volumeSnapshotApi.delete

// ============ VolumeSnapshotClass API ============

export const volumeSnapshotClassApi = {
  list:    (data?: any) => request.post('/k8s/volumesnapshotclass/list', data),
  detail:  (params: { name: string }) => request.get('/k8s/volumesnapshotclass/detail', { params }),
  getYaml: (params: { name: string }) => request.get('/k8s/volumesnapshotclass/get-yaml', { params }),
  create:  (data: { yaml: string }) => request.post('/k8s/volumesnapshotclass/create', data),
  updateYaml: (data: { yaml: string }) => request.put('/k8s/volumesnapshotclass/update', data),
  delete:  (params: { name: string }) => request.delete('/k8s/volumesnapshotclass/delete', { params }),
}

export const getVolumeSnapshotClassList = volumeSnapshotClassApi.list
export const getVolumeSnapshotClassDetail = volumeSnapshotClassApi.detail
export const getVolumeSnapshotClassYaml = volumeSnapshotClassApi.getYaml
export const createVolumeSnapshotClass = volumeSnapshotClassApi.create
export const updateVolumeSnapshotClass = volumeSnapshotClassApi.updateYaml
export const deleteVolumeSnapshotClass = volumeSnapshotClassApi.delete
