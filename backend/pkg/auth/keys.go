package auth

import (
	"fmt"
	"gkube/config"
	"gkube/pkg/logger"
)

// aesKeyNew 是当前生效的加密密钥,启动时从配置加载。
var aesKeyNew []byte

// aesKeyLegacy 是可选的旧版密钥,仅用于解密历史数据。为 nil 时跳过回退。
var aesKeyLegacy []byte

// InitKeys 在启动时从配置加载并校验密钥。缺失或不合法直接 Fatal 退出。
func InitKeys() {
	raw := config.Conf.Security.AESKey
	if len(raw) != 32 {
		logger.Fatal(fmt.Sprintf("配置 security.aes_key 必须为 32 字节,当前长度:%d", len(raw)))
	}
	aesKeyNew = []byte(raw)

	// 可选旧版密钥,仅用于兼容历史已加密数据。空值或不足 32 字节则跳过。
	if legacy := config.Conf.Security.LegacyAESKey; legacy != "" {
		if len(legacy) != 32 {
			logger.Fatal(fmt.Sprintf("配置 security.legacy_aes_key 必须为 32 字节,当前长度:%d", len(legacy)))
		}
		aesKeyLegacy = []byte(legacy)
		logger.Info("已加载旧版 AES 密钥,将用于解密历史数据")
	}

	if config.Conf.Security.JWTSecret == "" {
		logger.Fatal("配置 security.jwt_secret 不能为空")
	}
}

// ActiveAESKey 返回当前生效的加密密钥。
func ActiveAESKey() []byte {
	return aesKeyNew
}

// LegacyAESKey 返回旧版解密密钥(仅用于历史数据回退)。未配置时返回 nil。
func LegacyAESKey() []byte {
	return aesKeyLegacy
}

// JWTSecret 返回 JWT 签名密钥。
func JWTSecret() []byte {
	return []byte(config.Conf.Security.JWTSecret)
}

// IsAdmin 判断用户名是否在管理员白名单内。
func IsAdmin(username string) bool {
	for _, u := range config.Conf.Security.AdminUsers {
		if u == username {
			return true
		}
	}
	return false
}
