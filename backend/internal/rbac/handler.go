package rbac

import (
	"encoding/json"
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

// nsRegex 命名空间名称校验（编译一次，复用）
var nsRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9\-]*[a-z0-9])?$`)

// resourceGroupDefs 预构建的资源组字典 map（运行时不变，避免每次校验重建）
var resourceGroupDefs = func() map[string]model.ResourceGroupDef {
	m := make(map[string]model.ResourceGroupDef, len(model.ResourceVerbDict))
	for _, d := range model.ResourceVerbDict {
		m[d.Group] = d
	}
	return m
}()

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
		Description string            `json:"description"`
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
			Description: r.Description,
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
	for _, ns := range nsList {
		if ns != "" && (len(ns) > 63 || !nsRegex.MatchString(ns)) {
			response.Fail(c, fmt.Sprintf("命名空间 %q 格式不合法", ns))
			return
		}
	}

	// 拒绝显式空串：namespaces=[""] 会绕过下方 scope 校验，静默创建集群级绑定
	for _, ns := range nsList {
		if ns == "" && role.ScopeType == "namespace" {
			response.Fail(c, "命名空间角色必须指定非空命名空间")
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

// ClusterMembers 返回某集群的成员（按用户分组），所有已认证用户可访问。
// P1 返回结构：items[] = { userId, username, displayName, isSuperAdmin, bindings[] }
// bindings[] 每项 = { roleId, roleName, roleDisplayName, scopeType, namespace }
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

	// 按用户分组（保持首次出现顺序）
	type bindingInfo struct {
		BindingID       uint   `json:"bindingId"`
		RoleID          uint   `json:"roleId"`
		RoleName        string `json:"roleName"`
		RoleDisplayName string `json:"roleDisplayName"`
		ScopeType       string `json:"scopeType"`  // cluster|namespace
		Namespace       string `json:"namespace"`  // cluster 级为 ""
	}
	type memberInfo struct {
		UserID       uint          `json:"userId"`
		Username     string        `json:"username"`
		DisplayName  string        `json:"displayName"`
		IsSuperAdmin bool          `json:"isSuperAdmin"`
		Bindings     []bindingInfo `json:"bindings"`
	}

	order := make([]uint, 0, 8)
	members := make(map[uint]*memberInfo, 8)
	for _, b := range bindings {
		m, exists := members[b.UserID]
		if !exists {
			m = &memberInfo{
				UserID:       b.UserID,
				Username:     b.User.Username,
				DisplayName:  b.User.DisplayName,
				IsSuperAdmin: b.User.IsSuperAdmin,
				Bindings:     make([]bindingInfo, 0, 4),
			}
			members[b.UserID] = m
			order = append(order, b.UserID)
		}
		m.Bindings = append(m.Bindings, bindingInfo{
			BindingID:       b.ID,
			RoleID:          b.RoleID,
			RoleName:        b.Role.Name,
			RoleDisplayName: b.Role.DisplayName,
			ScopeType:       b.Role.ScopeType,
			Namespace:       b.Namespace,
		})
	}

	items := make([]memberInfo, 0, len(order))
	for _, uid := range order {
		items = append(items, *members[uid])
	}

	response.Success(c, "获取集群成员成功", gin.H{
		"items": items,
		"total": len(items),
	})
}

// --- P1 集合级批量操作 ---

// UpdateBindingBatchParams 按组换角色参数
type UpdateBindingBatchParams struct {
	UserID    uint `json:"userId" binding:"required" label:"用户ID"`
	ClusterID uint `json:"clusterId" binding:"required" label:"集群ID"`
	FromRoleID uint `json:"fromRoleId" binding:"required" label:"原角色ID"`
	ToRoleID  uint `json:"toRoleId" binding:"required" label:"新角色ID"`
}

// UpdateBindingBatch 将某用户在某集群下指定角色的全部绑定，统一换成新角色。
// 前端"按组换角色"入口：一次调用覆盖 N 个 namespace 的 N 条绑定。
func (h *rbacHandler) UpdateBindingBatch(c *gin.Context) {
	var p UpdateBindingBatchParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	// 校验目标角色存在
	var toRole model.Role
	if err := database.DB.First(&toRole, p.ToRoleID).Error; err != nil {
		response.Fail(c, "角色不存在")
		return
	}

	// 查出待更新绑定（拿 scopeType 做校验）
	var bindings []model.PermissionBinding
	if err := database.DB.
		Where("user_id = ? AND cluster_id = ? AND role_id = ?", p.UserID, p.ClusterID, p.FromRoleID).
		Find(&bindings).Error; err != nil {
		logger.Error(fmt.Sprintf("批量查询绑定失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "批量更新绑定失败")
		return
	}
	if len(bindings) == 0 {
		response.FailWithStatus(c, http.StatusNotFound, "未找到匹配的绑定")
		return
	}

	// scopeType 一致性校验（与单条 UpdateBinding 规则相同）:
	// 集群级绑定只能换 cluster 角色；空间级绑定只能换 namespace 角色。
	for _, b := range bindings {
		if b.Namespace == "" && toRole.ScopeType != "cluster" {
			response.Fail(c, "角色类型与绑定作用域不匹配")
			return
		}
		if b.Namespace != "" && toRole.ScopeType != "namespace" {
			response.Fail(c, "角色类型与绑定作用域不匹配")
			return
		}
	}

	// 一条 SQL 批量更新
	res := database.DB.Model(&model.PermissionBinding{}).
		Where("user_id = ? AND cluster_id = ? AND role_id = ?", p.UserID, p.ClusterID, p.FromRoleID).
		Update("role_id", p.ToRoleID)
	if res.Error != nil {
		logger.Error(fmt.Sprintf("批量更新绑定失败: %v", res.Error))
		response.FailWithStatus(c, http.StatusInternalServerError, "批量更新绑定失败")
		return
	}

	// 失效缓存
	auth.InvalidateUserPermissions(p.UserID)

	response.Success(c, "批量更新成功", gin.H{"updated": res.RowsAffected})
}

// DeleteBindingBatchParams 按组删除参数
type DeleteBindingBatchParams struct {
	UserID    uint `json:"userId" binding:"required" label:"用户ID"`
	ClusterID uint `json:"clusterId" binding:"required" label:"集群ID"`
	RoleID    uint `json:"roleId" binding:"required" label:"角色ID"`
}

// DeleteBindingBatch 删除某用户在某集群下指定角色的全部绑定（按组删除）。
func (h *rbacHandler) DeleteBindingBatch(c *gin.Context) {
	var p DeleteBindingBatchParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	res := database.DB.Where("user_id = ? AND cluster_id = ? AND role_id = ?",
		p.UserID, p.ClusterID, p.RoleID).Delete(&model.PermissionBinding{})
	if res.Error != nil {
		logger.Error(fmt.Sprintf("批量删除绑定失败: %v", res.Error))
		response.FailWithStatus(c, http.StatusInternalServerError, "批量删除绑定失败")
		return
	}
	if res.RowsAffected == 0 {
		response.FailWithStatus(c, http.StatusNotFound, "未找到匹配的绑定")
		return
	}

	// 失效缓存
	auth.InvalidateUserPermissions(p.UserID)

	response.Success(c, "批量删除成功", gin.H{"deleted": res.RowsAffected})
}

// RemoveClusterMember 移出集群：删除该用户在该集群下的所有绑定（不限角色）。
// 路由为 DELETE /rbac/bindings/user/:userId，clusterId 走 query。
func (h *rbacHandler) RemoveClusterMember(c *gin.Context) {
	userID, ok := parseID(c)
	if !ok {
		return
	}
	clusterIDStr := c.Query("clusterId")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 64)
	if err != nil || clusterID == 0 {
		response.Fail(c, "无效的集群ID")
		return
	}

	res := database.DB.Where("user_id = ? AND cluster_id = ?", userID, clusterID).
		Delete(&model.PermissionBinding{})
	if res.Error != nil {
		logger.Error(fmt.Sprintf("移出集群失败: %v", res.Error))
		response.FailWithStatus(c, http.StatusInternalServerError, "移出集群失败")
		return
	}

	// 失效缓存（即使 RowsAffected==0 也无害）
	auth.InvalidateUserPermissions(uint(userID))

	response.Success(c, "已移出集群", gin.H{"deleted": res.RowsAffected})
}

// --- P2 角色管理 ---

type CreateRoleParams struct {
	Name        string              `json:"name" binding:"required,max=50" label:"角色标识"`
	DisplayName string              `json:"displayName" binding:"required,max=100" label:"角色名称"`
	ScopeType   string              `json:"scopeType" binding:"required,oneof=cluster namespace" label:"作用域类型"`
	Description string              `json:"description" binding:"max=255" label:"描述"`
	Permissions map[string][]string `json:"permissions" binding:"required" label:"权限定义"`
}

type UpdateRoleParams struct {
	DisplayName string              `json:"displayName" binding:"required,max=100" label:"角色名称"`
	Description string              `json:"description" binding:"max=255" label:"描述"`
	Permissions map[string][]string `json:"permissions" binding:"required" label:"权限定义"`
}

// validateRolePermissions 校验权限矩阵合法性：资源组与动词必须在 ResourceVerbDict 内。
// clusterOnly 资源组只能出现在 cluster 作用域角色中。
func validateRolePermissions(scopeType string, perms map[string][]string) string {
	for group, verbs := range perms {
		d, ok := resourceGroupDefs[group]
		if !ok {
			return fmt.Sprintf("未知的资源组 %q", group)
		}
		if d.ClusterOnly && scopeType != "cluster" {
			return fmt.Sprintf("资源组 %q 仅集群级角色可配置", group)
		}
		for _, v := range verbs {
			found := false
			for _, allowed := range d.Verbs {
				if v == allowed {
					found = true
					break
				}
			}
			if !found {
				return fmt.Sprintf("资源组 %q 不支持动词 %q", group, v)
			}
		}
	}
	return ""
}

// CreateRole 创建自定义角色
func (h *rbacHandler) CreateRole(c *gin.Context) {
	var p CreateRoleParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	if msg := validateRolePermissions(p.ScopeType, p.Permissions); msg != "" {
		response.Fail(c, msg)
		return
	}

	// 角色名不能与预置角色重名（预置名占用 name 唯一索引，DB 兜底）
	var cnt int64
	database.DB.Model(&model.Role{}).Where("name = ?", p.Name).Count(&cnt)
	if cnt > 0 {
		response.Fail(c, "角色标识已存在")
		return
	}

	permsJSON, err := json.Marshal(p.Permissions)
	if err != nil {
		response.Fail(c, "权限定义序列化失败")
		return
	}

	role := model.Role{
		Name: p.Name, DisplayName: p.DisplayName,
		ScopeType: p.ScopeType, Description: p.Description,
		IsSystem: false, Permissions: string(permsJSON),
	}
	if err := database.DB.Create(&role).Error; err != nil {
		if isDuplicateErr(err) {
			response.Fail(c, "角色标识已存在")
			return
		}
		logger.Error(fmt.Sprintf("创建角色失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "创建角色失败")
		return
	}

	response.Success(c, "创建成功", role)
}

// UpdateRole 更新自定义角色（预置角色只读）
func (h *rbacHandler) UpdateRole(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var role model.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		response.FailWithStatus(c, http.StatusNotFound, "角色不存在")
		return
	}
	if role.IsSystem {
		response.Fail(c, "系统预置角色不可修改")
		return
	}

	var p UpdateRoleParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	if msg := validateRolePermissions(role.ScopeType, p.Permissions); msg != "" {
		response.Fail(c, msg)
		return
	}

	permsJSON, err := json.Marshal(p.Permissions)
	if err != nil {
		response.Fail(c, "权限定义序列化失败")
		return
	}

	if err := database.DB.Model(&role).Updates(map[string]any{
		"DisplayName": p.DisplayName,
		"Description": p.Description,
		"Permissions": string(permsJSON),
	}).Error; err != nil {
		logger.Error(fmt.Sprintf("更新角色失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "更新角色失败")
		return
	}

	// 角色权限变更：失效全部用户权限缓存
	auth.InvalidateAllPermissions()

	database.DB.First(&role, role.ID)
	response.Success(c, "更新成功", role)
}

// DeleteRole 删除自定义角色（预置角色只读；存在绑定时拒绝）
func (h *rbacHandler) DeleteRole(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var role model.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		response.FailWithStatus(c, http.StatusNotFound, "角色不存在")
		return
	}
	if role.IsSystem {
		response.Fail(c, "系统预置角色不可删除")
		return
	}

	var cnt int64
	database.DB.Model(&model.PermissionBinding{}).Where("role_id = ?", role.ID).Count(&cnt)
	if cnt > 0 {
		response.Fail(c, fmt.Sprintf("该角色仍有 %d 条绑定引用，请先移除相关授权", cnt))
		return
	}

	if err := database.DB.Delete(&role).Error; err != nil {
		logger.Error(fmt.Sprintf("删除角色失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "删除角色失败")
		return
	}

	response.Success(c, "删除成功", nil)
}

// ListResourceDict 返回资源组×动词字典（矩阵编辑器数据源）
func (h *rbacHandler) ListResourceDict(c *gin.Context) {
	response.Success(c, "获取资源字典成功", model.ResourceVerbDict)
}

// --- P2 权限诊断 ---

type CanIParams struct {
	UserID        uint   `json:"userId" binding:"required" label:"用户ID"`
	ClusterID     uint   `json:"clusterId" binding:"required" label:"集群ID"`
	Namespace     string `json:"namespace"`
	ResourceGroup string `json:"resourceGroup" binding:"required" label:"资源组"`
	Verb          string `json:"verb" binding:"required" label:"动词"`
}

// CanI 诊断：模拟中间件鉴权逻辑，返回允许与否 + 命中的绑定。
// 与 middleware.RequirePermission 的匹配语义保持一致（超管放行、两层作用域匹配）。
func (h *rbacHandler) CanI(c *gin.Context) {
	var p CanIParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	// 资源组与动词合法性（用字典校验，避免任意字符串）
	validGroup := false
	var validVerbs []string
	clusterOnly := false
	for _, d := range model.ResourceVerbDict {
		if d.Group == p.ResourceGroup {
			validGroup = true
			validVerbs = d.Verbs
			clusterOnly = d.ClusterOnly
			break
		}
	}
	if !validGroup {
		response.Fail(c, "未知的资源组")
		return
	}
	validVerb := false
	for _, v := range validVerbs {
		if v == p.Verb {
			validVerb = true
			break
		}
	}
	if !validVerb {
		response.Fail(c, "该资源组不支持此动词")
		return
	}

	// 超管
	var user authmodel.User
	if err := database.DB.First(&user, p.UserID).Error; err != nil {
		response.Fail(c, "用户不存在")
		return
	}
	if auth.IsAdmin(user.Username) {
		response.Success(c, "诊断完成", gin.H{"allow": true, "reason": "super_admin_config", "matchedBinding": nil})
		return
	}

	cached := auth.GetUserPermissions(p.UserID)
	if cached.IsSuperAdmin {
		response.Success(c, "诊断完成", gin.H{"allow": true, "reason": "super_admin_db", "matchedBinding": nil})
		return
	}

	// 两层作用域匹配（与 middleware.RequirePermission 现有逻辑一致，无分组分支）
	type matchInfo struct {
		BindingID uint   `json:"bindingId"`
		ScopeType string `json:"scopeType"` // cluster|namespace
		Namespace string `json:"namespace"`
		RoleName  string `json:"roleName"`
	}

	for _, b := range cached.Bindings {
		if b.ClusterID != p.ClusterID {
			continue
		}
		scope := ""
		matched := false
		switch {
		case b.Namespace == "":
			// 集群级绑定：覆盖全集群
			scope, matched = "cluster", true
		case b.Namespace == p.Namespace:
			// 空间级绑定：精确 ns 匹配
			scope, matched = "namespace", true
		}
		if !matched {
			continue
		}
		perms := b.Role.GetPermissionsMap()
		if verbs, ok := perms[p.ResourceGroup]; ok {
			for _, v := range verbs {
				if v == p.Verb {
					response.Success(c, "诊断完成", gin.H{
						"allow": true,
						"reason": "role_permission",
						"matchedBinding": matchInfo{
							BindingID: b.ID, ScopeType: scope,
							Namespace: b.Namespace,
							RoleName:  b.Role.Name,
						},
					})
					return
				}
			}
		}
	}

	_ = clusterOnly // 已在上方校验中隐式覆盖（ResourceVerbDict 约束角色创建时不可为 namespace 角色配置 clusterOnly 组）
	response.Success(c, "诊断完成", gin.H{"allow": false, "reason": "no_matching_binding", "matchedBinding": nil, "resourceGroupClusterOnly": clusterOnly})
}
