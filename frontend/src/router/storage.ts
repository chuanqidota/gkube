import type { RouteRecordRaw } from 'vue-router'

export const storageRoutes: RouteRecordRaw[] = [
  // PV
  {
    path: 'storage/pvs',
    name: 'PVList',
    component: () => import('@/views/storage/PVList.vue'),
    meta: { title: 'PersistentVolume', icon: 'Coin' },
  },
  {
    path: 'storage/pvs/create',
    name: 'PVCreate',
    component: () => import('@/views/storage/PVCreate.vue'),
    meta: { title: '创建PV', parent: 'PVList' },
  },
  {
    path: 'storage/pvs/:name',
    name: 'PVDetail',
    component: () => import('@/views/storage/PVDetail.vue'),
    props: true,
    meta: { title: 'PV详情', parent: 'PVList' },
  },
  // PVC
  {
    path: 'storage/pvcs',
    name: 'PVCList',
    component: () => import('@/views/storage/PVCList.vue'),
    meta: { title: 'PVC', icon: 'Box' },
  },
  {
    path: 'storage/pvcs/create',
    name: 'PVCCreate',
    component: () => import('@/views/storage/PVCCreate.vue'),
    meta: { title: '创建PVC', parent: 'PVCList' },
  },
  {
    path: 'storage/pvcs/:namespace/:name',
    name: 'PVCDetail',
    component: () => import('@/views/storage/PVCDetail.vue'),
    props: true,
    meta: { title: 'PVC详情', parent: 'PVCList' },
  },
  // StorageClass
  {
    path: 'storage/storageclasses',
    name: 'StorageClassList',
    component: () => import('@/views/storage/StorageClassList.vue'),
    meta: { title: 'StorageClass', icon: 'Files' },
  },
  {
    path: 'storage/storageclasses/create',
    name: 'StorageClassCreate',
    component: () => import('@/views/storage/StorageClassCreate.vue'),
    meta: { title: '创建StorageClass', parent: 'StorageClassList' },
  },
  {
    path: 'storage/storageclasses/:name',
    name: 'StorageClassDetail',
    component: () => import('@/views/storage/StorageClassDetail.vue'),
    props: true,
    meta: { title: 'StorageClass详情', parent: 'StorageClassList' },
  },
  // VolumeSnapshot
  {
    path: 'storage/volumesnapshots',
    name: 'VolumeSnapshotList',
    component: () => import('@/views/storage/VolumeSnapshotList.vue'),
    meta: { title: 'VolumeSnapshot', icon: 'Camera' },
  },
  {
    path: 'storage/volumesnapshots/create',
    name: 'VolumeSnapshotCreate',
    component: () => import('@/views/storage/VolumeSnapshotCreate.vue'),
    meta: { title: '创建VolumeSnapshot', parent: 'VolumeSnapshotList' },
  },
  {
    path: 'storage/volumesnapshots/:namespace/:name',
    name: 'VolumeSnapshotDetail',
    component: () => import('@/views/storage/VolumeSnapshotDetail.vue'),
    props: true,
    meta: { title: 'VolumeSnapshot详情', parent: 'VolumeSnapshotList' },
  },
  // VolumeSnapshotClass
  {
    path: 'storage/volumesnapshotclasses',
    name: 'VolumeSnapshotClassList',
    component: () => import('@/views/storage/VolumeSnapshotClassList.vue'),
    meta: { title: 'VolumeSnapshotClass', icon: 'CameraFilled' },
  },
  {
    path: 'storage/volumesnapshotclasses/create',
    name: 'VolumeSnapshotClassCreate',
    component: () => import('@/views/storage/VolumeSnapshotClassCreate.vue'),
    meta: { title: '创建VolumeSnapshotClass', parent: 'VolumeSnapshotClassList' },
  },
  {
    path: 'storage/volumesnapshotclasses/:name',
    name: 'VolumeSnapshotClassDetail',
    component: () => import('@/views/storage/VolumeSnapshotClassDetail.vue'),
    props: true,
    meta: { title: 'VolumeSnapshotClass详情', parent: 'VolumeSnapshotClassList' },
  },
]
