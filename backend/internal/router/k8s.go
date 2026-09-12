package router

import (
	k8s "gkube/internal/k8s"
	"gkube/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// registerK8sRoutes 注册所有K8s资源路由
func registerK8sRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("k8s")
	grp.Use(middleware.RequirePermission()) // 统一权限检查，无需每条路由单独挂载
	{
		// Labels (shared across all resource types)
		grp.GET("labels", k8s.Label.GetLabels)

		// ---- Core ----
		registerCoreRoutes(grp)
		// ---- Workload ----
		registerWorkloadRoutes(grp)
		// ---- Network ----
		registerNetworkRoutes(grp)
		// ---- Storage ----
		registerStorageRoutes(grp)
		// ---- Config ----
		registerConfigRoutes(grp)
		// ---- CRD ----
		registerCrdRoutes(grp)
		// ---- Audit ----
		registerAuditRoutes(grp)
	}
}

func registerCoreRoutes(rg *gin.RouterGroup) {
	// Cluster (read-only)
	rg.GET("cluster/version", k8s.Cluster.GetClusterVersion)
	rg.GET("cluster/nodes", k8s.Cluster.GetClusterNodesInfo)

	// Node
	rg.GET("node/detail", k8s.Node.GetNodeDetail)
	rg.GET("node/get-yaml", k8s.Node.GetNodeYaml)
	rg.GET("node/pods", k8s.Node.GetNodePods)
	rg.GET("node/events", k8s.Node.GetNodeEvents)
	rg.PUT("node/cordon", k8s.Node.CordonNode)
	rg.PUT("node/taints", k8s.Node.UpdateNodeTaints)
	rg.PUT("node/labels", k8s.Node.UpdateNodeLabels)
	rg.PUT("node/drain", k8s.Node.DrainNode)
	rg.PUT("node/update-yaml", k8s.Node.UpdateNodeYaml)
	rg.DELETE("node/delete", k8s.Node.DeleteNode)

	// Namespace
	rg.GET("namespace/list", k8s.Namespace.GetNamespaceList)
	rg.POST("namespace/list", k8s.Namespace.GetNamespaceList)
	rg.GET("namespace/detail", k8s.Namespace.GetNamespaceDetail)
	rg.GET("namespace/get-yaml", k8s.Namespace.GetNamespaceYaml)
	rg.POST("namespace/create", k8s.Namespace.CreateNamespace)
	rg.PUT("namespace/update", k8s.Namespace.UpdateNamespace)
	rg.PUT("namespace/labels", k8s.Namespace.UpdateNamespaceLabels)
	rg.DELETE("namespace/delete", k8s.Namespace.DeleteNamespace)

	// Pod
	rg.GET("pod/list", k8s.GetPodList)
	rg.POST("pod/list", k8s.GetPodList)
	rg.GET("pod/detail", k8s.GetPodByName)
	rg.GET("pod/get-yaml", k8s.GetPodYaml)
	rg.GET("pod/events", k8s.ListPodEvents)
	rg.POST("pod/create", k8s.CreatePod)
	rg.PUT("pod/update-yaml", k8s.PatchPodMetadata)
	rg.DELETE("pod/delete", k8s.DeletePodByName)

	// Event (read-only)
	rg.GET("event/list", k8s.Event.ListEvents)
	rg.GET("event/watch", k8s.Event.WatchEvents)
	// Container
	rg.GET("container/exec", k8s.HandleWebSocket)
	rg.GET("log", k8s.PodContainerLog)
	rg.GET("log/stream", k8s.StreamPodContainerLogs)
}

