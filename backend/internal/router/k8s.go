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
	// Cluster
	rg.GET("cluster/version", k8s.Cluster.GetClusterVersion)
	rg.GET("cluster/nodes", k8s.Cluster.GetClusterNodesInfo)

	// Node
	rg.GET("node/detail", k8s.Node.GetNodeDetail)
	rg.GET("node/get-yaml", k8s.Node.GetNodeYaml)
	rg.GET("node/pods", k8s.Node.GetNodePods)
	rg.GET("node/events", k8s.Node.GetNodeEvents)
	rg.PUT("node/cordon", k8s.Node.CordonNode)
	rg.PUT("node/taint", k8s.Node.SetTaintNode)
	rg.PUT("node/taints", k8s.Node.UpdateNodeTaints)
	rg.PUT("node/labels", k8s.Node.UpdateNodeLabels)
	rg.PUT("node/drain", k8s.Node.DrainNode)
	rg.PUT("node/update-yaml", k8s.Node.UpdateNodeYaml)
	rg.DELETE("node/delete", k8s.Node.DeleteNode)

	// Namespace
	rg.GET("namespace/list", k8s.Namespace.GetNamespaceList)
	rg.GET("namespace/detail", k8s.Namespace.GetNamespaceDetail)
	rg.GET("namespace/get-yaml", k8s.Namespace.GetNamespaceYaml)
	rg.POST("namespace/create", k8s.Namespace.CreateNamespace)
	rg.PUT("namespace/update", k8s.Namespace.UpdateNamespace)
	rg.PUT("namespace/labels", k8s.Namespace.UpdateNamespaceLabels)
	rg.DELETE("namespace/delete", k8s.Namespace.DeleteNamespace)

	// Pod
	rg.GET("pod/list", k8s.Pod.GetPodList)
	rg.GET("pod/detail", k8s.Pod.GetPodByName)
	rg.GET("pod/get-yaml", k8s.Pod.GetPodYaml)
	rg.GET("pod/events", k8s.Pod.WatchPodEvent)
	rg.POST("pod/create", k8s.Pod.CreatePod)
	rg.PUT("pod/update", k8s.Pod.UpdatePod)
	rg.PUT("pod/update-yaml", k8s.Pod.UpdatePod)
	rg.DELETE("pod/delete", k8s.Pod.DeletePodByName)

	// Event
	rg.GET("event/list", k8s.Event.ListEvents)
	rg.GET("event/watch", k8s.Event.WatchEvents)
	// Container
	rg.GET("container/exec", k8s.HandleWebSocket)
	rg.GET("container/record/list", k8s.RecordList)
	rg.GET("container/record/url", k8s.RecordUrl)
	rg.GET("log", k8s.PodContainerLog)
	rg.GET("log/stream", k8s.StreamPodContainerLogs)
}

