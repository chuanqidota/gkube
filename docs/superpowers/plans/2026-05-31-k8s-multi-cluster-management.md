# K8s 多集群管理功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 gkube 添加完整的 K8s 多集群管理能力，包括认证鉴权、集群管理、全局 Dashboard 和前端界面。

**Architecture:** 采用渐进增强策略，在现有 Go 后端代码基础上新增 auth/cluster/dashboard 模块，前端新建 Vue 3 项目。后端保持现有目录结构（不迁移到 backend/），前端放在 frontend/ 目录。API 重新规划为 `/v1/auth/*`、`/v1/clusters/*`、`/v1/dashboard/*` 等路由组。

**Tech Stack:** Go + Gin + GORM + client-go + JWT + bcrypt | Vue 3 + TypeScript + Element Plus + Vite + Pinia + Axios

---

## 文件结构总览

### 后端新增/修改文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `go.mod` | 修改 | 添加 golang-jwt 和 bcrypt 依赖 |
| `app/auth/model/user.go` | 新建 | User GORM 模型 |
| `app/auth/model/role.go` | 新建 | Role GORM 模型 |
| `app/auth/model/permission.go` | 新建 | Permission GORM 模型 |
| `app/auth/params/auth.go` | 新建 | 登录/注册请求参数 |
| `app/auth/params/user.go` | 新建 | 用户 CRUD 请求参数 |
| `app/auth/params/role.go` | 新建 | 角色 CRUD 请求参数 |
| `app/auth/api/auth.go` | 新建 | 登录/注册/Token刷新 Handler |
| `app/auth/api/user.go` | 新建 | 用户 CRUD Handler |
| `app/auth/api/role.go` | 新建 | 角色 CRUD Handler |
| `app/cluster/model/cluster.go` | 新建 | 增强版 K8SCluster 模型 |
| `app/cluster/params/cluster.go` | 新建 | 集群 CRUD 请求参数 |
| `app/cluster/api/cluster.go` | 新建 | 集群 CRUD + 健康检测 Handler |
| `app/cluster/service/health.go` | 新建 | 集群健康检测服务 |
| `app/dashboard/api/dashboard.go` | 新建 | Dashboard 数据聚合 Handler |
| `app/dashboard/params/dashboard.go` | 新建 | Dashboard 请求参数 |
| `pkg/auth/jwt.go` | 新建 | JWT 工具函数 |
| `pkg/auth/password.go` | 新建 | bcrypt 密码工具 |
| `pkg/auth/encrypt.go` | 新建 | AES kubeconfig 加密 |
| `pkg/middleware/jwt.go` | 新建 | JWT 认证中间件 |
| `pkg/middleware/rbac.go` | 新建 | RBAC 权限校验中间件 |
| `router/router.go` | 重写 | 重新规划路由结构 |
| `cmd/root.go` | 修改 | 添加健康检测服务启动 |
| `cmd/migrate.go` | 修改 | 添加新模型迁移 |
| `cmd/seed.go` | 新建 | 初始化管理员账号和默认角色 |

### 前端文件（frontend/）

| 文件 | 说明 |
|------|------|
| `frontend/package.json` | 项目配置 |
| `frontend/vite.config.ts` | Vite 配置 |
| `frontend/tsconfig.json` | TypeScript 配置 |
| `frontend/src/main.ts` | 入口文件 |
| `frontend/src/App.vue` | 根组件 |
| `frontend/src/api/request.ts` | Axios 实例 + 拦截器 |
| `frontend/src/api/auth.ts` | 认证 API |
| `frontend/src/api/cluster.ts` | 集群管理 API |
| `frontend/src/api/dashboard.ts` | Dashboard API |
| `frontend/src/api/resource.ts` | K8s 资源 API |
| `frontend/src/stores/auth.ts` | 认证状态管理 |
| `frontend/src/stores/cluster.ts` | 集群状态管理 |
| `frontend/src/router/index.ts` | 路由配置 |
| `frontend/src/utils/auth.ts` | Token 工具函数 |
| `frontend/src/views/login/LoginView.vue` | 登录页 |
| `frontend/src/views/dashboard/DashboardView.vue` | Dashboard 页 |
| `frontend/src/views/cluster/ClusterList.vue` | 集群列表 |
| `frontend/src/views/cluster/ClusterForm.vue` | 添加/编辑集群 |
| `frontend/src/views/cluster/ClusterDetail.vue` | 集群详情 |
| `frontend/src/views/user/UserList.vue` | 用户管理 |
| `frontend/src/views/user/RoleList.vue` | 角色管理 |
| `frontend/src/components/Layout/AppLayout.vue` | 整体布局 |
| `frontend/src/components/Layout/Sidebar.vue` | 侧边栏 |
| `frontend/src/components/Layout/Header.vue` | 顶部栏 |

---

## Task 1: 后端依赖添加

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: 添加 JWT 和 bcrypt 依赖**

```bash
cd /Users/zqqzqq/05_github/gkube
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
```

- [ ] **Step 2: 验证依赖安装**

```bash
go mod tidy
go build ./...
```

Expected: 编译成功，无报错

- [ ] **Step 3: 提交**

```bash
git add go.mod go.sum
git commit -m "deps: add golang-jwt and bcrypt dependencies"
```

---

## Task 2: 认证工具包 — JWT

**Files:**
- Create: `pkg/auth/jwt.go`

- [ ] **Step 1: 创建 JWT 工具函数**

```go
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT 密钥，生产环境应从配置读取
var jwtSecret = []byte("gkube-jwt-secret-change-in-production")

type Claims struct {
	UserID      uint   `json:"userId"`
	Username    string `json:"username"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 access_token 和 refresh_token
func GenerateToken(userID uint, username string, isSuperAdmin bool) (accessToken string, refreshToken string, err error) {
	// Access Token - 2 小时过期
	accessClaims := Claims{
		UserID:       userID,
		Username:     username,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gkube",
		},
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = access.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token - 7 天过期
	refreshClaims := Claims{
		UserID:       userID,
		Username:     username,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gkube",
		},
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refresh.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./pkg/auth/...
```

Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add pkg/auth/jwt.go
git commit -m "feat(auth): add JWT token generation and parsing"
```

---

## Task 3: 认证工具包 — 密码与加密

**Files:**
- Create: `pkg/auth/password.go`
- Create: `pkg/auth/encrypt.go`

- [ ] **Step 1: 创建密码工具**

```go
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
```

- [ ] **Step 2: 创建 AES 加密工具**

```go
package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// AES 密钥，生产环境应从配置读取（必须 16/24/32 字节）
var aesKey = []byte("gkube-aes-key-32byte-padding!!")

// EncryptAES 使用 AES-256-GCM 加密
func EncryptAES(plaintext string) (string, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES 使用 AES-256-GCM 解密
func DecryptAES(encoded string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", err
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
```

- [ ] **Step 3: 验证编译**

```bash
go build ./pkg/auth/...
```

- [ ] **Step 4: 提交**

```bash
git add pkg/auth/password.go pkg/auth/encrypt.go
git commit -m "feat(auth): add password hashing and AES encryption utilities"
```

---

## Task 4: 数据模型 — User / Role / Permission

**Files:**
- Create: `app/auth/model/user.go`
- Create: `app/auth/model/role.go`
- Create: `app/auth/model/permission.go`
- Modify: `cmd/migrate.go`

- [ ] **Step 1: 创建 Permission 模型**

```go
// app/auth/model/permission.go
package model

import "time"

type Permission struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Resource    string `gorm:"column:resource;size:100;not null" json:"resource"`    // cluster, pod, deployment, ...
	Action      string `gorm:"column:action;size:50;not null" json:"action"`         // create, read, update, delete, exec
	ClusterID   *uint  `gorm:"column:cluster_id" json:"clusterId"`                   // 可选：限定集群
	Namespace   string `gorm:"column:namespace;size:100" json:"namespace"`            // 可选：限定命名空间
	Description string `gorm:"column:description;size:200" json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Permission) TableName() string {
	return "permission"
}
```

- [ ] **Step 2: 创建 Role 模型**

```go
// app/auth/model/role.go
package model

import "time"

