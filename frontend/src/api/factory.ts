import request from './request'

export interface ResourceApiOptions {
  /** 更新 YAML 的端点路径，默认 '${basePath}/update-yaml'，部分资源用 '${basePath}/update' */
  updatePath?: string
}

/**
 * 为一种 K8s 资源生成标准 CRUD API 函数。
 * 资源特有的操作（scale/restart/rollback）在各域文件中单独定义。
 *
 * @param basePath API 路径前缀，如 '/k8s/deployment'
 * @param options 可选配置
 */
export function createResourceApi(basePath: string, options?: ResourceApiOptions) {
  const updateEndpoint = options?.updatePath ?? `${basePath}/update-yaml`
  return {
    list:       (params?: any) => request.get(`${basePath}/list`, { params }),
    detail:     (params: any) => request.get(`${basePath}/detail`, { params }),
    getYaml:    (params: any) => request.get(`${basePath}/get-yaml`, { params }),
    create:     (data: any) => request.post(`${basePath}/create`, data),
    updateYaml: (data: any) => request.put(updateEndpoint, data),
    delete:     (data: any) => request.delete(`${basePath}/delete`, { data }),
    events:     (params: any) => request.get(`${basePath}/events`, { params }),
  }
}