func registerWorkloadRoutes(rg *gin.RouterGroup) {
	// Deployment
	rg.GET("deployment/list", k8s.Deployment.GetDeploymentList)
	rg.GET("deployment/detail", k8s.Deployment.GetDeploymentDetail)
	rg.GET("deployment/get-yaml", k8s.Deployment.GetDeploymentYaml)
	rg.GET("deployment/events", k8s.Deployment.GetDeploymentEvents)
	rg.POST("deployment/create", k8s.Deployment.CreateDeployment)
	rg.PUT("deployment/update-yaml", k8s.Deployment.UpdateDeployment)
	rg.DELETE("deployment/delete", k8s.Deployment.DeleteDeployment)
	rg.PUT("deployment/scale", k8s.Deployment.ScaleDeployment)
	rg.POST("deployment/restart", k8s.Deployment.RestartDeployment)
	rg.POST("deployment/rollback", k8s.Deployment.RollbackDeployment)
	rg.PUT("deployment/update-image", k8s.Deployment.UpdateDeploymentImage)
	rg.GET("deployment/pods", k8s.Deployment.DeploymentPodList)
	rg.GET("deployment/replicasets", k8s.Deployment.GetDeploymentReplicaSets)

	// StatefulSet
	rg.GET("statefulset/list", k8s.StatefulSet.GetStatefulSetList)
	rg.GET("statefulset/detail", k8s.StatefulSet.GetStatefulSetByName)
	rg.GET("statefulset/get-yaml", k8s.StatefulSet.GetStatefulSetYaml)
	rg.GET("statefulset/events", k8s.StatefulSet.GetStatefulSetEvents)
	rg.GET("statefulset/pods", k8s.StatefulSet.StatefulSetPodList)
	rg.POST("statefulset/create", k8s.StatefulSet.CreateStatefulSet)
	rg.PUT("statefulset/update", k8s.StatefulSet.UpdateStatefulSet)
	rg.DELETE("statefulset/delete", k8s.StatefulSet.DeleteStatefulSetByName)
	rg.PUT("statefulset/scale", k8s.StatefulSet.ScaleStatefulSet)
	rg.PUT("statefulset/restart", k8s.StatefulSet.RestartStatefulSet)
	rg.POST("statefulset/rollback", k8s.StatefulSet.RollbackStatefulSet)
	rg.PUT("statefulset/update-image", k8s.StatefulSet.UpdateStatefulSetImage)

	// DaemonSet
	rg.GET("daemonset/list", k8s.DaemonSet.GetDaemonSetList)
	rg.GET("daemonset/detail", k8s.DaemonSet.GetDaemonSetByName)
	rg.GET("daemonset/get-yaml", k8s.DaemonSet.GetDaemonSetYaml)
	rg.GET("daemonset/events", k8s.DaemonSet.GetDaemonSetEvents)
	rg.GET("daemonset/pods", k8s.DaemonSet.DaemonSetPodList)
	rg.POST("daemonset/create", k8s.DaemonSet.CreateDaemonSet)
	rg.PUT("daemonset/update", k8s.DaemonSet.UpdateDaemonSet)
	rg.DELETE("daemonset/delete", k8s.DaemonSet.DeleteDaemonSetByName)
	rg.POST("daemonset/restart", k8s.DaemonSet.RestartDaemonSet)
	rg.POST("daemonset/rollback", k8s.DaemonSet.RollbackDaemonSet)
	rg.PUT("daemonset/update-image", k8s.DaemonSet.UpdateDaemonSetImage)

	// Job
	rg.GET("job/list", k8s.Job.GetJobList)
	rg.GET("job/detail", k8s.Job.GetJobByName)
	rg.GET("job/get-yaml", k8s.Job.GetJobYaml)
	rg.GET("job/events", k8s.Job.GetJobEvents)
	rg.GET("job/pods", k8s.Job.JobPodList)
	rg.POST("job/create", k8s.Job.CreateJob)
	rg.PUT("job/update", k8s.Job.UpdateJob)
	rg.DELETE("job/delete", k8s.Job.DeleteJob)

	// CronJob
	rg.GET("cronjob/list", k8s.Cronjob.GetCronJobList)
	rg.GET("cronjob/detail", k8s.Cronjob.GetCronJobByName)
	rg.GET("cronjob/get-yaml", k8s.Cronjob.GetCronJobYaml)
	rg.GET("cronjob/events", k8s.Cronjob.GetCronJobEvents)
	rg.GET("cronjob/jobs", k8s.Cronjob.CronJobJobsList)
	rg.POST("cronjob/create", k8s.Cronjob.CreateCronJob)
	rg.PUT("cronjob/update", k8s.Cronjob.UpdateCronJob)
	rg.DELETE("cronjob/delete", k8s.Cronjob.DeleteCronJobByName)

	rg.PUT("cronjob/suspend", k8s.Cronjob.SuspendCronJob)
	rg.PUT("cronjob/resume", k8s.Cronjob.ResumeCronJob)
	rg.POST("cronjob/trigger", k8s.Cronjob.TriggerCronJob)

	// ReplicaSet
	rg.GET("replicaset/list", k8s.ReplicaSet.GetReplicaSetList)
	rg.GET("replicaset/get-yaml", k8s.ReplicaSet.GetReplicaSetYaml)
	rg.GET("replicaset/detail", k8s.ReplicaSet.GetReplicaSetDetail)
	rg.GET("replicaset/pods", k8s.ReplicaSet.GetReplicaSetPodList)
	rg.GET("replicaset/events", k8s.ReplicaSet.GetReplicaSetEvents)
	rg.DELETE("replicaset/delete", k8s.ReplicaSet.DeleteReplicaSet)

	// HPA
	rg.GET("hpa/list", k8s.Hpa.GetHPAList)
	rg.GET("hpa/detail", k8s.Hpa.GetHPADetail)
	rg.GET("hpa/get-yaml", k8s.Hpa.GetHPAYaml)
	rg.POST("hpa/create", k8s.Hpa.CreateHPA)
	rg.PUT("hpa/update", k8s.Hpa.UpdateHPA)
	rg.DELETE("hpa/delete", k8s.Hpa.DeleteHPA)

	// PDB
	rg.GET("pdb/list", k8s.Pdb.GetPDBList)
	rg.GET("pdb/detail", k8s.Pdb.GetPDBDetail)
	rg.GET("pdb/get-yaml", k8s.Pdb.GetPDBYaml)
	rg.POST("pdb/create", k8s.Pdb.CreatePDB)
	rg.PUT("pdb/update", k8s.Pdb.UpdatePDB)
	rg.DELETE("pdb/delete", k8s.Pdb.DeletePDB)
}

