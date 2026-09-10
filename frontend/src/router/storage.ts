import type { RouteRecordRaw } from 'vue-router'

export const storageRoutes: RouteRecordRaw[] = [
  // PV
  {
    path: 'storage/pvs',
    name: 'PVList',
    component: () => import('@/views/storage/PVList.vue'),
    meta: { titleKey: 'sidebar.pvs', icon: 'Coin' },
  },
  {
    path: 'storage/pvs/create',
    name: 'PVCreate',
    component: () => import('@/views/storage/PVCreate.vue'),
    meta: { titleKey: 'storage.pvCreate', parent: 'PVList' },
  },
  {
    path: 'storage/pvs/:name',
    name: 'PVDetail',
    component: () => import('@/views/storage/PVDetail.vue'),
    props: true,
    meta: { titleKey: 'storage.pvDetail', parent: 'PVList' },
  },
  // PVC
  {
    path: 'storage/pvcs',
    name: 'PVCList',
    component: () => import('@/views/storage/PVCList.vue'),
    meta: { titleKey: 'sidebar.pvcs', icon: 'Box' },
  },
  {
    path: 'storage/pvcs/create',
    name: 'PVCCreate',
    component: () => import('@/views/storage/PVCCreate.vue'),
    meta: { titleKey: 'storage.pvcCreate', parent: 'PVCList' },
  },
  {
    path: 'storage/pvcs/:namespace/:name',
    name: 'PVCDetail',
    component: () => import('@/views/storage/PVCDetail.vue'),
    props: true,
    meta: { titleKey: 'storage.pvcDetail', parent: 'PVCList' },
  },
  // StorageClass
  {
    path: 'storage/storageclasses',
    name: 'StorageClassList',
    component: () => import('@/views/storage/StorageClassList.vue'),
    meta: { titleKey: 'sidebar.storageclasses', icon: 'Files' },
  },
  {
    path: 'storage/storageclasses/create',
    name: 'StorageClassCreate',
    component: () => import('@/views/storage/StorageClassCreate.vue'),
    meta: { titleKey: 'storage.storageClassCreate', parent: 'StorageClassList' },
  },
  {
    path: 'storage/storageclasses/:name',
    name: 'StorageClassDetail',
    component: () => import('@/views/storage/StorageClassDetail.vue'),
    props: true,
    meta: { titleKey: 'storage.storageClassDetail', parent: 'StorageClassList' },
  },
  // VolumeSnapshot
  {
    path: 'storage/volumesnapshots',
    name: 'VolumeSnapshotList',
    component: () => import('@/views/storage/VolumeSnapshotList.vue'),
    meta: { titleKey: 'storage.volumeSnapshot', icon: 'Camera' },
  },
  {
    path: 'storage/volumesnapshots/create',
    name: 'VolumeSnapshotCreate',
    component: () => import('@/views/storage/VolumeSnapshotCreate.vue'),
    meta: { titleKey: 'storage.volumeSnapshotCreate', parent: 'VolumeSnapshotList' },
  },
  {
    path: 'storage/volumesnapshots/:namespace/:name',
    name: 'VolumeSnapshotDetail',
    component: () => import('@/views/storage/VolumeSnapshotDetail.vue'),
    props: true,
    meta: { titleKey: 'storage.volumeSnapshotDetail', parent: 'VolumeSnapshotList' },
  },
  // VolumeSnapshotClass
  {
    path: 'storage/volumesnapshotclasses',
    name: 'VolumeSnapshotClassList',
    component: () => import('@/views/storage/VolumeSnapshotClassList.vue'),
    meta: { titleKey: 'storage.volumeSnapshotClass', icon: 'CameraFilled' },
  },
  {
    path: 'storage/volumesnapshotclasses/create',
    name: 'VolumeSnapshotClassCreate',
    component: () => import('@/views/storage/VolumeSnapshotClassCreate.vue'),
    meta: { titleKey: 'storage.volumeSnapshotClassCreate', parent: 'VolumeSnapshotClassList' },
  },
  {
    path: 'storage/volumesnapshotclasses/:name',
    name: 'VolumeSnapshotClassDetail',
    component: () => import('@/views/storage/VolumeSnapshotClassDetail.vue'),
    props: true,
    meta: { titleKey: 'storage.volumeSnapshotClassDetail', parent: 'VolumeSnapshotClassList' },
  },
]
