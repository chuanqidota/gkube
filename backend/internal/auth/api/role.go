package api

import (
	"fmt"

	"gkube/internal/auth/model"
	"gkube/internal/auth/params"
	"gkube/pkg/database"
	"gkube/pkg/response"

	"github.com/gin-gonic/gin"
)

type role struct{}

var Role = new(role)

// List
//
//	@Description: 获取角色列表（分页）
//	@receiver r
//	@param c
func (r *role) List(c *gin.Context) {
	var query params.RoleQueryParams
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

	db := database.DB.Model(&model.Role{})
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询角色总数失败:%s", err.Error()))
		return
	}

	var roles []model.Role
	offset := (query.Page - 1) * query.Size
	if err := db.Preload("Permissions").Offset(offset).Limit(query.Size).Order("id DESC").Find(&roles).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询角色列表失败:%s", err.Error()))
		return
	}

	response.Success(c, "获取角色列表成功", gin.H{
		"items": roles,
		"total": total,
	})
}

// Create
//
//	@Description: 创建角色
//	@receiver r
//	@param c
func (r *role) Create(c *gin.Context) {
	var p params.CreateRoleParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	// 检查角色名唯一性
	var count int64
	if err := database.DB.Model(&model.Role{}).Where("name = ?", p.Name).Count(&count).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询角色失败:%s", err.Error()))
		return
	}
	if count > 0 {
		response.Fail(c, "角色名已存在")
		return
	}

	role := model.Role{
		Name:        p.Name,
		Description: p.Description,
	}

	// 关联权限
	if len(p.PermissionIDs) > 0 {
		var permissions []model.Permission
		if err := database.DB.Where("id IN ?", p.PermissionIDs).Find(&permissions).Error; err != nil {
			response.Fail(c, fmt.Sprintf("查询权限失败:%s", err.Error()))
			return
		}
		role.Permissions = permissions
	}

	if err := database.DB.Create(&role).Error; err != nil {
		response.Fail(c, fmt.Sprintf("创建角色失败:%s", err.Error()))
		return
	}

	response.Success(c, "创建角色成功", role)
}

// Update
//
//	@Description: 更新角色
//	@receiver r
//	@param c
func (r *role) Update(c *gin.Context) {
	var p params.UpdateRoleParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	var role model.Role
	if err := database.DB.First(&role, p.ID).Error; err != nil {
		response.Fail(c, "角色不存在")
		return
	}

	// 更新描述
	if err := database.DB.Model(&role).Update("description", p.Description).Error; err != nil {
		response.Fail(c, fmt.Sprintf("更新角色失败:%s", err.Error()))
		return
	}

	// 替换权限关联
	if p.PermissionIDs != nil {
		var permissions []model.Permission
		if len(p.PermissionIDs) > 0 {
			if err := database.DB.Where("id IN ?", p.PermissionIDs).Find(&permissions).Error; err != nil {
				response.Fail(c, fmt.Sprintf("查询权限失败:%s", err.Error()))
				return
			}
		}
		if err := database.DB.Model(&role).Association("Permissions").Replace(permissions); err != nil {
			response.Fail(c, fmt.Sprintf("更新角色权限失败:%s", err.Error()))
			return
		}
	}

	// 重新加载角色（含权限）
	database.DB.Preload("Permissions").First(&role, role.ID)

	response.Success(c, "更新角色成功", role)
}

// Delete
//
//	@Description: 删除角色
//	@receiver r
//	@param c
func (r *role) Delete(c *gin.Context) {
	var body struct {
		ID uint `json:"id" binding:"required" label:"角色ID"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	var role model.Role
	if err := database.DB.First(&role, body.ID).Error; err != nil {
		response.Fail(c, "角色不存在")
		return
	}

	// 禁止删除 super_admin 角色
	if role.Name == "super_admin" {
		response.Fail(c, "禁止删除超级管理员角色")
		return
	}

	// 清除权限关联
	if err := database.DB.Model(&role).Association("Permissions").Clear(); err != nil {
		response.Fail(c, fmt.Sprintf("清除角色权限关联失败:%s", err.Error()))
		return
	}

	if err := database.DB.Delete(&role).Error; err != nil {
		response.Fail(c, fmt.Sprintf("删除角色失败:%s", err.Error()))
		return
	}

	response.Success(c, "删除角色成功", nil)
}
