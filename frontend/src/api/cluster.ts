import request from '@/api/request'

export function getClusterList(params: { page: number; size: number }) {
  return request.get('/clusters', { params })
}

export function getCluster(id: string) {
  return request.get(`/clusters/${id}`)
}

export function createCluster(data: any) {
  return request.post('/clusters', data)
}

export function updateCluster(id: string, data: any) {
  return request.put(`/clusters/${id}`, data)
}

export function deleteCluster(id: string) {
  return request.delete(`/clusters/${id}`)
}
