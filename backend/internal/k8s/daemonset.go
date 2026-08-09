package k8s

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sDaemonSet "gkube/pkg/k8s/daemonset"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type daemonSet struct {
}

var DaemonSet = new(daemonSet)

func (d *daemonSet) GetDaemonSetList(c *gin.Context) {
	var query DaemonSetQueryListParams
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
	if limit > 0 {
		dsList, err := k8sDaemonSet.ListDaemonSets(client, query.Namespace, limit, continueToken)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取DaemonSet列表失败")
			return
		}
		remaining := int64(0)
		if dsList.RemainingItemCount != nil {
			remaining = *dsList.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(dsList.Items, dsList.Continue, remaining, limit)
		data.Total = len(dsList.Items) + int(remaining)
		response.Success(c, "执行成功", data)
	} else {
		daemonSets, err := k8sDaemonSet.GetDaemonSetList(client, query.Namespace)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取DaemonSet列表失败")
			return
		}
		response.Success(c, "执行成功", daemonSets)
	}
}

func (d *daemonSet) GetDaemonSetByName(c *gin.Context) {
	var query DaemonSetQueryByNameParams
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

	daemonSet, err := k8sDaemonSet.GetDaemonSetByName(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取DaemonSet失败")
		return
	}
	response.Success(c, "执行成功", daemonSet)
}

func (d *daemonSet) GetDaemonSetYaml(c *gin.Context) {
	var query DaemonSetQueryByNameParams
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

	daemonSetYaml, err := k8sDaemonSet.GetDaemonSetYaml(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取DaemonSet YAML失败")
		return
	}
	response.Success(c, "执行成功", daemonSetYaml)
}

func (d *daemonSet) CreateDaemonSet(c *gin.Context) {
	var body DaemonSetCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sDaemonSet.CreateDaemonSet(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建DaemonSet失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

func (d *daemonSet) UpdateDaemonSet(c *gin.Context) {
	var body DaemonSetUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sDaemonSet.UpdateDaemonSet(client, body.Namespace, body.Name, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新DaemonSet失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

func (d *daemonSet) DeleteDaemonSetByName(c *gin.Context) {
	var body DaemonSetDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}

	if err := k8sDaemonSet.DeleteDaemonSetByName(client, body.Namespace, body.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除DaemonSet失败")
		return
	}

	response.Success(c, "执行成功", nil)
}

// GetDaemonSetEvents
//
//	@Description: 获取daemonset事件
//	@receiver d
//	@param c
func (d *daemonSet) GetDaemonSetEvents(c *gin.Context) {
	var query DaemonSetQueryByNameParams
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
	events, err := client.CoreV1().Events(query.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=DaemonSet", query.Name),
	})
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取DaemonSet事件失败")
		return
	}
	var result []map[string]any
	for _, event := range events.Items {
		lastSeen := ""
		if !event.LastTimestamp.IsZero() {
			lastSeen = event.LastTimestamp.Time.Format("2006-01-02 15:04:05")
		}
		result = append(result, map[string]any{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"last_seen": lastSeen,
		})
	}
	response.Success(c, "执行成功", result)
}

// DaemonSetPodList
//
//	@Description: 获取daemonset关联的pod列表
//	@receiver d
//	@param c
func (d *daemonSet) DaemonSetPodList(c *gin.Context) {
	var query DaemonSetQueryByNameParams
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
	podList, err := k8sDaemonSet.DaemonSetPodList(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取DaemonSet Pod列表失败")
		return
	}
	response.Success(c, "执行成功", podList)
}

// RestartDaemonSet
//
//	@Description: 重启daemonset
//	@receiver d
//	@param c
func (d *daemonSet) RestartDaemonSet(c *gin.Context) {
	var body DaemonSetRestartParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	ok, err := k8sDaemonSet.RestartDaemonSet(client, body.Namespace, body.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "重启DaemonSet失败")
		return
	}
	if !ok {
		response.Fail(c, "重启DaemonSet失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

type DaemonSetRollbackParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Revision    int64  `form:"revision" json:"revision" binding:"required" label:"版本号"`
}

type DaemonSetUpdateImageParams struct {
	ClusterName   string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace     string `form:"namespace" json:"namespace" label:"命名空间"`
	Name          string `form:"name" json:"name" binding:"required" label:"名称"`
	ContainerName string `form:"containerName" json:"containerName" binding:"required" label:"容器名称"`
	Image         string `form:"image" json:"image" binding:"required" label:"镜像地址"`
}

func (d *daemonSet) RollbackDaemonSet(c *gin.Context) {
	var body DaemonSetRollbackParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	result, err := k8sDaemonSet.RollbackDaemonSet(client, body.Namespace, body.Name, body.Revision)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "回滚DaemonSet失败")
		return
	}
	response.Success(c, "回滚成功", result)
}

func (d *daemonSet) UpdateDaemonSetImage(c *gin.Context) {
	var body DaemonSetUpdateImageParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	result, err := k8sDaemonSet.UpdateDaemonSetImage(client, body.Namespace, body.Name, body.ContainerName, body.Image)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新DaemonSet镜像失败")
		return
	}
	response.Success(c, "更新镜像成功", result)
}

func (d *daemonSet) GetDaemonSetRollbacks(c *gin.Context) {
	var query DaemonSetQueryByNameParams
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
	rollbacks, err := k8sDaemonSet.GetDaemonSetRollbacks(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取DaemonSet回滚列表失败")
		return
	}
	response.Success(c, "执行成功", rollbacks)
}

type DaemonSetQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type DaemonSetQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type DaemonSetCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"Yaml"`
}

type DaemonSetUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"Yaml"`
}

type DaemonSetDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type DaemonSetRestartParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}
