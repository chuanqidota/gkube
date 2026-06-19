package router

import (
	"github.com/gin-gonic/gin"

	authApi "gkube/app/auth/api"
	clusterApi "gkube/app/cluster/api"
	dashboardApi "gkube/app/dashboard/api"
	k8sApi "gkube/app/k8s/api"
	settingsApi "gkube/app/settings/api"
	"gkube/pkg/middleware"
)

func Engine() *gin.Engine {
	router := gin.Default()

	// 全局CORS中间件
	router.Use(middleware.CORSMiddleware())

	v1 := router.Group("v1")

	// ============ 公开路由（无需认证）============
	v1.POST("auth/login", authApi.Auth.Login)
	v1.POST("auth/refresh", authApi.Auth.Refresh)
	v1.GET("auth/oidc/login", authApi.OIDC.GetLoginURL)
	v1.GET("auth/oidc/callback", authApi.OIDC.HandleCallback)

	// ============ 需要JWT认证的路由============
	authorized := v1.Group("", middleware.JWTAuth())
	{
		authorized.GET("auth/me", authApi.Auth.GetMe)

		// Users
		users := authorized.Group("users", middleware.RBAC("user", "*"))
		{
			users.GET("", authApi.User.List)
			users.POST("", authApi.User.Create)
			users.PUT("", authApi.User.Update)
			users.DELETE("", authApi.User.Delete)
			users.PUT("change-password", authApi.User.ChangePassword)
		}

		// Roles
		roles := authorized.Group("roles", middleware.RBAC("role", "*"))
		{
			roles.GET("", authApi.Role.List)
			roles.POST("", authApi.Role.Create)
			roles.PUT("", authApi.Role.Update)
			roles.DELETE("", authApi.Role.Delete)
		}

		// RBAC Matrix
		authorized.GET("rbac/matrix", authApi.RBACMatrix.GetMatrix)
		authorized.GET("rbac/matrix/role", authApi.RBACMatrix.GetRoleMatrix)
		authorized.PUT("rbac/matrix", authApi.RBACMatrix.UpdateMatrix)
		authorized.POST("rbac/matrix/toggle", authApi.RBACMatrix.TogglePermission)
		authorized.POST("rbac/matrix/init", authApi.RBACMatrix.InitializeMatrix)

		// Clusters
		clusters := authorized.Group("clusters", middleware.RBAC("cluster", "*"))
		{
			clusters.GET("", clusterApi.Cluster.List)
			clusters.POST("", clusterApi.Cluster.Create)
			clusters.GET("/:id", clusterApi.Cluster.Detail)
			clusters.PUT("", clusterApi.Cluster.Update)
			clusters.DELETE("", clusterApi.Cluster.Delete)
			clusters.GET("/:id/check", clusterApi.Cluster.Check)
		}

		// Dashboard
		dashboard := authorized.Group("dashboard")
		{
			dashboard.GET("overview", dashboardApi.Dashboard.Overview)
			dashboard.GET("resources", dashboardApi.Dashboard.Resources)
			dashboard.GET("workloads", dashboardApi.Dashboard.Workloads)
			dashboard.GET("events", dashboardApi.Dashboard.Events)
		}

		// Settings
		settings := authorized.Group("settings")
		{
			settings.GET("auth", settingsApi.Settings.GetAuthSettings)
			settings.PUT("auth", settingsApi.Settings.UpdateAuthSettings)
		}

		// K8s resources - standardized URL patterns matching frontend
		k8s := authorized.Group("k8s")
		{
			// Cluster
			k8s.GET("cluster/version", k8sApi.Cluster.GetClusterVersion)
			k8s.GET("cluster/nodes", k8sApi.Cluster.GetClusterNodesInfo) // plural: /nodes

			// Node
			k8s.GET("node/detail", k8sApi.Node.GetNodeDetail)          // NEW
			k8s.GET("node/get-yaml", k8sApi.Node.GetNodeYaml)           // /get-yaml
			k8s.GET("node/pods", k8sApi.Node.GetNodePods)
			k8s.GET("node/events", k8sApi.Node.GetNodeEvents)           // NEW
			k8s.PUT("node/cordon", k8sApi.Node.UnscheduledNode)         // /cordon
			k8s.PUT("node/taint", k8sApi.Node.SetTaintNode)
			k8s.PUT("node/update-yaml", k8sApi.Node.UpdateNodeYaml)     // NEW

			// Namespace
			k8s.GET("namespace/list", k8sApi.Namespace.GetNamespaceList)
			k8s.GET("namespace/detail", k8sApi.Namespace.GetNamespaceDetail)
			k8s.GET("namespace/get-yaml", k8sApi.Namespace.GetNamespaceYaml)
			k8s.POST("namespace/create", k8sApi.Namespace.CreateNamespace)
			k8s.PUT("namespace/update", k8sApi.Namespace.UpdateNamespace)
			k8s.DELETE("namespace/delete", k8sApi.Namespace.DeleteNamespace)

			// Deployment
			k8s.GET("deployment/list", k8sApi.Deployment.GetDeploymentList)
			k8s.GET("deployment/detail", k8sApi.Deployment.GetDeploymentDetail)
			k8s.GET("deployment/get-yaml", k8sApi.Deployment.GetDeploymentYaml)
			k8s.GET("deployment/events", k8sApi.Deployment.GetDeploymentEvents) // NEW
			k8s.POST("deployment/create", k8sApi.Deployment.CreateDeployment)
			k8s.PUT("deployment/update", k8sApi.Deployment.UpdateDeployment)
			k8s.PUT("deployment/update-yaml", k8sApi.Deployment.UpdateDeployment) // alias
			k8s.DELETE("deployment/delete", k8sApi.Deployment.DeleteDeployment)   // /delete
			k8s.PUT("deployment/scale", k8sApi.Deployment.ScaleDeployment)       // PUT not POST
			k8s.POST("deployment/restart", k8sApi.Deployment.RestartDeployment)
			k8s.POST("deployment/rollback", k8sApi.Deployment.RollbackDeployment)
			k8s.GET("deployment/pods", k8sApi.Deployment.DeploymentPodList)
			k8s.GET("deployment/replicasets", k8sApi.Deployment.GetDeploymentReplicaSets)

			// StatefulSet (lowercase: statefulset)
			k8s.GET("statefulset/list", k8sApi.StatefulSet.GetStatefulSetList)
			k8s.GET("statefulset/detail", k8sApi.StatefulSet.GetStatefulSetByName) // /detail
			k8s.GET("statefulset/get-yaml", k8sApi.StatefulSet.GetStatefulSetYaml) // /get-yaml
			k8s.POST("statefulset/create", k8sApi.StatefulSet.CreateStatefulSet)
			k8s.PUT("statefulset/update", k8sApi.StatefulSet.UpdateStatefulSet)
			k8s.DELETE("statefulset/delete", k8sApi.StatefulSet.DeleteStatefulSetByName) // /delete

			// DaemonSet (lowercase: daemonset)
			k8s.GET("daemonset/list", k8sApi.DaemonSet.GetDaemonSetList)
			k8s.GET("daemonset/detail", k8sApi.DaemonSet.GetDaemonSetByName) // /detail
			k8s.GET("daemonset/get-yaml", k8sApi.DaemonSet.GetDaemonSetYaml) // /get-yaml
			k8s.POST("daemonset/create", k8sApi.DaemonSet.CreateDaemonSet)
			k8s.PUT("daemonset/update", k8sApi.DaemonSet.UpdateDaemonSet)
			k8s.DELETE("daemonset/delete", k8sApi.DaemonSet.DeleteDaemonSetByName) // /delete

			// Job
			k8s.GET("job/list", k8sApi.Job.GetJobList)
			k8s.GET("job/detail", k8sApi.Job.GetJobByName)           // /detail
			k8s.GET("job/get-yaml", k8sApi.Job.GetJobYaml)           // /get-yaml
			k8s.POST("job/create", k8sApi.Job.CreateJob)
			k8s.PUT("job/update", k8sApi.Job.UpdateJob)
			k8s.DELETE("job/delete", k8sApi.Job.DeleteJob)

			// CronJob
			k8s.GET("cronjob/list", k8sApi.Cronjob.GetCronJobList)
			k8s.GET("cronjob/detail", k8sApi.Cronjob.GetCronJobByName) // /detail
			k8s.GET("cronjob/get-yaml", k8sApi.Cronjob.GetCronJobYaml) // /get-yaml
			k8s.POST("cronjob/create", k8sApi.Cronjob.CreateCronJob)
			k8s.PUT("cronjob/update", k8sApi.Cronjob.UpdateCronJob)
			k8s.DELETE("cronjob/delete", k8sApi.Cronjob.DeleteCronJobByName) // /delete

			// Service
			k8s.GET("service/list", k8sApi.Service.GetServicesList)
			k8s.GET("service/detail", k8sApi.Service.GetServicesByName) // /detail
			k8s.GET("service/get-yaml", k8sApi.Service.GetServicesYaml) // /get-yaml
			k8s.POST("service/create", k8sApi.Service.CreateService)
			k8s.PUT("service/update", k8sApi.Service.UpdateService)
			k8s.DELETE("service/delete", k8sApi.Service.DeleteService)

			// Ingress
			k8s.GET("ingress/list", k8sApi.Ingress.GetIngressList)
			k8s.GET("ingress/detail", k8sApi.Ingress.GetIngressByName) // /detail
			k8s.GET("ingress/get-yaml", k8sApi.Ingress.GetIngressYaml) // /get-yaml
			k8s.POST("ingress/create", k8sApi.Ingress.CreateIngress)
			k8s.PUT("ingress/update", k8sApi.Ingress.UpdateIngress)
			k8s.DELETE("ingress/delete", k8sApi.Ingress.DeleteIngressByName) // /delete

			// ConfigMap
			k8s.GET("configmap/list", k8sApi.ConfigMap.GetConfigMapList)
			k8s.GET("configmap/detail", k8sApi.ConfigMap.GetConfigMapByName) // /detail
			k8s.GET("configmap/get-yaml", k8sApi.ConfigMap.GetConfigMapYaml) // /get-yaml
			k8s.POST("configmap/create", k8sApi.ConfigMap.CreateConfigMap)
			k8s.PUT("configmap/update", k8sApi.ConfigMap.UpdateConfigMap)
			k8s.DELETE("configmap/delete", k8sApi.ConfigMap.DeleteConfigMapByName) // /delete

			// Secret
			k8s.GET("secret/list", k8sApi.Secret.GetSecretsList)
			k8s.GET("secret/detail", k8sApi.Secret.GetSecretByName) // /detail
			k8s.GET("secret/get-yaml", k8sApi.Secret.GetSecretYaml) // /get-yaml
			k8s.POST("secret/create", k8sApi.Secret.CreateSecret)
			k8s.PUT("secret/update", k8sApi.Secret.UpdateSecret)
			k8s.DELETE("secret/delete", k8sApi.Secret.DeleteSecret)

			// Pod
			k8s.GET("pod/list", k8sApi.Pod.GetPodList)
			k8s.GET("pod/detail", k8sApi.Pod.GetPodByName)             // /detail
			k8s.GET("pod/get-yaml", k8sApi.Pod.GetPodYaml)             // /get-yaml
			k8s.GET("pod/events", k8sApi.Pod.WatchPodEvent)            // GET not POST
			k8s.POST("pod/create", k8sApi.Pod.CreatePod)               // POST not GET
			k8s.PUT("pod/update", k8sApi.Pod.UpdatePod)                 // PUT not GET
			k8s.PUT("pod/update-yaml", k8sApi.Pod.UpdatePod)            // alias
			k8s.DELETE("pod/delete", k8sApi.Pod.DeletePodByName)        // DELETE not GET

			// PV
			k8s.GET("pv/list", k8sApi.Pv.GetPVList)
			k8s.GET("pv/detail", k8sApi.Pv.GetPVByName)             // /detail
			k8s.GET("pv/get-yaml", k8sApi.Pv.GetPVYaml)             // /get-yaml
			k8s.POST("pv/create", k8sApi.Pv.CreatePV)
			k8s.PUT("pv/update", k8sApi.Pv.UpdatePV)
			k8s.DELETE("pv/delete", k8sApi.Pv.DeletePVByName)        // /delete

			// PVC
			k8s.GET("pvc/list", k8sApi.Pvc.GetPVCList)
			k8s.GET("pvc/detail", k8sApi.Pvc.GetPVCByName)             // /detail
			k8s.GET("pvc/get-yaml", k8sApi.Pvc.GetPVCYaml)             // /get-yaml
			k8s.POST("pvc/create", k8sApi.Pvc.CreatePVC)
			k8s.DELETE("pvc/delete", k8sApi.Pvc.DeletePVCByName)        // /delete

			// StorageClass (lowercase: storageclass)
			k8s.GET("storageclass/list", k8sApi.StorageClass.GetStorageClassList)
			k8s.GET("storageclass/detail", k8sApi.StorageClass.GetStorageClassByName) // /detail
			k8s.GET("storageclass/get-yaml", k8sApi.StorageClass.GetStorageClassYaml) // /get-yaml
			k8s.POST("storageclass/create", k8sApi.StorageClass.CreateStorageClass)
			k8s.PUT("storageclass/update", k8sApi.StorageClass.UpdateStorageClass)
			k8s.DELETE("storageclass/delete", k8sApi.StorageClass.DeleteStorageClassByName)

			// Container
			k8s.GET("container/exec", k8sApi.HandleWebSocket)
			k8s.GET("container/record/list", k8sApi.RecordList)
			k8s.GET("container/record/url", k8sApi.RecordUrl)
			k8s.GET("log", k8sApi.PodContainerLog)
			k8s.GET("log/stream", k8sApi.StreamPodContainerLogs)

			// HPA
			k8s.GET("hpa/list", k8sApi.Hpa.GetHPAList)
			k8s.GET("hpa/detail", k8sApi.Hpa.GetHPADetail)
			k8s.GET("hpa/yaml", k8sApi.Hpa.GetHPAYaml)
			k8s.GET("hpa/get-yaml", k8sApi.Hpa.GetHPAYaml) // alias
			k8s.POST("hpa/create", k8sApi.Hpa.CreateHPA)
			k8s.PUT("hpa/update", k8sApi.Hpa.UpdateHPA)
			k8s.DELETE("hpa/delete", k8sApi.Hpa.DeleteHPA)

			// NetworkPolicy
			k8s.GET("networkpolicy/list", k8sApi.NetworkPolicy.GetNetworkPolicyList)
			k8s.GET("networkpolicy/detail", k8sApi.NetworkPolicy.GetNetworkPolicyDetail)
			k8s.GET("networkpolicy/yaml", k8sApi.NetworkPolicy.GetNetworkPolicyYaml)
			k8s.GET("networkpolicy/get-yaml", k8sApi.NetworkPolicy.GetNetworkPolicyYaml) // alias
			k8s.POST("networkpolicy/create", k8sApi.NetworkPolicy.CreateNetworkPolicy)
			k8s.PUT("networkpolicy/update", k8sApi.NetworkPolicy.UpdateNetworkPolicy)
			k8s.DELETE("networkpolicy/delete", k8sApi.NetworkPolicy.DeleteNetworkPolicy)

			// PDB
			k8s.GET("pdb/list", k8sApi.Pdb.GetPDBList)
			k8s.GET("pdb/detail", k8sApi.Pdb.GetPDBDetail)
			k8s.GET("pdb/yaml", k8sApi.Pdb.GetPDBYaml)
			k8s.GET("pdb/get-yaml", k8sApi.Pdb.GetPDBYaml) // alias
			k8s.POST("pdb/create", k8sApi.Pdb.CreatePDB)
			k8s.PUT("pdb/update", k8sApi.Pdb.UpdatePDB)
			k8s.DELETE("pdb/delete", k8sApi.Pdb.DeletePDB)

			// ResourceQuota
			k8s.GET("resourcequota/list", k8sApi.ResourceQuota.GetResourceQuotaList)
			k8s.GET("resourcequota/detail", k8sApi.ResourceQuota.GetResourceQuotaDetail)
			k8s.GET("resourcequota/yaml", k8sApi.ResourceQuota.GetResourceQuotaYaml)
			k8s.GET("resourcequota/get-yaml", k8sApi.ResourceQuota.GetResourceQuotaYaml) // alias
			k8s.POST("resourcequota/create", k8sApi.ResourceQuota.CreateResourceQuota)
			k8s.PUT("resourcequota/update", k8sApi.ResourceQuota.UpdateResourceQuota)
			k8s.DELETE("resourcequota/delete", k8sApi.ResourceQuota.DeleteResourceQuota)

			// LimitRange
			k8s.GET("limitrange/list", k8sApi.LimitRange.GetLimitRangeList)
			k8s.GET("limitrange/detail", k8sApi.LimitRange.GetLimitRangeDetail)
			k8s.GET("limitrange/yaml", k8sApi.LimitRange.GetLimitRangeYaml)
			k8s.GET("limitrange/get-yaml", k8sApi.LimitRange.GetLimitRangeYaml) // alias
			k8s.POST("limitrange/create", k8sApi.LimitRange.CreateLimitRange)
			k8s.PUT("limitrange/update", k8sApi.LimitRange.UpdateLimitRange)
			k8s.DELETE("limitrange/delete", k8sApi.LimitRange.DeleteLimitRange)

			// CRD
			k8s.GET("crd/list", k8sApi.Crd.GetCRDList)
			k8s.GET("crd/detail", k8sApi.Crd.GetCRDDetail)
			k8s.GET("crd/yaml", k8sApi.Crd.GetCRDYaml)
			k8s.GET("crd/get-yaml", k8sApi.Crd.GetCRDYaml) // alias
			k8s.POST("crd/create", k8sApi.Crd.CreateCRD)
			k8s.PUT("crd/update", k8sApi.Crd.UpdateCRD)
			k8s.DELETE("crd/delete", k8sApi.Crd.DeleteCRD)
			k8s.GET("crd/resources", k8sApi.Crd.GetCustomResourceList)
			k8s.GET("crd/resource/yaml", k8sApi.Crd.GetCustomResourceYaml)
			k8s.POST("crd/resource/create", k8sApi.Crd.CreateCustomResource)
			k8s.DELETE("crd/resource", k8sApi.Crd.DeleteCustomResource)

			// Audit Logs
			k8s.GET("audit/list", k8sApi.Audit.ListAuditLogs)
			k8s.GET("audit/detail", k8sApi.Audit.GetAuditLog)
			k8s.POST("audit/create", k8sApi.Audit.CreateAuditLog)
			k8s.GET("audit/stats", k8sApi.Audit.GetAuditStats)
			k8s.DELETE("audit/clear", k8sApi.Audit.ClearAuditLogs)
		}
	}

	return router
}
