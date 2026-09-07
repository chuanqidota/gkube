import request from './request'

// ============ Event API ============

export function getEventList(params: {
  namespace?: string
  fieldSelector?: string
  limit?: number
  continue?: string
}) {
  return request.get('/k8s/event/list', { params })
}

// ============ Dashboard Events API ============

export function getDashboardEvents(params?: {
  clusterId?: number
  type?: string
  namespace?: string
  limit?: number
  continue?: string
  fieldSelector?: string
}) {
  return request.get('/dashboard/events', { params })
}
