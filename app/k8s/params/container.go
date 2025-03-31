package params


type ContainerQueryParams struct {
    ClusterName string `form:"clusterName" json:"clusterName"`
    Container   string `form:"container" json:"container"`
    Namespace   string `form:"namespace" json:"namespace"`
    PodName     string `form:"podName" json:"podName"`
}