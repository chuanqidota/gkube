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
		k8sRouter.POST("cluster") // 集群创建
		k8sRouter.GET("cluster") // 获取所有集群信息
		k8sRouter.GET("cluster/:id") // 获取集群信息
		k8sRouter.DELETE("cluster") // 删除集群

		k8sRouter.GET("namespace") // 获取命名空间

		k8sRouter.GET("events") // 事件

		k8sRouter.GET("node") // 获取所有节点信息
		k8sRouter.GET("node/detail") // 获取节点信息
		k8sRouter.DELETE("node") // 删除节点
		k8sRouter.POST("node/schedule") // 节点调度
		k8sRouter.POST("node/cordon") // 排空所有pod

		k8sRouter.GET("deployment") // 获取deployment
		k8sRouter.GET("deployment/detail")
		k8sRouter.POST("deployment") // 创建deployment
		k8sRouter.DELETE("deployment") // 删除deployment
		
		k8sRouter.POST("deployment/scale") // 扩容
		k8sRouter.POST("deployment/restart") // 重启

		k8sRouter.GET("pod") //  获取pod列表
		k8sRouter.GET("pod/detail") // 获取pod详情
		k8sRouter.GET("pod/selector") // 通过标签查询pod


		k8sRouter.GET("statefulset") // 获取有状态服务
		k8sRouter.GET("statefulset/detail") // 获取有状态服务详情
		k8sRouter.POST("statefulset") // 创建有状态服务
		k8sRouter.DELETE("statefulset") // 删除有状态服务
		k8sRouter.POST("statefulset/restart") // 重启有状态服务
		k8sRouter.POST("statefulset/scale") // 扩所容有状态服务

		k8sRouter.GET("daemonset") // 获取守护进程集
		k8sRouter.GET("daemonset/detail") // 获取守护进程集详情
		k8sRouter.POST("daemonset") // 创建守护进程集
		k8sRouter.DELETE("daemonset") // 删除守护进程集
		k8sRouter.POST("daemonset/restart") // 重启守护进程集


		k8sRouter.GET("job") // 获取任务
		k8sRouter.GET("job/detail") // 获取任务详情
		k8sRouter.POST("job") // 创建任务
		k8sRouter.DELETE("job") // 删除任务
		k8sRouter.POST("job/scale") // 扩容任务

		k8sRouter.GET("cronjob") // 获取定时任务
		k8sRouter.GET("cronjob/detail") // 获取定时任务详情
		k8sRouter.POST("cronjob") // 创建定时任务
		k8sRouter.DELETE("cronjob") // 删除定时任务

		k8sRouter.GET("storage/pvc") // 获取存储pvc
		k8sRouter.GET("storage/pvc/detail") // 获取存储pvc详情
		k8sRouter.DELETE("storage/pvc") // 删除存储pvc

		k8sRouter.GET("storage/pv") // 获取存储pv
		k8sRouter.GET("storage/pv/detail") // 获取存储pv详情
		k8sRouter.DELETE("storage/pv") // 删除存储pv


		k8sRouter.GET("storage/storageclass") // 获取存储storageclass
		k8sRouter.GET("storage/storageclass/detail") // 获取存储storageclass详情
		k8sRouter.DELETE("storage/storageclass") // 删除存储storageclass

		k8sRouter.GET("service") // 获取服务
		k8sRouter.GET("service/detail") // 获取服务详情
		k8sRouter.POST("service") // 创建服务
		k8sRouter.DELETE("service") // 删除服务

		k8sRouter.GET("ingress") // 获取ingress
		k8sRouter.GET("ingress/detail") // 获取ingress详情
		k8sRouter.POST("ingress") // 创建ingress
		k8sRouter.DELETE("ingress") // 删除ingress


		k8sRouter.GET("configmap") // 获取configmap
		k8sRouter.GET("configmap/detail") // 获取configmap详情
		k8sRouter.POST("configmap") // 创建configmap
		k8sRouter.DELETE("configmap") // 删除configmap


		k8sRouter.GET("secret") // 获取secret
		k8sRouter.GET("secret/detail") // 获取secret详情
		k8sRouter.POST("secret") // 创建secret
		k8sRouter.DELETE("secret") // 删除secret


		k8sRouter.GET("container/exec",api.HandleWebSocket) // websocket container
		k8sRouter.GET("container/record/list",api.RecordList) // 操作记录列表
		k8sRouter.GET("container/record/url",api.RecordUrl) // 操作审计
		
		k8sRouter.GET("/log/:namespace/:pod/:container") // 获取容器日志

	}
	
	return router
}
