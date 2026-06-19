package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/app/auth/model"
	"gkube/app/auth/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/response"
)

type authHandler struct{}

var Auth = new(authHandler)

// Login
//
//	@Description: 用户登录
//	@receiver h
//	@param c
func (h *authHandler) Login(c *gin.Context) {
	var p params.LoginParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	// 查询用户
	var user model.User
	if err := database.DB.Preload("Roles").Where("username = ? AND status = 1", p.Username).First(&user).Error; err != nil {
		response.Fail(c, "用户名或密码错误")
		return
	}

	// 验证密码
	if !auth.CheckPassword(p.Password, user.PasswordHash) {
		response.Fail(c, "用户名或密码错误")
		return
	}

	// 判断是否为超级管理员
	var isSuperAdmin bool
	for _, role := range user.Roles {
		if role.Name == "super_admin" {
			isSuperAdmin = true
			break
		}
	}

	// 生成 Token
	tokenPair, err := auth.GenerateToken(user.ID, user.Username, isSuperAdmin)
	if err != nil {
		response.Fail(c, fmt.Sprintf("生成Token失败:%s", err.Error()))
		return
	}

	response.Success(c, "登录成功", gin.H{
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"roles":        user.Roles,
		},
	})
}

// Refresh
//
//	@Description: 刷新Token
//	@receiver h
//	@param c
func (h *authHandler) Refresh(c *gin.Context) {
	var p params.RefreshParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	// 解析 Refresh Token
	claims, err := auth.ParseToken(p.RefreshToken)
	if err != nil {
		response.Fail(c, "Refresh Token 无效或已过期")
		return
	}

	// 验证用户是否仍然存在且处于活跃状态
	var user model.User
	if err := database.DB.Where("id = ? AND status = 1", claims.UserID).First(&user).Error; err != nil {
		response.Fail(c, "用户不存在或已被禁用")
		return
	}

	// 生成新的 Token 对
	tokenPair, err := auth.GenerateToken(claims.UserID, claims.Username, claims.IsSuperAdmin)
	if err != nil {
		response.Fail(c, fmt.Sprintf("生成Token失败:%s", err.Error()))
		return
	}

	response.Success(c, "刷新成功", gin.H{
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
	})
}

// GetMe
//
//	@Description: 获取当前用户信息
//	@receiver h
//	@param c
func (h *authHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Fail(c, "未获取到用户信息")
		return
	}

	var user model.User
	if err := database.DB.Preload("Roles.Permissions").Where("id = ?", userID).First(&user).Error; err != nil {
		response.Fail(c, fmt.Sprintf("查询用户失败:%s", err.Error()))
		return
	}

	response.Success(c, "获取成功", gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"status":       user.Status,
		"roles":        user.Roles,
		"created_at":   user.CreatedAt,
	})
}