type Role struct {
	ID          uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"column:name;unique;size:50;not null" json:"name"`
	Description string       `gorm:"column:description;size:200" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

func (Role) TableName() string {
	return "role"
}
```

- [ ] **Step 3: 创建 User 模型**

```go
// app/auth/model/user.go
package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"column:username;unique;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"-"`
	Email        string    `gorm:"column:email;size:100" json:"email"`
	DisplayName  string    `gorm:"column:display_name;size:100" json:"displayName"`
	Status       int       `gorm:"column:status;default:1" json:"status"` // 1=启用 0=禁用
	Roles        []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (User) TableName() string {
	return "user"
}
```

- [ ] **Step 4: 添加模型引用目录文件**

```go
// app/auth/model/model.go
package model

// 包级导入，确保所有模型被引用
var Models = []interface{}{
	&User{},
	&Role{},
	&Permission{},
}
```

- [ ] **Step 5: 更新迁移命令**

在 `cmd/migrate.go` 的 `AutoMigrate` 调用中添加新模型：

```go
import authModel "gkube/app/auth/model"

// 在 AutoMigrate 中添加:
&authModel.User{},
&authModel.Role{},
&authModel.Permission{},
```

- [ ] **Step 6: 验证编译**

```bash
go build ./...
```

- [ ] **Step 7: 提交**

```bash
git add app/auth/model/ cmd/migrate.go
git commit -m "feat(auth): add User, Role, Permission data models"
```

---

## Task 5: 数据模型 — 增强版 K8SCluster

**Files:**
- Create: `app/cluster/model/cluster.go`
- Modify: `cmd/migrate.go`

- [ ] **Step 1: 创建增强版集群模型**

```go
// app/cluster/model/cluster.go
package model

import (
	"time"

	"gorm.io/gorm"
)

type K8SCluster struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ClusterName     string         `gorm:"column:cluster_name;unique;size:100;not null" json:"clusterName"`
	DisplayName     string         `gorm:"column:display_name;size:100" json:"displayName"`
	Description     string         `gorm:"column:description;size:500" json:"description"`
	KubeConfig      string         `gorm:"column:kube_config;size:12800;not null" json:"-"` // AES 加密存储，API 不返回
	Status          string         `gorm:"column:status;size:20;default:unknown" json:"status"` // online/offline/unknown
	ClusterVersion  string         `gorm:"column:cluster_version;size:100" json:"clusterVersion"`
	NodeCount       int            `gorm:"column:node_count;default:0" json:"nodeCount"`
	LastHealthCheck time.Time      `gorm:"column:last_health_check" json:"lastHealthCheck"`
	Labels          string         `gorm:"column:labels;size:500" json:"labels"` // JSON: {"env":"prod","region":"cn-east"}
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (K8SCluster) TableName() string {
	return "k8s_cluster"
}
```

- [ ] **Step 2: 更新迁移命令**

在 `cmd/migrate.go` 中用新的 `app/cluster/model.K8SCluster` 替换旧的 `app/k8s/model.K8SCluster`：

```go
import clusterModel "gkube/app/cluster/model"

// 替换旧模型引用
&clusterModel.K8SCluster{},
```

- [ ] **Step 3: 验证编译**

```bash
go build ./...
```

- [ ] **Step 4: 提交**

```bash
git add app/cluster/model/ cmd/migrate.go
git commit -m "feat(cluster): add enhanced K8SCluster model with health status and labels"
```

---

## Task 6: 请求参数定义

**Files:**
- Create: `app/auth/params/auth.go`
- Create: `app/auth/params/user.go`
- Create: `app/auth/params/role.go`
- Create: `app/cluster/params/cluster.go`
- Create: `app/dashboard/params/dashboard.go`

- [ ] **Step 1: 创建认证参数**

```go
// app/auth/params/auth.go
package params

type LoginParams struct {
	Username string `json:"username" binding:"required" label:"用户名"`
	Password string `json:"password" binding:"required" label:"密码"`
}

type RefreshParams struct {
	RefreshToken string `json:"refreshToken" binding:"required" label:"Refresh Token"`
}
```

- [ ] **Step 2: 创建用户管理参数**

```go
// app/auth/params/user.go
package params

type CreateUserParams struct {
	Username    string `json:"username" binding:"required" label:"用户名"`
	Password    string `json:"password" binding:"required" label:"密码"`
	Email       string `json:"email" label:"邮箱"`
	DisplayName string `json:"displayName" label:"显示名称"`
	RoleIDs     []uint `json:"roleIds" label:"角色ID列表"`
}

type UpdateUserParams struct {
	ID          uint   `json:"id" binding:"required" label:"用户ID"`
	Email       string `json:"email" label:"邮箱"`
	DisplayName string `json:"displayName" label:"显示名称"`
	Status      *int   `json:"status" label:"状态"`
	RoleIDs     []uint `json:"roleIds" label:"角色ID列表"`
}

type UserQueryParams struct {
	Page     int    `form:"page" json:"page" label:"页码"`
	Size     int    `form:"size" json:"size" label:"每页数量"`
	Username string `form:"username" json:"username" label:"用户名"`
	Status   *int   `form:"status" json:"status" label:"状态"`
}

type ChangePasswordParams struct {
	OldPassword string `json:"oldPassword" binding:"required" label:"旧密码"`
	NewPassword string `json:"newPassword" binding:"required" label:"新密码"`
}
```

- [ ] **Step 3: 创建角色管理参数**

```go
// app/auth/params/role.go
package params

type CreateRoleParams struct {
	Name          string `json:"name" binding:"required" label:"角色名"`
	Description   string `json:"description" label:"描述"`
	PermissionIDs []uint `json:"permissionIds" label:"权限ID列表"`
}

type UpdateRoleParams struct {
	ID            uint   `json:"id" binding:"required" label:"角色ID"`
	Description   string `json:"description" label:"描述"`
	PermissionIDs []uint `json:"permissionIds" label:"权限ID列表"`
}

type RoleQueryParams struct {
	Page int    `form:"page" json:"page" label:"页码"`
	Size int    `form:"size" json:"size" label:"每页数量"`
	Name string `form:"name" json:"name" label:"角色名"`
}
```

- [ ] **Step 4: 创建集群管理参数**

```go
// app/cluster/params/cluster.go
package params

type CreateClusterParams struct {
	ClusterName string `json:"clusterName" binding:"required" label:"集群名称"`
	DisplayName string `json:"displayName" label:"显示名称"`
	Description string `json:"description" label:"描述"`
	KubeConfig  string `json:"kubeConfig" binding:"required" label:"KubeConfig"`
	Labels      string `json:"labels" label:"标签JSON"`
}

type UpdateClusterParams struct {
	ID          uint   `json:"id" binding:"required" label:"集群ID"`
	DisplayName string `json:"displayName" label:"显示名称"`
	Description string `json:"description" label:"描述"`
	Labels      string `json:"labels" label:"标签JSON"`
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
```

- [ ] **Step 5: 创建 Dashboard 参数**

```go
// app/dashboard/params/dashboard.go
package params

type EventQueryParams struct {
	ClusterID *uint  `form:"clusterId" json:"clusterId" label:"集群ID"`
	Type      string `form:"type" json:"type" label:"事件类型"`
	Limit     int    `form:"limit" json:"limit" label:"数量限制"`
}
```

- [ ] **Step 6: 验证编译**

```bash
go build ./app/auth/params/... ./app/cluster/params/... ./app/dashboard/params/...
```

- [ ] **Step 7: 提交**

```bash
git add app/auth/params/ app/cluster/params/ app/dashboard/params/
git commit -m "feat: add request parameter structs for auth, cluster, dashboard"
```

---

## Task 7: JWT 认证中间件

**Files:**
- Create: `pkg/middleware/jwt.go`

- [ ] **Step 1: 创建 JWT 中间件**

```go
// pkg/middleware/jwt.go
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"gkube/pkg/auth"
	"gkube/pkg/response"
)

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, "未提供认证 Token")
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, "Token 格式错误")
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			response.Fail(c, "Token 无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("isSuperAdmin", claims.IsSuperAdmin)

		c.Next()
	}
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./pkg/middleware/...
```

- [ ] **Step 3: 提交**

```bash
git add pkg/middleware/jwt.go
git commit -m "feat(auth): add JWT authentication middleware"
```

---

## Task 8: RBAC 权限中间件

**Files:**
- Create: `pkg/middleware/rbac.go`

- [ ] **Step 1: 创建 RBAC 中间件**

```go
// pkg/middleware/rbac.go
package middleware

import (
	"github.com/gin-gonic/gin"

	"gkube/pkg/database"
	"gkube/pkg/response"

	authModel "gkube/app/auth/model"
)

// RBAC 权限校验中间件
// resource: 资源类型（如 "cluster", "pod", "deployment"）
// action: 操作类型（如 "create", "read", "update", "delete"）
func RBAC(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 超级管理员跳过检查
		isSuperAdmin, exists := c.Get("isSuperAdmin")
		if exists && isSuperAdmin.(bool) {
			c.Next()
			return
		}

		// 获取用户 ID
		userID, exists := c.Get("userID")
		if !exists {
			response.Fail(c, "未获取到用户信息")
			c.Abort()
			return
		}

		// 查询用户角色和权限
		var user authModel.User
		if err := database.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
			response.Fail(c, "用户不存在")
			c.Abort()
			return
		}

		// 检查权限
		authorized := false
		for _, role := range user.Roles {
			for _, perm := range role.Permissions {
				if perm.Resource == resource && perm.Action == action {
					authorized = true
					break
				}
				// 通配符权限
				if perm.Resource == "*" && perm.Action == "*" {
					authorized = true
					break
				}
			}
			if authorized {
				break
			}
		}

		if !authorized {
			response.Fail(c, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./pkg/middleware/...
```

- [ ] **Step 3: 提交**

```bash
git add pkg/middleware/rbac.go
git commit -m "feat(auth): add RBAC permission middleware"
```

---

## Task 9: 认证 API — 登录/注册/Token刷新

**Files:**
- Create: `app/auth/api/auth.go`

- [ ] **Step 1: 创建认证 Handler**

```go
// app/auth/api/auth.go
package api

import (
	"github.com/gin-gonic/gin"

	"gkube/app/auth/model"
	"gkube/app/auth/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/response"
)

type authHandler struct{}

var Auth = new(authHandler)

// Login 用户登录
func (a *authHandler) Login(c *gin.Context) {
	var p params.LoginParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 查询用户
	var user model.User
	if err := database.DB.Where("username = ? AND status = 1", p.Username).First(&user).Error; err != nil {
		response.Fail(c, "用户名或密码错误")
		return
	}

	// 验证密码
	if !auth.CheckPassword(p.Password, user.PasswordHash) {
		response.Fail(c, "用户名或密码错误")
		return
	}

	// 查询用户角色，判断是否超级管理员
	database.DB.Preload("Roles").First(&user, user.ID)
	isSuperAdmin := false
	for _, role := range user.Roles {
		if role.Name == "super_admin" {
			isSuperAdmin = true
			break
		}
	}

	// 生成 Token
	accessToken, refreshToken, err := auth.GenerateToken(user.ID, user.Username, isSuperAdmin)
	if err != nil {
		response.Fail(c, "生成 Token 失败")
		return
	}

	response.Success(c, "登录成功", gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user": gin.H{
			"id":          user.ID,
			"username":    user.Username,
			"displayName": user.DisplayName,
			"email":       user.Email,
			"isSuperAdmin": isSuperAdmin,
		},
	})
}

// Refresh 刷新 Token
func (a *authHandler) Refresh(c *gin.Context) {
	var p params.RefreshParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 解析 Refresh Token
	claims, err := auth.ParseToken(p.RefreshToken)
	if err != nil {
		response.Fail(c, "Refresh Token 无效或已过期")
		return
	}

	// 生成新 Token
	accessToken, refreshToken, err := auth.GenerateToken(claims.UserID, claims.Username, claims.IsSuperAdmin)
	if err != nil {
		response.Fail(c, "生成 Token 失败")
		return
	}

	response.Success(c, "刷新成功", gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// GetMe 获取当前用户信息
func (a *authHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user model.User
	if err := database.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		response.Fail(c, "用户不存在")
		return
	}

	response.Success(c, "获取成功", gin.H{
		"id":          user.ID,
		"username":    user.Username,
		"displayName": user.DisplayName,
		"email":       user.Email,
		"status":      user.Status,
		"roles":       user.Roles,
	})
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./app/auth/api/...
```

- [ ] **Step 3: 提交**

```bash
git add app/auth/api/auth.go
git commit -m "feat(auth): add login, refresh, and get-me API handlers"
```

---

## Task 10: 用户管理 API

**Files:**
- Create: `app/auth/api/user.go`

- [ ] **Step 1: 创建用户管理 Handler**

```go
// app/auth/api/user.go
package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gkube/app/auth/model"
	"gkube/app/auth/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/response"
)

type userHandler struct{}

var User = new(userHandler)

// List 用户列表
func (u *userHandler) List(c *gin.Context) {
	var p params.UserQueryParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 20
	}

	query := database.DB.Model(&model.User{})
	if p.Username != "" {
		query = query.Where("username LIKE ?", "%"+p.Username+"%")
	}
	if p.Status != nil {
		query = query.Where("status = ?", *p.Status)
	}

	var total int64
	query.Count(&total)

	var users []model.User
	query.Preload("Roles").Offset((p.Page - 1) * p.Size).Limit(p.Size).Find(&users)

	response.Success(c, "获取成功", gin.H{
		"total": total,
		"items": users,
	})
}

