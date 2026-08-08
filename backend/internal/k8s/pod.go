package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sEvent "gkube/pkg/k8s/event"
	k8sPod "gkube/pkg/k8s/pod"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/apimachinery/pkg/fields"
)

type pod struct{}

var Pod = new(pod)

// GetPodList 获取pod列表（支持分页）
func (p *pod) GetPodList(c *gin.Context) {
	var query PodQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}

	limit, continueToken := k8sclient.GetPaginationParams(c)
	podList, err := k8sPod.ListPods(client, query.Namespace, limit, continueToken)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pod列表失败")
		return
	}
	remaining := int64(0)
	if podList.RemainingItemCount != nil {
		remaining = *podList.RemainingItemCount
	}
	// Total = 当前页条数 + 剩余条数,接近集群内真实总数(分页时);非分页时 remaining=0,Total=全量条数
	data := k8sclient.BuildPaginatedData(podList.Items, podList.Continue, remaining, limit)
	data.Total = len(podList.Items) + int(remaining)
	response.Success(c, "获取pod列表成功", data)
}

// GetPodByName 获取pod根据名称
func (p *pod) GetPodByName(c *gin.Context) {
	var query PodQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	pod, err := k8sPod.GetPodByName(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pod失败")
		return
	}
	response.Success(c, "获取pod成功", pod)
}

// GetPodYaml 获取pod的yaml
func (p *pod) GetPodYaml(c *gin.Context) {
	var query PodQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	podYaml, err := k8sPod.GetPodYaml(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pod失败")
		return
	}
	response.Success(c, "获取pod成功", gin.H{"yaml": podYaml})
}

// CreatePod 创建pod
func (p *pod) CreatePod(c *gin.Context) {
	var query PodCreateParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sPod.CreatePod(client, query.Namespace, query.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建pod失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdatePod 更新pod
func (p *pod) UpdatePod(c *gin.Context) {
	var query PodUpdateParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sPod.UpdatePod(client, query.Namespace, query.Name, query.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新pod失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeletePodByName 删除pod根据名称
func (p *pod) DeletePodByName(c *gin.Context) {
	var query PodDeleteByNameParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sPod.DeletePodByName(client, query.Namespace, query.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除pod失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// ListPodEvents 返回与指定 Pod 关联的 K8s 事件列表(JSON)。
// 用 fields.Selector 构造 involvedObject.name 过滤,避免 Name 注入非法字段选择器语法。
// 前端 PodDetail 总是传 name;name 为空时选择器匹配 involvedObject.name 为空的事件,结果为空集。
func (p *pod) ListPodEvents(c *gin.Context) {
	var query PodEventQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	selector := fields.OneTermEqualSelector("involvedObject.name", query.Name).String()
	events, _, _, err := k8sEvent.ListEvents(client, query.Namespace, selector, 0, "")
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pod事件失败")
		return
	}
	response.Success(c, "获取pod事件成功", events)
}

type PodQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"Pod Yaml"`
}

type PodUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"Pod Yaml"`
}

type PodDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodEventQueryParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"Pod名称"`
}
