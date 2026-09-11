import request from './request'

export interface ResourceApiOptions {
  /** 更新 YAML 的端点路径，默认 '${basePath}/update'，部分资源用 '${basePath}/update-yaml' */
  updatePath?: string
  /** 删除时使用 query params 而非 request body，默认 false（使用 body） */
  deleteUseParams?: boolean
}

/**
 * 为一种 K8s 资源生成标准 CRUD API 函数。
 * 资源特有的操作（scale/restart/rollback）在各域文件中单独定义。
 *
 * @param basePath API 路径前缀，如 '/k8s/deployment'
 * @param options 可选配置
 */
export function createResourceApi<TList = any, TDetail = any>(basePath: string, options?: ResourceApiOptions) {
  const updateEndpoint = options?.updatePath ?? `${basePath}/update`
  const useParamsForDelete = options?.deleteUseParams ?? false
  return {
    /** List — POST, body 可含 namespace/limit/continue/labelFilters */
    list:       (data?: any) => request.post<TList>(`${basePath}/list`, data),
    detail:     (params: any) => request.get<TDetail>(`${basePath}/detail`, { params }),
    getYaml:    (params: any) => request.get<string>(`${basePath}/get-yaml`, { params }),
    create:     (data: any) => request.post(`${basePath}/create`, data),
    updateYaml: (data: any) => request.put(updateEndpoint, data),
    delete:     useParamsForDelete
      ? (params: any) => request.delete(`${basePath}/delete`, { params })
      : (data: any) => request.delete(`${basePath}/delete`, { data }),
    events:     (params: any) => request.get(`${basePath}/events`, { params }),
  }
}
