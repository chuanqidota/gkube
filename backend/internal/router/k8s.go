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
	rg.GET("pod/list", middleware.RequirePermission(), k8s.Pod.GetPodList)
	rg.POST("pod/list", middleware.RequirePermission(), k8s.Pod.GetPodList)
	rg.GET("pod/detail", middleware.RequirePermission(), k8s.Pod.GetPodByName)
	rg.GET("pod/get-yaml", middleware.RequirePermission(), k8s.Pod.GetPodYaml)
	rg.GET("pod/events", middleware.RequirePermission(), k8s.Pod.ListPodEvents)
	rg.POST("pod/create", middleware.RequirePermission(), k8s.Pod.CreatePod)
	rg.PUT("pod/update-yaml", middleware.RequirePermission(), k8s.Pod.PatchPodMetadata)
	rg.DELETE("pod/delete", middleware.RequirePermission(), k8s.Pod.DeletePodByName)

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
	rg.GET("deployment/list", middleware.RequirePermission(), k8s.Deployment.GetDeploymentList)
	rg.POST("deployment/list", middleware.RequirePermission(), k8s.Deployment.GetDeploymentList)
	rg.GET("deployment/detail", middleware.RequirePermission(), k8s.Deployment.GetDeploymentDetail)
	rg.GET("deployment/get-yaml", middleware.RequirePermission(), k8s.Deployment.GetDeploymentYaml)
	rg.GET("deployment/events", middleware.RequirePermission(), k8s.Deployment.GetDeploymentEvents)
	rg.POST("deployment/create", middleware.RequirePermission(), k8s.Deployment.CreateDeployment)
	rg.PUT("deployment/update-yaml", middleware.RequirePermission(), k8s.Deployment.UpdateDeployment)
	rg.DELETE("deployment/delete", middleware.RequirePermission(), k8s.Deployment.DeleteDeployment)
	rg.PUT("deployment/scale", middleware.RequirePermission(), k8s.Deployment.ScaleDeployment)
	rg.POST("deployment/restart", middleware.RequirePermission(), k8s.Deployment.RestartDeployment)
	rg.POST("deployment/rollback", middleware.RequirePermission(), k8s.Deployment.RollbackDeployment)
	rg.PUT("deployment/update-image", middleware.RequirePermission(), k8s.Deployment.UpdateDeploymentImage)
	rg.GET("deployment/pods", middleware.RequirePermission(), k8s.Deployment.DeploymentPodList)
	rg.GET("deployment/replicasets", middleware.RequirePermission(), k8s.Deployment.GetDeploymentReplicaSets)

	// StatefulSet
	rg.GET("statefulset/list", middleware.RequirePermission(), k8s.StatefulSet.GetStatefulSetList)
	rg.POST("statefulset/list", middleware.RequirePermission(), k8s.StatefulSet.GetStatefulSetList)
	rg.GET("statefulset/detail", middleware.RequirePermission(), k8s.StatefulSet.GetStatefulSetByName)
	rg.GET("statefulset/get-yaml", middleware.RequirePermission(), k8s.StatefulSet.GetStatefulSetYaml)
	rg.GET("statefulset/events", middleware.RequirePermission(), k8s.StatefulSet.GetStatefulSetEvents)
	rg.GET("statefulset/pods", middleware.RequirePermission(), k8s.StatefulSet.StatefulSetPodList)
	rg.POST("statefulset/create", middleware.RequirePermission(), k8s.StatefulSet.CreateStatefulSet)
	rg.PUT("statefulset/update", middleware.RequirePermission(), k8s.StatefulSet.UpdateStatefulSet)
	rg.DELETE("statefulset/delete", middleware.RequirePermission(), k8s.StatefulSet.DeleteStatefulSetByName)
	rg.PUT("statefulset/scale", middleware.RequirePermission(), k8s.StatefulSet.ScaleStatefulSet)
	rg.POST("statefulset/restart", middleware.RequirePermission(), k8s.StatefulSet.RestartStatefulSet)
	rg.POST("statefulset/rollback", middleware.RequirePermission(), k8s.StatefulSet.RollbackStatefulSet)
	rg.PUT("statefulset/update-image", middleware.RequirePermission(), k8s.StatefulSet.UpdateStatefulSetImage)
	rg.GET("statefulset/rollbacks", middleware.RequirePermission(), k8s.StatefulSet.GetStatefulSetRollbacks)
	rg.GET("statefulset/pvcs", middleware.RequirePermission(), k8s.StatefulSet.GetStatefulSetPVCs)

	// DaemonSet
	rg.GET("daemonset/list", middleware.RequirePermission(), k8s.DaemonSet.GetDaemonSetList)
	rg.POST("daemonset/list", middleware.RequirePermission(), k8s.DaemonSet.GetDaemonSetList)
	rg.GET("daemonset/detail", middleware.RequirePermission(), k8s.DaemonSet.GetDaemonSetByName)
	rg.GET("daemonset/get-yaml", middleware.RequirePermission(), k8s.DaemonSet.GetDaemonSetYaml)
	rg.GET("daemonset/events", middleware.RequirePermission(), k8s.DaemonSet.GetDaemonSetEvents)
	rg.GET("daemonset/pods", middleware.RequirePermission(), k8s.DaemonSet.DaemonSetPodList)
	rg.POST("daemonset/create", middleware.RequirePermission(), k8s.DaemonSet.CreateDaemonSet)
	rg.PUT("daemonset/update", middleware.RequirePermission(), k8s.DaemonSet.UpdateDaemonSet)
	rg.DELETE("daemonset/delete", middleware.RequirePermission(), k8s.DaemonSet.DeleteDaemonSetByName)
	rg.POST("daemonset/restart", middleware.RequirePermission(), k8s.DaemonSet.RestartDaemonSet)
	rg.POST("daemonset/rollback", middleware.RequirePermission(), k8s.DaemonSet.RollbackDaemonSet)
	rg.PUT("daemonset/update-image", middleware.RequirePermission(), k8s.DaemonSet.UpdateDaemonSetImage)
	rg.GET("daemonset/rollbacks", middleware.RequirePermission(), k8s.DaemonSet.GetDaemonSetRollbacks)

	// Job
	rg.GET("job/list", middleware.RequirePermission(), k8s.Job.GetJobList)
	rg.POST("job/list", middleware.RequirePermission(), k8s.Job.GetJobList)
	rg.GET("job/detail", middleware.RequirePermission(), k8s.Job.GetJobByName)
	rg.GET("job/get-yaml", middleware.RequirePermission(), k8s.Job.GetJobYaml)
	rg.GET("job/events", middleware.RequirePermission(), k8s.Job.GetJobEvents)
	rg.GET("job/pods", middleware.RequirePermission(), k8s.Job.JobPodList)
	rg.POST("job/create", middleware.RequirePermission(), k8s.Job.CreateJob)
	rg.PUT("job/update", middleware.RequirePermission(), k8s.Job.UpdateJob)
	rg.DELETE("job/delete", middleware.RequirePermission(), k8s.Job.DeleteJob)
	rg.POST("job/rerun", middleware.RequirePermission(), k8s.Job.RerunJob)

	// CronJob
	rg.GET("cronjob/list", middleware.RequirePermission(), k8s.Cronjob.GetCronJobList)
	rg.POST("cronjob/list", middleware.RequirePermission(), k8s.Cronjob.GetCronJobList)
	rg.GET("cronjob/detail", middleware.RequirePermission(), k8s.Cronjob.GetCronJobByName)
	rg.GET("cronjob/get-yaml", middleware.RequirePermission(), k8s.Cronjob.GetCronJobYaml)
	rg.GET("cronjob/events", middleware.RequirePermission(), k8s.Cronjob.GetCronJobEvents)
	rg.GET("cronjob/jobs", middleware.RequirePermission(), k8s.Cronjob.CronJobJobsList)
	rg.POST("cronjob/create", middleware.RequirePermission(), k8s.Cronjob.CreateCronJob)
	rg.PUT("cronjob/update", middleware.RequirePermission(), k8s.Cronjob.UpdateCronJob)
	rg.DELETE("cronjob/delete", middleware.RequirePermission(), k8s.Cronjob.DeleteCronJobByName)
	rg.PUT("cronjob/suspend", middleware.RequirePermission(), k8s.Cronjob.SuspendCronJob)
	rg.PUT("cronjob/resume", middleware.RequirePermission(), k8s.Cronjob.ResumeCronJob)
	rg.POST("cronjob/trigger", middleware.RequirePermission(), k8s.Cronjob.TriggerCronJob)

	// ReplicaSet
	rg.GET("replicaset/list", middleware.RequirePermission(), k8s.ReplicaSet.GetReplicaSetList)
	rg.POST("replicaset/list", middleware.RequirePermission(), k8s.ReplicaSet.GetReplicaSetList)
	rg.GET("replicaset/get-yaml", middleware.RequirePermission(), k8s.ReplicaSet.GetReplicaSetYaml)
	rg.GET("replicaset/detail", middleware.RequirePermission(), k8s.ReplicaSet.GetReplicaSetDetail)
	rg.GET("replicaset/pods", middleware.RequirePermission(), k8s.ReplicaSet.GetReplicaSetPodList)
	rg.GET("replicaset/events", middleware.RequirePermission(), k8s.ReplicaSet.GetReplicaSetEvents)
	rg.DELETE("replicaset/delete", middleware.RequirePermission(), k8s.ReplicaSet.DeleteReplicaSet)

	// HPA
	rg.GET("hpa/list", middleware.RequirePermission(), k8s.Hpa.GetHPAList)
	rg.POST("hpa/list", middleware.RequirePermission(), k8s.Hpa.GetHPAList)
	rg.GET("hpa/detail", middleware.RequirePermission(), k8s.Hpa.GetHPADetail)
	rg.GET("hpa/get-yaml", middleware.RequirePermission(), k8s.Hpa.GetHPAYaml)
	rg.POST("hpa/create", middleware.RequirePermission(), k8s.Hpa.CreateHPA)
	rg.PUT("hpa/update", middleware.RequirePermission(), k8s.Hpa.UpdateHPA)
	rg.DELETE("hpa/delete", middleware.RequirePermission(), k8s.Hpa.DeleteHPA)
	rg.GET("hpa/events", middleware.RequirePermission(), k8s.Hpa.GetHPAEvents)
	rg.POST("hpa/pause", middleware.RequirePermission(), k8s.Hpa.PauseHPA)
	rg.POST("hpa/resume", middleware.RequirePermission(), k8s.Hpa.ResumeHPA)
}