func registerWorkloadRoutes(rg *gin.RouterGroup) {
	// Deployment
	rg.GET("deployment/list", k8s.GetDeploymentList)
	rg.POST("deployment/list", k8s.GetDeploymentList)
	rg.GET("deployment/detail", k8s.GetDeploymentDetail)
	rg.GET("deployment/get-yaml", k8s.GetDeploymentYaml)
	rg.GET("deployment/events", k8s.GetDeploymentEvents)
	rg.POST("deployment/create", k8s.CreateDeployment)
	rg.PUT("deployment/update-yaml", k8s.UpdateDeployment)
	rg.DELETE("deployment/delete", k8s.DeleteDeployment)
	rg.PUT("deployment/scale", k8s.ScaleDeployment)
	rg.POST("deployment/restart", k8s.RestartDeployment)
	rg.POST("deployment/rollback", k8s.RollbackDeployment)
	rg.PUT("deployment/update-image", k8s.UpdateDeploymentImage)
	rg.GET("deployment/pods", k8s.DeploymentPodList)
	rg.GET("deployment/replicasets", k8s.GetDeploymentReplicaSets)

	// StatefulSet
	rg.GET("statefulset/list", k8s.GetStatefulSetList)
	rg.POST("statefulset/list", k8s.GetStatefulSetList)
	rg.GET("statefulset/detail", k8s.GetStatefulSetByName)
	rg.GET("statefulset/get-yaml", k8s.GetStatefulSetYaml)
	rg.GET("statefulset/events", k8s.GetStatefulSetEvents)
	rg.GET("statefulset/pods", k8s.StatefulSetPodList)
	rg.POST("statefulset/create", k8s.CreateStatefulSet)
	rg.PUT("statefulset/update", k8s.UpdateStatefulSet)
	rg.DELETE("statefulset/delete", k8s.DeleteStatefulSetByName)
	rg.PUT("statefulset/scale", k8s.ScaleStatefulSet)
	rg.POST("statefulset/restart", k8s.RestartStatefulSet)
	rg.POST("statefulset/rollback", k8s.RollbackStatefulSet)
	rg.PUT("statefulset/update-image", k8s.UpdateStatefulSetImage)
	rg.GET("statefulset/rollbacks", k8s.GetStatefulSetRollbacks)
	rg.GET("statefulset/pvcs", k8s.GetStatefulSetPVCs)

	// DaemonSet
	rg.GET("daemonset/list", k8s.GetDaemonSetList)
	rg.POST("daemonset/list", k8s.GetDaemonSetList)
	rg.GET("daemonset/detail", k8s.GetDaemonSetByName)
	rg.GET("daemonset/get-yaml", k8s.GetDaemonSetYaml)
	rg.GET("daemonset/events", k8s.GetDaemonSetEvents)
	rg.GET("daemonset/pods", k8s.DaemonSetPodList)
	rg.POST("daemonset/create", k8s.CreateDaemonSet)
	rg.PUT("daemonset/update", k8s.UpdateDaemonSet)
	rg.DELETE("daemonset/delete", k8s.DeleteDaemonSetByName)
	rg.POST("daemonset/restart", k8s.RestartDaemonSet)
	rg.POST("daemonset/rollback", k8s.RollbackDaemonSet)
	rg.PUT("daemonset/update-image", k8s.UpdateDaemonSetImage)
	rg.GET("daemonset/rollbacks", k8s.GetDaemonSetRollbacks)

	// Job
	rg.GET("job/list", k8s.GetJobList)
	rg.POST("job/list", k8s.GetJobList)
	rg.GET("job/detail", k8s.GetJobByName)
	rg.GET("job/get-yaml", k8s.GetJobYaml)
	rg.GET("job/events", k8s.GetJobEvents)
	rg.GET("job/pods", k8s.JobPodList)
	rg.POST("job/create", k8s.CreateJob)
	rg.PUT("job/update", k8s.UpdateJob)
	rg.DELETE("job/delete", k8s.DeleteJob)
	rg.POST("job/rerun", k8s.RerunJob)

	// CronJob
	rg.GET("cronjob/list", k8s.GetCronJobList)
	rg.POST("cronjob/list", k8s.GetCronJobList)
	rg.GET("cronjob/detail", k8s.GetCronJobByName)
	rg.GET("cronjob/get-yaml", k8s.GetCronJobYaml)
	rg.GET("cronjob/events", k8s.GetCronJobEvents)
	rg.GET("cronjob/jobs", k8s.CronJobJobsList)
	rg.POST("cronjob/create", k8s.CreateCronJob)
	rg.PUT("cronjob/update", k8s.UpdateCronJob)
	rg.DELETE("cronjob/delete", k8s.DeleteCronJobByName)
	rg.PUT("cronjob/suspend", k8s.SuspendCronJob)
	rg.PUT("cronjob/resume", k8s.ResumeCronJob)
	rg.POST("cronjob/trigger", k8s.TriggerCronJob)

	// ReplicaSet
	rg.GET("replicaset/list", k8s.GetReplicaSetList)
	rg.POST("replicaset/list", k8s.GetReplicaSetList)
	rg.GET("replicaset/get-yaml", k8s.GetReplicaSetYaml)
	rg.GET("replicaset/detail", k8s.GetReplicaSetDetail)
	rg.GET("replicaset/pods", k8s.GetReplicaSetPodList)
	rg.GET("replicaset/events", k8s.GetReplicaSetEvents)
	rg.DELETE("replicaset/delete", k8s.DeleteReplicaSet)

	// HPA
	rg.GET("hpa/list", k8s.GetHPAList)
	rg.POST("hpa/list", k8s.GetHPAList)
	rg.GET("hpa/detail", k8s.GetHPADetail)
	rg.GET("hpa/get-yaml", k8s.GetHPAYaml)
	rg.POST("hpa/create", k8s.CreateHPA)
	rg.PUT("hpa/update", k8s.UpdateHPA)
	rg.DELETE("hpa/delete", k8s.DeleteHPA)
	rg.GET("hpa/events", k8s.GetHPAEvents)
	rg.POST("hpa/pause", k8s.PauseHPA)
	rg.POST("hpa/resume", k8s.ResumeHPA)
}

