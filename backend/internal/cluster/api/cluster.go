package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/internal/cluster/model"
	"gkube/internal/cluster/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/k8s"
	k8sCluster "gkube/pkg/k8s/cluster"
	"gkube/pkg/response"
)

type clusterHandler struct{}

var Cluster = new(clusterHandler)

// List
//
//	@Description: 获取集群列表（分页）
//	@receiver cl
//	@param c
func (cl *clusterHandler) List(c *gin.Context) {
	var query params.ClusterQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 10
	}
	if query.Size > 100 {
		query.Size = 100
	}

	db := database.DB.Model(&model.K8SCluster{})
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		keyword := escapeLike(query.Keyword)
		db = db.Where("cluster_name LIKE ? OR display_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询集群总数失败:%s", err.Error()))
		return
	}

	var clusters []model.K8SCluster
	if err := db.Offset((query.Page - 1) * query.Size).
		Limit(query.Size).
		Order("id DESC").
		Find(&clusters).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询集群列表失败:%s", err.Error()))
		return
	}

	response.Success(c, "获取集群列表成功", gin.H{
		"items": clusters,
		"total": total,
	})
}

// Create
//
//	@Description: 创建集群
//	@receiver cl
//	@param c
func (cl *clusterHandler) Create(c *gin.Context) {
	var p params.CreateClusterParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	// 检查集群名称唯一性
	var count int64
	if err := database.DB.Model(&model.K8SCluster{}).Where("cluster_name = ?", p.ClusterName).Count(&count).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询集群名称失败:%s", err.Error()))
		return
	}
	if count > 0 {
		response.Fail(c, "集群名称已存在")
		return
	}

	// 验证kubeconfig连通性
	client, err := k8s.GetK8sClient(p.KubeConfig)
	if err != nil {
		response.Fail(c, fmt.Sprintf("kubeconfig验证失败:%s", err.Error()))
		return
	}

	// 获取集群版本
	version, err := k8sCluster.GetClusterVersion(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群版本失败:%s", err.Error()))
		return
	}

	// 获取节点数量
	nodes, err := k8sCluster.GetClusterNodesInfo(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群节点信息失败:%s", err.Error()))
		return
	}
	nodeCount := len(nodes)

	// 加密kubeconfig
	encryptedConfig, err := auth.EncryptAES(p.KubeConfig)
	if err != nil {
		response.Fail(c, fmt.Sprintf("加密kubeconfig失败:%s", err.Error()))
		return
	}

	// 序列化标签为JSON字符串
	labelsJSON := ""
	if len(p.Labels) > 0 {
		if b, err := json.Marshal(p.Labels); err == nil {
			labelsJSON = string(b)
		}
	}

	cluster := model.K8SCluster{
		ClusterName:     p.ClusterName,
		DisplayName:     p.DisplayName,
		Description:     p.Description,
		KubeConfig:      encryptedConfig,
		Status:          "online",
		ClusterVersion:  version,
		NodeCount:       nodeCount,
		Labels:          labelsJSON,
		LastHealthCheck: time.Now(),
	}

	if err := database.DB.Create(&cluster).Error; err != nil {
		response.Fail(c, fmt.Sprintf("创建集群失败:%s", err.Error()))
		return
	}

	response.Success(c, "创建集群成功", cluster)
}