// Create 创建用户
func (u *userHandler) Create(c *gin.Context) {
	var p params.CreateUserParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", p.Username).Count(&count)
	if count > 0 {
		response.Fail(c, "用户名已存在")
		return
	}

	// 加密密码
	hash, err := auth.HashPassword(p.Password)
	if err != nil {
		response.Fail(c, "密码加密失败")
		return
	}

	user := model.User{
		Username:     p.Username,
		PasswordHash: hash,
		Email:        p.Email,
		DisplayName:  p.DisplayName,
		Status:       1,
	}

	// 关联角色
	if len(p.RoleIDs) > 0 {
		var roles []model.Role
		database.DB.Where("id IN ?", p.RoleIDs).Find(&roles)
		user.Roles = roles
	}

	if err := database.DB.Create(&user).Error; err != nil {
		response.Fail(c, "创建用户失败: " + err.Error())
		return
	}

	response.Success(c, "创建成功", user)
}

// Update 更新用户
func (u *userHandler) Update(c *gin.Context) {
	var p params.UpdateUserParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	var user model.User
	if err := database.DB.First(&user, p.ID).Error; err != nil {
		response.Fail(c, "用户不存在")
		return
	}

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

	database.DB.Model(&user).Updates(updates)

	// 更新角色
	if p.RoleIDs != nil {
		var roles []model.Role
		database.DB.Where("id IN ?", p.RoleIDs).Find(&roles)
		database.DB.Model(&user).Association("Roles").Replace(roles)
	}

	response.Success(c, "更新成功", user)
}

// Delete 删除用户
func (u *userHandler) Delete(c *gin.Context) {
	var p struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 不能删除自己
	currentUserID, _ := c.Get("userID")
	if currentUserID.(uint) == p.ID {
		response.Fail(c, "不能删除当前登录用户")
		return
	}

	if err := database.DB.Delete(&model.User{}, p.ID).Error; err != nil {
		response.Fail(c, "删除用户失败")
		return
	}

	response.Success(c, "删除成功", nil)
}

// ChangePassword 修改密码
func (u *userHandler) ChangePassword(c *gin.Context) {
	var p params.ChangePasswordParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		response.Fail(c, "用户不存在")
		return
	}

	if !auth.CheckPassword(p.OldPassword, user.PasswordHash) {
		response.Fail(c, "旧密码错误")
		return
	}

	hash, err := auth.HashPassword(p.NewPassword)
	if err != nil {
		response.Fail(c, "密码加密失败")
		return
	}

	database.DB.Model(&user).Update("password_hash", hash)
	response.Success(c, "密码修改成功", nil)
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./app/auth/api/...
```

- [ ] **Step 3: 提交**

```bash
git add app/auth/api/user.go
git commit -m "feat(auth): add user CRUD API handlers"
```

---

## Task 11: 角色管理 API

**Files:**
- Create: `app/auth/api/role.go`

- [ ] **Step 1: 创建角色管理 Handler**

```go
// app/auth/api/role.go
package api

import (
	"github.com/gin-gonic/gin"

	"gkube/app/auth/model"
	"gkube/app/auth/params"
	"gkube/pkg/database"
	"gkube/pkg/response"
)

type roleHandler struct{}

var Role = new(roleHandler)

// List 角色列表
func (r *roleHandler) List(c *gin.Context) {
	var p params.RoleQueryParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 20
	}

	query := database.DB.Model(&model.Role{})
	if p.Name != "" {
		query = query.Where("name LIKE ?", "%"+p.Name+"%")
	}

	var total int64
	query.Count(&total)

	var roles []model.Role
	query.Preload("Permissions").Offset((p.Page - 1) * p.Size).Limit(p.Size).Find(&roles)

	response.Success(c, "获取成功", gin.H{
		"total": total,
		"items": roles,
	})
}

// Create 创建角色
func (r *roleHandler) Create(c *gin.Context) {
	var p params.CreateRoleParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 检查角色名是否已存在
	var count int64
	database.DB.Model(&model.Role{}).Where("name = ?", p.Name).Count(&count)
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
		database.DB.Where("id IN ?", p.PermissionIDs).Find(&permissions)
		role.Permissions = permissions
	}

	if err := database.DB.Create(&role).Error; err != nil {
		response.Fail(c, "创建角色失败: " + err.Error())
		return
	}

	response.Success(c, "创建成功", role)
}

// Update 更新角色
func (r *roleHandler) Update(c *gin.Context) {
	var p params.UpdateRoleParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	var role model.Role
	if err := database.DB.First(&role, p.ID).Error; err != nil {
		response.Fail(c, "角色不存在")
		return
	}

	if p.Description != "" {
		database.DB.Model(&role).Update("description", p.Description)
	}

	// 更新权限
	if p.PermissionIDs != nil {
		var permissions []model.Permission
		database.DB.Where("id IN ?", p.PermissionIDs).Find(&permissions)
		database.DB.Model(&role).Association("Permissions").Replace(permissions)
	}

	response.Success(c, "更新成功", role)
}

// Delete 删除角色
func (r *roleHandler) Delete(c *gin.Context) {
	var p struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 不能删除 super_admin 角色
	var role model.Role
	if err := database.DB.First(&role, p.ID).Error; err != nil {
		response.Fail(c, "角色不存在")
		return
	}
	if role.Name == "super_admin" {
		response.Fail(c, "不能删除超级管理员角色")
		return
	}

	if err := database.DB.Model(&role).Association("Permissions").Clear(); err != nil {
		response.Fail(c, "清除角色权限失败")
		return
	}
	if err := database.DB.Delete(&role).Error; err != nil {
		response.Fail(c, "删除角色失败")
		return
	}

	response.Success(c, "删除成功", nil)
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./app/auth/api/...
```

- [ ] **Step 3: 提交**

```bash
git add app/auth/api/role.go
git commit -m "feat(auth): add role CRUD API handlers"
```

---

## Task 12: 集群管理 API

**Files:**
- Create: `app/cluster/api/cluster.go`

- [ ] **Step 1: 创建集群管理 Handler**

```go
// app/cluster/api/cluster.go
package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gkube/app/cluster/model"
	"gkube/app/cluster/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/k8s"
	"gkube/pkg/response"
)

type clusterHandler struct{}

var Cluster = new(clusterHandler)

// List 集群列表
func (cl *clusterHandler) List(c *gin.Context) {
	var p params.ClusterQueryParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 20
	}

	query := database.DB.Model(&model.K8SCluster{})
	if p.Status != "" {
		query = query.Where("status = ?", p.Status)
	}
	if p.Keyword != "" {
		query = query.Where("cluster_name LIKE ? OR display_name LIKE ?", "%"+p.Keyword+"%", "%"+p.Keyword+"%")
	}

	var total int64
	query.Count(&total)

	var clusters []model.K8SCluster
	query.Offset((p.Page - 1) * p.Size).Limit(p.Size).Find(&clusters)

	response.Success(c, "获取成功", gin.H{
		"total": total,
		"items": clusters,
	})
}

// Create 注册集群
func (cl *clusterHandler) Create(c *gin.Context) {
	var p params.CreateClusterParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 检查集群名是否已存在
	var count int64
	database.DB.Model(&model.K8SCluster{}).Where("cluster_name = ?", p.ClusterName).Count(&count)
	if count > 0 {
		response.Fail(c, "集群名称已存在")
		return
	}

	// 验证 kubeconfig 连通性
	client, err := k8s.GetK8sClient(p.KubeConfig)
	if err != nil {
		response.Fail(c, "KubeConfig 无效或集群不可达: " + err.Error())
		return
	}

	// 获取集群版本
	version, err := client.Discovery().ServerVersion()
	if err != nil {
		response.Fail(c, "无法获取集群版本: " + err.Error())
		return
	}

	// 加密 kubeconfig
	encryptedConfig, err := auth.EncryptAES(p.KubeConfig)
	if err != nil {
		response.Fail(c, "KubeConfig 加密失败")
		return
	}

	// 获取节点数
	nodes, _ := client.CoreV1().Nodes().List(c, metav1.ListOptions{})
	nodeCount := 0
	if nodes != nil {
		nodeCount = len(nodes.Items)
	}

	cluster := model.K8SCluster{
		ClusterName:     p.ClusterName,
		DisplayName:     p.DisplayName,
		Description:     p.Description,
		KubeConfig:      encryptedConfig,
		Status:          "online",
		ClusterVersion:  version.GitVersion,
		NodeCount:       nodeCount,
		LastHealthCheck: time.Now(),
		Labels:          p.Labels,
	}

	if err := database.DB.Create(&cluster).Error; err != nil {
		response.Fail(c, "注册集群失败: " + err.Error())
		return
	}

	response.Success(c, "注册成功", cluster)
}