func registerNetworkRoutes(rg *gin.RouterGroup) {
	// Service
	rg.GET("service/list", k8s.GetServicesList)
	rg.POST("service/list", k8s.GetServicesList)
	rg.GET("service/detail", k8s.GetServicesByName)
	rg.GET("service/get-yaml", k8s.GetServicesYaml)
	rg.GET("service/events", k8s.GetServiceEvents)
	rg.GET("service/pods", k8s.ServicePodList)
	rg.GET("service/endpoints", k8s.GetServiceEndpoints)
	rg.POST("service/create", k8s.CreateService)
	rg.PUT("service/update", k8s.UpdateService)
	rg.DELETE("service/delete", k8s.DeleteService)

	// Ingress
	rg.GET("ingress/list", k8s.GetIngressList)
	rg.POST("ingress/list", k8s.GetIngressList)
	rg.GET("ingress/detail", k8s.GetIngressByName)
	rg.GET("ingress/get-yaml", k8s.GetIngressYaml)
	rg.GET("ingress/events", k8s.GetIngressEvents)
	rg.GET("ingress/tls-status", k8s.CheckIngressTLSCertStatus)
	rg.GET("ingress/ingressclasses", k8s.GetIngressClassList)
	rg.POST("ingress/create", k8s.CreateIngress)
	rg.PUT("ingress/update", k8s.UpdateIngress)
	rg.DELETE("ingress/delete", k8s.DeleteIngressByName)

	// NetworkPolicy
	rg.GET("networkpolicy/list", k8s.GetNetworkPolicyList)
	rg.POST("networkpolicy/list", k8s.GetNetworkPolicyList)
	rg.GET("networkpolicy/detail", k8s.GetNetworkPolicyDetail)
	rg.GET("networkpolicy/get-yaml", k8s.GetNetworkPolicyYaml)
	rg.GET("networkpolicy/events", k8s.GetNetworkPolicyEvents)
	rg.GET("networkpolicy/pods", k8s.GetNetworkPolicyPods)
	rg.POST("networkpolicy/create", k8s.CreateNetworkPolicy)
	rg.PUT("networkpolicy/update", k8s.UpdateNetworkPolicy)
	rg.DELETE("networkpolicy/delete", k8s.DeleteNetworkPolicy)
}

