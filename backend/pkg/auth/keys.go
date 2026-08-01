package auth

import (
	"fmt"
	"gkube/config"
	"gkube/pkg/logger"
)

// legacyAESKey 是历史版本硬编码的 AES-256-GCM 密钥,仅用于解密旧数据时的兜底回退。
// 新数据一律用 config.Conf.Security.AESKey 加密。
const legacyAESKey = "gkube-aes-256-gcm-secret-key-32b" // exactly 32 bytes

// aesKeyNew 是当前生效的加密密钥,启动时从配置加载。
var aesKeyNew []byte

// InitKeys 在启动时从配置加载并校验密钥。缺失或不合法直接 Fatal 退出。
func InitKeys() {
	raw := config.Conf.Security.AESKey
	if len(raw) != 32 {
		logger.Fatal(fmt.Sprintf("配置 security.aes_key 必须为 32 字节,当前长度:%d", len(raw)))
	}
	aesKeyNew = []byte(raw)

	if config.Conf.Security.JWTSecret == "" {
		logger.Fatal("配置 security.jwt_secret 不能为空")
	}
}

// ActiveAESKey 返回当前生效的加密密钥。
func ActiveAESKey() []byte {
	return aesKeyNew
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
