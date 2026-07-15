package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/internal/auth/model"
	"gkube/internal/auth/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/response"
)

type userHandler struct{}

var User = new(userHandler)

// List
//
//	@Description: 获取用户列表（分页）
//	@receiver u
//	@param c
func (u *userHandler) List(c *gin.Context) {
	var query params.UserQueryParams
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

	db := database.DB.Model(&model.User{})
	if query.Username != "" {
		db = db.Where("username LIKE ?", "%"+query.Username+"%")
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询用户总数失败:%s", err.Error()))
		return
	}

	var users []model.User
	if err := db.Offset((query.Page - 1) * query.Size).
		Limit(query.Size).
		Order("id DESC").
		Find(&users).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询用户列表失败:%s", err.Error()))
		return
	}

	response.Success(c, "获取用户列表成功", gin.H{
		"items": users,
		"total": total,
	})
}

// Create
//
//	@Description: 创建用户
//	@receiver u
//	@param c
func (u *userHandler) Create(c *gin.Context) {
	var p params.CreateUserParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	// 检查用户名唯一性
	var count int64
	if err := database.DB.Model(&model.User{}).Where("username = ?", p.Username).Count(&count).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询用户名失败:%s", err.Error()))
		return
	}
	if count > 0 {
		response.Fail(c, "用户名已存在")
		return
	}

	// 密码哈希
	hashedPassword, err := auth.HashPassword(p.Password)
	if err != nil {
		response.Fail(c, fmt.Sprintf("密码加密失败:%s", err.Error()))
		return
	}

	user := model.User{
		Username:     p.Username,
		PasswordHash: hashedPassword,
		Email:        p.Email,
		DisplayName:  p.DisplayName,
		Status:       1,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		response.Fail(c, fmt.Sprintf("创建用户失败:%s", err.Error()))
		return
	}

	response.Success(c, "创建用户成功", user)
}

// Update
//
//	@Description: 更新用户
//	@receiver u
//	@param c
func (u *userHandler) Update(c *gin.Context) {
	var p params.UpdateUserParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	var user model.User
	if err := database.DB.First(&user, p.ID).Error; err != nil {
		response.Fail(c, "用户不存在")
		return
	}

	// 更新字段
	updates := map[string]interface{}{}
	if p.Email != "" {
		updates["email"] = p.Email
	}
	if p.DisplayName != "" {
		updates["display_name"] = p.DisplayName
	}
	if p.Status != nil {
		updates["status"] = *p.Status
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
			response.Fail(c, fmt.Sprintf("更新用户失败:%s", err.Error()))
			return
		}
	}

	database.DB.First(&user, user.ID)
	response.Success(c, "更新用户成功", user)
}

// Delete
//
//	@Description: 删除用户（软删除）
//	@receiver u
//	@param c
func (u *userHandler) Delete(c *gin.Context) {
	var body struct {
		ID uint `json:"id" binding:"required" label:"用户ID"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	// 防止删除自己
	currentUserID, exists := c.Get("userID")
	if exists && currentUserID.(uint) == body.ID {
		response.Fail(c, "不能删除当前登录用户")
		return
	}

	var user model.User
	if err := database.DB.First(&user, body.ID).Error; err != nil {
		response.Fail(c, "用户不存在")
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		response.Fail(c, fmt.Sprintf("删除用户失败:%s", err.Error()))
		return
	}

	response.Success(c, "删除用户成功", nil)
}

// ChangePassword
//
//	@Description: 修改密码
//	@receiver u
//	@param c
func (u *userHandler) ChangePassword(c *gin.Context) {
	var p params.ChangePasswordParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		response.Fail(c, "未获取到当前用户信息")
		return
	}

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		response.Fail(c, "用户不存在")
		return
	}

	// 验证旧密码
	if !auth.CheckPassword(p.OldPassword, user.PasswordHash) {
		response.Fail(c, "旧密码不正确")
		return
	}

	// 哈希新密码
	hashedPassword, err := auth.HashPassword(p.NewPassword)
	if err != nil {
		response.Fail(c, fmt.Sprintf("密码加密失败:%s", err.Error()))
		return
	}

	if err := database.DB.Model(&user).Update("password_hash", hashedPassword).Error; err != nil {
		response.Fail(c, fmt.Sprintf("修改密码失败:%s", err.Error()))
		return
	}

	response.Success(c, "修改密码成功", nil)
}