func registerStorageRoutes(rg *gin.RouterGroup) {
	// PV
	rg.GET("pv/list", k8s.GetPVList)
	rg.POST("pv/list", k8s.GetPVList)
	rg.GET("pv/detail", k8s.GetPVByName)
	rg.GET("pv/get-yaml", k8s.GetPVYaml)
	rg.POST("pv/create", k8s.CreatePV)
	rg.PUT("pv/update", k8s.UpdatePV)
	rg.DELETE("pv/delete", k8s.DeletePVByName)

	// PVC
	rg.GET("pvc/list", k8s.GetPVCList)
	rg.POST("pvc/list", k8s.GetPVCList)
	rg.GET("pvc/list-by-storageclass", k8s.GetPVCListByStorageClass)
	rg.GET("pvc/detail", k8s.GetPVCByName)
	rg.GET("pvc/get-yaml", k8s.GetPVCYaml)
	rg.POST("pvc/create", k8s.CreatePVC)
	rg.PUT("pvc/update", k8s.UpdatePVC)
	rg.DELETE("pvc/delete", k8s.DeletePVCByName)

	// StorageClass
	rg.GET("storageclass/list", k8s.GetStorageClassList)
	rg.POST("storageclass/list", k8s.GetStorageClassList)
	rg.GET("storageclass/detail", k8s.GetStorageClassByName)
	rg.GET("storageclass/get-yaml", k8s.GetStorageClassYaml)
	rg.POST("storageclass/create", k8s.CreateStorageClass)
	rg.PUT("storageclass/update", k8s.UpdateStorageClass)
	rg.DELETE("storageclass/delete", k8s.DeleteStorageClassByName)
	rg.GET("storageclass/events", k8s.GetStorageClassEvents)

	// VolumeSnapshot
	rg.GET("volumesnapshot/list", k8s.GetVolumeSnapshotList)
	rg.POST("volumesnapshot/list", k8s.GetVolumeSnapshotList)
	rg.GET("volumesnapshot/detail", k8s.GetVolumeSnapshotByName)
	rg.GET("volumesnapshot/get-yaml", k8s.GetVolumeSnapshotYaml)
	rg.POST("volumesnapshot/create", k8s.CreateVolumeSnapshot)
	rg.PUT("volumesnapshot/update", k8s.UpdateVolumeSnapshot)
	rg.DELETE("volumesnapshot/delete", k8s.DeleteVolumeSnapshotByName)

	// VolumeSnapshotClass
	rg.GET("volumesnapshotclass/list", k8s.GetVolumeSnapshotClassList)
	rg.POST("volumesnapshotclass/list", k8s.GetVolumeSnapshotClassList)
	rg.GET("volumesnapshotclass/detail", k8s.GetVolumeSnapshotClassByName)
	rg.GET("volumesnapshotclass/get-yaml", k8s.GetVolumeSnapshotClassYaml)
	rg.POST("volumesnapshotclass/create", k8s.CreateVolumeSnapshotClass)
	rg.PUT("volumesnapshotclass/update", k8s.UpdateVolumeSnapshotClass)
	rg.DELETE("volumesnapshotclass/delete", k8s.DeleteVolumeSnapshotClassByName)
}

