package api

import (
	"github.com/gin-gonic/gin"
	"gkube/app/auth/model"
	"gkube/pkg/database"
	"gkube/pkg/response"
)

type rbacMatrixHandler struct{}

var RBACMatrix = new(rbacMatrixHandler)

type MatrixPermission struct {
	Resource string `json:"resource"`
	Verb     string `json:"verb"`
	Allowed  bool   `json:"allowed"`
}

type UpdateMatrixRequest struct {
	RoleID      uint               `json:"roleId"`
	Permissions []MatrixPermission `json:"permissions"`
}

// GetMatrix gets the RBAC permission matrix
func (h *rbacMatrixHandler) GetMatrix(c *gin.Context) {
	var roles []model.Role
	if err := database.DB.Find(&roles).Error; err != nil {
		response.Fail(c, "获取角色列表失败: "+err.Error())
		return
	}

	var matrices []model.RBACMatrix
	if err := database.DB.Find(&matrices).Error; err != nil {
		response.Fail(c, "获取权限矩阵失败: "+err.Error())
		return
	}

	// Build matrix data
	matrixData := make(map[uint]map[string][]string)
	for _, m := range matrices {
		if _, ok := matrixData[m.RoleID]; !ok {
			matrixData[m.RoleID] = make(map[string][]string)
		}
		if m.Allowed {
			matrixData[m.RoleID][m.Resource] = append(matrixData[m.RoleID][m.Resource], m.Verb)
		}
	}

	// Build response
	type RoleMatrix struct {
		RoleID      uint              `json:"roleId"`
		RoleName    string            `json:"roleName"`
		Permissions map[string][]string `json:"permissions"`
	}

	var result []RoleMatrix
	for _, role := range roles {
		rm := RoleMatrix{
			RoleID:      role.ID,
			RoleName:    role.Name,
			Permissions: make(map[string][]string),
		}
		if perms, ok := matrixData[role.ID]; ok {
			rm.Permissions = perms
		}
		result = append(result, rm)
	}

	response.Success(c, "获取成功", result)
}

// GetRoleMatrix gets the RBAC permission matrix for a specific role
func (h *rbacMatrixHandler) GetRoleMatrix(c *gin.Context) {
	roleID := c.Query("roleId")

	var matrices []model.RBACMatrix
	if err := database.DB.Where("role_id = ?", roleID).Find(&matrices).Error; err != nil {
		response.Fail(c, "获取权限矩阵失败: "+err.Error())
		return
	}

	// Build permissions map
	permissions := make(map[string][]string)
	for _, m := range matrices {
		if m.Allowed {
			permissions[m.Resource] = append(permissions[m.Resource], m.Verb)
		}
	}

	response.Success(c, "获取成功", permissions)
}

// UpdateMatrix updates the RBAC permission matrix
func (h *rbacMatrixHandler) UpdateMatrix(c *gin.Context) {
	var req UpdateMatrixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// Delete existing permissions for this role
	if err := database.DB.Where("role_id = ?", req.RoleID).Delete(&model.RBACMatrix{}).Error; err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}

	// Insert new permissions
	for _, p := range req.Permissions {
		matrix := model.RBACMatrix{
			RoleID:   req.RoleID,
			Resource: p.Resource,
			Verb:     p.Verb,
			Allowed:  p.Allowed,
		}
		if err := database.DB.Create(&matrix).Error; err != nil {
			response.Fail(c, "更新失败: "+err.Error())
			return
		}
	}

	response.Success(c, "更新成功", nil)
}

// TogglePermission toggles a specific permission
func (h *rbacMatrixHandler) TogglePermission(c *gin.Context) {
	var req struct {
		RoleID   uint   `json:"roleId"`
		Resource string `json:"resource"`
		Verb     string `json:"verb"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	var matrix model.RBACMatrix
	result := database.DB.Where("role_id = ? AND resource = ? AND verb = ?", req.RoleID, req.Resource, req.Verb).First(&matrix)

	if result.Error != nil {
		// Create new permission
		matrix = model.RBACMatrix{
			RoleID:   req.RoleID,
			Resource: req.Resource,
			Verb:     req.Verb,
			Allowed:  true,
		}
		if err := database.DB.Create(&matrix).Error; err != nil {
			response.Fail(c, "更新失败: "+err.Error())
			return
		}
	} else {
		// Toggle existing permission
		matrix.Allowed = !matrix.Allowed
		if err := database.DB.Save(&matrix).Error; err != nil {
			response.Fail(c, "更新失败: "+err.Error())
			return
		}
	}

	response.Success(c, "更新成功", matrix)
}

// InitializeMatrix initializes the RBAC matrix with default resources and verbs
func (h *rbacMatrixHandler) InitializeMatrix(c *gin.Context) {
	resources := []string{
		"pods", "deployments", "statefulsets", "daemonsets", "services", "ingresses",
		"configmaps", "secrets", "persistentvolumeclaims", "persistentvolumes",
		"namespaces", "nodes", "serviceaccounts", "roles", "rolebindings",
		"clusterroles", "clusterrolebindings", "resourcequotas", "limitranges",
		"horizontalpodautoscalers", "networkpolicies", "poddisruptionbudgets",
	}

	verbs := []string{"get", "list", "watch", "create", "update", "patch", "delete"}

	var roles []model.Role
	if err := database.DB.Find(&roles).Error; err != nil {
		response.Fail(c, "获取角色列表失败: "+err.Error())
		return
	}

	// Initialize permissions for each role
	for _, role := range roles {
		for _, resource := range resources {
			for _, verb := range verbs {
				allowed := false
				// Super admin gets all permissions
				if role.Name == "super_admin" {
					allowed = true
				}
				// Admin gets most permissions
				if role.Name == "admin" && verb != "delete" {
					allowed = true
				}
				// Viewer gets read-only permissions
				if role.Name == "viewer" && (verb == "get" || verb == "list" || verb == "watch") {
					allowed = true
				}

				matrix := model.RBACMatrix{
					RoleID:   role.ID,
					Resource: resource,
					Verb:     verb,
					Allowed:  allowed,
				}
				database.DB.FirstOrCreate(&matrix, model.RBACMatrix{
					RoleID:   role.ID,
					Resource: resource,
					Verb:     verb,
				})
			}
		}
	}

	response.Success(c, "初始化成功", nil)
}
