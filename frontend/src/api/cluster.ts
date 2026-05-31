import request from './request'

export interface Cluster {
  id: number
  name: string
  description: string
  api_server: string
  status: string
  created_at: string
  updated_at: string
}

export interface ClusterCreateParams {
  name: string
  description?: string
  api_server: string
  token: string
  ca_cert?: string
}

export interface ClusterUpdateParams {
  id: number
  name?: string
  description?: string
  api_server?: string
  token?: string
  ca_cert?: string
}

export function getClusterList() {
  return request.get<Cluster[]>('/clusters')
}

export function createCluster(data: ClusterCreateParams) {
  return request.post<Cluster>('/clusters', data)
}

export function getClusterDetail(id: number) {
  return request.get<Cluster>(`/clusters/${id}`)
}

export function updateCluster(data: ClusterUpdateParams) {
  return request.put<Cluster>('/clusters', data)
}

export function deleteCluster(id: number) {
  return request.delete('/clusters', { data: { id } })
}

export function checkCluster(id: number) {
  return request.get(`/clusters/${id}/check`)
}