func registerNetworkRoutes(rg *gin.RouterGroup) {
	// Service
	rg.GET("service/list", k8s.Service.GetServicesList)
	rg.GET("service/detail", k8s.Service.GetServicesByName)
	rg.GET("service/get-yaml", k8s.Service.GetServicesYaml)
	rg.GET("service/events", k8s.Service.GetServiceEvents)
	rg.GET("service/pods", k8s.Service.ServicePodList)
	rg.POST("service/create", k8s.Service.CreateService)
	rg.PUT("service/update", k8s.Service.UpdateService)
	rg.DELETE("service/delete", k8s.Service.DeleteService)

	// Ingress
	rg.GET("ingress/list", k8s.Ingress.GetIngressList)
	rg.GET("ingress/detail", k8s.Ingress.GetIngressByName)
	rg.GET("ingress/get-yaml", k8s.Ingress.GetIngressYaml)
	rg.GET("ingress/events", k8s.Ingress.GetIngressEvents)
	rg.POST("ingress/create", k8s.Ingress.CreateIngress)
	rg.PUT("ingress/update", k8s.Ingress.UpdateIngress)
	rg.DELETE("ingress/delete", k8s.Ingress.DeleteIngressByName)

	// NetworkPolicy
	rg.GET("networkpolicy/list", k8s.NetworkPolicy.GetNetworkPolicyList)
	rg.GET("networkpolicy/detail", k8s.NetworkPolicy.GetNetworkPolicyDetail)
	rg.GET("networkpolicy/get-yaml", k8s.NetworkPolicy.GetNetworkPolicyYaml)
	rg.GET("networkpolicy/events", k8s.NetworkPolicy.GetNetworkPolicyEvents)
	rg.GET("networkpolicy/pods", k8s.NetworkPolicy.GetNetworkPolicyPods)
	rg.POST("networkpolicy/create", k8s.NetworkPolicy.CreateNetworkPolicy)
	rg.PUT("networkpolicy/update", k8s.NetworkPolicy.UpdateNetworkPolicy)
	rg.DELETE("networkpolicy/delete", k8s.NetworkPolicy.DeleteNetworkPolicy)
}

func registerStorageRoutes(rg *gin.RouterGroup) {
	// PV
	rg.GET("pv/list", k8s.Pv.GetPVList)
	rg.GET("pv/detail", k8s.Pv.GetPVByName)
	rg.GET("pv/get-yaml", k8s.Pv.GetPVYaml)
	rg.POST("pv/create", k8s.Pv.CreatePV)
	rg.PUT("pv/update", k8s.Pv.UpdatePV)
	rg.DELETE("pv/delete", k8s.Pv.DeletePVByName)

	// PVC
	rg.GET("pvc/list", k8s.Pvc.GetPVCList)
	rg.GET("pvc/detail", k8s.Pvc.GetPVCByName)
	rg.GET("pvc/get-yaml", k8s.Pvc.GetPVCYaml)
	rg.POST("pvc/create", k8s.Pvc.CreatePVC)
	rg.PUT("pvc/update", k8s.Pvc.UpdatePVC)
	rg.DELETE("pvc/delete", k8s.Pvc.DeletePVCByName)

	// StorageClass
	rg.GET("storageclass/list", k8s.StorageClass.GetStorageClassList)
	rg.GET("storageclass/detail", k8s.StorageClass.GetStorageClassByName)
	rg.GET("storageclass/get-yaml", k8s.StorageClass.GetStorageClassYaml)
	rg.POST("storageclass/create", k8s.StorageClass.CreateStorageClass)
	rg.PUT("storageclass/update", k8s.StorageClass.UpdateStorageClass)
	rg.DELETE("storageclass/delete", k8s.StorageClass.DeleteStorageClassByName)
	rg.GET("storageclass/events", k8s.StorageClass.GetStorageClassEvents)

	// VolumeSnapshot
	rg.GET("volumesnapshot/list", k8s.VolumeSnapshot.GetVolumeSnapshotList)
	rg.GET("volumesnapshot/detail", k8s.VolumeSnapshot.GetVolumeSnapshotByName)
	rg.GET("volumesnapshot/get-yaml", k8s.VolumeSnapshot.GetVolumeSnapshotYaml)
	rg.POST("volumesnapshot/create", k8s.VolumeSnapshot.CreateVolumeSnapshot)
	rg.PUT("volumesnapshot/update", k8s.VolumeSnapshot.UpdateVolumeSnapshot)
	rg.DELETE("volumesnapshot/delete", k8s.VolumeSnapshot.DeleteVolumeSnapshotByName)

	// VolumeSnapshotClass
	rg.GET("volumesnapshotclass/list", k8s.VolumeSnapshotClass.GetVolumeSnapshotClassList)
	rg.GET("volumesnapshotclass/detail", k8s.VolumeSnapshotClass.GetVolumeSnapshotClassByName)
	rg.GET("volumesnapshotclass/get-yaml", k8s.VolumeSnapshotClass.GetVolumeSnapshotClassYaml)
	rg.POST("volumesnapshotclass/create", k8s.VolumeSnapshotClass.CreateVolumeSnapshotClass)
	rg.PUT("volumesnapshotclass/update", k8s.VolumeSnapshotClass.UpdateVolumeSnapshotClass)
	rg.DELETE("volumesnapshotclass/delete", k8s.VolumeSnapshotClass.DeleteVolumeSnapshotClassByName)
}

