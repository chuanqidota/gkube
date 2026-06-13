package router

import (
	"github.com/gin-gonic/gin"

	authApi "gkube/app/auth/api"
	clusterApi "gkube/app/cluster/api"
	dashboardApi "gkube/app/dashboard/api"
	k8sApi "gkube/app/k8s/api"
	"gkube/pkg/middleware"
)

func Engine() *gin.Engine {
	router := gin.Default()

	// 全局CORS中间件
	router.Use(middleware.CORSMiddleware())

	v1 := router.Group("v1")

	// ============ 公开路由（无需认证）============
	// 登录 / 刷新Token
	v1.POST("auth/login", authApi.Auth.Login)
	v1.POST("auth/refresh", authApi.Auth.Refresh)

	// ============ 需要JWT认证的路由============
	authorized := v1.Group("", middleware.JWTAuth())
	{
		// ----- Auth: 获取当前用户信息 -----
		authorized.GET("auth/me", authApi.Auth.GetMe)

		// ----- Users: 用户管理 (RBAC user) -----
		users := authorized.Group("users", middleware.RBAC("user", "*"))
		{
			users.GET("", authApi.User.List)
			users.POST("", authApi.User.Create)
			users.PUT("", authApi.User.Update)
			users.DELETE("", authApi.User.Delete)
			users.PUT("change-password", authApi.User.ChangePassword)
		}

		// ----- Roles: 角色管理 (RBAC role) -----
		roles := authorized.Group("roles", middleware.RBAC("role", "*"))
		{
			roles.GET("", authApi.Role.List)
			roles.POST("", authApi.Role.Create)
			roles.PUT("", authApi.Role.Update)
			roles.DELETE("", authApi.Role.Delete)
		}

		// ----- Clusters: 集群管理 (RBAC cluster) -----
		clusters := authorized.Group("clusters", middleware.RBAC("cluster", "*"))
		{
			clusters.GET("", clusterApi.Cluster.List)
			clusters.POST("", clusterApi.Cluster.Create)
			clusters.GET("/:id", clusterApi.Cluster.Detail)
			clusters.PUT("", clusterApi.Cluster.Update)
			clusters.DELETE("", clusterApi.Cluster.Delete)
			clusters.GET("/:id/check", clusterApi.Cluster.Check)
		}

		// ----- Dashboard: 仪表盘 -----
		dashboard := authorized.Group("dashboard")
		{
			dashboard.GET("overview", dashboardApi.Dashboard.Overview)
			dashboard.GET("resources", dashboardApi.Dashboard.Resources)
			dashboard.GET("workloads", dashboardApi.Dashboard.Workloads)
			dashboard.GET("events", dashboardApi.Dashboard.Events)
		}

		// ----- K8s: Kubernetes资源管理（暂不加RBAC）-----
		k8s := authorized.Group("k8s")
		{
			k8s.GET("cluster/version", k8sApi.Cluster.GetClusterVersion) // 获取集群版本信息
			k8s.GET("cluster/node", k8sApi.Cluster.GetClusterNodesInfo)  // 获取所有集群信息

			k8s.GET("node/yaml", k8sApi.Node.GetNodeYaml)                  // 获取节点yaml
			k8s.GET("node/pods", k8sApi.Node.GetNodePods)                  // 获取节点的pods
			k8s.PUT("node/unscheduled", k8sApi.Node.UnscheduledNode)       // 禁止调度
			k8s.POST("node/evict-all", k8sApi.Node.EvictsNodeAllPods)      // 驱逐节点上所有的pod
			k8s.POST("node/evict-single", k8sApi.Node.EvictsNodeSinglePod) // 驱逐节点上的单个pod
			k8s.PUT("node/taint", k8sApi.Node.SetTaintNode)                // 给节点设置污点

			k8s.GET("namespace/list", k8sApi.Namespace.GetNamespaceList)   // 获取命名空间
			k8s.POST("namespace/create", k8sApi.Namespace.CreateNamespace) // 创建命名空间

			k8s.GET("job/list", k8sApi.Job.GetJobList)                     // 获取任务列表
			k8s.GET("job/list-by-name", k8sApi.Job.GetJobByName)           // 根据名称获取任务
			k8s.POST("job/list-by-field", k8sApi.Job.GetJobByFiled)        // 根据字段获取任务
			k8s.POST("job/list-by-label", k8sApi.Job.GetJobByLabel)        // 根据标签获取任务
			k8s.GET("job/yaml", k8sApi.Job.GetJobYaml)                     // 获取任务的yaml
			k8s.POST("job/create", k8sApi.Job.CreateJob)                   // 创建任务
			k8s.PUT("job/update", k8sApi.Job.UpdateJob)                    // 更新任务
			k8s.DELETE("job/delete", k8sApi.Job.DeleteJob)                 // 删除任务
			k8s.DELETE("job/delete/by-field", k8sApi.Job.DeleteJobByField) // 根据字段删除任务
			k8s.DELETE("job/delete/by-label", k8sApi.Job.DeleteJobByLabel) // 根据标签删除任务

			k8s.GET("cronjob/list", k8sApi.Cronjob.GetCronJobList)                     // 获取定时任务
			k8s.GET("cronjob/list-by-name", k8sApi.Cronjob.GetCronJobByName)           // 获取定时任务根据名称
			k8s.POST("cronjob/list-by-label", k8sApi.Cronjob.GetCronJobByLabel)        // 获取定时任务根据标签
			k8s.POST("cronjob/list-by-field", k8sApi.Cronjob.GetCronJobByField)        // 获取定时任务根据字段
			k8s.GET("cronjob/yaml", k8sApi.Cronjob.GetCronJobYaml)                     // 获取定时任务详情
			k8s.POST("cronjob/create", k8sApi.Cronjob.CreateCronJob)                   // 创建定时任务
			k8s.PUT("cronjob/update", k8sApi.Cronjob.UpdateCronJob)                    // 更新定时任务
			k8s.DELETE("cronjob/delete-by-name", k8sApi.Cronjob.DeleteCronJobByName)   // 删除定时任务根据名称
			k8s.DELETE("cronjob/delete-by-label", k8sApi.Cronjob.DeleteCronJobByLabel) // 删除定时任务根据标签
			k8s.DELETE("cronjob/delete-by-field", k8sApi.Cronjob.DeleteCronJobByField) // 删除定时任务根据字段

			k8s.GET("ingress/list", k8sApi.Ingress.GetIngressList)                     // 获取ingress列表
			k8s.GET("ingress/list-by-name", k8sApi.Ingress.GetIngressByName)           // 获取ingress根据名称
			k8s.POST("ingress/list-by-label", k8sApi.Ingress.GetIngressByLabel)        // 获取ingress根据标签
			k8s.POST("ingress/list-by-field", k8sApi.Ingress.GetIngressByField)        // 获取ingress根据字段
			k8s.GET("ingress/yaml", k8sApi.Ingress.GetIngressYaml)                     // 获取ingress的yaml
			k8s.POST("ingress/create", k8sApi.Ingress.CreateIngress)                   // 创建ingress
			k8s.PUT("ingress/update", k8sApi.Ingress.UpdateIngress)                    // 更新ingress
			k8s.DELETE("ingress/delete-by-name", k8sApi.Ingress.DeleteIngressByName)   // 删除ingress根据名称
			k8s.DELETE("ingress/delete-by-label", k8sApi.Ingress.DeleteIngressByLabel) // 删除ingress根据标签
			k8s.DELETE("ingress/delete-by-field", k8sApi.Ingress.DeleteIngressByField) // 删除ingress根据字段

			k8s.GET("service/list", k8sApi.Service.GetServicesList)                // 获取服务
			k8s.GET("service/list-by-name", k8sApi.Service.GetServicesByName)      // 获取服务根据名称
			k8s.POST("service/list-by-label", k8sApi.Service.GetServicesByLabel)   // 获取服务根据标签
			k8s.POST("service/list-by-field", k8sApi.Service.GetServicesByField)   // 获取服务根据字段
			k8s.GET("service/yaml", k8sApi.Service.GetServicesYaml)                // 获取服务的yaml
			k8s.POST("service/create", k8sApi.Service.CreateService)               // 创建服务
			k8s.PUT("service/update", k8sApi.Service.UpdateService)                // 更新服务
			k8s.DELETE("service/delete", k8sApi.Service.DeleteService)              // 删除服务

			k8s.GET("deployment/list", k8sApi.Deployment.GetDeploymentList)                     // 获取deployment
			k8s.GET("deployment/detail", k8sApi.Deployment.GetDeploymentDetail)                // 获取deployment详情
			k8s.GET("deployment/get-yaml", k8sApi.Deployment.GetDeploymentYaml)                // 获取deployment yaml
			k8s.POST("deployment/list-by-field", k8sApi.Deployment.GetDeploymentByField)        // 根据字段获取deployment
			k8s.POST("deployment/list-by-label", k8sApi.Deployment.GetDeploymentByLabel)        // 根据标签获取deployment
			k8s.POST("deployment/create", k8sApi.Deployment.CreateDeployment)                   // 创建deployment
			k8s.PUT("deployment/update", k8sApi.Deployment.UpdateDeployment)                    // 更新deployment
			k8s.DELETE("deployment/delete-by-name", k8sApi.Deployment.DeleteDeployment)         // 删除deployment
			k8s.DELETE("deployment/delete-by-field", k8sApi.Deployment.DeleteDeploymentByField) // 根据字段删除deployment
			k8s.DELETE("deployment/delete-by-label", k8sApi.Deployment.DeleteDeploymentByLabel) // 根据标签删除deployment
			k8s.POST("deployment/scale", k8sApi.Deployment.ScaleDeployment)                     // 扩容deployment
			k8s.POST("deployment/restart", k8sApi.Deployment.RestartDeployment)                 // 重启deployment
			k8s.POST("deployment/rollback", k8sApi.Deployment.RollbackDeployment)               // 回滚deployment
			k8s.GET("deployment/pods", k8sApi.Deployment.DeploymentPodList)                     // 获取deployment pods

			k8s.GET("statefulSet/list", k8sApi.StatefulSet.GetStatefulSetList)                     // 获取有状态服务列表
			k8s.GET("statefulSet/list-by-name", k8sApi.StatefulSet.GetStatefulSetByName)           // 获取有状态服务根据名称
			k8s.POST("statefulSet/list-by-field", k8sApi.StatefulSet.GetStatefulSetByField)        // 获取有状态服务根据字段
			k8s.POST("statefulSet/list-by-label", k8sApi.StatefulSet.GetStatefulSetByLabel)        // 获取有状态服务根据标签
			k8s.GET("statefulSet/yaml", k8sApi.StatefulSet.GetStatefulSetYaml)                     // 获取有状态服务的yaml
			k8s.POST("statefulSet/create", k8sApi.StatefulSet.CreateStatefulSet)                   // 创建有状态服务
			k8s.PUT("statefulSet/update", k8sApi.StatefulSet.UpdateStatefulSet)                    // 更新有状态服务
			k8s.DELETE("statefulSet/delete-by-name", k8sApi.StatefulSet.DeleteStatefulSetByName)   // 删除有状态服务根据名称
			k8s.DELETE("statefulSet/delete-by-label", k8sApi.StatefulSet.DeleteStatefulSetByLabel) // 删除有状态服务根据标签
			k8s.DELETE("statefulSet/delete-by-field", k8sApi.StatefulSet.DeleteStatefulSetByField) // 删除有状态服务根据字段

			k8s.GET("daemonSet/list", k8sApi.DaemonSet.GetDaemonSetList)                     // 获取守护进程集
			k8s.GET("daemonSet/list-by-name", k8sApi.DaemonSet.GetDaemonSetByName)           // 获取守护进程根据名称
			k8s.POST("daemonSet/list-by-label", k8sApi.DaemonSet.GetDaemonSetByLabel)        // 获取守护进程根据标签
			k8s.POST("daemonSet/list-by-field", k8sApi.DaemonSet.GetDaemonSetByField)        // 获取守护进程根据字段
			k8s.GET("daemonSet/yaml", k8sApi.DaemonSet.GetDaemonSetYaml)                     // 获取守护进程的yaml
			k8s.POST("daemonSet/create", k8sApi.DaemonSet.CreateDaemonSet)                   // 创建守护进程集
			k8s.PUT("daemonSet/update", k8sApi.DaemonSet.UpdateDaemonSet)                    // 更新守护进程
			k8s.DELETE("daemonSet/delete-by-name", k8sApi.DaemonSet.DeleteDaemonSetByName)   // 删除守护进程根据名称
			k8s.DELETE("daemonSet/delete-by-label", k8sApi.DaemonSet.DeleteDaemonSetByLabel) // 删除守护进程根据标签
			k8s.DELETE("daemonSet/delete-by-field", k8sApi.DaemonSet.DeleteDaemonSetByField) // 删除守护进程根据字段

			k8s.GET("pv/list", k8sApi.Pv.GetPVList)                     // 获取存储pv
			k8s.GET("pv/list/by-name", k8sApi.Pv.GetPVByName)           // 获取存储pv根据名称
			k8s.POST("pv/list/by-label", k8sApi.Pv.GetPVByLabel)        // 获取存储pv根据标签
			k8s.POST("pv/list/by-field", k8sApi.Pv.GetPVByField)        // 获取存储pv根据字段
			k8s.GET("pv/yaml", k8sApi.Pv.GetPVYaml)                     // 获取存储pv的yaml
			k8s.POST("pv/create", k8sApi.Pv.CreatePV)                   // 创建存储pv
			k8s.PUT("pv/update", k8sApi.Pv.UpdatePV)                    // 更新存储pv
			k8s.DELETE("pv/delete/by-name", k8sApi.Pv.DeletePVByName)   // 删除存储pv根据名称
			k8s.DELETE("pv/delete/by-label", k8sApi.Pv.DeletePVByLabel) // 删除存储pv根据标签
			k8s.DELETE("pv/delete/by-field", k8sApi.Pv.DeletePVByField) // 删除存储pv根据字段

			k8s.GET("pvc/list", k8sApi.Pvc.GetPVCList)                     //获取存储pvc
			k8s.GET("pvc/list/by-name", k8sApi.Pvc.GetPVCByName)           // 获取存储pvc根据名称
			k8s.POST("pvc/list/by-label", k8sApi.Pvc.GetPVCByLabel)        // 获取存储pvc根据标签
			k8s.POST("pvc/list/by-field", k8sApi.Pvc.GetPVCByField)        // 获取存储pvc根据字段
			k8s.GET("pvc/yaml", k8sApi.Pvc.GetPVCYaml)                     // 获取存储pvc的yaml
			k8s.POST("pvc/create", k8sApi.Pvc.CreatePVC)                   // 创建存储pvc
			k8s.DELETE("pvc/delete/by-name", k8sApi.Pvc.DeletePVCByName)   // 删除存储pvc根据名称
			k8s.DELETE("pvc/delete/by-label", k8sApi.Pvc.DeletePVCByLabel) // 删除存储pvc根据标签
			k8s.DELETE("pvc/delete/by-field", k8sApi.Pvc.DeletePVCByField) // 删除存储pvc根据字段

			k8s.GET("storageClass/list", k8sApi.StorageClass.GetStorageClassList)                     // 获取存储sc列表
			k8s.GET("storageClass/list-by-name", k8sApi.StorageClass.GetStorageClassByName)           // 获取存储sc根据名称
			k8s.POST("storageClass/list-by-field", k8sApi.StorageClass.GetStorageClassByField)        // 获取存储sc根据字段
			k8s.POST("storageClass/list-by-label", k8sApi.StorageClass.GetStorageClassByLabel)        // 获取存储sc根据标签
			k8s.GET("storageClass/yaml", k8sApi.StorageClass.GetStorageClassYaml)                     // 获取存储sc的yaml
			k8s.POST("storageClass/create", k8sApi.StorageClass.CreateStorageClass)                   // 创建存储sc
			k8s.GET("storageClass/update", k8sApi.StorageClass.UpdateStorageClass)                    // 更新存储sc
			k8s.DELETE("storageClass/delete-by-name", k8sApi.StorageClass.DeleteStorageClassByName)   // 删除存储sc根据名称
			k8s.DELETE("storageClass/delete-by-field", k8sApi.StorageClass.DeleteStorageClassByField) // 删除存储sc根据字段
			k8s.DELETE("storageClass/delete-by-label", k8sApi.StorageClass.DeleteStorageClassByLabel) // 删除存储sc根据标签

			k8s.GET("configmap/list", k8sApi.ConfigMap.GetConfigMapList)                   // 获取configmap
			k8s.GET("configmap/by-name", k8sApi.ConfigMap.GetConfigMapByName)              // 获取configmap根据名称
			k8s.GET("configmap/yaml", k8sApi.ConfigMap.GetConfigMapYaml)                   // 获取configmap的yaml
			k8s.POST("configmap/create", k8sApi.ConfigMap.CreateConfigMap)                 // 创建configmap
			k8s.PUT("configmap/update", k8sApi.ConfigMap.UpdateConfigMap)                  // 更新configmap
			k8s.DELETE("configmap/delete-by-name", k8sApi.ConfigMap.DeleteConfigMapByName) // 删除configmap根据名称

			k8s.GET("secret/list", k8sApi.Secret.GetSecretsList)          // 获取secret列表
			k8s.GET("secret/list-by-name", k8sApi.Secret.GetSecretByName) // 获取secret根据名称
			k8s.GET("secret/yaml", k8sApi.Secret.GetSecretYaml)           // 获取secret的yaml
			k8s.POST("secret/create", k8sApi.Secret.CreateSecret)         // 创建secret
			k8s.PUT("secret/update", k8sApi.Secret.UpdateSecret)          // 更新secret
			k8s.DELETE("secret/delete", k8sApi.Secret.DeleteSecret)       // 删除secret

			k8s.GET("pod/list", k8sApi.Pod.GetPodList)                  // 获取pod列表
			k8s.GET("pod/list-by-name", k8sApi.Pod.GetPodByName)        // 获取pod根据名称
			k8s.GET("pod/list-by-label", k8sApi.Pod.GetPodByLabel)      // 获取pod根据标签
			k8s.GET("pod/list-by-field", k8sApi.Pod.GetPodByField)      // 获取pod根据字段
			k8s.GET("pod/yaml", k8sApi.Pod.GetPodYaml)                  // 获取pod的yaml
			k8s.GET("pod/create", k8sApi.Pod.CreatePod)                 // 创建pod
			k8s.GET("pod/update", k8sApi.Pod.UpdatePod)                 // 更新pod
			k8s.GET("pod/delete-by-name", k8sApi.Pod.DeletePodByName)   // 删除pod根据名称
			k8s.GET("pod/delete-by-label", k8sApi.Pod.DeletePodByLabel) // 删除pod根据标签
			k8s.GET("pod/delete-by-field", k8sApi.Pod.DeletePodByField) // 删除pod根据字段
			k8s.POST("pod/events", k8sApi.Pod.WatchPodEvent)            // 监听pod事件

			// container资源
			k8s.GET("container/exec", k8sApi.HandleWebSocket)     // websocket container
			k8s.GET("container/record/list", k8sApi.RecordList)   // 操作记录列表
			k8s.GET("container/record/url", k8sApi.RecordUrl)     // 操作审计
			k8s.GET("/log", k8sApi.PodContainerLog)               // 获取容器日志
			k8s.GET("/log/stream", k8sApi.StreamPodContainerLogs) // 获取容器日志流

			// HPA资源
			k8s.GET("hpa/list", k8sApi.Hpa.GetHPAList)       // 获取HPA列表
			k8s.GET("hpa/detail", k8sApi.Hpa.GetHPADetail)   // 获取HPA详情
			k8s.GET("hpa/yaml", k8sApi.Hpa.GetHPAYaml)       // 获取HPA YAML
			k8s.POST("hpa/create", k8sApi.Hpa.CreateHPA)     // 创建HPA
			k8s.PUT("hpa/update", k8sApi.Hpa.UpdateHPA)      // 更新HPA
			k8s.DELETE("hpa/delete", k8sApi.Hpa.DeleteHPA)   // 删除HPA

			// NetworkPolicy资源
			k8s.GET("networkpolicy/list", k8sApi.NetworkPolicy.GetNetworkPolicyList)       // 获取NetworkPolicy列表
			k8s.GET("networkpolicy/detail", k8sApi.NetworkPolicy.GetNetworkPolicyDetail)   // 获取NetworkPolicy详情
			k8s.GET("networkpolicy/yaml", k8sApi.NetworkPolicy.GetNetworkPolicyYaml)       // 获取NetworkPolicy YAML
			k8s.POST("networkpolicy/create", k8sApi.NetworkPolicy.CreateNetworkPolicy)     // 创建NetworkPolicy
			k8s.PUT("networkpolicy/update", k8sApi.NetworkPolicy.UpdateNetworkPolicy)      // 更新NetworkPolicy
			k8s.DELETE("networkpolicy/delete", k8sApi.NetworkPolicy.DeleteNetworkPolicy)   // 删除NetworkPolicy

			// RBAC资源
			k8s.GET("serviceaccount/list", k8sApi.Rbac.GetServiceAccountList)         // 获取ServiceAccount列表
			k8s.GET("serviceaccount/yaml", k8sApi.Rbac.GetServiceAccountYaml)         // 获取ServiceAccount YAML
			k8s.DELETE("serviceaccount/delete", k8sApi.Rbac.DeleteServiceAccount)     // 删除ServiceAccount

			k8s.GET("clusterrole/list", k8sApi.Rbac.GetClusterRoleList)              // 获取ClusterRole列表
			k8s.GET("clusterrole/yaml", k8sApi.Rbac.GetClusterRoleYaml)              // 获取ClusterRole YAML
			k8s.DELETE("clusterrole/delete", k8sApi.Rbac.DeleteClusterRole)          // 删除ClusterRole

			k8s.GET("role/list", k8sApi.Rbac.GetRoleList)                            // 获取Role列表
			k8s.GET("role/yaml", k8sApi.Rbac.GetRoleYaml)                            // 获取Role YAML
			k8s.DELETE("role/delete", k8sApi.Rbac.DeleteRole)                        // 删除Role

			k8s.GET("clusterrolebinding/list", k8sApi.Rbac.GetClusterRoleBindingList)    // 获取ClusterRoleBinding列表
			k8s.GET("clusterrolebinding/yaml", k8sApi.Rbac.GetClusterRoleBindingYaml)    // 获取ClusterRoleBinding YAML
			k8s.DELETE("clusterrolebinding/delete", k8sApi.Rbac.DeleteClusterRoleBinding) // 删除ClusterRoleBinding

			k8s.GET("rolebinding/list", k8sApi.Rbac.GetRoleBindingList)              // 获取RoleBinding列表
			k8s.GET("rolebinding/yaml", k8sApi.Rbac.GetRoleBindingYaml)              // 获取RoleBinding YAML
			k8s.DELETE("rolebinding/delete", k8sApi.Rbac.DeleteRoleBinding)          // 删除RoleBinding

			// PDB资源
			k8s.GET("pdb/list", k8sApi.Pdb.GetPDBList)       // 获取PDB列表
			k8s.GET("pdb/detail", k8sApi.Pdb.GetPDBDetail)   // 获取PDB详情
			k8s.GET("pdb/yaml", k8sApi.Pdb.GetPDBYaml)       // 获取PDB YAML
			k8s.POST("pdb/create", k8sApi.Pdb.CreatePDB)     // 创建PDB
			k8s.PUT("pdb/update", k8sApi.Pdb.UpdatePDB)      // 更新PDB
			k8s.DELETE("pdb/delete", k8sApi.Pdb.DeletePDB)   // 删除PDB

			// CRD资源
			k8s.GET("crd/list", k8sApi.Crd.GetCRDList)                           // 获取CRD列表
			k8s.GET("crd/detail", k8sApi.Crd.GetCRDDetail)                       // 获取CRD详情
			k8s.GET("crd/yaml", k8sApi.Crd.GetCRDYaml)                           // 获取CRD YAML
			k8s.GET("crd/resources", k8sApi.Crd.GetCustomResourceList)            // 获取自定义资源列表
			k8s.GET("crd/resource/yaml", k8sApi.Crd.GetCustomResourceYaml)        // 获取自定义资源YAML
			k8s.DELETE("crd/resource", k8sApi.Crd.DeleteCustomResource)            // 删除自定义资源
		}
	}

	return router
}