func registerNetworkRoutes(rg *gin.RouterGroup) {
	// Service
	rg.GET("service/list", middleware.RequirePermission(), k8s.Service.GetServicesList)
	rg.POST("service/list", middleware.RequirePermission(), k8s.Service.GetServicesList)
	rg.GET("service/detail", middleware.RequirePermission(), k8s.Service.GetServicesByName)
	rg.GET("service/get-yaml", middleware.RequirePermission(), k8s.Service.GetServicesYaml)
	rg.GET("service/events", middleware.RequirePermission(), k8s.Service.GetServiceEvents)
	rg.GET("service/pods", middleware.RequirePermission(), k8s.Service.ServicePodList)
	rg.GET("service/endpoints", middleware.RequirePermission(), k8s.Service.GetServiceEndpoints)
	rg.POST("service/create", middleware.RequirePermission(), k8s.Service.CreateService)
	rg.PUT("service/update", middleware.RequirePermission(), k8s.Service.UpdateService)
	rg.DELETE("service/delete", middleware.RequirePermission(), k8s.Service.DeleteService)

	// Ingress
	rg.GET("ingress/list", middleware.RequirePermission(), k8s.Ingress.GetIngressList)
	rg.POST("ingress/list", middleware.RequirePermission(), k8s.Ingress.GetIngressList)
	rg.GET("ingress/detail", middleware.RequirePermission(), k8s.Ingress.GetIngressByName)
	rg.GET("ingress/get-yaml", middleware.RequirePermission(), k8s.Ingress.GetIngressYaml)
	rg.GET("ingress/events", middleware.RequirePermission(), k8s.Ingress.GetIngressEvents)
	rg.GET("ingress/tls-status", middleware.RequirePermission(), k8s.Ingress.CheckIngressTLSCertStatus)
	rg.GET("ingress/ingressclasses", middleware.RequirePermission(), k8s.Ingress.GetIngressClassList)
	rg.POST("ingress/create", middleware.RequirePermission(), k8s.Ingress.CreateIngress)
	rg.PUT("ingress/update", middleware.RequirePermission(), k8s.Ingress.UpdateIngress)
	rg.DELETE("ingress/delete", middleware.RequirePermission(), k8s.Ingress.DeleteIngressByName)

	// NetworkPolicy
	rg.GET("networkpolicy/list", middleware.RequirePermission(), k8s.NetworkPolicy.GetNetworkPolicyList)
	rg.POST("networkpolicy/list", middleware.RequirePermission(), k8s.NetworkPolicy.GetNetworkPolicyList)
	rg.GET("networkpolicy/detail", middleware.RequirePermission(), k8s.NetworkPolicy.GetNetworkPolicyDetail)
	rg.GET("networkpolicy/get-yaml", middleware.RequirePermission(), k8s.NetworkPolicy.GetNetworkPolicyYaml)
	rg.GET("networkpolicy/events", middleware.RequirePermission(), k8s.NetworkPolicy.GetNetworkPolicyEvents)
	rg.GET("networkpolicy/pods", middleware.RequirePermission(), k8s.NetworkPolicy.GetNetworkPolicyPods)
	rg.POST("networkpolicy/create", middleware.RequirePermission(), k8s.NetworkPolicy.CreateNetworkPolicy)
	rg.PUT("networkpolicy/update", middleware.RequirePermission(), k8s.NetworkPolicy.UpdateNetworkPolicy)
	rg.DELETE("networkpolicy/delete", middleware.RequirePermission(), k8s.NetworkPolicy.DeleteNetworkPolicy)
}

