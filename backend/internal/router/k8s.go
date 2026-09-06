package router

import (
	k8s "gkube/internal/k8s"
	"gkube/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// registerK8sRoutes 注册所有K8s资源路由
func registerK8sRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("k8s")
	{
		// Labels (shared across all resource types)
		grp.GET("labels", middleware.RequirePermission(), k8s.Label.GetLabels)

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
	rg.GET("cluster/version", middleware.RequirePermission(), k8s.Cluster.GetClusterVersion)
	rg.GET("cluster/nodes", middleware.RequirePermission(), k8s.Cluster.GetClusterNodesInfo)

	// Node
	rg.GET("node/detail", middleware.RequirePermission(), k8s.Node.GetNodeDetail)
	rg.GET("node/get-yaml", middleware.RequirePermission(), k8s.Node.GetNodeYaml)
	rg.GET("node/pods", middleware.RequirePermission(), k8s.Node.GetNodePods)
	rg.GET("node/events", middleware.RequirePermission(), k8s.Node.GetNodeEvents)
	rg.PUT("node/cordon", middleware.RequirePermission(), k8s.Node.CordonNode)
	rg.PUT("node/taints", middleware.RequirePermission(), k8s.Node.UpdateNodeTaints)
	rg.PUT("node/labels", middleware.RequirePermission(), k8s.Node.UpdateNodeLabels)
	rg.PUT("node/drain", middleware.RequirePermission(), k8s.Node.DrainNode)
	rg.PUT("node/update-yaml", middleware.RequirePermission(), k8s.Node.UpdateNodeYaml)
	rg.DELETE("node/delete", middleware.RequirePermission(), k8s.Node.DeleteNode)

	// Namespace
	rg.GET("namespace/list", middleware.RequirePermission(), k8s.Namespace.GetNamespaceList)
	rg.POST("namespace/list", middleware.RequirePermission(), k8s.Namespace.GetNamespaceList)
	rg.GET("namespace/detail", middleware.RequirePermission(), k8s.Namespace.GetNamespaceDetail)
	rg.GET("namespace/get-yaml", middleware.RequirePermission(), k8s.Namespace.GetNamespaceYaml)
	rg.POST("namespace/create", middleware.RequirePermission(), k8s.Namespace.CreateNamespace)
	rg.PUT("namespace/update", middleware.RequirePermission(), k8s.Namespace.UpdateNamespace)
	rg.PUT("namespace/labels", middleware.RequirePermission(), k8s.Namespace.UpdateNamespaceLabels)
	rg.DELETE("namespace/delete", middleware.RequirePermission(), k8s.Namespace.DeleteNamespace)

	// Pod
	rg.GET("pod/list", middleware.RequirePermission(), k8s.GetPodList)
	rg.POST("pod/list", middleware.RequirePermission(), k8s.GetPodList)
	rg.GET("pod/detail", middleware.RequirePermission(), k8s.GetPodByName)
	rg.GET("pod/get-yaml", middleware.RequirePermission(), k8s.GetPodYaml)
	rg.GET("pod/events", middleware.RequirePermission(), k8s.ListPodEvents)
	rg.POST("pod/create", middleware.RequirePermission(), k8s.CreatePod)
	rg.PUT("pod/update-yaml", middleware.RequirePermission(), k8s.PatchPodMetadata)
	rg.DELETE("pod/delete", middleware.RequirePermission(), k8s.DeletePodByName)

	// Event (read-only)
	rg.GET("event/list", middleware.RequirePermission(), k8s.Event.ListEvents)
	rg.GET("event/watch", middleware.RequirePermission(), k8s.Event.WatchEvents)
	// Container
	rg.GET("container/exec", middleware.RequirePermission(), k8s.HandleWebSocket)
	rg.GET("log", middleware.RequirePermission(), k8s.PodContainerLog)
	rg.GET("log/stream", middleware.RequirePermission(), k8s.StreamPodContainerLogs)
}

func registerWorkloadRoutes(rg *gin.RouterGroup) {
	// Deployment
	rg.GET("deployment/list", middleware.RequirePermission(), k8s.GetDeploymentList)
	rg.POST("deployment/list", middleware.RequirePermission(), k8s.GetDeploymentList)
	rg.GET("deployment/detail", middleware.RequirePermission(), k8s.GetDeploymentDetail)
	rg.GET("deployment/get-yaml", middleware.RequirePermission(), k8s.GetDeploymentYaml)
	rg.GET("deployment/events", middleware.RequirePermission(), k8s.GetDeploymentEvents)
	rg.POST("deployment/create", middleware.RequirePermission(), k8s.CreateDeployment)
	rg.PUT("deployment/update-yaml", middleware.RequirePermission(), k8s.UpdateDeployment)
	rg.DELETE("deployment/delete", middleware.RequirePermission(), k8s.DeleteDeployment)
	rg.PUT("deployment/scale", middleware.RequirePermission(), k8s.ScaleDeployment)
	rg.POST("deployment/restart", middleware.RequirePermission(), k8s.RestartDeployment)
	rg.POST("deployment/rollback", middleware.RequirePermission(), k8s.RollbackDeployment)
	rg.PUT("deployment/update-image", middleware.RequirePermission(), k8s.UpdateDeploymentImage)
	rg.GET("deployment/pods", middleware.RequirePermission(), k8s.DeploymentPodList)
	rg.GET("deployment/replicasets", middleware.RequirePermission(), k8s.GetDeploymentReplicaSets)

	// StatefulSet
	rg.GET("statefulset/list", middleware.RequirePermission(), k8s.GetStatefulSetList)
	rg.POST("statefulset/list", middleware.RequirePermission(), k8s.GetStatefulSetList)
	rg.GET("statefulset/detail", middleware.RequirePermission(), k8s.GetStatefulSetByName)
	rg.GET("statefulset/get-yaml", middleware.RequirePermission(), k8s.GetStatefulSetYaml)
	rg.GET("statefulset/events", middleware.RequirePermission(), k8s.GetStatefulSetEvents)
	rg.GET("statefulset/pods", middleware.RequirePermission(), k8s.StatefulSetPodList)
	rg.POST("statefulset/create", middleware.RequirePermission(), k8s.CreateStatefulSet)
	rg.PUT("statefulset/update", middleware.RequirePermission(), k8s.UpdateStatefulSet)
	rg.DELETE("statefulset/delete", middleware.RequirePermission(), k8s.DeleteStatefulSetByName)
	rg.PUT("statefulset/scale", middleware.RequirePermission(), k8s.ScaleStatefulSet)
	rg.POST("statefulset/restart", middleware.RequirePermission(), k8s.RestartStatefulSet)
	rg.POST("statefulset/rollback", middleware.RequirePermission(), k8s.RollbackStatefulSet)
	rg.PUT("statefulset/update-image", middleware.RequirePermission(), k8s.UpdateStatefulSetImage)
	rg.GET("statefulset/rollbacks", middleware.RequirePermission(), k8s.GetStatefulSetRollbacks)
	rg.GET("statefulset/pvcs", middleware.RequirePermission(), k8s.GetStatefulSetPVCs)

	// DaemonSet
	rg.GET("daemonset/list", middleware.RequirePermission(), k8s.GetDaemonSetList)
	rg.POST("daemonset/list", middleware.RequirePermission(), k8s.GetDaemonSetList)
	rg.GET("daemonset/detail", middleware.RequirePermission(), k8s.GetDaemonSetByName)
	rg.GET("daemonset/get-yaml", middleware.RequirePermission(), k8s.GetDaemonSetYaml)
	rg.GET("daemonset/events", middleware.RequirePermission(), k8s.GetDaemonSetEvents)
	rg.GET("daemonset/pods", middleware.RequirePermission(), k8s.DaemonSetPodList)
	rg.POST("daemonset/create", middleware.RequirePermission(), k8s.CreateDaemonSet)
	rg.PUT("daemonset/update", middleware.RequirePermission(), k8s.UpdateDaemonSet)
	rg.DELETE("daemonset/delete", middleware.RequirePermission(), k8s.DeleteDaemonSetByName)
	rg.POST("daemonset/restart", middleware.RequirePermission(), k8s.RestartDaemonSet)
	rg.POST("daemonset/rollback", middleware.RequirePermission(), k8s.RollbackDaemonSet)
	rg.PUT("daemonset/update-image", middleware.RequirePermission(), k8s.UpdateDaemonSetImage)
	rg.GET("daemonset/rollbacks", middleware.RequirePermission(), k8s.GetDaemonSetRollbacks)

	// Job
	rg.GET("job/list", middleware.RequirePermission(), k8s.GetJobList)
	rg.POST("job/list", middleware.RequirePermission(), k8s.GetJobList)
	rg.GET("job/detail", middleware.RequirePermission(), k8s.GetJobByName)
	rg.GET("job/get-yaml", middleware.RequirePermission(), k8s.GetJobYaml)
	rg.GET("job/events", middleware.RequirePermission(), k8s.GetJobEvents)
	rg.GET("job/pods", middleware.RequirePermission(), k8s.JobPodList)
	rg.POST("job/create", middleware.RequirePermission(), k8s.CreateJob)
	rg.PUT("job/update", middleware.RequirePermission(), k8s.UpdateJob)
	rg.DELETE("job/delete", middleware.RequirePermission(), k8s.DeleteJob)
	rg.POST("job/rerun", middleware.RequirePermission(), k8s.RerunJob)

	// CronJob
	rg.GET("cronjob/list", middleware.RequirePermission(), k8s.GetCronJobList)
	rg.POST("cronjob/list", middleware.RequirePermission(), k8s.GetCronJobList)
	rg.GET("cronjob/detail", middleware.RequirePermission(), k8s.GetCronJobByName)
	rg.GET("cronjob/get-yaml", middleware.RequirePermission(), k8s.GetCronJobYaml)
	rg.GET("cronjob/events", middleware.RequirePermission(), k8s.GetCronJobEvents)
	rg.GET("cronjob/jobs", middleware.RequirePermission(), k8s.CronJobJobsList)
	rg.POST("cronjob/create", middleware.RequirePermission(), k8s.CreateCronJob)
	rg.PUT("cronjob/update", middleware.RequirePermission(), k8s.UpdateCronJob)
	rg.DELETE("cronjob/delete", middleware.RequirePermission(), k8s.DeleteCronJobByName)
	rg.PUT("cronjob/suspend", middleware.RequirePermission(), k8s.SuspendCronJob)
	rg.PUT("cronjob/resume", middleware.RequirePermission(), k8s.ResumeCronJob)
	rg.POST("cronjob/trigger", middleware.RequirePermission(), k8s.TriggerCronJob)

	// ReplicaSet
	rg.GET("replicaset/list", middleware.RequirePermission(), k8s.GetReplicaSetList)
	rg.POST("replicaset/list", middleware.RequirePermission(), k8s.GetReplicaSetList)
	rg.GET("replicaset/get-yaml", middleware.RequirePermission(), k8s.GetReplicaSetYaml)
	rg.GET("replicaset/detail", middleware.RequirePermission(), k8s.GetReplicaSetDetail)
	rg.GET("replicaset/pods", middleware.RequirePermission(), k8s.GetReplicaSetPodList)
	rg.GET("replicaset/events", middleware.RequirePermission(), k8s.GetReplicaSetEvents)
	rg.DELETE("replicaset/delete", middleware.RequirePermission(), k8s.DeleteReplicaSet)

	// HPA
	rg.GET("hpa/list", middleware.RequirePermission(), k8s.GetHPAList)
	rg.POST("hpa/list", middleware.RequirePermission(), k8s.GetHPAList)
	rg.GET("hpa/detail", middleware.RequirePermission(), k8s.GetHPADetail)
	rg.GET("hpa/get-yaml", middleware.RequirePermission(), k8s.GetHPAYaml)
	rg.POST("hpa/create", middleware.RequirePermission(), k8s.CreateHPA)
	rg.PUT("hpa/update", middleware.RequirePermission(), k8s.UpdateHPA)
	rg.DELETE("hpa/delete", middleware.RequirePermission(), k8s.DeleteHPA)
	rg.GET("hpa/events", middleware.RequirePermission(), k8s.GetHPAEvents)
	rg.POST("hpa/pause", middleware.RequirePermission(), k8s.PauseHPA)
	rg.POST("hpa/resume", middleware.RequirePermission(), k8s.ResumeHPA)
}

func registerNetworkRoutes(rg *gin.RouterGroup) {
	// Service
	rg.GET("service/list", middleware.RequirePermission(), k8s.GetServicesList)
	rg.POST("service/list", middleware.RequirePermission(), k8s.GetServicesList)
	rg.GET("service/detail", middleware.RequirePermission(), k8s.GetServicesByName)
	rg.GET("service/get-yaml", middleware.RequirePermission(), k8s.GetServicesYaml)
	rg.GET("service/events", middleware.RequirePermission(), k8s.GetServiceEvents)
	rg.GET("service/pods", middleware.RequirePermission(), k8s.ServicePodList)
	rg.GET("service/endpoints", middleware.RequirePermission(), k8s.GetServiceEndpoints)
	rg.POST("service/create", middleware.RequirePermission(), k8s.CreateService)
	rg.PUT("service/update", middleware.RequirePermission(), k8s.UpdateService)
	rg.DELETE("service/delete", middleware.RequirePermission(), k8s.DeleteService)

	// Ingress
	rg.GET("ingress/list", middleware.RequirePermission(), k8s.GetIngressList)
	rg.POST("ingress/list", middleware.RequirePermission(), k8s.GetIngressList)
	rg.GET("ingress/detail", middleware.RequirePermission(), k8s.GetIngressByName)
	rg.GET("ingress/get-yaml", middleware.RequirePermission(), k8s.GetIngressYaml)
	rg.GET("ingress/events", middleware.RequirePermission(), k8s.GetIngressEvents)
	rg.GET("ingress/tls-status", middleware.RequirePermission(), k8s.CheckIngressTLSCertStatus)
	rg.GET("ingress/ingressclasses", middleware.RequirePermission(), k8s.GetIngressClassList)
	rg.POST("ingress/create", middleware.RequirePermission(), k8s.CreateIngress)
	rg.PUT("ingress/update", middleware.RequirePermission(), k8s.UpdateIngress)
	rg.DELETE("ingress/delete", middleware.RequirePermission(), k8s.DeleteIngressByName)

	// NetworkPolicy
	rg.GET("networkpolicy/list", middleware.RequirePermission(), k8s.GetNetworkPolicyList)
	rg.POST("networkpolicy/list", middleware.RequirePermission(), k8s.GetNetworkPolicyList)
	rg.GET("networkpolicy/detail", middleware.RequirePermission(), k8s.GetNetworkPolicyDetail)
	rg.GET("networkpolicy/get-yaml", middleware.RequirePermission(), k8s.GetNetworkPolicyYaml)
	rg.GET("networkpolicy/events", middleware.RequirePermission(), k8s.GetNetworkPolicyEvents)
	rg.GET("networkpolicy/pods", middleware.RequirePermission(), k8s.GetNetworkPolicyPods)
	rg.POST("networkpolicy/create", middleware.RequirePermission(), k8s.CreateNetworkPolicy)
	rg.PUT("networkpolicy/update", middleware.RequirePermission(), k8s.UpdateNetworkPolicy)
	rg.DELETE("networkpolicy/delete", middleware.RequirePermission(), k8s.DeleteNetworkPolicy)
}

func registerStorageRoutes(rg *gin.RouterGroup) {
	// PV
	rg.GET("pv/list", middleware.RequirePermission(), k8s.GetPVList)
	rg.POST("pv/list", middleware.RequirePermission(), k8s.GetPVList)
	rg.GET("pv/detail", middleware.RequirePermission(), k8s.GetPVByName)
	rg.GET("pv/get-yaml", middleware.RequirePermission(), k8s.GetPVYaml)
	rg.POST("pv/create", middleware.RequirePermission(), k8s.CreatePV)
	rg.PUT("pv/update", middleware.RequirePermission(), k8s.UpdatePV)
	rg.DELETE("pv/delete", middleware.RequirePermission(), k8s.DeletePVByName)

	// PVC
	rg.GET("pvc/list", middleware.RequirePermission(), k8s.GetPVCList)
	rg.POST("pvc/list", middleware.RequirePermission(), k8s.GetPVCList)
	rg.GET("pvc/list-by-storageclass", middleware.RequirePermission(), k8s.GetPVCListByStorageClass)
	rg.GET("pvc/detail", middleware.RequirePermission(), k8s.GetPVCByName)
	rg.GET("pvc/get-yaml", middleware.RequirePermission(), k8s.GetPVCYaml)
	rg.POST("pvc/create", middleware.RequirePermission(), k8s.CreatePVC)
	rg.PUT("pvc/update", middleware.RequirePermission(), k8s.UpdatePVC)
	rg.DELETE("pvc/delete", middleware.RequirePermission(), k8s.DeletePVCByName)

	// StorageClass
	rg.GET("storageclass/list", middleware.RequirePermission(), k8s.GetStorageClassList)
	rg.POST("storageclass/list", middleware.RequirePermission(), k8s.GetStorageClassList)
	rg.GET("storageclass/detail", middleware.RequirePermission(), k8s.GetStorageClassByName)
	rg.GET("storageclass/get-yaml", middleware.RequirePermission(), k8s.GetStorageClassYaml)
	rg.POST("storageclass/create", middleware.RequirePermission(), k8s.CreateStorageClass)
	rg.PUT("storageclass/update", middleware.RequirePermission(), k8s.UpdateStorageClass)
	rg.DELETE("storageclass/delete", middleware.RequirePermission(), k8s.DeleteStorageClassByName)
	rg.GET("storageclass/events", middleware.RequirePermission(), k8s.GetStorageClassEvents)

	// VolumeSnapshot
	rg.GET("volumesnapshot/list", middleware.RequirePermission(), k8s.GetVolumeSnapshotList)
	rg.POST("volumesnapshot/list", middleware.RequirePermission(), k8s.GetVolumeSnapshotList)
	rg.GET("volumesnapshot/detail", middleware.RequirePermission(), k8s.GetVolumeSnapshotByName)
	rg.GET("volumesnapshot/get-yaml", middleware.RequirePermission(), k8s.GetVolumeSnapshotYaml)
	rg.POST("volumesnapshot/create", middleware.RequirePermission(), k8s.CreateVolumeSnapshot)
	rg.PUT("volumesnapshot/update", middleware.RequirePermission(), k8s.UpdateVolumeSnapshot)
	rg.DELETE("volumesnapshot/delete", middleware.RequirePermission(), k8s.DeleteVolumeSnapshotByName)

	// VolumeSnapshotClass
	rg.GET("volumesnapshotclass/list", middleware.RequirePermission(), k8s.GetVolumeSnapshotClassList)
	rg.POST("volumesnapshotclass/list", middleware.RequirePermission(), k8s.GetVolumeSnapshotClassList)
	rg.GET("volumesnapshotclass/detail", middleware.RequirePermission(), k8s.GetVolumeSnapshotClassByName)
	rg.GET("volumesnapshotclass/get-yaml", middleware.RequirePermission(), k8s.GetVolumeSnapshotClassYaml)
	rg.POST("volumesnapshotclass/create", middleware.RequirePermission(), k8s.CreateVolumeSnapshotClass)
	rg.PUT("volumesnapshotclass/update", middleware.RequirePermission(), k8s.UpdateVolumeSnapshotClass)
	rg.DELETE("volumesnapshotclass/delete", middleware.RequirePermission(), k8s.DeleteVolumeSnapshotClassByName)
}

func registerConfigRoutes(rg *gin.RouterGroup) {
	// ConfigMap
	rg.GET("configmap/list", middleware.RequirePermission(), k8s.GetConfigMapList)
	rg.POST("configmap/list", middleware.RequirePermission(), k8s.GetConfigMapList)
	rg.GET("configmap/detail", middleware.RequirePermission(), k8s.GetConfigMapByName)
	rg.GET("configmap/get-yaml", middleware.RequirePermission(), k8s.GetConfigMapYaml)
	rg.POST("configmap/create", middleware.RequirePermission(), k8s.CreateConfigMapFromYaml)
	rg.PUT("configmap/update", middleware.RequirePermission(), k8s.UpdateConfigMapFromYaml)
	rg.DELETE("configmap/delete", middleware.RequirePermission(), k8s.DeleteConfigMapByName)

	// Secret
	rg.GET("secret/list", middleware.RequirePermission(), k8s.GetSecretsList)
	rg.POST("secret/list", middleware.RequirePermission(), k8s.GetSecretsList)
	rg.GET("secret/detail", middleware.RequirePermission(), k8s.GetSecretByName)
	rg.GET("secret/get-yaml", middleware.RequirePermission(), k8s.GetSecretYaml)
	rg.POST("secret/create", middleware.RequirePermission(), k8s.CreateSecretFromYaml)
	rg.PUT("secret/update", middleware.RequirePermission(), k8s.UpdateSecretFromYaml)
	rg.DELETE("secret/delete", middleware.RequirePermission(), k8s.DeleteSecret)

	// ResourceQuota
	rg.GET("resourcequota/list", middleware.RequirePermission(), k8s.GetResourceQuotaList)
	rg.POST("resourcequota/list", middleware.RequirePermission(), k8s.GetResourceQuotaList)
	rg.GET("resourcequota/detail", middleware.RequirePermission(), k8s.GetResourceQuotaDetail)
	rg.GET("resourcequota/get-yaml", middleware.RequirePermission(), k8s.GetResourceQuotaYaml)
	rg.POST("resourcequota/create", middleware.RequirePermission(), k8s.CreateResourceQuota)
	rg.PUT("resourcequota/update", middleware.RequirePermission(), k8s.UpdateResourceQuota)
	rg.DELETE("resourcequota/delete", middleware.RequirePermission(), k8s.DeleteResourceQuota)

	// LimitRange
	rg.GET("limitrange/list", middleware.RequirePermission(), k8s.GetLimitRangeList)
	rg.POST("limitrange/list", middleware.RequirePermission(), k8s.GetLimitRangeList)
	rg.GET("limitrange/detail", middleware.RequirePermission(), k8s.GetLimitRangeDetail)
	rg.GET("limitrange/get-yaml", middleware.RequirePermission(), k8s.GetLimitRangeYaml)
	rg.POST("limitrange/create", middleware.RequirePermission(), k8s.CreateLimitRange)
	rg.PUT("limitrange/update", middleware.RequirePermission(), k8s.UpdateLimitRange)
	rg.DELETE("limitrange/delete", middleware.RequirePermission(), k8s.DeleteLimitRange)
}

func registerCrdRoutes(rg *gin.RouterGroup) {
	rg.GET("crd/list", middleware.RequirePermission(), k8s.Crd.GetCRDList)
	rg.POST("crd/list", middleware.RequirePermission(), k8s.Crd.GetCRDList)
	rg.GET("crd/detail", middleware.RequirePermission(), k8s.Crd.GetCRDDetail)
	rg.GET("crd/get-yaml", middleware.RequirePermission(), k8s.Crd.GetCRDYaml)
	rg.POST("crd/create", middleware.RequirePermission(), k8s.Crd.CreateCRD)
	rg.PUT("crd/update", middleware.RequirePermission(), k8s.Crd.UpdateCRD)
	rg.DELETE("crd/delete", middleware.RequirePermission(), k8s.Crd.DeleteCRD)
	rg.GET("crd/resources", middleware.RequirePermission(), k8s.Crd.GetCustomResourceList)
	rg.POST("crd/resources", middleware.RequirePermission(), k8s.Crd.GetCustomResourceList)
	rg.GET("crd/resource/detail", middleware.RequirePermission(), k8s.Crd.GetCustomResourceDetail)
	rg.GET("crd/resource/yaml", middleware.RequirePermission(), k8s.Crd.GetCustomResourceYaml)
	rg.POST("crd/resource/create", middleware.RequirePermission(), k8s.Crd.CreateCustomResource)
	rg.DELETE("crd/resource", middleware.RequirePermission(), k8s.Crd.DeleteCustomResource)
	rg.PUT("crd/resource/update", middleware.RequirePermission(), k8s.Crd.UpdateCustomResource)
	rg.PATCH("crd/resource/patch", middleware.RequirePermission(), k8s.Crd.PatchCustomResource)
}

func registerAuditRoutes(rg *gin.RouterGroup) {
	// 审计属集群级只读资源：集群级角色有 audit:read 权限，ns 角色无
	rg.GET("audit/list", middleware.RequirePermission(), k8s.Audit.ListAuditLogs)
	rg.GET("audit/detail", middleware.RequirePermission(), k8s.Audit.GetAuditLog)
	rg.GET("audit/stats", middleware.RequirePermission(), k8s.Audit.GetAuditStats)
	// 审计清除属高危操作,需管理员
	rg.DELETE("audit/clear", middleware.RequireAdmin(), k8s.Audit.ClearAuditLogs)
}
