import request from '@/api/request'

export function getClusterList(params?: { page?: number; size?: number }) {
  return request.get('/clusters', { params })
}

export function getCluster(id: string) {
  return request.get(`/clusters/${id}`)
}

export function createCluster(data: any) {
  return request.post('/clusters', data)
}

export function updateCluster(data: { id: number; displayName?: string; description?: string; labels?: Record<string, string> }) {
  return request.put('/clusters', data)
}

export function deleteCluster(id: string) {
  return request.delete(`/clusters/${id}`)
}

export function checkCluster(id: string) {
  return request.get(`/clusters/${id}/check`)
}
