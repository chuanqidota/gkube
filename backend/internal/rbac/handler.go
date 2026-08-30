package rbac

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	authmodel "gkube/internal/auth/model"
	"gkube/internal/rbac/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

type rbacHandler struct{}

var RbacHandler = new(rbacHandler)

// --- 参数结构体 ---

type CreateBindingParams struct {
	UserID      uint     `json:"userId" binding:"required" label:"用户ID"`
	RoleID      uint     `json:"roleId" binding:"required" label:"角色ID"`
	ClusterID   uint     `json:"clusterId" binding:"required" label:"集群ID"`
	Namespace   string   `json:"namespace"`  // 单个命名空间（兼容旧接口）
	Namespaces  []string `json:"namespaces"` // 多个命名空间
}

type UpdateBindingParams struct {
	RoleID uint `json:"roleId" binding:"required" label:"角色ID"`
}

type ListBindingsQuery struct {
	ClusterID *uint `form:"clusterId" json:"clusterId"`
	UserID    *uint `form:"userId" json:"userId"`
	Page      int   `form:"page" json:"page"`
	Size      int   `form:"size" json:"size"`
}

type ClusterMembersQuery struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required" label:"集群ID"`
}

// parseID 从 URL path 参数解析 uint ID，失败返回 0 + false
func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, "无效的ID")
		return 0, false
	}
	return uint(id), true
}

// isDuplicateErr 检查 MySQL 唯一约束冲突
func isDuplicateErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}

