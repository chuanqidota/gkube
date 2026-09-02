package auth

import (
	"fmt"
	"sync"
	"time"

	authmodel "gkube/internal/auth/model"
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

// permCache 为进程内缓存。注意：多副本部署时 InvalidateUserPermissions 仅对本进程
// 生效，其他副本最长 5 分钟（TTL）后自然过期。横向扩容前需迁移到 Redis pub/sub 失效。
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

	// 2. 查 DB：先校验用户可用（status=1 且未软删除，GORM DeletedAt 自动过滤），
	//    不可用用户视为无任何权限
	var user authmodel.User
	if err := database.DB.Select("id", "is_super_admin").
		Where("id = ? AND status = 1", userID).
		First(&user).Error; err != nil {
		// 用户不存在/被禁用/被删除：返回空权限，短 TTL（1 分钟）以便恢复后快速生效
		cp := &CachedPermissions{
			IsSuperAdmin: false,
			Bindings:     nil,
			ExpireAt:     time.Now().Add(time.Minute),
		}
		permCache.Store(userID, cp)
		return cp
	}

	var bindings []rbacmodel.PermissionBinding
	database.DB.Preload("Role").Where("user_id = ?", userID).Find(&bindings)

	// 3. 写入缓存
	cp := &CachedPermissions{
		IsSuperAdmin: user.IsSuperAdmin,
		Bindings:     bindings,
		ExpireAt:     time.Now().Add(permCacheTTL),
	}
	permCache.Store(userID, cp)
	logger.Info(fmt.Sprintf("权限缓存已加载: userID=%d, isSuperAdmin=%v, bindings=%d",
		userID, user.IsSuperAdmin, len(bindings)))
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

// SetPermissionsForTest 测试辅助：直接写入权限缓存（仅测试用）。
func SetPermissionsForTest(userID uint, cp *CachedPermissions) {
	permCache.Store(userID, cp)
}
