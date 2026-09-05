import request from './request'
import { calcAge } from './core'

// ============ 类型定义 ============

export interface Service {
  name: string
  namespace: string
  type: string
  cluster_ip: string
  external_ip: string
  ports: string
  age: string
}

export interface Ingress {
  name: string
  namespace: string
  hosts: string
  address: string
  age: string
}

// ============ Transform 函数 ============

export function transformServices(items: any[]): Service[] {
  if (!Array.isArray(items)) return []
  return items.map((svc: any) => {
    const ports = (svc.spec?.ports || [])
      .map((p: any) => {
        let s = `${p.port}`
        if (p.targetPort && p.targetPort !== p.port) s += `:${p.targetPort}`
        if (p.nodePort) s += `/${p.nodePort}`
        if (p.protocol && p.protocol !== 'TCP') s += `/${p.protocol}`
        return s
      })
      .join(', ')
    const externalIps = svc.spec?.externalIPs?.join(', ') || svc.status?.loadBalancer?.ingress?.map((i: any) => i.ip || i.hostname).join(', ') || ''
    return {
      name: svc.metadata?.name || '',
      namespace: svc.metadata?.namespace || '',
      type: svc.spec?.type || 'ClusterIP',
      cluster_ip: svc.spec?.clusterIP || '',
      external_ip: externalIps,
      ports,
      age: calcAge(svc.metadata?.creationTimestamp),
    }
  })
}

export function transformIngresses(items: any[]): Ingress[] {
  if (!Array.isArray(items)) return []
  return items.map((ing: any) => {
    const hosts = (ing.spec?.rules || []).map((r: any) => r.host || '*').join(', ')
    const address = ing.status?.loadBalancer?.ingress?.map((i: any) => i.ip || i.hostname).join(', ') || ''
    return {
      name: ing.metadata?.name || '',
      namespace: ing.metadata?.namespace || '',
      hosts,
      address,
      age: calcAge(ing.metadata?.creationTimestamp),
    }
  })
}

// ============ Service API ============

export const serviceApi = {
  list:    (data?: any) => request.post('/k8s/service/list', data),
  detail:  (params: { namespace: string; name: string }) => request.get('/k8s/service/detail', { params }),
  getYaml: (params: { namespace: string; name: string }) => request.get('/k8s/service/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/service/create', data),
  updateYaml: (data: { namespace: string; name: string; yaml: string }) => request.put('/k8s/service/update', data),
  delete:  (data: { namespace: string; name: string }) => request.delete('/k8s/service/delete', { data }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/service/events', { params }),
}

export const getServiceList = serviceApi.list
export const getServiceDetail = serviceApi.detail
export const getServiceYaml = serviceApi.getYaml
export const createService = serviceApi.create
export const updateService = serviceApi.updateYaml
export const deleteService = serviceApi.delete
export const getServiceEvents = serviceApi.events
export const getServicePods = (params: { namespace: string; name: string }) => request.get('/k8s/service/pods', { params })
export const getServiceEndpoints = (params: { namespace: string; name: string }) => request.get('/k8s/service/endpoints', { params })

// ============ Ingress API ============

export const ingressApi = {
  list:    (data?: any) => request.post('/k8s/ingress/list', data),
  detail:  (params: { namespace: string; name: string }) => request.get('/k8s/ingress/detail', { params }),
  getYaml: (params: { namespace: string; name: string }) => request.get('/k8s/ingress/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/ingress/create', data),
  updateYaml: (data: { namespace: string; name: string; yaml: string }) => request.put('/k8s/ingress/update', data),
  delete:  (data: { namespace: string; name: string }) => request.delete('/k8s/ingress/delete', { data }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/ingress/events', { params }),
}

export const getIngressList = ingressApi.list
export const getIngressDetail = ingressApi.detail
export const getIngressYaml = ingressApi.getYaml
export const createIngress = ingressApi.create
export const updateIngress = ingressApi.updateYaml
export const deleteIngress = ingressApi.delete
export const getIngressEvents = ingressApi.events
export const getIngressClassList = (params: { clusterName?: string }) => request.get('/k8s/ingress/ingressclasses', { params })
export const getIngressTLSCertStatus = (params: { namespace: string; name: string }) => request.get('/k8s/ingress/tls-status', { params })

// ============ NetworkPolicy API ============

export const networkPolicyApi = {
  list:    (data?: any) => request.post('/k8s/networkpolicy/list', data),
  detail:  (params: { namespace: string; name: string }) => request.get('/k8s/networkpolicy/detail', { params }),
  getYaml: (params: { namespace: string; name: string }) => request.get('/k8s/networkpolicy/get-yaml', { params }),
  create:  (data: { namespace: string; yaml: string }) => request.post('/k8s/networkpolicy/create', data),
  updateYaml: (data: { namespace: string; name: string; yaml: string }) => request.put('/k8s/networkpolicy/update', data),
  delete:  (params: { namespace: string; name: string }) => request.delete('/k8s/networkpolicy/delete', { params }),
  events:  (params: { namespace: string; name: string }) => request.get('/k8s/networkpolicy/events', { params }),
}

export const getNetworkPolicyList = networkPolicyApi.list
export const getNetworkPolicyDetail = networkPolicyApi.detail
export const getNetworkPolicyYaml = networkPolicyApi.getYaml
export const createNetworkPolicy = networkPolicyApi.create
export const updateNetworkPolicyYaml = networkPolicyApi.updateYaml
export const deleteNetworkPolicy = networkPolicyApi.delete
export const getNetworkPolicyEvents = networkPolicyApi.events
export const getNetworkPolicyPods = (params: { namespace: string; name: string }) => request.get('/k8s/networkpolicy/pods', { params })
