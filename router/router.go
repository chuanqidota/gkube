package router

import (
	"gkube/app/k8s/api"

	"github.com/gin-gonic/gin"
)

func Engine() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("v1")

	k8sRouter := v1.Group("k8s")
	{
		k8sRouter.GET("cluster/version", api.Cluster.GetClusterVersion) // 获取集群版本信息
		k8sRouter.GET("cluster/node", api.Cluster.GetClusterNodesInfo)  // 获取所有集群信息

		k8sRouter.GET("node/yaml", api.Node.GetNodeYaml)                  // 获取节点yaml
		k8sRouter.GET("node/pods", api.Node.GetNodePods)                  // 获取节点的pods
		k8sRouter.PUT("node/unscheduled", api.Node.UnscheduledNode)       // 禁止调度
		k8sRouter.POST("node/evict-all", api.Node.EvictsNodeAllPods)      // 驱逐节点上所有的pod
		k8sRouter.POST("node/evict-single", api.Node.EvictsNodeSinglePod) // 驱逐节点上的单个pod
		k8sRouter.PUT("node/taint", api.Node.SetTaintNode)                // 给节点设置污点

		k8sRouter.GET("namespace/list", api.Namespace.GetNamespaceList)   // 获取命名空间
		k8sRouter.POST("namespace/create", api.Namespace.CreateNamespace) // 创建命名空间

		k8sRouter.GET("job/list", api.Job.GetJobList)                     // 获取任务列表
		k8sRouter.GET("job/by-name", api.Job.GetJobByName)                // 根据名称获取任务
		k8sRouter.POST("job/by-field", api.Job.GetJobByFiled)             // 根据字段获取任务
		k8sRouter.POST("job/by-label", api.Job.GetJobByLabel)             // 根据标签获取任务
		k8sRouter.GET("job/yaml", api.Job.GetJobYaml)                     // 获取任务的yaml
		k8sRouter.POST("job/create", api.Job.CreateJob)                   // 创建任务
		k8sRouter.PUT("job/update", api.Job.UpdateJob)                    // 更新任务
		k8sRouter.DELETE("job/delete", api.Job.DeleteJob)                 // 删除任务
		k8sRouter.DELETE("job/delete/by-field", api.Job.DeleteJobByField) // 根据字段删除任务
		k8sRouter.DELETE("job/delete/by-label", api.Job.DeleteJobByLabel) // 根据标签删除任务

		k8sRouter.GET("cronjob")        // 获取定时任务
		k8sRouter.GET("cronjob/detail") // 获取定时任务详情
		k8sRouter.POST("cronjob")       // 创建定时任务
		k8sRouter.DELETE("cronjob")     // 删除定时任务

		k8sRouter.GET("ingress")        // 获取ingress
		k8sRouter.GET("ingress/detail") // 获取ingress详情
		k8sRouter.POST("ingress")       // 创建ingress
		k8sRouter.DELETE("ingress")     // 删除ingress

		k8sRouter.GET("service")        // 获取服务
		k8sRouter.GET("service/detail") // 获取服务详情
		k8sRouter.POST("service")       // 创建服务
		k8sRouter.DELETE("service")     // 删除服务

		k8sRouter.GET("deployment/list", api.Deployment.GetDeploymentList)                     // 获取deployment
		k8sRouter.POST("deployment/list-by-field", api.Deployment.GetDeploymentByField)        // 根据字段获取deployment
		k8sRouter.POST("deployment/list-by-label", api.Deployment.GetDeploymentByLabel)        // 根据标签获取deployment
		k8sRouter.POST("deployment/create", api.Deployment.CreateDeployment)                   // 创建deployment                                        // 创建deployment
		k8sRouter.PUT("deployment/update", api.Deployment.UpdateDeployment)                    // 更新deployment                                        // 创建deployment
		k8sRouter.DELETE("deployment/delete-by-name", api.Deployment.DeleteDeployment)         // 删除deployment                                         // 删除deployment
		k8sRouter.DELETE("deployment/delete-by-field", api.Deployment.DeleteDeploymentByField) // 根据字段删除deployment
		k8sRouter.DELETE("deployment/delete-by-label", api.Deployment.DeleteDeploymentByLabel) // 根据标签删除deployment
		k8sRouter.POST("deployment/scale", api.Deployment.ScaleDeployment)                     // 扩容deployment
		k8sRouter.POST("deployment/restart", api.Deployment.RestartDeployment)                 // 重启deployment
		k8sRouter.GET("deployment/pods", api.Deployment.DeploymentPodList)                     // 获取deployment pods

		k8sRouter.GET("statefulset")          // 获取有状态服务
		k8sRouter.GET("statefulset/detail")   // 获取有状态服务详情
		k8sRouter.POST("statefulset")         // 创建有状态服务
		k8sRouter.DELETE("statefulset")       // 删除有状态服务
		k8sRouter.POST("statefulset/restart") // 重启有状态服务
		k8sRouter.POST("statefulset/scale")   // 扩所容有状态服务

		k8sRouter.GET("daemonset")          // 获取守护进程集
		k8sRouter.GET("daemonset/detail")   // 获取守护进程集详情
		k8sRouter.POST("daemonset")         // 创建守护进程集
		k8sRouter.DELETE("daemonset")       // 删除守护进程集
		k8sRouter.POST("daemonset/restart") // 重启守护进程集

		k8sRouter.GET("pv/list", api.Pv.GetPVList)                     // 获取存储pv
		k8sRouter.GET("pv/list/by-name", api.Pv.GetPVByName)           // 获取存储pv根据名称
		k8sRouter.POST("pv/list/by-label", api.Pv.GetPVByLabel)        // 获取存储pv根据标签
		k8sRouter.POST("pv/list/by-field", api.Pv.GetPVByField)        // 获取存储pv根据字段
		k8sRouter.GET("pv/yaml", api.Pv.GetPVYaml)                     // 获取存储pv的yaml
		k8sRouter.POST("pv/create", api.Pv.CreatePV)                   // 创建存储pv
		k8sRouter.PUT("pv/update", api.Pv.UpdatePV)                    // 更新存储pv
		k8sRouter.DELETE("pv/delete/by-name", api.Pv.DeletePVByName)   // 删除存储pv根据名称
		k8sRouter.DELETE("pv/delete/by-label", api.Pv.DeletePVByLabel) // 删除存储pv根据标签
		k8sRouter.DELETE("pv/delete/by-field", api.Pv.DeletePVByField) // 删除存储pv根据字段

		k8sRouter.GET("storage/storageclass")        // 获取存储storageclass
		k8sRouter.GET("storage/storageclass/detail") // 获取存储storageclass详情
		k8sRouter.DELETE("storage/storageclass")     // 删除存储storageclass

		k8sRouter.GET("configmap")        // 获取configmap
		k8sRouter.GET("configmap/detail") // 获取configmap详情
		k8sRouter.POST("configmap")       // 创建configmap
		k8sRouter.DELETE("configmap")     // 删除configmap

		k8sRouter.GET("secret")        // 获取secret
		k8sRouter.GET("secret/detail") // 获取secret详情
		k8sRouter.POST("secret")       // 创建secret
		k8sRouter.DELETE("secret")     // 删除secret

		k8sRouter.GET("events") // 事件

		k8sRouter.GET("pod")          //  获取pod列表
		k8sRouter.GET("pod/detail")   // 获取pod详情
		k8sRouter.GET("pod/selector") // 通过标签查询pod

		// container资源
		k8sRouter.GET("container/exec", api.HandleWebSocket)     // websocket container
		k8sRouter.GET("container/record/list", api.RecordList)   // 操作记录列表
		k8sRouter.GET("container/record/url", api.RecordUrl)     // 操作审计
		k8sRouter.GET("/log", api.PodContainerLog)               // 获取容器日志
		k8sRouter.GET("/log/stream", api.StreamPodContainerLogs) // 获取容器日志流

	}

	return router
}