// ListRoles 返回所有角色列表
func (h *rbacHandler) ListRoles(c *gin.Context) {
	var roles []model.Role
	if err := database.DB.Find(&roles).Error; err != nil {
		logger.Error(fmt.Sprintf("查询角色列表失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "查询角色列表失败")
		return
	}

	// 解析 permissions JSON 为 map 返回
	type roleResp struct {
		ID          uint              `json:"id"`
		Name        string            `json:"name"`
		DisplayName string            `json:"displayName"`
		ScopeType   string            `json:"scopeType"`
		IsSystem    bool              `json:"isSystem"`
		Permissions map[string][]string `json:"permissions"`
	}
	items := make([]roleResp, 0, len(roles))
	for _, r := range roles {
		items = append(items, roleResp{
			ID:          r.ID,
			Name:        r.Name,
			DisplayName: r.DisplayName,
			ScopeType:   r.ScopeType,
			IsSystem:    r.IsSystem,
			Permissions: r.GetPermissionsMap(),
		})
	}

	response.Success(c, "获取角色列表成功", items)
}

// ListBindings 查询权限绑定列表（管理员）
func (h *rbacHandler) ListBindings(c *gin.Context) {
	var q ListBindingsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 {
		q.Size = 20
	}
	if q.Size > 100 {
		q.Size = 100
	}

	db := database.DB.Model(&model.PermissionBinding{})
	if q.ClusterID != nil {
		db = db.Where("cluster_id = ?", *q.ClusterID)
	}
	if q.UserID != nil {
		db = db.Where("user_id = ?", *q.UserID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		logger.Error(fmt.Sprintf("查询绑定总数失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "查询绑定列表失败")
		return
	}

	var bindings []model.PermissionBinding
	if err := db.Preload("User").Preload("Role").Preload("Cluster").
		Offset((q.Page - 1) * q.Size).Limit(q.Size).
		Order("id DESC").Find(&bindings).Error; err != nil {
		logger.Error(fmt.Sprintf("查询绑定列表失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "查询绑定列表失败")
		return
	}

	response.Success(c, "获取绑定列表成功", gin.H{
		"items": bindings,
		"total": total,
	})
}

// CreateBinding 创建权限绑定（支持批量命名空间）
func (h *rbacHandler) CreateBinding(c *gin.Context) {
	var p CreateBindingParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	// 校验用户存在
	var user authmodel.User
	if err := database.DB.First(&user, p.UserID).Error; err != nil {
		response.Fail(c, "用户不存在")
		return
	}

	// 校验角色存在
	var role model.Role
	if err := database.DB.First(&role, p.RoleID).Error; err != nil {
		response.Fail(c, "角色不存在")
		return
	}

	// 收集命名空间列表：优先用 Namespaces 数组，兼容单个 Namespace
	nsList := p.Namespaces
	if len(nsList) == 0 && p.Namespace != "" {
		nsList = []string{p.Namespace}
	}

	// 校验命名空间格式
	nsRegex := regexp.MustCompile(`^[a-z0-9]([a-z0-9\-]*[a-z0-9])?$`)
	for _, ns := range nsList {
		if ns != "" && (len(ns) > 63 || !nsRegex.MatchString(ns)) {
			response.Fail(c, fmt.Sprintf("命名空间 %q 格式不合法", ns))
			return
		}
	}

	// 校验角色与作用域匹配
	if role.ScopeType == "namespace" && len(nsList) == 0 {
		response.Fail(c, "命名空间角色必须指定命名空间")
		return
	}

	// 集群级角色且未指定命名空间：创建一条集群级绑定
	if role.ScopeType == "cluster" && len(nsList) == 0 {
		nsList = []string{""}
	}

	// 批量创建绑定，跳过已存在的
	var created []model.PermissionBinding
	var skipped int
	for _, ns := range nsList {
		binding := model.PermissionBinding{
			UserID:    p.UserID,
			RoleID:    p.RoleID,
			ClusterID: p.ClusterID,
			Namespace: ns,
		}
		if err := database.DB.Create(&binding).Error; err != nil {
			if isDuplicateErr(err) {
				skipped++
				continue
			}
			logger.Error(fmt.Sprintf("创建绑定失败: %v", err))
			response.FailWithStatus(c, http.StatusInternalServerError, "创建绑定失败")
			return
		}
		created = append(created, binding)
	}

	// 失效缓存
	auth.InvalidateUserPermissions(p.UserID)

	// 返回结果
	if len(created) == 1 {
		database.DB.Preload("User").Preload("Role").Preload("Cluster").First(&created[0], created[0].ID)
		response.Success(c, "创建成功", created[0])
	} else {
		response.Success(c, fmt.Sprintf("创建成功 %d 条, 跳过 %d 条", len(created), skipped), gin.H{
			"items":   created,
			"total":   len(created),
			"skipped": skipped,
		})
	}
}

// UpdateBinding 修改权限绑定（只能改角色）
func (h *rbacHandler) UpdateBinding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var p UpdateBindingParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	// 查找原绑定
	var binding model.PermissionBinding
	if err := database.DB.First(&binding, id).Error; err != nil {
		response.FailWithStatus(c, http.StatusNotFound, "绑定不存在")
		return
	}

	// 校验新角色存在
	var role model.Role
	if err := database.DB.First(&role, p.RoleID).Error; err != nil {
		response.Fail(c, "角色不存在")
		return
	}

	// 校验 scopeType 一致
	if binding.Namespace == "" && role.ScopeType != "cluster" {
		response.Fail(c, "角色类型与绑定作用域不匹配")
		return
	}
	if binding.Namespace != "" && role.ScopeType != "namespace" {
		response.Fail(c, "角色类型与绑定作用域不匹配")
		return
	}

	// 更新
	if err := database.DB.Model(&binding).Update("role_id", p.RoleID).Error; err != nil {
		logger.Error(fmt.Sprintf("更新绑定失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "更新绑定失败")
		return
	}

	// 失效缓存
	auth.InvalidateUserPermissions(binding.UserID)

	// 重新加载关联数据返回
	database.DB.Preload("User").Preload("Role").First(&binding, binding.ID)
	response.Success(c, "更新成功", binding)
}

// DeleteBinding 删除权限绑定
func (h *rbacHandler) DeleteBinding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var binding model.PermissionBinding
	if err := database.DB.First(&binding, id).Error; err != nil {
		response.FailWithStatus(c, http.StatusNotFound, "绑定不存在")
		return
	}

	if err := database.DB.Delete(&binding).Error; err != nil {
		logger.Error(fmt.Sprintf("删除绑定失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "删除绑定失败")
		return
	}

	// 失效缓存
	auth.InvalidateUserPermissions(binding.UserID)

	response.Success(c, "删除成功", nil)
}

// MyPermissions 返回当前用户的所有集群权限
func (h *rbacHandler) MyPermissions(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.FailWithStatus(c, http.StatusUnauthorized, "未认证")
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		response.FailWithStatus(c, http.StatusUnauthorized, "未认证")
		return
	}

	// 检查是否为超级管理员
	username, _ := c.Get("username")
	name, _ := username.(string)
	isSuperAdmin := auth.IsAdmin(name)

	cached := auth.GetUserPermissions(userID)
	isSuperAdmin = isSuperAdmin || cached.IsSuperAdmin

	// 构造精简的 bindings 列表
	type bindingInfo struct {
		ClusterID uint   `json:"clusterId"`
		Namespace string `json:"namespace"`
		RoleName  string `json:"roleName"`
	}
	bindings := make([]bindingInfo, 0, len(cached.Bindings))
	for _, b := range cached.Bindings {
		bindings = append(bindings, bindingInfo{
			ClusterID: b.ClusterID,
			Namespace: b.Namespace,
			RoleName:  b.Role.Name,
		})
	}

	response.Success(c, "获取权限成功", gin.H{
		"isSuperAdmin": isSuperAdmin,
		"bindings":     bindings,
	})
}

// ClusterMembers 返回某集群的全部绑定（所有用户可访问）
func (h *rbacHandler) ClusterMembers(c *gin.Context) {
	var q ClusterMembersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	var bindings []model.PermissionBinding
	if err := database.DB.Preload("User").Preload("Role").
		Where("cluster_id = ?", q.ClusterID).
		Order("id ASC").Find(&bindings).Error; err != nil {
		logger.Error(fmt.Sprintf("查询集群成员失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "查询集群成员失败")
		return
	}

	// 构造展平的成员列表
	type memberInfo struct {
		ID              uint   `json:"id"`
		UserID          uint   `json:"userId"`
		Username        string `json:"username"`
		DisplayName     string `json:"displayName"`
		ScopeType       string `json:"scopeType"`
		Namespace       string `json:"namespace"`
		RoleName        string `json:"roleName"`
		RoleDisplayName string `json:"roleDisplayName"`
	}
	items := make([]memberInfo, 0, len(bindings))
	for _, b := range bindings {
		items = append(items, memberInfo{
			ID:              b.ID,
			UserID:          b.UserID,
			Username:        b.User.Username,
			DisplayName:     b.User.DisplayName,
			ScopeType:       b.Role.ScopeType,
			Namespace:       b.Namespace,
			RoleName:        b.Role.Name,
			RoleDisplayName: b.Role.DisplayName,
		})
	}

	response.Success(c, "获取集群成员成功", gin.H{
		"items": items,
		"total": len(items),
	})
}