func registerConfigRoutes(rg *gin.RouterGroup) {
	// ConfigMap
	rg.GET("configmap/list", k8s.ConfigMap.GetConfigMapList)
	rg.GET("configmap/detail", k8s.ConfigMap.GetConfigMapByName)
	rg.GET("configmap/get-yaml", k8s.ConfigMap.GetConfigMapYaml)
	rg.POST("configmap/create", k8s.ConfigMap.CreateConfigMapFromYaml)
	rg.PUT("configmap/update", k8s.ConfigMap.UpdateConfigMapFromYaml)
	rg.DELETE("configmap/delete", k8s.ConfigMap.DeleteConfigMapByName)

	// Secret
	rg.GET("secret/list", k8s.Secret.GetSecretsList)
	rg.GET("secret/detail", k8s.Secret.GetSecretByName)
	rg.GET("secret/get-yaml", k8s.Secret.GetSecretYaml)
	rg.POST("secret/create", k8s.Secret.CreateSecretFromYaml)
	rg.PUT("secret/update", k8s.Secret.UpdateSecretFromYaml)
	rg.DELETE("secret/delete", k8s.Secret.DeleteSecret)

	// ResourceQuota
	rg.GET("resourcequota/list", k8s.ResourceQuota.GetResourceQuotaList)
	rg.GET("resourcequota/detail", k8s.ResourceQuota.GetResourceQuotaDetail)
	rg.GET("resourcequota/get-yaml", k8s.ResourceQuota.GetResourceQuotaYaml)
	rg.POST("resourcequota/create", k8s.ResourceQuota.CreateResourceQuota)
	rg.PUT("resourcequota/update", k8s.ResourceQuota.UpdateResourceQuota)
	rg.DELETE("resourcequota/delete", k8s.ResourceQuota.DeleteResourceQuota)

	// LimitRange
	rg.GET("limitrange/list", k8s.LimitRange.GetLimitRangeList)
	rg.GET("limitrange/detail", k8s.LimitRange.GetLimitRangeDetail)
	rg.GET("limitrange/get-yaml", k8s.LimitRange.GetLimitRangeYaml)
	rg.POST("limitrange/create", k8s.LimitRange.CreateLimitRange)
	rg.PUT("limitrange/update", k8s.LimitRange.UpdateLimitRange)
	rg.DELETE("limitrange/delete", k8s.LimitRange.DeleteLimitRange)
}

func registerCrdRoutes(rg *gin.RouterGroup) {
	rg.GET("crd/list", k8s.Crd.GetCRDList)
	rg.GET("crd/detail", k8s.Crd.GetCRDDetail)
	rg.GET("crd/get-yaml", k8s.Crd.GetCRDYaml)
	rg.POST("crd/create", k8s.Crd.CreateCRD)
	rg.PUT("crd/update", k8s.Crd.UpdateCRD)
	rg.DELETE("crd/delete", k8s.Crd.DeleteCRD)
	rg.GET("crd/resources", k8s.Crd.GetCustomResourceList)
	rg.GET("crd/resource/yaml", k8s.Crd.GetCustomResourceYaml)
	rg.POST("crd/resource/create", k8s.Crd.CreateCustomResource)
	rg.DELETE("crd/resource", k8s.Crd.DeleteCustomResource)
	rg.PUT("crd/resource/update", k8s.Crd.UpdateCustomResource)
	rg.PATCH("crd/resource/patch", k8s.Crd.PatchCustomResource)
}

func registerAuditRoutes(rg *gin.RouterGroup) {
	rg.GET("audit/list", k8s.Audit.ListAuditLogs)
	rg.GET("audit/detail", k8s.Audit.GetAuditLog)
	rg.POST("audit/create", k8s.Audit.CreateAuditLog)
	rg.GET("audit/stats", k8s.Audit.GetAuditStats)
	// 审计清除属高危操作,需管理员
	rg.DELETE("audit/clear", middleware.RequireAdmin(), k8s.Audit.ClearAuditLogs)
}
