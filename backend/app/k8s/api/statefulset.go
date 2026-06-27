package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
	k8sStatefulSet "gkube/pkg/k8s/statefulset"
	"gkube/pkg/response"
)

type statefulSet struct {
}

var StatefulSet = new(statefulSet)

// GetStatefulSetList
//
//	@Description: 获取statefulset列表
//	@receiver s
//	@param c
func (s *statefulSet) GetStatefulSetList(c *gin.Context) {
	var query params.StatefulSetQueryListParams
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
		ssList, err := k8sStatefulSet.ListStatefulSets(client, query.Namespace, limit, continueToken)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取statefulset列表失败:%v", err.Error()))
			return
		}
		remaining := int64(0)
		if ssList.RemainingItemCount != nil {
			remaining = *ssList.RemainingItemCount
		}
		data := k8s.BuildPaginatedData(ssList.Items, ssList.Continue, remaining, limit)
		data.Total = len(ssList.Items)
		response.Success(c, "执行成功", data)
	} else {
		statefulSets, err := k8sStatefulSet.GetStatefulSetList(client, query.Namespace)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取statefulset列表失败:%v", err.Error()))
			return
		}
		response.Success(c, "执行成功", statefulSets)
	}
}

// GetStatefulSetByName
//
//	@Description: 获取statefulset根据名称
//	@receiver s
//	@param c
func (s *statefulSet) GetStatefulSetByName(c *gin.Context) {
	var query params.StatefulSetQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	statefulSet, err := k8sStatefulSet.GetStatefulSetByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取statefulset失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", statefulSet)
}

// GetStatefulSetYaml
//
//	@Description: 获取statefulset的yaml
//	@receiver s
//	@param c
func (s *statefulSet) GetStatefulSetYaml(c *gin.Context) {
	var query params.StatefulSetQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	statefulSetYaml, err := k8sStatefulSet.GetStatefulSetYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取statefulset yaml失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", statefulSetYaml)
}

// GetStatefulSetByField
//
//	@Description: 获取statefulset根据字段查询
//	@receiver s
//	@param c
func (s *statefulSet) GetStatefulSetByField(c *gin.Context) {
	var body params.StatefulSetQueryByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	statefulSets, err := k8sStatefulSet.GetStatefulSetByField(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取statefulset列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", statefulSets)
}

// GetStatefulSetByLabel
//
//	@Description: 获取statefulset根据标签查询
//	@receiver s
//	@param c
func (s *statefulSet) GetStatefulSetByLabel(c *gin.Context) {
	var body params.StatefulSetQueryByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	statefulSets, err := k8sStatefulSet.GetStatefulSetByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取statefulset列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", statefulSets)
}

// CreateStatefulSet
//
//	@Description: 创建statefulset
//	@receiver s
//	@param c
func (s *statefulSet) CreateStatefulSet(c *gin.Context) {
	var body params.StatefulSetCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	err = k8sStatefulSet.CreateStatefulSet(client, body.Namespace, body.Yaml)
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建statefulset失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateStatefulSet
//
//	@Description: 更新statefulset
//	@receiver s
//	@param c
func (s *statefulSet) UpdateStatefulSet(c *gin.Context) {
	var body params.StatefulSetUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sStatefulSet.UpdateStatefulSet(client, body.Namespace, body.Yaml)

	if err != nil {
		response.Fail(c, fmt.Sprintf("更新statefulset失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)

}

// DeleteStatefulSetByName
//
//	@Description: 删除statefulset根据名称
//	@receiver s
//	@param c
func (s *statefulSet) DeleteStatefulSetByName(c *gin.Context) {
	var body params.StatefulSetDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	err = k8sStatefulSet.DeleteStatefulSetByName(client, body.Namespace, body.Name)

	if err != nil {
		response.Fail(c, fmt.Sprintf("删除statefulset失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteStatefulSetByLabel
//
//	@Description: 删除statefulset根据标签
//	@receiver s
//	@param c
func (s *statefulSet) DeleteStatefulSetByLabel(c *gin.Context) {
	var body params.StatefulSetDeleteByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sStatefulSet.DeleteStatefulSetByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("删除statefulset失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteStatefulSetByField
//
//	@Description: 删除statefulset根据字段
//	@receiver s
//	@param c
func (s *statefulSet) DeleteStatefulSetByField(c *gin.Context) {
	var body params.StatefulSetDeleteByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sStatefulSet.DeleteStatefulSetByField(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("删除statefulset失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