// Detail 集群详情
func (cl *clusterHandler) Detail(c *gin.Context) {
	var p params.ClusterIDParams
	if err := c.ShouldBindUri(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	var cluster model.K8SCluster
	if err := database.DB.First(&cluster, p.ID).Error; err != nil {
		response.Fail(c, "集群不存在")
		return
	}

	response.Success(c, "获取成功", cluster)
}

// Update 更新集群
func (cl *clusterHandler) Update(c *gin.Context) {
	var p params.UpdateClusterParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	var cluster model.K8SCluster
	if err := database.DB.First(&cluster, p.ID).Error; err != nil {
		response.Fail(c, "集群不存在")
		return
	}

	updates := map[string]interface{}{}
	if p.DisplayName != "" {
		updates["display_name"] = p.DisplayName
	}
	if p.Description != "" {
		updates["description"] = p.Description
	}
	if p.Labels != "" {
		updates["labels"] = p.Labels
	}

	database.DB.Model(&cluster).Updates(updates)
	response.Success(c, "更新成功", cluster)
}

// Delete 删除集群
func (cl *clusterHandler) Delete(c *gin.Context) {
	var p struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if err := database.DB.Delete(&model.K8SCluster{}, p.ID).Error; err != nil {
		response.Fail(c, "删除集群失败")
		return
	}

	response.Success(c, "删除成功", nil)
}

// Check 连通性检测
func (cl *clusterHandler) Check(c *gin.Context) {
	var p params.ClusterIDParams
	if err := c.ShouldBindUri(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	var cluster model.K8SCluster
	if err := database.DB.First(&cluster, p.ID).Error; err != nil {
		response.Fail(c, "集群不存在")
		return
	}

	// 解密 kubeconfig
	kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
	if err != nil {
		response.Fail(c, "KubeConfig 解密失败")
		return
	}

	// 测试连通性
	start := time.Now()
	client, err := k8s.GetK8sClient(kubeConfig)
	if err != nil {
		database.DB.Model(&cluster).Update("status", "offline")
		response.Success(c, "检测完成", gin.H{
			"connected": false,
			"message":   err.Error(),
		})
		return
	}

	version, err := client.Discovery().ServerVersion()
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		database.DB.Model(&cluster).Update("status", "offline")
		response.Success(c, "检测完成", gin.H{
			"connected": false,
			"message":   err.Error(),
		})
		return
	}

	// 更新集群状态
	nodes, _ := client.CoreV1().Nodes().List(c, metav1.ListOptions{})
	nodeCount := 0
	if nodes != nil {
		nodeCount = len(nodes.Items)
	}

	database.DB.Model(&cluster).Updates(map[string]interface{}{
		"status":           "online",
		"cluster_version":  version.GitVersion,
		"node_count":       nodeCount,
		"last_health_check": time.Now(),
	})

	response.Success(c, "检测完成", gin.H{
		"connected":     true,
		"version":       version.GitVersion,
		"responseTimeMs": elapsed,
		"nodeCount":     nodeCount,
		"message":       "集群健康",
	})
}
```

- [ ] **Step 2: 验证编译**

需要在文件顶部添加 `metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"` 导入。

```bash
go build ./app/cluster/api/...
```

- [ ] **Step 3: 提交**

```bash
git add app/cluster/api/cluster.go
git commit -m "feat(cluster): add cluster CRUD and connectivity check API"
```

---

## Task 13: 集群健康检测服务

**Files:**
- Create: `app/cluster/service/health.go`

- [ ] **Step 1: 创建健康检测服务**

```go
// app/cluster/service/health.go
package service

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"gkube/app/cluster/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HealthChecker struct {
	interval time.Duration
	stopCh   chan struct{}
}

func NewHealthChecker(interval time.Duration) *HealthChecker {
	return &HealthChecker{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动健康检测
func (hc *HealthChecker) Start() {
	logrus.Info("集群健康检测服务启动，检测间隔: ", hc.interval)
	ticker := time.NewTicker(hc.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				hc.checkAll()
			case <-hc.stopCh:
				logrus.Info("集群健康检测服务停止")
				return
			}
		}
	}()
}

// Stop 停止健康检测
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

// checkAll 检测所有集群
func (hc *HealthChecker) checkAll() {
	var clusters []model.K8SCluster
	database.DB.Find(&clusters)

	var wg sync.WaitGroup
	for _, cluster := range clusters {
		wg.Add(1)
		go func(c model.K8SCluster) {
			defer wg.Done()
			hc.checkOne(c)
		}(cluster)
	}
	wg.Wait()
}

// checkOne 检测单个集群
func (hc *HealthChecker) checkOne(cluster model.K8SCluster) {
	// 解密 kubeconfig
	kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
	if err != nil {
		logrus.Errorf("集群 %s kubeconfig 解密失败: %v", cluster.ClusterName, err)
		hc.updateStatus(cluster.ID, "offline")
		return
	}

	// 测试连通性
	client, err := k8s.GetK8sClient(kubeConfig)
	if err != nil {
		logrus.Errorf("集群 %s 连接失败: %v", cluster.ClusterName, err)
		hc.updateStatus(cluster.ID, "offline")
		return
	}

	// 获取版本和节点信息
	version, err := client.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("集群 %s 获取版本失败: %v", cluster.ClusterName, err)
		hc.updateStatus(cluster.ID, "offline")
		return
	}

	nodes, _ := client.CoreV1().Nodes().List(nil, metav1.ListOptions{})
	nodeCount := 0
	if nodes != nil {
		nodeCount = len(nodes.Items)
	}

	// 更新集群状态
	database.DB.Model(&model.K8SCluster{}).Where("id = ?", cluster.ID).Updates(map[string]interface{}{
		"status":           "online",
		"cluster_version":  version.GitVersion,
		"node_count":       nodeCount,
		"last_health_check": time.Now(),
	})
}

func (hc *HealthChecker) updateStatus(id uint, status string) {
	database.DB.Model(&model.K8SCluster{}).Where("id = ?", id).Update("status", status)
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./app/cluster/service/...
```

- [ ] **Step 3: 提交**

```bash
git add app/cluster/service/health.go
git commit -m "feat(cluster): add health check background service"
```

---

## Task 14: Dashboard API

**Files:**
- Create: `app/dashboard/api/dashboard.go`

- [ ] **Step 1: 创建 Dashboard Handler**

```go
// app/dashboard/api/dashboard.go
package api

import (
	"github.com/gin-gonic/gin"

	"gkube/app/cluster/model"
	"gkube/app/dashboard/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/k8s"
	"gkube/pkg/response"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type dashboardHandler struct{}

var Dashboard = new(dashboardHandler)

// Overview 全局概览
func (d *dashboardHandler) Overview(c *gin.Context) {
	var clusters []model.K8SCluster
	database.DB.Find(&clusters)

	total := len(clusters)
	online := 0
	offline := 0
	totalNodes := 0
	totalPods := 0

	for _, cluster := range clusters {
		if cluster.Status == "online" {
			online++
		} else {
			offline++
		}
		totalNodes += cluster.NodeCount

		// 获取 Pod 数量（仅在线集群）
		if cluster.Status == "online" {
			kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
			if err != nil {
				continue
			}
			client, err := k8s.GetK8sClient(kubeConfig)
			if err != nil {
				continue
			}
			pods, _ := client.CoreV1().Pods("").List(c, metav1.ListOptions{})
			if pods != nil {
				totalPods += len(pods.Items)
			}
		}
	}

	response.Success(c, "获取成功", gin.H{
		"totalClusters": total,
		"online":        online,
		"offline":       offline,
		"totalNodes":    totalNodes,
		"totalPods":     totalPods,
	})
}

// Resources 资源使用率
func (d *dashboardHandler) Resources(c *gin.Context) {
	var clusters []model.K8SCluster
	database.DB.Find(&clusters)

	var clusterResources []gin.H
	for _, cluster := range clusters {
		info := gin.H{
			"clusterId":   cluster.ID,
			"clusterName": cluster.ClusterName,
			"displayName": cluster.DisplayName,
			"status":      cluster.Status,
			"nodeCount":   cluster.NodeCount,
		}

		if cluster.Status == "online" {
			kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
			if err != nil {
				clusterResources = append(clusterResources, info)
				continue
			}
			client, err := k8s.GetK8sClient(kubeConfig)
			if err != nil {
				clusterResources = append(clusterResources, info)
				continue
			}

			// 获取节点资源
			nodes, _ := client.CoreV1().Nodes().List(c, metav1.ListOptions{})
			if nodes != nil {
				info["nodesReady"] = len(nodes.Items)
			}

			// 获取 Pod 数量
			pods, _ := client.CoreV1().Pods("").List(c, metav1.ListOptions{})
			runningPods := 0
			if pods != nil {
				for _, pod := range pods.Items {
					if pod.Status.Phase == "Running" {
						runningPods++
					}
				}
				info["totalPods"] = len(pods.Items)
				info["runningPods"] = runningPods
			}
		}

		clusterResources = append(clusterResources, info)
	}

	response.Success(c, "获取成功", gin.H{
		"clusters": clusterResources,
	})
}

// Workloads 工作负载统计
func (d *dashboardHandler) Workloads(c *gin.Context) {
	var clusters []model.K8SCluster
	database.DB.Find(&clusters)

	totalDeployments := 0
	totalStatefulSets := 0
	totalDaemonSets := 0

	for _, cluster := range clusters {
		if cluster.Status != "online" {
			continue
		}
		kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
		if err != nil {
			continue
		}
		client, err := k8s.GetK8sClient(kubeConfig)
		if err != nil {
			continue
		}

		deploys, _ := client.AppsV1().Deployments("").List(c, metav1.ListOptions{})
		if deploys != nil {
			totalDeployments += len(deploys.Items)
		}
		statefulsets, _ := client.AppsV1().StatefulSets("").List(c, metav1.ListOptions{})
		if statefulsets != nil {
			totalStatefulSets += len(statefulsets.Items)
		}
		daemonsets, _ := client.AppsV1().DaemonSets("").List(c, metav1.ListOptions{})
		if daemonsets != nil {
			totalDaemonSets += len(daemonsets.Items)
		}
	}

	response.Success(c, "获取成功", gin.H{
		"deployments":  totalDeployments,
		"statefulSets": totalStatefulSets,
		"daemonSets":   totalDaemonSets,
	})
}

// Events 跨集群事件聚合
func (d *dashboardHandler) Events(c *gin.Context) {
	var p params.EventQueryParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}

	var clusters []model.K8SCluster
	if p.ClusterID != nil {
		database.DB.Where("id = ?", *p.ClusterID).Find(&clusters)
	} else {
		database.DB.Find(&clusters)
	}

	type EventInfo struct {
		ClusterName string `json:"clusterName"`
		Type        string `json:"type"`
		Reason      string `json:"reason"`
		Message     string `json:"message"`
		Object      string `json:"object"`
		Timestamp   string `json:"timestamp"`
	}
	var allEvents []EventInfo

	for _, cluster := range clusters {
		if cluster.Status != "online" {
			continue
		}
		kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
		if err != nil {
			continue
		}
		client, err := k8s.GetK8sClient(kubeConfig)
		if err != nil {
			continue
		}

		events, _ := client.CoreV1().Events("").List(c, metav1.ListOptions{})
		if events == nil {
			continue
		}

		for _, event := range events.Items {
			if p.Type != "" && event.Type != p.Type {
				continue
			}
			allEvents = append(allEvents, EventInfo{
				ClusterName: cluster.ClusterName,
				Type:        event.Type,
				Reason:      event.Reason,
				Message:     event.Message,
				Object:      event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
				Timestamp:   event.LastTimestamp.Time.Format("2006-01-02 15:04:05"),
			})
		}
	}

	// 截断到 limit
	if len(allEvents) > p.Limit {
		allEvents = allEvents[:p.Limit]
	}

	response.Success(c, "获取成功", gin.H{
		"events": allEvents,
	})
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./app/dashboard/api/...
```

- [ ] **Step 3: 提交**

```bash
git add app/dashboard/api/dashboard.go
git commit -m "feat(dashboard): add global overview, resources, workloads, events APIs"
```

---

## Task 15: 路由重写

**Files:**
- Rewrite: `router/router.go`

- [ ] **Step 1: 重写路由文件**

```go
// router/router.go
package router

import (
	"github.com/gin-gonic/gin"

	authApi "gkube/app/auth/api"
	clusterApi "gkube/app/cluster/api"
	dashboardApi "gkube/app/dashboard/api"
	k8sApi "gkube/app/k8s/api"
	"gkube/pkg/middleware"
)

func Engine() *gin.Engine {
	r := gin.Default()

	// CORS 中间件
	r.Use(middleware.Cors())

	v1 := r.Group("/v1")

	// ===== 公开路由（无需认证） =====
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authApi.Auth.Login)
		auth.POST("/refresh", authApi.Auth.Refresh)
	}

	// ===== 需要认证的路由 =====
	authorized := v1.Group("")
	authorized.Use(middleware.JWTAuth())
	{
		// 认证
		authorized.GET("/auth/me", authApi.Auth.GetMe)

		// 用户管理
		users := authorized.Group("/users")
		{
			users.GET("", middleware.RBAC("user", "read"), authApi.User.List)
			users.POST("", middleware.RBAC("user", "create"), authApi.User.Create)
			users.PUT("/:id", middleware.RBAC("user", "update"), authApi.User.Update)
			users.DELETE("/:id", middleware.RBAC("user", "delete"), authApi.User.Delete)
			users.POST("/change-password", authApi.User.ChangePassword)
		}

		// 角色管理
		roles := authorized.Group("/roles")
		{
			roles.GET("", middleware.RBAC("role", "read"), authApi.Role.List)
			roles.POST("", middleware.RBAC("role", "create"), authApi.Role.Create)
			roles.PUT("/:id", middleware.RBAC("role", "update"), authApi.Role.Update)
			roles.DELETE("/:id", middleware.RBAC("role", "delete"), authApi.Role.Delete)
		}

		// 集群管理
		clusters := authorized.Group("/clusters")
		{
			clusters.GET("", middleware.RBAC("cluster", "read"), clusterApi.Cluster.List)
			clusters.POST("", middleware.RBAC("cluster", "create"), clusterApi.Cluster.Create)
			clusters.GET("/:id", middleware.RBAC("cluster", "read"), clusterApi.Cluster.Detail)
			clusters.PUT("/:id", middleware.RBAC("cluster", "update"), clusterApi.Cluster.Update)
			clusters.DELETE("/:id", middleware.RBAC("cluster", "delete"), clusterApi.Cluster.Delete)
			clusters.POST("/:id/check", middleware.RBAC("cluster", "read"), clusterApi.Cluster.Check)
		}

		// Dashboard
		dashboard := authorized.Group("/dashboard")
		{
			dashboard.GET("/overview", dashboardApi.Dashboard.Overview)
			dashboard.GET("/resources", dashboardApi.Dashboard.Resources)
			dashboard.GET("/workloads", dashboardApi.Dashboard.Workloads)
			dashboard.GET("/events", dashboardApi.Dashboard.Events)
		}

		// K8s 资源（挂在集群路由下，保持现有功能）
		k8sGroup := authorized.Group("/k8s")
		{
			// 集群信息
			k8sGroup.GET("/cluster/version", k8sApi.Cluster.GetClusterVersion)
			k8sGroup.GET("/cluster/nodes", k8sApi.Cluster.GetClusterNodesInfo)

			// Pod
			k8sGroup.GET("/pod/list", k8sApi.Pod.GetPodList)
			k8sGroup.GET("/pod/detail", k8sApi.Pod.GetPodByName)
			k8sGroup.GET("/pod/get-by-label", k8sApi.Pod.GetPodByLabels)
			k8sGroup.GET("/pod/get-by-field", k8sApi.Pod.GetPodByField)
			k8sGroup.GET("/pod/get-yaml", k8sApi.Pod.GetPodYaml)
			k8sGroup.POST("/pod/create", k8sApi.Pod.CreatePod)
			k8sGroup.PUT("/pod/update", k8sApi.Pod.UpdatePod)
			k8sGroup.DELETE("/pod/delete", k8sApi.Pod.DeletePod)
			k8sGroup.GET("/pod/events", k8sApi.Pod.GetPodEvents)

			// Deployment
			k8sGroup.GET("/deployment/list", k8sApi.Deployment.GetDeploymentList)
			k8sGroup.GET("/deployment/detail", k8sApi.Deployment.GetDeploymentByName)
			k8sGroup.GET("/deployment/get-by-label", k8sApi.Deployment.GetDeploymentByLabels)
			k8sGroup.GET("/deployment/get-by-field", k8sApi.Deployment.GetDeploymentByField)
			k8sGroup.GET("/deployment/get-yaml", k8sApi.Deployment.GetDeploymentYaml)
			k8sGroup.POST("/deployment/create", k8sApi.Deployment.CreateDeployment)
			k8sGroup.PUT("/deployment/update", k8sApi.Deployment.UpdateDeployment)
			k8sGroup.DELETE("/deployment/delete", k8sApi.Deployment.DeleteDeployment)
			k8sGroup.PUT("/deployment/scale", k8sApi.Deployment.ScaleDeployment)
			k8sGroup.POST("/deployment/restart", k8sApi.Deployment.RestartDeployment)
			k8sGroup.GET("/deployment/pods", k8sApi.Deployment.GetDeploymentPods)

			// Namespace
			k8sGroup.GET("/namespace/list", k8sApi.Namespace.GetNamespaceList)
			k8sGroup.POST("/namespace/create", k8sApi.Namespace.CreateNamespace)

			// Node
			k8sGroup.GET("/node/get-yaml", k8sApi.Node.GetNodeYaml)
			k8sGroup.GET("/node/pods", k8sApi.Node.GetNodePods)
			k8sGroup.PUT("/node/cordon", k8sApi.Node.CordonNode)
			k8sGroup.POST("/node/evict-all-pods", k8sApi.Node.EvictAllPods)
			k8sGroup.POST("/node/evict-pod", k8sApi.Node.EvictPod)
			k8sGroup.PUT("/node/taint", k8sApi.Node.TaintNode)

			// Service
			k8sGroup.GET("/service/list", k8sApi.Service.GetServiceList)
			k8sGroup.GET("/service/detail", k8sApi.Service.GetServiceByName)
			k8sGroup.GET("/service/get-yaml", k8sApi.Service.GetServiceYaml)
			k8sGroup.POST("/service/create", k8sApi.Service.CreateService)
			k8sGroup.PUT("/service/update", k8sApi.Service.UpdateService)
			k8sGroup.DELETE("/service/delete", k8sApi.Service.DeleteService)

			// Ingress
			k8sGroup.GET("/ingress/list", k8sApi.Ingress.GetIngressList)
			k8sGroup.GET("/ingress/detail", k8sApi.Ingress.GetIngressByName)
			k8sGroup.GET("/ingress/get-yaml", k8sApi.Ingress.GetIngressYaml)
			k8sGroup.POST("/ingress/create", k8sApi.Ingress.CreateIngress)
			k8sGroup.PUT("/ingress/update", k8sApi.Ingress.UpdateIngress)
			k8sGroup.DELETE("/ingress/delete", k8sApi.Ingress.DeleteIngress)

			// ConfigMap
			k8sGroup.GET("/configmap/list", k8sApi.ConfigMap.GetConfigMapList)
			k8sGroup.GET("/configmap/detail", k8sApi.ConfigMap.GetConfigMapByName)
			k8sGroup.GET("/configmap/get-yaml", k8sApi.ConfigMap.GetConfigMapYaml)
			k8sGroup.POST("/configmap/create", k8sApi.ConfigMap.CreateConfigMap)
			k8sGroup.PUT("/configmap/update", k8sApi.ConfigMap.UpdateConfigMap)
			k8sGroup.DELETE("/configmap/delete", k8sApi.ConfigMap.DeleteConfigMap)

			// Secret
			k8sGroup.GET("/secret/list", k8sApi.Secret.GetSecretList)
			k8sGroup.GET("/secret/detail", k8sApi.Secret.GetSecretByName)
			k8sGroup.GET("/secret/get-yaml", k8sApi.Secret.GetSecretYaml)
			k8sGroup.POST("/secret/create", k8sApi.Secret.CreateSecret)
			k8sGroup.PUT("/secret/update", k8sApi.Secret.UpdateSecret)
			k8sGroup.DELETE("/secret/delete", k8sApi.Secret.DeleteSecret)

			// PV
			k8sGroup.GET("/pv/list", k8sApi.Pv.GetPvList)
			k8sGroup.GET("/pv/detail", k8sApi.Pv.GetPvByName)
			k8sGroup.GET("/pv/get-yaml", k8sApi.Pv.GetPvYaml)
			k8sGroup.POST("/pv/create", k8sApi.Pv.CreatePv)
			k8sGroup.PUT("/pv/update", k8sApi.Pv.UpdatePv)
			k8sGroup.DELETE("/pv/delete", k8sApi.Pv.DeletePv)

			// PVC
			k8sGroup.GET("/pvc/list", k8sApi.Pvc.GetPvcList)
			k8sGroup.GET("/pvc/detail", k8sApi.Pvc.GetPvcByName)
			k8sGroup.GET("/pvc/get-yaml", k8sApi.Pvc.GetPvcYaml)
			k8sGroup.POST("/pvc/create", k8sApi.Pvc.CreatePvc)
			k8sGroup.PUT("/pvc/update", k8sApi.Pvc.UpdatePvc)
			k8sGroup.DELETE("/pvc/delete", k8sApi.Pvc.DeletePvc)

			// StorageClass
			k8sGroup.GET("/storageclass/list", k8sApi.StorageClass.GetStorageClassList)
			k8sGroup.GET("/storageclass/detail", k8sApi.StorageClass.GetStorageClassByName)
			k8sGroup.GET("/storageclass/get-yaml", k8sApi.StorageClass.GetStorageClassYaml)
			k8sGroup.POST("/storageclass/create", k8sApi.StorageClass.CreateStorageClass)
			k8sGroup.PUT("/storageclass/update", k8sApi.StorageClass.UpdateStorageClass)
			k8sGroup.DELETE("/storageclass/delete", k8sApi.StorageClass.DeleteStorageClass)

			// StatefulSet
			k8sGroup.GET("/statefulset/list", k8sApi.StatefulSet.GetStatefulSetList)
			k8sGroup.GET("/statefulset/detail", k8sApi.StatefulSet.GetStatefulSetByName)
			k8sGroup.GET("/statefulset/get-yaml", k8sApi.StatefulSet.GetStatefulSetYaml)
			k8sGroup.POST("/statefulset/create", k8sApi.StatefulSet.CreateStatefulSet)
			k8sGroup.PUT("/statefulset/update", k8sApi.StatefulSet.UpdateStatefulSet)
			k8sGroup.DELETE("/statefulset/delete", k8sApi.StatefulSet.DeleteStatefulSet)

			// DaemonSet
			k8sGroup.GET("/daemonset/list", k8sApi.DaemonSet.GetDaemonSetList)
			k8sGroup.GET("/daemonset/detail", k8sApi.DaemonSet.GetDaemonSetByName)
			k8sGroup.GET("/daemonset/get-yaml", k8sApi.DaemonSet.GetDaemonSetYaml)
			k8sGroup.POST("/daemonset/create", k8sApi.DaemonSet.CreateDaemonSet)
			k8sGroup.PUT("/daemonset/update", k8sApi.DaemonSet.UpdateDaemonSet)
			k8sGroup.DELETE("/daemonset/delete", k8sApi.DaemonSet.DeleteDaemonSet)

			// Job
			k8sGroup.GET("/job/list", k8sApi.Job.GetJobList)
			k8sGroup.GET("/job/detail", k8sApi.Job.GetJobByName)
			k8sGroup.GET("/job/get-yaml", k8sApi.Job.GetJobYaml)
			k8sGroup.POST("/job/create", k8sApi.Job.CreateJob)
			k8sGroup.PUT("/job/update", k8sApi.Job.UpdateJob)
			k8sGroup.DELETE("/job/delete", k8sApi.Job.DeleteJob)

			// CronJob
			k8sGroup.GET("/cronjob/list", k8sApi.CronJob.GetCronJobList)
			k8sGroup.GET("/cronjob/detail", k8sApi.CronJob.GetCronJobByName)
			k8sGroup.GET("/cronjob/get-yaml", k8sApi.CronJob.GetCronJobYaml)
			k8sGroup.POST("/cronjob/create", k8sApi.CronJob.CreateCronJob)
			k8sGroup.PUT("/cronjob/update", k8sApi.CronJob.UpdateCronJob)
			k8sGroup.DELETE("/cronjob/delete", k8sApi.CronJob.DeleteCronJob)

			// Container (exec, logs)
			k8sGroup.GET("/container/logs", k8sApi.Container.GetContainerLogs)
			k8sGroup.GET("/container/log-stream", k8sApi.Container.LogStream)
			k8sGroup.GET("/container/exec", k8sApi.Container.HandleTerminal)
		}
	}

	return r
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./router/...
```

- [ ] **Step 3: 提交**

```bash
git add router/router.go
git commit -m "feat(router): rewrite routes with auth, cluster, dashboard modules"
```

---

## Task 16: 初始化种子数据

**Files:**
- Create: `cmd/seed.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: 创建种子数据命令**

```go
// cmd/seed.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	authModel "gkube/app/auth/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "初始化默认数据（管理员账号、默认角色、默认权限）",
	Run: func(cmd *cobra.Command, args []string) {
		seedPermissions()
		seedRoles()
		seedAdmin()
	},
}