func registerConfigRoutes(rg *gin.RouterGroup) {
	// ConfigMap
	rg.GET("configmap/list", k8s.GetConfigMapList)
	rg.POST("configmap/list", k8s.GetConfigMapList)
	rg.GET("configmap/detail", k8s.GetConfigMapByName)
	rg.GET("configmap/get-yaml", k8s.GetConfigMapYaml)
	rg.POST("configmap/create", k8s.CreateConfigMapFromYaml)
	rg.PUT("configmap/update", k8s.UpdateConfigMapFromYaml)
	rg.DELETE("configmap/delete", k8s.DeleteConfigMapByName)

	// Secret
	rg.GET("secret/list", k8s.GetSecretsList)
	rg.POST("secret/list", k8s.GetSecretsList)
	rg.GET("secret/detail", k8s.GetSecretByName)
	rg.GET("secret/get-yaml", k8s.GetSecretYaml)
	rg.POST("secret/create", k8s.CreateSecretFromYaml)
	rg.PUT("secret/update", k8s.UpdateSecretFromYaml)
	rg.DELETE("secret/delete", k8s.DeleteSecret)

	// ResourceQuota
	rg.GET("resourcequota/list", k8s.GetResourceQuotaList)
	rg.POST("resourcequota/list", k8s.GetResourceQuotaList)
	rg.GET("resourcequota/detail", k8s.GetResourceQuotaDetail)
	rg.GET("resourcequota/get-yaml", k8s.GetResourceQuotaYaml)
	rg.POST("resourcequota/create", k8s.CreateResourceQuota)
	rg.PUT("resourcequota/update", k8s.UpdateResourceQuota)
	rg.DELETE("resourcequota/delete", k8s.DeleteResourceQuota)

	// LimitRange
	rg.GET("limitrange/list", k8s.GetLimitRangeList)
	rg.POST("limitrange/list", k8s.GetLimitRangeList)
	rg.GET("limitrange/detail", k8s.GetLimitRangeDetail)
	rg.GET("limitrange/get-yaml", k8s.GetLimitRangeYaml)
	rg.POST("limitrange/create", k8s.CreateLimitRange)
	rg.PUT("limitrange/update", k8s.UpdateLimitRange)
	rg.DELETE("limitrange/delete", k8s.DeleteLimitRange)
}

func registerCrdRoutes(rg *gin.RouterGroup) {
	rg.GET("crd/list", k8s.Crd.GetCRDList)
	rg.POST("crd/list", k8s.Crd.GetCRDList)
	rg.GET("crd/detail", k8s.Crd.GetCRDDetail)
	rg.GET("crd/get-yaml", k8s.Crd.GetCRDYaml)
	rg.POST("crd/create", k8s.Crd.CreateCRD)
	rg.PUT("crd/update", k8s.Crd.UpdateCRD)
	rg.DELETE("crd/delete", k8s.Crd.DeleteCRD)
	rg.GET("crd/resources", k8s.Crd.GetCustomResourceList)
	rg.POST("crd/resources", k8s.Crd.GetCustomResourceList)
	rg.GET("crd/resource/detail", k8s.Crd.GetCustomResourceDetail)
	rg.GET("crd/resource/yaml", k8s.Crd.GetCustomResourceYaml)
	rg.POST("crd/resource/create", k8s.Crd.CreateCustomResource)
	rg.DELETE("crd/resource", k8s.Crd.DeleteCustomResource)
	rg.PUT("crd/resource/update", k8s.Crd.UpdateCustomResource)
	rg.PATCH("crd/resource/patch", k8s.Crd.PatchCustomResource)
}

func registerAuditRoutes(rg *gin.RouterGroup) {
	// 审计属集群级只读资源：集群级角色有 audit:read 权限，ns 角色无
	rg.GET("audit/list", k8s.Audit.ListAuditLogs)
	rg.GET("audit/detail", k8s.Audit.GetAuditLog)
	rg.GET("audit/stats", k8s.Audit.GetAuditStats)
	// 审计清除属高危操作,需管理员（RequireAdmin 叠加在 group 级 RequirePermission 之上）
	rg.DELETE("audit/clear", middleware.RequireAdmin(), k8s.Audit.ClearAuditLogs)
}
