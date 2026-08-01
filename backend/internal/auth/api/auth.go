package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gkube/internal/auth/model"
	"gkube/internal/auth/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

type authHandler struct{}

var Auth = new(authHandler)

// Login 用户登录,响应 data 增加 isAdmin(是否在管理员白名单内)。
func (h *authHandler) Login(c *gin.Context) {
	var p params.LoginParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	// 查询用户(用户不存在与密码错误统一文案,避免账号枚举)
	var user model.User
	if err := database.DB.Where("username = ? AND status = 1", p.Username).First(&user).Error; err != nil {
		logger.Error(err.Error())
		response.Fail(c, "用户名或密码错误")
		return
	}

	// 验证密码
	if !auth.CheckPassword(p.Password, user.PasswordHash) {
		response.Fail(c, "用户名或密码错误")
		return
	}

	// 生成 Token
	tokenPair, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "生成Token失败")
		return
	}

	response.Success(c, "登录成功", gin.H{
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"isAdmin":      auth.IsAdmin(user.Username),
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.DisplayName,
		},
	})
}

// Refresh 刷新Token
func (h *authHandler) Refresh(c *gin.Context) {
	var p params.RefreshParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
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
	tokenPair, err := auth.GenerateToken(claims.UserID, claims.Username)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "生成Token失败")
		return
	}

	response.Success(c, "刷新成功", gin.H{
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
	})
}

// WsTicket 签发一次性短期 WebSocket ticket(30s),供 WS 连接鉴权使用。需 JWT。
func (h *authHandler) WsTicket(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.FailWithStatus(c, http.StatusUnauthorized, "未认证")
		return
	}
	username, _ := c.Get("username")
	uid, _ := userID.(uint)
	name, _ := username.(string)

	ticket := auth.IssueTicket(uid, name)
	response.Success(c, "ok", gin.H{"ticket": ticket})
}
