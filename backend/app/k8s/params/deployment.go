package params

type DeploymentListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type DeploymentQueryByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
	FieldMap    map[string]string `form:"fieldMap" json:"field" label:"字段"`
}

type DeploymentQueryByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
	LabelMap    map[string]string `form:"labelMap" json:"label" label:"标签"`
}

type DeploymentCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"yaml"`
}

type DeploymentUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"yaml"`
}

type DeploymentDeleteParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"名称"`
}

type DeploymentDeleteByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
	FieldMap    map[string]string `form:"fieldMap" json:"field" label:"字段"`
}

type DeploymentDeleteByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
	LabelMap    map[string]string `form:"labelMap" json:"label" label:"标签"`
}

type DeploymentScaleParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"名称"`
	Replicas    int32  `form:"replicas" json:"replicas" label:"副本数"`
}

type DeploymentRestartParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"名称"`
}

type DeploymentPodParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"名称"`
}

type DeploymentRollbackParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"名称"`
	Revision    int64  `form:"revision" json:"revision" label:"回滚版本"`
}
