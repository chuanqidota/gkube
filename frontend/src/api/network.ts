import request from './request'
import { formatAge } from '@/utils/helpers'
import { createResourceApi } from './factory'

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
    const externalIps =
      svc.spec?.externalIPs?.join(', ') ||
      svc.status?.loadBalancer?.ingress?.map((i: any) => i.ip || i.hostname).join(', ') ||
      ''
    return {
      name: svc.metadata?.name || '',
      namespace: svc.metadata?.namespace || '',
      type: svc.spec?.type || 'ClusterIP',
      cluster_ip: svc.spec?.clusterIP || '',
      external_ip: externalIps,
      ports,
      age: formatAge(svc.metadata?.creationTimestamp, false),
    }
  })
}

export function transformIngresses(items: any[]): Ingress[] {
  if (!Array.isArray(items)) return []
  return items.map((ing: any) => {
    const hosts = (ing.spec?.rules || []).map((r: any) => r.host || '*').join(', ')
    const address =
      ing.status?.loadBalancer?.ingress?.map((i: any) => i.ip || i.hostname).join(', ') || ''
    return {
      name: ing.metadata?.name || '',
      namespace: ing.metadata?.namespace || '',
      hosts,
      address,
      age: formatAge(ing.metadata?.creationTimestamp, false),
    }
  })
}

// ============ 标准 CRUD（工厂生成） ============

export const serviceApi = createResourceApi('/k8s/service')

export const ingressApi = createResourceApi('/k8s/ingress')

export const networkPolicyApi = createResourceApi('/k8s/networkpolicy', {
  deleteUseParams: true,
})

// ============ 向后兼容函数别名（deprecated） ============

// Service
/** @deprecated 使用 serviceApi.list() */
export const getServiceList = serviceApi.list
/** @deprecated 使用 serviceApi.detail() */
export const getServiceDetail = serviceApi.detail
/** @deprecated 使用 serviceApi.getYaml() */
export const getServiceYaml = serviceApi.getYaml
/** @deprecated 使用 serviceApi.create() */
export const createService = serviceApi.create
/** @deprecated 使用 serviceApi.updateYaml() */
export const updateService = serviceApi.updateYaml
/** @deprecated 使用 serviceApi.delete() */
export const deleteService = serviceApi.delete
/** @deprecated 使用 serviceApi.events() */
export const getServiceEvents = serviceApi.events

// Ingress
/** @deprecated 使用 ingressApi.list() */
export const getIngressList = ingressApi.list
/** @deprecated 使用 ingressApi.detail() */
export const getIngressDetail = ingressApi.detail
/** @deprecated 使用 ingressApi.getYaml() */
export const getIngressYaml = ingressApi.getYaml
/** @deprecated 使用 ingressApi.create() */
export const createIngress = ingressApi.create
/** @deprecated 使用 ingressApi.updateYaml() */
export const updateIngress = ingressApi.updateYaml
/** @deprecated 使用 ingressApi.delete() */
export const deleteIngress = ingressApi.delete
/** @deprecated 使用 ingressApi.events() */
export const getIngressEvents = ingressApi.events

// NetworkPolicy
/** @deprecated 使用 networkPolicyApi.list() */
export const getNetworkPolicyList = networkPolicyApi.list
/** @deprecated 使用 networkPolicyApi.detail() */
export const getNetworkPolicyDetail = networkPolicyApi.detail
/** @deprecated 使用 networkPolicyApi.getYaml() */
export const getNetworkPolicyYaml = networkPolicyApi.getYaml
/** @deprecated 使用 networkPolicyApi.create() */
export const createNetworkPolicy = networkPolicyApi.create
/** @deprecated 使用 networkPolicyApi.updateYaml() */
export const updateNetworkPolicyYaml = networkPolicyApi.updateYaml
/** @deprecated 使用 networkPolicyApi.delete() */
export const deleteNetworkPolicy = networkPolicyApi.delete
/** @deprecated 使用 networkPolicyApi.events() */
export const getNetworkPolicyEvents = networkPolicyApi.events

// ============ 资源特有操作 ============

export const getServicePods = (params: { namespace: string; name: string }) =>
  request.get('/k8s/service/pods', { params })
export const getServiceEndpoints = (params: { namespace: string; name: string }) =>
  request.get('/k8s/service/endpoints', { params })
export const getIngressClassList = (params: { clusterName?: string }) =>
  request.get('/k8s/ingress/ingressclasses', { params })
export const getIngressTLSCertStatus = (params: { namespace: string; name: string }) =>
  request.get('/k8s/ingress/tls-status', { params })
export const getNetworkPolicyPods = (params: { namespace: string; name: string }) =>
  request.get('/k8s/networkpolicy/pods', { params })
