package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

const versionPrefix = "v1:"

var (
	ErrMissingEncryptionKey = errors.New("ENCRYPTION_KEY environment variable is not set")
	ErrInvalidEncryptionKey = errors.New("ENCRYPTION_KEY must be 32 bytes (64 hex chars or 44 base64 chars)")
	ErrInvalidCiphertext    = errors.New("invalid ciphertext format")
	ErrDecryptionFailed     = errors.New("decryption failed: invalid key or corrupted data")
)

func loadKey() ([]byte, error) {
	raw := os.Getenv("ENCRYPTION_KEY")
	if raw == "" {
		return nil, ErrMissingEncryptionKey
	}
	if key, err := hex.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	if key, err := base64.StdEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	return nil, ErrInvalidEncryptionKey
}

func Encrypt(plaintext, tenantID string) (string, error) {
	key, err := loadKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	aad := []byte(tenantID)
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), aad)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return versionPrefix + encoded, nil
}

func Decrypt(ciphertext, tenantID string) (string, error) {
	if !strings.HasPrefix(ciphertext, versionPrefix) {
		return "", ErrInvalidCiphertext
	}

	key, err := loadKey()
	if err != nil {
		return "", err
	}

	encoded := ciphertext[len(versionPrefix):]
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, sealed := data[:nonceSize], data[nonceSize:]
	aad := []byte(tenantID)
	plaintext, err := aesGCM.Open(nil, nonce, sealed, aad)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}