func seedPermissions() {
	permissions := []authModel.Permission{
		{Resource: "*", Action: "*", Description: "全部权限"},
		{Resource: "cluster", Action: "create", Description: "创建集群"},
		{Resource: "cluster", Action: "read", Description: "查看集群"},
		{Resource: "cluster", Action: "update", Description: "更新集群"},
		{Resource: "cluster", Action: "delete", Description: "删除集群"},
		{Resource: "user", Action: "create", Description: "创建用户"},
		{Resource: "user", Action: "read", Description: "查看用户"},
		{Resource: "user", Action: "update", Description: "更新用户"},
		{Resource: "user", Action: "delete", Description: "删除用户"},
		{Resource: "role", Action: "create", Description: "创建角色"},
		{Resource: "role", Action: "read", Description: "查看角色"},
		{Resource: "role", Action: "update", Description: "更新角色"},
		{Resource: "role", Action: "delete", Description: "删除角色"},
		{Resource: "pod", Action: "read", Description: "查看Pod"},
		{Resource: "pod", Action: "create", Description: "创建Pod"},
		{Resource: "pod", Action: "delete", Description: "删除Pod"},
		{Resource: "pod", Action: "exec", Description: "Pod终端"},
		{Resource: "deployment", Action: "read", Description: "查看Deployment"},
		{Resource: "deployment", Action: "create", Description: "创建Deployment"},
		{Resource: "deployment", Action: "update", Description: "更新Deployment"},
		{Resource: "deployment", Action: "delete", Description: "删除Deployment"},
	}

	for _, p := range permissions {
		database.DB.Where("resource = ? AND action = ?", p.Resource, p.Action).FirstOrCreate(&p)
	}
	fmt.Println("权限数据初始化完成")
}

