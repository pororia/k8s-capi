package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// AES는 AES-256-GCM 암호화/복호화를 제공합니다.
type AES struct {
	key []byte
}

// NewAES는 base64 인코딩된 32바이트 키로 AES 인스턴스를 생성합니다.
func NewAES(base64Key string) (*AES, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode AES key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("AES key must be 32 bytes, got %d", len(key))
	}
	return &AES{key: key}, nil
}

// Encrypt는 평문을 AES-256-GCM으로 암호화하고 base64 인코딩된 문자열을 반환합니다.
func (a *AES) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt는 base64 인코딩된 암호문을 복호화하여 평문을 반환합니다.
func (a *AES) Decrypt(encoded string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
