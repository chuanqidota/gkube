import request from './request'

// ============ Label Selector 类型定义 ============

/** 单个 Label 过滤条件 */
export interface LabelFilter {
  key: string
  operator: '=' | '!=' | 'in' | 'notin'
  values: string[]
}

/** 标签自动补全数据 */
export interface LabelData {
  keys: string[]
  values: Record<string, string[]>
}

// ============ Label API ============

/** 获取资源可用标签（keys + values），用于自动补全 */
export function getAvailableLabels(params: {
  clusterName: string
  namespace?: string
  resourceType: string
}) {
  return request.get<LabelData>('/k8s/labels', { params })
}