func seedRoles() {
	// 超级管理员角色
	var adminRole authModel.Role
	result := database.DB.Where("name = ?", "super_admin").First(&adminRole)
	if result.RowsAffected == 0 {
		var allPerms []authModel.Permission
		database.DB.Find(&allPerms)
		adminRole = authModel.Role{
			Name:        "super_admin",
			Description: "超级管理员，拥有全部权限",
			Permissions: allPerms,
		}
		database.DB.Create(&adminRole)
	}

	// 只读角色
	var viewerRole authModel.Role
	result = database.DB.Where("name = ?", "viewer").First(&viewerRole)
	if result.RowsAffected == 0 {
		var readPerms []authModel.Permission
		database.DB.Where("action = ?", "read").Find(&readPerms)
		viewerRole = authModel.Role{
			Name:        "viewer",
			Description: "只读用户，可查看所有资源",
			Permissions: readPerms,
		}
		database.DB.Create(&viewerRole)
	}

	fmt.Println("角色数据初始化完成")
}

func seedAdmin() {
	var count int64
	database.DB.Model(&authModel.User{}).Count(&count)
	if count > 0 {
		fmt.Println("已存在用户，跳过管理员初始化")
		return
	}

	hash, err := auth.HashPassword("admin123")
	if err != nil {
		fmt.Printf("密码加密失败: %v\n", err)
		return
	}

	var adminRole authModel.Role
	database.DB.Where("name = ?", "super_admin").First(&adminRole)

	admin := authModel.User{
		Username:     "admin",
		PasswordHash: hash,
		Email:        "admin@gkube.local",
		DisplayName:  "管理员",
		Status:       1,
		Roles:        []authModel.Role{adminRole},
	}
	database.DB.Create(&admin)
	fmt.Println("管理员账号初始化完成 (admin / admin123)")
}

func init() {
	rootCmd.AddCommand(seedCmd)
}
```

- [ ] **Step 2: 更新 root.go 启动健康检测**

在 `cmd/root.go` 的 `Run` 函数中，HTTP 服务启动前添加健康检测服务启动：

```go
import clusterService "gkube/app/cluster/service"

// 在 Run 函数中，server.ListenAndServe() 之前添加：
healthChecker := clusterService.NewHealthChecker(30 * time.Second)
healthChecker.Start()
defer healthChecker.Stop()
```

- [ ] **Step 3: 验证编译**

```bash
go build ./...
```

- [ ] **Step 4: 提交**

```bash
git add cmd/seed.go cmd/root.go
git commit -m "feat: add seed command for default data and health checker startup"
```

---

## Task 17: 后端验证与数据库迁移

**Files:**
- Modify: `cmd/migrate.go` (已在 Task 4/5 中修改)

- [ ] **Step 1: 编译全部后端代码**

```bash
cd /Users/zqqzqq/05_github/gkube
go build ./...
```

Expected: 编译成功，无报错

- [ ] **Step 2: 执行数据库迁移**

```bash
go run main.go migrate
```

Expected: 所有表创建成功

- [ ] **Step 3: 初始化种子数据**

```bash
go run main.go seed
```

Expected: 输出 "权限数据初始化完成"、"角色数据初始化完成"、"管理员账号初始化完成 (admin / admin123)"

- [ ] **Step 4: 启动服务验证**

```bash
go run main.go
```

Expected: 服务在 0.0.0.0:8080 启动

- [ ] **Step 5: 测试登录 API**

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Expected: 返回 `{"code":1,"msg":"登录成功","data":{"accessToken":"...","refreshToken":"...","user":{...}}}`

- [ ] **Step 6: 提交**

```bash
git add -A
git commit -m "feat: complete backend auth, cluster management, dashboard APIs"
```

---

## Task 18: 前端项目初始化

**Files:**
- Create: `frontend/` (整个目录)

- [ ] **Step 1: 使用 Vite 创建 Vue 3 + TypeScript 项目**

```bash
cd /Users/zqqzqq/05_github/gkube
npm create vite@latest frontend -- --template vue-ts
cd frontend
npm install
```

- [ ] **Step 2: 安装依赖**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm install element-plus @element-plus/icons-vue
npm install vue-router@4 pinia axios
npm install -D @types/node
```

- [ ] **Step 3: 验证项目启动**

```bash
npm run dev
```

Expected: 开发服务器在 http://localhost:5173 启动

- [ ] **Step 4: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/
git commit -m "feat(frontend): initialize Vue 3 + TypeScript + Element Plus project"
```

---

## Task 19: 前端 — Axios 请求封装

**Files:**
- Create: `frontend/src/utils/auth.ts`
- Create: `frontend/src/api/request.ts`
- Create: `frontend/src/api/auth.ts`
- Create: `frontend/src/api/cluster.ts`
- Create: `frontend/src/api/dashboard.ts`
- Create: `frontend/src/api/resource.ts`

- [ ] **Step 1: 创建 Token 工具函数**

```typescript
// frontend/src/utils/auth.ts
const TOKEN_KEY = 'gkube_access_token'
const REFRESH_KEY = 'gkube_refresh_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY)
}

export function setRefreshToken(token: string): void {
  localStorage.setItem(REFRESH_KEY, token)
}
```

- [ ] **Step 2: 创建 Axios 实例**

```typescript
// frontend/src/api/request.ts
import axios from 'axios'
import { getToken, removeToken } from '@/utils/auth'
import { ElMessage } from 'element-plus'
import router from '@/router'

const service = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

