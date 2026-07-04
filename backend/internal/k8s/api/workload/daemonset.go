package workload

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/internal/k8s/params"
	"gkube/pkg/k8s"
	k8sDaemonSet "gkube/pkg/k8s/daemonset"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type daemonSet struct {
}

var DaemonSet = new(daemonSet)

func (d *daemonSet) GetDaemonSetList(c *gin.Context) {
	var query params.DaemonSetQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	limit, continueToken := k8s.GetPaginationParams(c)
	if limit > 0 {
		dsList, err := k8sDaemonSet.ListDaemonSets(client, query.Namespace, limit, continueToken)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取DaemonSet列表失败:%v", err.Error()))
			return
		}
		remaining := int64(0)
		if dsList.RemainingItemCount != nil {
			remaining = *dsList.RemainingItemCount
		}
		data := k8s.BuildPaginatedData(dsList.Items, dsList.Continue, remaining, limit)
		data.Total = len(dsList.Items)
		response.Success(c, "执行成功", data)
	} else {
		daemonSets, err := k8sDaemonSet.GetDaemonSetList(client, query.Namespace)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取DaemonSet列表失败:%v", err.Error()))
			return
		}
		response.Success(c, "执行成功", daemonSets)
	}
}

func (d *daemonSet) GetDaemonSetByName(c *gin.Context) {
	var query params.DaemonSetQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	daemonSet, err := k8sDaemonSet.GetDaemonSetByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取DaemonSet失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", daemonSet)
}

func (d *daemonSet) GetDaemonSetByLabel(c *gin.Context) {
	var body params.DaemonSetQueryByLabelParams
	if err := c.ShouldBindQuery(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	daemonSets, err := k8sDaemonSet.GetDaemonSetByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取DaemonSet失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", daemonSets)
}

func (d *daemonSet) GetDaemonSetByField(c *gin.Context) {
	var body params.DaemonSetQueryByFieldParams
	if err := c.ShouldBindQuery(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	daemonSets, err := k8sDaemonSet.GetDaemonSetByField(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取DaemonSet失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", daemonSets)
}

func (d *daemonSet) GetDaemonSetYaml(c *gin.Context) {
	var query params.DaemonSetQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	daemonSetYaml, err := k8sDaemonSet.GetDaemonSetYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取DaemonSet失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", daemonSetYaml)
}

func (d *daemonSet) CreateDaemonSet(c *gin.Context) {
	var body params.DaemonSetCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sDaemonSet.CreateDaemonSet(client, body.Namespace, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建DaemonSet失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

func (d *daemonSet) UpdateDaemonSet(c *gin.Context) {
	var body params.DaemonSetUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sDaemonSet.UpdateDaemonSet(client, body.Namespace, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新DaemonSet失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

func (d *daemonSet) DeleteDaemonSetByName(c *gin.Context) {
	var body params.DaemonSetDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sDaemonSet.DeleteDaemonSetByName(client, body.Namespace, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除DaemonSet失败:%v", err.Error()))
		return
	}

	response.Success(c, "执行成功", nil)
}

func (d *daemonSet) DeleteDaemonSetByLabel(c *gin.Context) {
	var body params.DaemonSetDeleteByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sDaemonSet.DeleteDaemonSetByLabel(client, body.Namespace, body.LabelMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除DaemonSet失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

func (d *daemonSet) DeleteDaemonSetByField(c *gin.Context) {
	var body params.DaemonSetDeleteByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sDaemonSet.DeleteDaemonSetByField(client, body.Namespace, body.FieldMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除DaemonSet失败:%v", err.Error()))
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
	var query params.DaemonSetQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	events, err := client.CoreV1().Events(query.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=DaemonSet", query.Name),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取daemonset事件失败:%v", err.Error()))
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
	var query params.DaemonSetQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	podList, err := k8sDaemonSet.DaemonSetPodList(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取daemonset pod列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", podList)
}
