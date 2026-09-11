import request from './request'

// ============ CRD API ============

export function getCrdList(data?: any) {
  return request.post('/k8s/crd/list', data)
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

export function getCustomResourceList(data: any) {
  return request.post('/k8s/crd/resources', data)
}

export function getCustomResourceYaml(params: {
  group: string
  version: string
  resource: string
  namespace?: string
  name: string
}) {
  return request.get('/k8s/crd/resource/yaml', { params })
}

export function getCustomResourceDetail(params: {
  group: string
  version: string
  resource: string
  namespace?: string
  name: string
}) {
  return request.get('/k8s/crd/resource/detail', { params })
}

export function createCustomResource(data: {
  group: string
  version: string
  resource: string
  namespace?: string
  yaml: string
}) {
  return request.post('/k8s/crd/resource/create', data)
}

export function deleteCustomResource(params: {
  group: string
  version: string
  resource: string
  namespace?: string
  name: string
}) {
  return request.delete('/k8s/crd/resource', { params })
}

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