// 请求拦截器
service.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code === 0) {
      ElMessage.error(res.msg || '请求失败')
      return Promise.reject(new Error(res.msg))
    }
    return res
  },
  (error) => {
    if (error.response?.status === 401) {
      removeToken()
      router.push('/login')
    } else {
      ElMessage.error(error.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

export default service
```

- [ ] **Step 3: 创建 API 模块**

```typescript
// frontend/src/api/auth.ts
import request from './request'

export function login(data: { username: string; password: string }) {
  return request.post('/auth/login', data)
}

export function refreshToken(data: { refreshToken: string }) {
  return request.post('/auth/refresh', data)
}

export function getMe() {
  return request.get('/auth/me')
}
```

```typescript
// frontend/src/api/cluster.ts
import request from './request'

export function getClusterList(params?: any) {
  return request.get('/clusters', { params })
}

export function createCluster(data: any) {
  return request.post('/clusters', data)
}

export function getClusterDetail(id: number) {
  return request.get(`/clusters/${id}`)
}

export function updateCluster(id: number, data: any) {
  return request.put(`/clusters/${id}`, data)
}

export function deleteCluster(id: number) {
  return request.delete('/clusters', { data: { id } })
}

export function checkCluster(id: number) {
  return request.post(`/clusters/${id}/check`)
}
```

```typescript
// frontend/src/api/dashboard.ts
import request from './request'

export function getOverview() {
  return request.get('/dashboard/overview')
}

export function getResources() {
  return request.get('/dashboard/resources')
}

export function getWorkloads() {
  return request.get('/dashboard/workloads')
}

export function getEvents(params?: any) {
  return request.get('/dashboard/events', { params })
}
```

```typescript
// frontend/src/api/resource.ts
import request from './request'

export function getPodList(params: any) {
  return request.get('/k8s/pod/list', { params })
}

export function getDeploymentList(params: any) {
  return request.get('/k8s/deployment/list', { params })
}

export function getServiceList(params: any) {
  return request.get('/k8s/service/list', { params })
}

export function getNamespaceList(params: any) {
  return request.get('/k8s/namespace/list', { params })
}
```

- [ ] **Step 4: 验证编译**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm run build
```

- [ ] **Step 5: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/api/ frontend/src/utils/
git commit -m "feat(frontend): add API request layer with JWT interceptor"
```

---

## Task 20: 前端 — Pinia 状态管理

**Files:**
- Create: `frontend/src/stores/auth.ts`
- Create: `frontend/src/stores/cluster.ts`
- Modify: `frontend/src/main.ts`

- [ ] **Step 1: 创建认证 Store**

```typescript
// frontend/src/stores/auth.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi, getMe } from '@/api/auth'
import { getToken, setToken, removeToken, setRefreshToken } from '@/utils/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const user = ref<any>(null)
  const roles = ref<string[]>([])

  async function login(form: { username: string; password: string }) {
    const res: any = await loginApi(form)
    token.value = res.data.accessToken
    setToken(res.data.accessToken)
    setRefreshToken(res.data.refreshToken)
    user.value = res.data.user
    return res
  }

  async function fetchUserInfo() {
    const res: any = await getMe()
    user.value = res.data
    roles.value = res.data.roles?.map((r: any) => r.name) || []
    return res
  }

  function logout() {
    token.value = null
    user.value = null
    roles.value = []
    removeToken()
  }

  const isSuperAdmin = () => roles.value.includes('super_admin')

  return { token, user, roles, login, fetchUserInfo, logout, isSuperAdmin }
})
```

- [ ] **Step 2: 创建集群 Store**

```typescript
// frontend/src/stores/cluster.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getClusterList } from '@/api/cluster'

export const useClusterStore = defineStore('cluster', () => {
  const clusterList = ref<any[]>([])
  const currentCluster = ref<any>(null)

  async function fetchClusters() {
    const res: any = await getClusterList({ page: 1, size: 100 })
    clusterList.value = res.data.items || []
  }

  function setCurrentCluster(cluster: any) {
    currentCluster.value = cluster
  }

  return { clusterList, currentCluster, fetchClusters, setCurrentCluster }
})
```

- [ ] **Step 3: 在 main.ts 中注册 Pinia**

```typescript
// frontend/src/main.ts — 添加 Pinia 注册
import { createPinia } from 'pinia'
// ...
app.use(createPinia())
```

- [ ] **Step 4: 验证编译**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm run build
```

- [ ] **Step 5: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/stores/ frontend/src/main.ts
git commit -m "feat(frontend): add Pinia stores for auth and cluster state"
```

---

## Task 21: 前端 — 路由配置

**Files:**
- Create: `frontend/src/router/index.ts`

- [ ] **Step 1: 创建路由配置**

```typescript
// frontend/src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/utils/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('@/components/Layout/AppLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/DashboardView.vue'),
        meta: { title: 'Dashboard' },
      },
      {
        path: 'clusters',
        name: 'ClusterList',
        component: () => import('@/views/cluster/ClusterList.vue'),
        meta: { title: '集群管理' },
      },
      {
        path: 'clusters/create',
        name: 'ClusterCreate',
        component: () => import('@/views/cluster/ClusterForm.vue'),
        meta: { title: '添加集群' },
      },
      {
        path: 'clusters/:id',
        name: 'ClusterDetail',
        component: () => import('@/views/cluster/ClusterDetail.vue'),
        meta: { title: '集群详情' },
      },
      {
        path: 'users',
        name: 'UserList',
        component: () => import('@/views/user/UserList.vue'),
        meta: { title: '用户管理', roles: ['super_admin'] },
      },
      {
        path: 'roles',
        name: 'RoleList',
        component: () => import('@/views/user/RoleList.vue'),
        meta: { title: '角色管理', roles: ['super_admin'] },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const token = getToken()
  if (to.path !== '/login' && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
```

- [ ] **Step 2: 更新 main.ts 注册路由**

```typescript
// frontend/src/main.ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.mount('#app')
```

- [ ] **Step 3: 配置 Vite 代理**

```typescript
// frontend/vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})
```

- [ ] **Step 4: 验证编译**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm run build
```

- [ ] **Step 5: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/router/ frontend/src/main.ts frontend/vite.config.ts
git commit -m "feat(frontend): add Vue Router with auth guard and Vite proxy config"
```

---

## Task 22: 前端 — 布局组件

**Files:**
- Create: `frontend/src/components/Layout/AppLayout.vue`
- Create: `frontend/src/components/Layout/Sidebar.vue`
- Create: `frontend/src/components/Layout/Header.vue`

- [ ] **Step 1: 创建整体布局**

```vue
<!-- frontend/src/components/Layout/AppLayout.vue -->
<template>
  <el-container style="height: 100vh">
    <el-aside :width="isCollapse ? '64px' : '220px'" style="transition: width 0.3s">
      <Sidebar :is-collapse="isCollapse" />
    </el-aside>
    <el-container>
      <el-header style="padding: 0; border-bottom: 1px solid #e4e7ed">
        <Header @toggle-collapse="isCollapse = !isCollapse" />
      </el-header>
      <el-main style="background: #f5f7fa">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'

const isCollapse = ref(false)
</script>
```

- [ ] **Step 2: 创建侧边栏**

```vue
<!-- frontend/src/components/Layout/Sidebar.vue -->
<template>
  <div style="height: 100%; background: #001529">
    <div style="height: 60px; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 18px; font-weight: bold">
      <span v-if="!isCollapse">☸ gkube</span>
      <span v-else>☸</span>
    </div>
    <el-menu
      :default-active="route.path"
      :collapse="isCollapse"
      background-color="#001529"
      text-color="#ffffffb3"
      active-text-color="#409eff"
      router
    >
      <el-menu-item index="/dashboard">
        <el-icon><DataBoard /></el-icon>
        <template #title>Dashboard</template>
      </el-menu-item>
      <el-menu-item index="/clusters">
        <el-icon><Monitor /></el-icon>
        <template #title>集群管理</template>
      </el-menu-item>
      <el-sub-menu index="user-mgmt">
        <template #title>
          <el-icon><User /></el-icon>
          <span>系统管理</span>
        </template>
        <el-menu-item index="/users">用户管理</el-menu-item>
        <el-menu-item index="/roles">角色管理</el-menu-item>
      </el-sub-menu>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

defineProps<{ isCollapse: boolean }>()
const route = useRoute()
</script>
```

- [ ] **Step 3: 创建顶部栏**

```vue
<!-- frontend/src/components/Layout/Header.vue -->
<template>
  <div style="height: 60px; display: flex; align-items: center; justify-content: space-between; padding: 0 20px; background: #fff">
    <div style="display: flex; align-items: center; gap: 16px">
      <el-icon style="cursor: pointer; font-size: 20px" @click="$emit('toggleCollapse')">
        <Fold />
      </el-icon>
      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
        <el-breadcrumb-item v-if="route.meta.title">{{ route.meta.title }}</el-breadcrumb-item>
      </el-breadcrumb>
    </div>
    <div style="display: flex; align-items: center; gap: 16px">
      <el-dropdown @command="handleCommand">
        <span style="cursor: pointer; display: flex; align-items: center; gap: 8px">
          <el-icon><User /></el-icon>
          {{ authStore.user?.displayName || authStore.user?.username || '用户' }}
          <el-icon><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="logout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

defineEmits(['toggleCollapse'])
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  }
}
</script>
```

- [ ] **Step 4: 验证编译**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm run build
```

- [ ] **Step 5: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/components/Layout/
git commit -m "feat(frontend): add layout components with sidebar and header"
```

---

## Task 23: 前端 — 登录页

**Files:**
- Create: `frontend/src/views/login/LoginView.vue`

- [ ] **Step 1: 创建登录页**

```vue
<!-- frontend/src/views/login/LoginView.vue -->
<template>
  <div style="min-height: 100vh; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%)">
    <el-card style="width: 400px" shadow="always">
      <template #header>
        <div style="text-align: center">
          <h2 style="margin: 0">☸ gkube</h2>
          <p style="color: #909399; margin: 8px 0 0">K8s 多集群管理平台</p>
        </div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="0" size="large">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width: 100%" :loading="loading" @click="handleLogin">登 录</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await authStore.login(form)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (e) {
    // 错误已在拦截器中处理
  } finally {
    loading.value = false
  }
}
</script>
```

- [ ] **Step 2: 验证编译**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm run build
```

- [ ] **Step 3: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/views/login/
git commit -m "feat(frontend): add login page"
```

---

## Task 24: 前端 — Dashboard 页面

**Files:**
- Create: `frontend/src/views/dashboard/DashboardView.vue`

- [ ] **Step 1: 创建 Dashboard 页面**

```vue
<!-- frontend/src/views/dashboard/DashboardView.vue -->
<template>
  <div>
    <!-- 统计卡片 -->
    <el-row :gutter="16" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card shadow="hover">
          <div style="text-align: center">
            <div style="font-size: 32px; font-weight: bold; color: #409eff">{{ overview.totalClusters }}</div>
            <div style="color: #909399; margin-top: 8px">集群总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div style="text-align: center">
            <div style="font-size: 32px; font-weight: bold; color: #67c23a">{{ overview.online }}</div>
            <div style="color: #909399; margin-top: 8px">在线集群</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div style="text-align: center">
            <div style="font-size: 32px; font-weight: bold; color: #e6a23c">{{ overview.totalNodes }}</div>
            <div style="color: #909399; margin-top: 8px">节点总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div style="text-align: center">
            <div style="font-size: 32px; font-weight: bold; color: #f56c6c">{{ overview.offline }}</div>
            <div style="color: #909399; margin-top: 8px">离线集群</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <!-- 工作负载统计 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header><span>工作负载统计</span></template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="Deployments">{{ workloads.deployments }}</el-descriptions-item>
            <el-descriptions-item label="StatefulSets">{{ workloads.statefulSets }}</el-descriptions-item>
            <el-descriptions-item label="DaemonSets">{{ workloads.daemonSets }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <!-- 最新事件 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header><span>最新事件</span></template>
          <el-table :data="events" style="width: 100%" max-height="300">
            <el-table-column prop="clusterName" label="集群" width="120" />
            <el-table-column prop="type" label="类型" width="80">
              <template #default="{ row }">
                <el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="原因" width="120" />
            <el-table-column prop="message" label="消息" show-overflow-tooltip />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getOverview, getWorkloads, getEvents } from '@/api/dashboard'

const overview = ref({ totalClusters: 0, online: 0, offline: 0, totalNodes: 0, totalPods: 0 })
const workloads = ref({ deployments: 0, statefulSets: 0, daemonSets: 0 })
const events = ref<any[]>([])

onMounted(async () => {
  try {
    const [overviewRes, workloadsRes, eventsRes]: any[] = await Promise.all([
      getOverview(),
      getWorkloads(),
      getEvents({ limit: 20 }),
    ])
    overview.value = overviewRes.data
    workloads.value = workloadsRes.data
    events.value = eventsRes.data.events || []
  } catch (e) {
    // 错误已在拦截器中处理
  }
})
</script>
```

- [ ] **Step 2: 验证编译**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm run build
```