func registerStorageRoutes(rg *gin.RouterGroup) {
	// PV
	rg.GET("pv/list", middleware.RequirePermission(), k8s.Pv.GetPVList)
	rg.POST("pv/list", middleware.RequirePermission(), k8s.Pv.GetPVList)
	rg.GET("pv/detail", middleware.RequirePermission(), k8s.Pv.GetPVByName)
	rg.GET("pv/get-yaml", middleware.RequirePermission(), k8s.Pv.GetPVYaml)
	rg.POST("pv/create", middleware.RequirePermission(), k8s.Pv.CreatePV)
	rg.PUT("pv/update", middleware.RequirePermission(), k8s.Pv.UpdatePV)
	rg.DELETE("pv/delete", middleware.RequirePermission(), k8s.Pv.DeletePVByName)

	// PVC
	rg.GET("pvc/list", middleware.RequirePermission(), k8s.Pvc.GetPVCList)
	rg.POST("pvc/list", middleware.RequirePermission(), k8s.Pvc.GetPVCList)
	rg.GET("pvc/list-by-storageclass", middleware.RequirePermission(), k8s.Pvc.GetPVCListByStorageClass)
	rg.GET("pvc/detail", middleware.RequirePermission(), k8s.Pvc.GetPVCByName)
	rg.GET("pvc/get-yaml", middleware.RequirePermission(), k8s.Pvc.GetPVCYaml)
	rg.POST("pvc/create", middleware.RequirePermission(), k8s.Pvc.CreatePVC)
	rg.PUT("pvc/update", middleware.RequirePermission(), k8s.Pvc.UpdatePVC)
	rg.DELETE("pvc/delete", middleware.RequirePermission(), k8s.Pvc.DeletePVCByName)

	// StorageClass
	rg.GET("storageclass/list", middleware.RequirePermission(), k8s.StorageClass.GetStorageClassList)
	rg.POST("storageclass/list", middleware.RequirePermission(), k8s.StorageClass.GetStorageClassList)
	rg.GET("storageclass/detail", middleware.RequirePermission(), k8s.StorageClass.GetStorageClassByName)
	rg.GET("storageclass/get-yaml", middleware.RequirePermission(), k8s.StorageClass.GetStorageClassYaml)
	rg.POST("storageclass/create", middleware.RequirePermission(), k8s.StorageClass.CreateStorageClass)
	rg.PUT("storageclass/update", middleware.RequirePermission(), k8s.StorageClass.UpdateStorageClass)
	rg.DELETE("storageclass/delete", middleware.RequirePermission(), k8s.StorageClass.DeleteStorageClassByName)
	rg.GET("storageclass/events", middleware.RequirePermission(), k8s.StorageClass.GetStorageClassEvents)

	// VolumeSnapshot
	rg.GET("volumesnapshot/list", middleware.RequirePermission(), k8s.VolumeSnapshot.GetVolumeSnapshotList)
	rg.POST("volumesnapshot/list", middleware.RequirePermission(), k8s.VolumeSnapshot.GetVolumeSnapshotList)
	rg.GET("volumesnapshot/detail", middleware.RequirePermission(), k8s.VolumeSnapshot.GetVolumeSnapshotByName)
	rg.GET("volumesnapshot/get-yaml", middleware.RequirePermission(), k8s.VolumeSnapshot.GetVolumeSnapshotYaml)
	rg.POST("volumesnapshot/create", middleware.RequirePermission(), k8s.VolumeSnapshot.CreateVolumeSnapshot)
	rg.PUT("volumesnapshot/update", middleware.RequirePermission(), k8s.VolumeSnapshot.UpdateVolumeSnapshot)
	rg.DELETE("volumesnapshot/delete", middleware.RequirePermission(), k8s.VolumeSnapshot.DeleteVolumeSnapshotByName)

	// VolumeSnapshotClass
	rg.GET("volumesnapshotclass/list", middleware.RequirePermission(), k8s.VolumeSnapshotClass.GetVolumeSnapshotClassList)
	rg.POST("volumesnapshotclass/list", middleware.RequirePermission(), k8s.VolumeSnapshotClass.GetVolumeSnapshotClassList)
	rg.GET("volumesnapshotclass/detail", middleware.RequirePermission(), k8s.VolumeSnapshotClass.GetVolumeSnapshotClassByName)
	rg.GET("volumesnapshotclass/get-yaml", middleware.RequirePermission(), k8s.VolumeSnapshotClass.GetVolumeSnapshotClassYaml)
	rg.POST("volumesnapshotclass/create", middleware.RequirePermission(), k8s.VolumeSnapshotClass.CreateVolumeSnapshotClass)
	rg.PUT("volumesnapshotclass/update", middleware.RequirePermission(), k8s.VolumeSnapshotClass.UpdateVolumeSnapshotClass)
	rg.DELETE("volumesnapshotclass/delete", middleware.RequirePermission(), k8s.VolumeSnapshotClass.DeleteVolumeSnapshotClassByName)
}

