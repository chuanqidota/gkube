package auth

import (
	"fmt"
	"sync"
	"time"

	rbacmodel "gkube/internal/rbac/model"
	"gkube/pkg/database"
	"gkube/pkg/logger"
)

// CachedPermissions 缓存的用户权限数据
type CachedPermissions struct {
	IsSuperAdmin bool
	Bindings     []rbacmodel.PermissionBinding
	ExpireAt     time.Time
}

var permCache sync.Map // key: userID (uint), value: *CachedPermissions

const permCacheTTL = 5 * time.Minute

// GetUserPermissions 获取用户权限（缓存优先，miss 时查 DB 并写入缓存）
func GetUserPermissions(userID uint) *CachedPermissions {
	// 1. 查缓存
	if v, ok := permCache.Load(userID); ok {
		cp := v.(*CachedPermissions)
		if time.Now().Before(cp.ExpireAt) {
			return cp
		}
		permCache.Delete(userID) // 过期删除
	}

	// 2. 查 DB：user.is_super_admin + 所有 bindings（含 Role 预加载）
	var isSuperAdmin bool
	database.DB.Table("user").Select("is_super_admin").Where("id = ?", userID).Scan(&isSuperAdmin)

	var bindings []rbacmodel.PermissionBinding
	database.DB.Preload("Role").Where("user_id = ?", userID).Find(&bindings)

	// 3. 写入缓存
	cp := &CachedPermissions{
		IsSuperAdmin: isSuperAdmin,
		Bindings:     bindings,
		ExpireAt:     time.Now().Add(permCacheTTL),
	}
	permCache.Store(userID, cp)
	logger.Info(fmt.Sprintf("权限缓存已加载: userID=%d, isSuperAdmin=%v, bindings=%d",
		userID, isSuperAdmin, len(bindings)))
	return cp
}

// InvalidateUserPermissions 失效指定用户的权限缓存
func InvalidateUserPermissions(userID uint) {
	permCache.Delete(userID)
}

// InvalidateAllPermissions 失效所有用户的权限缓存（角色变更时使用）
func InvalidateAllPermissions() {
	permCache.Range(func(key, _ any) bool {
		permCache.Delete(key)
		return true
	})
}
