package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptAES encrypts plaintext using AES-256-GCM with the active config key
// and returns a base64-encoded string containing the nonce prepended to the ciphertext.
func EncryptAES(plaintext string) (string, error) {
	block, err := aes.NewCipher(ActiveAESKey())
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

// decryptWithKey 尝试用指定密钥解密。
func decryptWithKey(encoded string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// DecryptAES decodes a base64-encoded string produced by EncryptAES and
// decrypts it using AES-256-GCM.优先用当前配置密钥,若配置了旧版密钥则回退尝试,
// 以兼容历史已加密的 kubeconfig,不破坏现有数据。
func DecryptAES(encoded string) (string, error) {
	if plaintext, err := decryptWithKey(encoded, ActiveAESKey()); err == nil {
		return plaintext, nil
	}
	// 回退到旧密钥(仅用于历史数据),未配置则直接返回错误
	if legacy := LegacyAESKey(); legacy != nil {
		return decryptWithKey(encoded, legacy)
	}
	return "", errors.New("decryption failed with active key and no legacy key configured")
}
