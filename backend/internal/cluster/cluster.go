package cluster

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/internal/cluster/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/k8s"
	k8sCluster "gkube/pkg/k8s/cluster"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"gorm.io/gorm"
)

type CreateClusterParams struct {
	ClusterName string            `json:"clusterName" binding:"required" label:"集群名称"`
	DisplayName string            `json:"displayName" label:"显示名称"`
	Description string            `json:"description" label:"描述"`
	KubeConfig  string            `json:"kubeConfig" binding:"required" label:"KubeConfig"`
	Labels      map[string]string `json:"labels" label:"标签"`
}

type UpdateClusterParams struct {
	DisplayName *string            `json:"displayName" label:"显示名称"`
	Description *string            `json:"description" label:"描述"`
	Labels      *map[string]string `json:"labels" label:"标签"`
}

type ClusterQueryParams struct {
	Page    int    `form:"page" json:"page" label:"页码"`
	Size    int    `form:"size" json:"size" label:"每页数量"`
	Status  string `form:"status" json:"status" label:"状态"`
	Keyword string `form:"keyword" json:"keyword" label:"关键词"`
}

type ClusterIDParams struct {
	ID uint `uri:"id" binding:"required" label:"集群ID"`
}

type clusterHandler struct{}

var Cluster = new(clusterHandler)

// List 获取集群列表（分页）
func (cl *clusterHandler) List(c *gin.Context) {
	var query ClusterQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
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
		if query.Status != "online" && query.Status != "offline" {
			response.Fail(c, "无效的状态参数,仅支持 online/offline")
			return
		}
		db = db.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		if len(query.Keyword) > 200 {
			response.Fail(c, "关键词过长,最多200个字符")
			return
		}
		keyword := escapeLike(query.Keyword)
		db = db.Where("cluster_name LIKE ? OR display_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%")
	}

	// Count 用独立 Session,避免污染后续 Find 的条件
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "查询集群总数失败")
		return
	}

	var clusters []model.K8SCluster
	if err := db.Offset((query.Page - 1) * query.Size).
		Limit(query.Size).
		Order("id DESC").
		Find(&clusters).Error; err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "查询集群列表失败")
		return
	}

	response.Success(c, "获取集群列表成功", gin.H{
		"items": clusters,
		"total": total,
	})
}

// Create 创建集群
func (cl *clusterHandler) Create(c *gin.Context) {
	var p CreateClusterParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	// 校验 kubeconfig 大小, 防止入过大被 DB 截断
	if len(p.KubeConfig) > 12800 {
		response.Fail(c, "kubeconfig 内容超出最大长度(12.8KB)")
		return
	}

	// 验证kubeconfig连通性
	client, err := k8s.GetK8sClient(p.KubeConfig)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "kubeconfig验证失败")
		return
	}

	// 获取集群版本
	version, err := k8sCluster.GetClusterVersion(client)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取集群版本失败")
		return
	}

	// 获取节点数量
	nodes, err := k8sCluster.GetClusterNodesInfo(client)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取集群节点信息失败")
		return
	}
	nodeCount := len(nodes)

	// 加密kubeconfig
	encryptedConfig, err := auth.EncryptAES(p.KubeConfig)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "加密kubeconfig失败")
		return
	}

	// 序列化标签为JSON字符串
	labelsJSON := ""
	if len(p.Labels) > 0 {
		b, err := json.Marshal(p.Labels)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusInternalServerError, "序列化标签失败")
			return
		}
		labelsJSON = string(b)
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
		logger.Error(err.Error())
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			response.Fail(c, "集群名称已存在")
			return
		}
		response.FailWithStatus(c, http.StatusInternalServerError, "创建集群失败")
		return
	}

	response.Success(c, "创建集群成功", cluster)
}

// Detail 获取集群详情
func (cl *clusterHandler) Detail(c *gin.Context) {
	var p ClusterIDParams
	if err := c.ShouldBindUri(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	var cluster model.K8SCluster
	if err := database.DB.First(&cluster, p.ID).Error; err != nil {
		response.Fail(c, "集群不存在")
		return
	}

	response.Success(c, "获取集群详情成功", cluster)
}

// Update 更新集群
func (cl *clusterHandler) Update(c *gin.Context) {
	var uriParams ClusterIDParams
	if err := c.ShouldBindUri(&uriParams); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	var p UpdateClusterParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
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
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusInternalServerError, "序列化标签失败")
			return
		}
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&cluster).Updates(updates).Error; err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusInternalServerError, "更新集群失败")
			return
		}
	}

	// 集群信息变更后失效客户端缓存
	k8s.InvalidateClient(cluster.ClusterName)

	if err := database.DB.First(&cluster, cluster.ID).Error; err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取更新后集群失败")
		return
	}
	response.Success(c, "更新集群成功", cluster)
}

// Delete 删除集群
func (cl *clusterHandler) Delete(c *gin.Context) {
	var p ClusterIDParams
	if err := c.ShouldBindUri(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	var cluster model.K8SCluster
	if err := database.DB.First(&cluster, p.ID).Error; err != nil {
		response.Fail(c, "集群不存在")
		return
	}

	if err := database.DB.Delete(&cluster).Error; err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "删除集群失败")
		return
	}

	// 删除后失效客户端缓存,释放连接
	k8s.InvalidateClient(cluster.ClusterName)

	response.Success(c, "删除集群成功", nil)
}

// Check 检查集群连通性
func (cl *clusterHandler) Check(c *gin.Context) {
	var p ClusterIDParams
	if err := c.ShouldBindUri(&p); err != nil {
		response.Fail(c, "参数校验失败")
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
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "解密kubeconfig失败")
		return
	}

	start := time.Now()

	// 测试连通性
	client, err := k8s.GetK8sClient(kubeConfig)
	if err != nil {
		logger.Error(err.Error())
		markOffline(&cluster, "")
		response.FailWithStatus(c, http.StatusBadGateway, "集群连接失败")
		return
	}

	// 获取集群版本
	version, err := k8sCluster.GetClusterVersion(client)
	if err != nil {
		logger.Error(err.Error())
		markOffline(&cluster, "")
		response.FailWithStatus(c, http.StatusBadGateway, "获取集群版本失败")
		return
	}

	// 获取节点数量
	nodes, err := k8sCluster.GetClusterNodesInfo(client)
	if err != nil {
		logger.Error(err.Error())
		markOffline(&cluster, version)
		response.FailWithStatus(c, http.StatusBadGateway, "获取集群节点信息失败")
		return
	}
	nodeCount := len(nodes)

	responseTimeMs := time.Since(start).Milliseconds()

	// 更新集群状态
	if err := database.DB.Model(&cluster).Updates(map[string]interface{}{
		"status":            "online",
		"cluster_version":   version,
		"node_count":        nodeCount,
		"last_health_check": time.Now(),
	}).Error; err != nil {
		logger.Error(err.Error())
	}

	response.Success(c, "集群连通性检查成功", gin.H{
		"status":         "online",
		"clusterVersion": version,
		"nodeCount":      nodeCount,
		"responseTimeMs": responseTimeMs,
	})
}

// markOffline 将集群标记为离线,保留已获取的版本信息。
func markOffline(cluster *model.K8SCluster, version string) {
	updates := map[string]interface{}{
		"status":            "offline",
		"last_health_check": time.Now(),
	}
	if version != "" {
		updates["cluster_version"] = version
	}
	if err := database.DB.Model(cluster).Updates(updates).Error; err != nil {
		logger.Error(err.Error())
	}
}

// escapeLike escapes SQL LIKE metacharacters (%, _, \) in user input.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