func registerConfigRoutes(rg *gin.RouterGroup) {
	// ConfigMap
	rg.GET("configmap/list", middleware.RequirePermission(), k8s.ConfigMap.GetConfigMapList)
	rg.POST("configmap/list", middleware.RequirePermission(), k8s.ConfigMap.GetConfigMapList)
	rg.GET("configmap/detail", middleware.RequirePermission(), k8s.ConfigMap.GetConfigMapByName)
	rg.GET("configmap/get-yaml", middleware.RequirePermission(), k8s.ConfigMap.GetConfigMapYaml)
	rg.POST("configmap/create", middleware.RequirePermission(), k8s.ConfigMap.CreateConfigMapFromYaml)
	rg.PUT("configmap/update", middleware.RequirePermission(), k8s.ConfigMap.UpdateConfigMapFromYaml)
	rg.DELETE("configmap/delete", middleware.RequirePermission(), k8s.ConfigMap.DeleteConfigMapByName)

	// Secret
	rg.GET("secret/list", middleware.RequirePermission(), k8s.Secret.GetSecretsList)
	rg.POST("secret/list", middleware.RequirePermission(), k8s.Secret.GetSecretsList)
	rg.GET("secret/detail", middleware.RequirePermission(), k8s.Secret.GetSecretByName)
	rg.GET("secret/get-yaml", middleware.RequirePermission(), k8s.Secret.GetSecretYaml)
	rg.POST("secret/create", middleware.RequirePermission(), k8s.Secret.CreateSecretFromYaml)
	rg.PUT("secret/update", middleware.RequirePermission(), k8s.Secret.UpdateSecretFromYaml)
	rg.DELETE("secret/delete", middleware.RequirePermission(), k8s.Secret.DeleteSecret)

	// ResourceQuota
	rg.GET("resourcequota/list", middleware.RequirePermission(), k8s.ResourceQuota.GetResourceQuotaList)
	rg.POST("resourcequota/list", middleware.RequirePermission(), k8s.ResourceQuota.GetResourceQuotaList)
	rg.GET("resourcequota/detail", middleware.RequirePermission(), k8s.ResourceQuota.GetResourceQuotaDetail)
	rg.GET("resourcequota/get-yaml", middleware.RequirePermission(), k8s.ResourceQuota.GetResourceQuotaYaml)
	rg.POST("resourcequota/create", middleware.RequirePermission(), k8s.ResourceQuota.CreateResourceQuota)
	rg.PUT("resourcequota/update", middleware.RequirePermission(), k8s.ResourceQuota.UpdateResourceQuota)
	rg.DELETE("resourcequota/delete", middleware.RequirePermission(), k8s.ResourceQuota.DeleteResourceQuota)

	// LimitRange
	rg.GET("limitrange/list", middleware.RequirePermission(), k8s.LimitRange.GetLimitRangeList)
	rg.POST("limitrange/list", middleware.RequirePermission(), k8s.LimitRange.GetLimitRangeList)
	rg.GET("limitrange/detail", middleware.RequirePermission(), k8s.LimitRange.GetLimitRangeDetail)
	rg.GET("limitrange/get-yaml", middleware.RequirePermission(), k8s.LimitRange.GetLimitRangeYaml)
	rg.POST("limitrange/create", middleware.RequirePermission(), k8s.LimitRange.CreateLimitRange)
	rg.PUT("limitrange/update", middleware.RequirePermission(), k8s.LimitRange.UpdateLimitRange)
	rg.DELETE("limitrange/delete", middleware.RequirePermission(), k8s.LimitRange.DeleteLimitRange)
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