- [ ] **Step 3: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/views/dashboard/
git commit -m "feat(frontend): add Dashboard page with overview, workloads, events"
```

---

## Task 25: 前端 — 集群管理页面

**Files:**
- Create: `frontend/src/views/cluster/ClusterList.vue`
- Create: `frontend/src/views/cluster/ClusterForm.vue`
- Create: `frontend/src/views/cluster/ClusterDetail.vue`

- [ ] **Step 1: 创建集群列表页**

```vue
<!-- frontend/src/views/cluster/ClusterList.vue -->
<template>
  <div>
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>集群管理</span>
          <el-button type="primary" @click="router.push('/clusters/create')">
            <el-icon><Plus /></el-icon> 添加集群
          </el-button>
        </div>
      </template>
      <el-table :data="clusters" style="width: 100%" v-loading="loading">
        <el-table-column prop="clusterName" label="集群名称" />
        <el-table-column prop="displayName" label="显示名称" />
        <el-table-column prop="clusterVersion" label="K8s 版本" />
        <el-table-column prop="nodeCount" label="节点数" width="80" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : row.status === 'offline' ? 'danger' : 'info'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastHealthCheck" label="最后检测" width="180">
          <template #default="{ row }">
            {{ row.lastHealthCheck ? new Date(row.lastHealthCheck).toLocaleString() : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250">
          <template #default="{ row }">
            <el-button size="small" @click="router.push(`/clusters/${row.id}`)">详情</el-button>
            <el-button size="small" type="warning" @click="handleCheck(row)" :loading="row._checking">检测</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getClusterList, deleteCluster, checkCluster } from '@/api/cluster'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()
const clusters = ref<any[]>([])
const loading = ref(false)

async function fetchClusters() {
  loading.value = true
  try {
    const res: any = await getClusterList({ page: 1, size: 100 })
    clusters.value = (res.data.items || []).map((c: any) => ({ ...c, _checking: false }))
  } finally {
    loading.value = false
  }
}

async function handleCheck(row: any) {
  row._checking = true
  try {
    const res: any = await checkCluster(row.id)
    ElMessage.success(res.data.message || '检测完成')
    await fetchClusters()
  } finally {
    row._checking = false
  }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确认删除集群 "${row.clusterName}" 吗？`, '确认')
  await deleteCluster(row.id)
  ElMessage.success('删除成功')
  await fetchClusters()
}

onMounted(fetchClusters)
</script>
```

- [ ] **Step 2: 创建添加/编辑集群表单**

```vue
<!-- frontend/src/views/cluster/ClusterForm.vue -->
<template>
  <el-card shadow="hover">
    <template #header><span>添加集群</span></template>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px" style="max-width: 600px">
      <el-form-item label="集群名称" prop="clusterName">
        <el-input v-model="form.clusterName" placeholder="如: prod-cluster-1" />
      </el-form-item>
      <el-form-item label="显示名称" prop="displayName">
        <el-input v-model="form.displayName" placeholder="如: 生产集群1" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="form.description" type="textarea" :rows="2" />
      </el-form-item>
      <el-form-item label="KubeConfig" prop="kubeConfig">
        <el-input v-model="form.kubeConfig" type="textarea" :rows="8" placeholder="粘贴 kubeconfig 内容" />
      </el-form-item>
      <el-form-item label="标签 (JSON)">
        <el-input v-model="form.labels" placeholder='{"env":"prod","region":"cn-east"}' />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="handleSubmit">注册集群</el-button>
        <el-button @click="router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { createCluster } from '@/api/cluster'
import { ElMessage } from 'element-plus'

const router = useRouter()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  clusterName: '',
  displayName: '',
  description: '',
  kubeConfig: '',
  labels: '',
})

const rules = {
  clusterName: [{ required: true, message: '请输入集群名称', trigger: 'blur' }],
  kubeConfig: [{ required: true, message: '请输入 KubeConfig', trigger: 'blur' }],
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await createCluster(form)
    ElMessage.success('集群注册成功')
    router.push('/clusters')
  } finally {
    loading.value = false
  }
}
</script>
```

- [ ] **Step 3: 创建集群详情页**

```vue
<!-- frontend/src/views/cluster/ClusterDetail.vue -->
<template>
  <div v-loading="loading">
    <el-card shadow="hover" style="margin-bottom: 16px">
      <template #header><span>集群详情</span></template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="集群名称">{{ cluster.clusterName }}</el-descriptions-item>
        <el-descriptions-item label="显示名称">{{ cluster.displayName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="K8s 版本">{{ cluster.clusterVersion || '-' }}</el-descriptions-item>
        <el-descriptions-item label="节点数">{{ cluster.nodeCount }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="cluster.status === 'online' ? 'success' : 'danger'">{{ cluster.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="最后检测">{{ cluster.lastHealthCheck ? new Date(cluster.lastHealthCheck).toLocaleString() : '-' }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ cluster.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="标签" :span="2">{{ cluster.labels || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
    <el-button type="primary" @click="router.push('/clusters')">返回列表</el-button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getClusterDetail } from '@/api/cluster'

const route = useRoute()
const router = useRouter()
const cluster = ref<any>({})
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const res: any = await getClusterDetail(Number(route.params.id))
    cluster.value = res.data
  } finally {
    loading.value = false
  }
})
</script>
```

- [ ] **Step 4: 验证编译**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm run build
```

- [ ] **Step 5: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/views/cluster/
git commit -m "feat(frontend): add cluster management pages (list, form, detail)"
```

---

## Task 26: 前端 — 用户管理页面

**Files:**
- Create: `frontend/src/views/user/UserList.vue`
- Create: `frontend/src/views/user/RoleList.vue`

- [ ] **Step 1: 创建用户管理页**

```vue
<!-- frontend/src/views/user/UserList.vue -->
<template>
  <el-card shadow="hover">
    <template #header>
      <div style="display: flex; justify-content: space-between; align-items: center">
        <span>用户管理</span>
        <el-button type="primary" @click="showCreateDialog">添加用户</el-button>
      </div>
    </template>
    <el-table :data="users" style="width: 100%" v-loading="loading">
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="displayName" label="显示名称" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="角色">
        <template #default="{ row }">
          <el-tag v-for="role in row.roles" :key="role.id" size="small" style="margin-right: 4px">{{ role.name }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="showEditDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '添加用户'" width="500px">
      <el-form ref="dialogFormRef" :model="dialogForm" :rules="dialogRules" label-width="100px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="dialogForm.username" :disabled="isEdit" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input v-model="dialogForm.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="显示名称">
          <el-input v-model="dialogForm.displayName" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="dialogForm.email" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="dialogLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

// 使用原生 axios 调用（因为 request 模块已创建）
import request from '@/api/request'

const users = ref<any[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogLoading = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const dialogFormRef = ref()

const dialogForm = reactive({
  username: '',
  password: '',
  displayName: '',
  email: '',
})

const dialogRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function fetchUsers() {
  loading.value = true
  try {
    const res: any = await request.get('/users', { params: { page: 1, size: 100 } })
    users.value = res.data.items || []
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  isEdit.value = false
  Object.assign(dialogForm, { username: '', password: '', displayName: '', email: '' })
  dialogVisible.value = true
}

function showEditDialog(row: any) {
  isEdit.value = true
  editId.value = row.id
  Object.assign(dialogForm, { username: row.username, password: '', displayName: row.displayName, email: row.email })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await dialogFormRef.value?.validate().catch(() => false)
  if (!valid) return

  dialogLoading.value = true
  try {
    if (isEdit.value) {
      await request.put(`/users/${editId.value}`, dialogForm)
      ElMessage.success('更新成功')
    } else {
      await request.post('/users', dialogForm)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    await fetchUsers()
  } finally {
    dialogLoading.value = false
  }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确认删除用户 "${row.username}" 吗？`, '确认')
  await request.delete('/users', { data: { id: row.id } })
  ElMessage.success('删除成功')
  await fetchUsers()
}

onMounted(fetchUsers)
</script>
```

- [ ] **Step 2: 创建角色管理页（简化版）**

```vue
<!-- frontend/src/views/user/RoleList.vue -->
<template>
  <el-card shadow="hover">
    <template #header><span>角色管理</span></template>
    <el-table :data="roles" style="width: 100%" v-loading="loading">
      <el-table-column prop="name" label="角色名" />
      <el-table-column prop="description" label="描述" />
      <el-table-column label="权限数量">
        <template #default="{ row }">{{ row.permissions?.length || 0 }}</template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间">
        <template #default="{ row }">{{ new Date(row.createdAt).toLocaleString() }}</template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/api/request'

const roles = ref<any[]>([])
const loading = ref(false)

async function fetchRoles() {
  loading.value = true
  try {
    const res: any = await request.get('/roles', { params: { page: 1, size: 100 } })
    roles.value = res.data.items || []
  } finally {
    loading.value = false
  }
}

onMounted(fetchRoles)
</script>
```

- [ ] **Step 3: 验证编译**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm run build
```

- [ ] **Step 4: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/views/user/
git commit -m "feat(frontend): add user and role management pages"
```

---

## Task 27: 前端 — 清理与全局样式

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/style.css`

- [ ] **Step 1: 更新 App.vue**

```vue
<!-- frontend/src/App.vue -->
<template>
  <router-view />
</template>

<script setup lang="ts">
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
</style>
```

- [ ] **Step 2: 清理默认样式**

删除或清空 `frontend/src/style.css` 中的默认样式。

- [ ] **Step 3: 验证编译**

```bash
cd /Users/zqqzqq/05_github/gkube/frontend
npm run build
```

- [ ] **Step 4: 提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add frontend/src/App.vue frontend/src/style.css
git commit -m "feat(frontend): clean up global styles and App.vue"
```

---

## Task 28: 全栈验证

- [ ] **Step 1: 编译后端**

```bash
cd /Users/zqqzqq/05_github/gkube
go build -o gkube ./main.go
```

- [ ] **Step 2: 执行迁移和种子数据**

```bash
./gkube migrate
./gkube seed
```

- [ ] **Step 3: 启动后端**

```bash
./gkube
```

- [ ] **Step 4: 启动前端开发服务器**

```bash
cd frontend && npm run dev
```

- [ ] **Step 5: 验证完整流程**

1. 打开 http://localhost:5173
2. 应自动跳转到登录页
3. 使用 admin / admin123 登录
4. 应看到 Dashboard 页面
5. 点击"集群管理"，应看到空列表
6. 点击"添加集群"，粘贴 kubeconfig 注册一个集群
7. 返回集群列表，应看到新注册的集群
8. 点击"检测"按钮，应显示连通性结果
9. 点击"用户管理"，应看到 admin 用户
10. 点击"角色管理"，应看到 super_admin 和 viewer 角色

- [ ] **Step 6: 最终提交**

```bash
cd /Users/zqqzqq/05_github/gkube
git add -A
git commit -m "feat: complete K8s multi-cluster management feature"
```
