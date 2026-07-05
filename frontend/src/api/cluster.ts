import request from '@/api/request'

export interface CreateClusterData {
  clusterName: string
  displayName?: string
  description?: string
  kubeConfig: string
  labels?: Record<string, string>
}

export interface UpdateClusterData {
  displayName?: string | null
  description?: string | null
  labels?: Record<string, string> | null
}

export function getClusterList(params?: { page?: number; size?: number }) {
  return request.get('/clusters', { params })
}

export function getCluster(id: number) {
  return request.get(`/clusters/${id}`)
}

export function createCluster(data: CreateClusterData) {
  return request.post('/clusters', data)
}

export function updateCluster(id: number, data: UpdateClusterData) {
  return request.put(`/clusters/${id}`, data)
}

export function deleteCluster(id: number) {
  return request.delete(`/clusters/${id}`)
}

export function checkCluster(id: number) {
  return request.get(`/clusters/${id}/check`)
}