// Detail
//
//	@Description: 获取集群详情
//	@receiver cl
//	@param c
func (cl *clusterHandler) Detail(c *gin.Context) {
	var p params.ClusterIDParams
	if err := c.ShouldBindUri(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	var cluster model.K8SCluster
	if err := database.DB.First(&cluster, p.ID).Error; err != nil {
		response.Fail(c, "集群不存在")
		return
	}

	response.Success(c, "获取集群详情成功", cluster)
}

// Update
//
//	@Description: 更新集群
//	@receiver cl
//	@param c
func (cl *clusterHandler) Update(c *gin.Context) {
	var uriParams params.ClusterIDParams
	if err := c.ShouldBindUri(&uriParams); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	var p params.UpdateClusterParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	var cluster model.K8SCluster
	if err := database.DB.First(&cluster, uriParams.ID).Error; err != nil {
		response.Fail(c, "集群不存在")
		return
	}

	updates := map[string]interface{}{}
	if p.DisplayName != nil {
		updates["display_name"] = *p.DisplayName
	}
	if p.Description != nil {
		updates["description"] = *p.Description
	}
	if p.Labels != nil {
		if b, err := json.Marshal(*p.Labels); err == nil {
			updates["labels"] = string(b)
		} else {
			response.Fail(c, fmt.Sprintf("序列化标签失败:%s", err.Error()))
			return
		}
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&cluster).Updates(updates).Error; err != nil {
			response.Fail(c, fmt.Sprintf("更新集群失败:%s", err.Error()))
			return
		}
	}

	if err := database.DB.First(&cluster, cluster.ID).Error; err != nil {
		response.Fail(c, fmt.Sprintf("获取更新后集群失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新集群成功", cluster)
}

// Delete
//
//	@Description: 删除集群（软删除）
//	@receiver cl
//	@param c
func (cl *clusterHandler) Delete(c *gin.Context) {
	var p params.ClusterIDParams
	if err := c.ShouldBindUri(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	var cluster model.K8SCluster
	if err := database.DB.First(&cluster, p.ID).Error; err != nil {
		response.Fail(c, "集群不存在")
		return
	}

	if err := database.DB.Delete(&cluster).Error; err != nil {
		response.Fail(c, fmt.Sprintf("删除集群失败:%s", err.Error()))
		return
	}

	response.Success(c, "删除集群成功", nil)
}

// Check
//
//	@Description: 检查集群连通性
//	@receiver cl
//	@param c
func (cl *clusterHandler) Check(c *gin.Context) {
	var p params.ClusterIDParams
	if err := c.ShouldBindUri(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	var cluster model.K8SCluster
	if err := database.DB.First(&cluster, p.ID).Error; err != nil {
		response.Fail(c, "集群不存在")
		return
	}

	// 解密kubeconfig
	kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
	if err != nil {
		response.Fail(c, fmt.Sprintf("解密kubeconfig失败:%s", err.Error()))
		return
	}

	start := time.Now()

	// 测试连通性
	client, err := k8s.GetK8sClient(kubeConfig)
	if err != nil {
		// 更新状态为offline
		database.DB.Model(&cluster).Updates(map[string]interface{}{
			"status":            "offline",
			"last_health_check": time.Now(),
		})
		response.Fail(c, fmt.Sprintf("集群连接失败:%s", err.Error()))
		return
	}

	// 获取集群版本
	version, err := k8sCluster.GetClusterVersion(client)
	if err != nil {
		database.DB.Model(&cluster).Updates(map[string]interface{}{
			"status":            "offline",
			"last_health_check": time.Now(),
		})
		response.Fail(c, fmt.Sprintf("获取集群版本失败:%s", err.Error()))
		return
	}

	// 获取节点数量
	nodes, err := k8sCluster.GetClusterNodesInfo(client)
	if err != nil {
		database.DB.Model(&cluster).Updates(map[string]interface{}{
			"status":            "offline",
			"last_health_check": time.Now(),
		})
		response.Fail(c, fmt.Sprintf("获取集群节点信息失败:%s", err.Error()))
		return
	}
	nodeCount := len(nodes)

	responseTimeMs := time.Since(start).Milliseconds()

	// 更新集群状态
	database.DB.Model(&cluster).Updates(map[string]interface{}{
		"status":            "online",
		"cluster_version":   version,
		"node_count":        nodeCount,
		"last_health_check": time.Now(),
	})

	response.Success(c, "集群连通性检查成功", gin.H{
		"status":         "online",
		"clusterVersion": version,
		"nodeCount":      nodeCount,
		"responseTimeMs": responseTimeMs,
	})
}

// escapeLike escapes SQL LIKE metacharacters (%, _, \) in user input.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
