package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/app/auth/model"
	"gkube/app/settings/api"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/response"
)

type oidcHandler struct{}

var OIDC = new(oidcHandler)

// OIDCState stores the state parameter for CSRF protection
var oidcStates = make(map[string]time.Time)

// GenerateOIDCLoginURL generates the OIDC authorization URL
func (h *oidcHandler) GetLoginURL(c *gin.Context) {
	settings := loadOIDCSettings()
	if settings == nil || !settings.Enabled {
		response.Fail(c, "OIDC 认证未启用")
		return
	}

	// Generate random state for CSRF protection
	state := generateRandomState()
	oidcStates[state] = time.Now().Add(10 * time.Minute)

	// Build authorization URL
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		settings.Issuer+"/authorize",
		settings.ClientID,
		settings.RedirectURI,
		settings.Scopes,
		state,
	)

	response.Success(c, "获取成功", gin.H{
		"url":   authURL,
		"state": state,
	})
}

// HandleCallback handles the OIDC callback
func (h *oidcHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		response.Fail(c, "无效的回调参数")
		return
	}

	// Verify state
	if _, exists := oidcStates[state]; !exists {
		response.Fail(c, "无效的状态参数")
		return
	}
	delete(oidcStates, state)

	settings := loadOIDCSettings()
	if settings == nil || !settings.Enabled {
		response.Fail(c, "OIDC 认证未启用")
		return
	}

	// Exchange code for tokens
	tokenResp, err := exchangeCodeForTokens(settings, code)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取 Token 失败: %s", err.Error()))
		return
	}

	// Get user info from ID token or userinfo endpoint
	userInfo, err := getUserInfo(settings, tokenResp.AccessToken)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取用户信息失败: %s", err.Error()))
		return
	}

	// Extract username from claims
	username := extractClaim(userInfo, settings.UsernameClaim)
	if username == "" {
		response.Fail(c, "无法从 Token 中获取用户名")
		return
	}

	email := extractClaim(userInfo, settings.EmailClaim)

	// Find or create user
	var user model.User
	result := database.DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		// Create new user
		user = model.User{
			Username:    username,
			Email:       email,
			DisplayName: username,
			Status:      1,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			response.Fail(c, fmt.Sprintf("创建用户失败: %s", err.Error()))
			return
		}
	}

	// Generate JWT tokens
	tokenPair, err := auth.GenerateToken(user.ID, user.Username, false)
	if err != nil {
		response.Fail(c, fmt.Sprintf("生成 Token 失败: %s", err.Error()))
		return
	}

	response.Success(c, "登录成功", gin.H{
		"accessToken":  tokenResp.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.DisplayName,
		},
	})
}

// TokenResponse represents the OIDC token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

// UserInfo represents the user info from OIDC
type UserInfo map[string]interface{}

func loadOIDCSettings() *api.OIDCConfig {
	data, err := os.ReadFile("config/auth-settings.json")
	if err != nil {
		return nil
	}

	var settings api.AuthSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil
	}

	return settings.OIDC
}

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func exchangeCodeForTokens(settings *api.OIDCConfig, code string) (*TokenResponse, error) {
	tokenURL := settings.Issuer + "/token"

	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("grant_type", "authorization_code")
	q.Set("code", code)
	q.Set("redirect_uri", settings.RedirectURI)
	q.Set("client_id", settings.ClientID)
	q.Set("client_secret", settings.ClientSecret)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 Token 失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析 Token 响应失败: %w", err)
	}

	return &tokenResp, nil
}

func getUserInfo(settings *api.OIDCConfig, accessToken string) (UserInfo, error) {
	userInfoURL := settings.Issuer + "/userinfo"

	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取用户信息失败, 状态码: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}

	return userInfo, nil
}

func extractClaim(userInfo UserInfo, claimName string) string {
	if claimName == "" {
		return ""
	}

	if val, ok := userInfo[claimName]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}

	return ""
}
