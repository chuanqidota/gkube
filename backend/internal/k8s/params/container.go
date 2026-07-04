package params


type ContainerQueryParams struct {
    ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
    Container   string `form:"container" json:"container" label:"容器名称"`
    Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
    PodName     string `form:"podName" json:"podName" label:"Pod名称"`
}


type ContainerLogQueryParams struct {
    ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
    Container   string `form:"container" json:"container" label:"容器名称"`
    Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
    PodName     string `form:"podName" json:"podName" label:"Pod名称"`
    TailLines   int64  `form:"tailLines" json:"tailLines" label:"日志行数"`
}
