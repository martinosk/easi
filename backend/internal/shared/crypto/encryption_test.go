package crypto

import (
	"encoding/hex"
	"os"
	"testing"
)

func setTestKey(t *testing.T) {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	t.Setenv("ENCRYPTION_KEY", hex.EncodeToString(key))
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	setTestKey(t)

	plaintext := "sk-secret-api-key-12345"
	tenantID := "tenant-abc"

	encrypted, err := Encrypt(plaintext, tenantID)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if encrypted[:3] != "v1:" {
		t.Errorf("expected v1: prefix, got %s", encrypted[:3])
	}

	decrypted, err := Decrypt(encrypted, tenantID)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecryptWrongTenantFails(t *testing.T) {
	setTestKey(t)

	encrypted, err := Encrypt("secret", "tenant-a")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(encrypted, "tenant-b")
	if err == nil {
		t.Fatal("expected decryption to fail with wrong tenant ID")
	}
}

func TestDecryptInvalidPrefix(t *testing.T) {
	setTestKey(t)

	_, err := Decrypt("invalid-ciphertext", "tenant-a")
	if err != ErrInvalidCiphertext {
		t.Errorf("expected ErrInvalidCiphertext, got %v", err)
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	setTestKey(t)

	enc1, _ := Encrypt("same-plaintext", "tenant-a")
	enc2, _ := Encrypt("same-plaintext", "tenant-a")

	if enc1 == enc2 {
		t.Error("expected different ciphertexts due to random nonce")
	}
}

func TestMissingEncryptionKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	os.Unsetenv("ENCRYPTION_KEY")

	_, err := Encrypt("test", "tenant")
	if err != ErrMissingEncryptionKey {
		t.Errorf("expected ErrMissingEncryptionKey, got %v", err)
	}
}

func TestInvalidEncryptionKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "too-short")

	_, err := Encrypt("test", "tenant")
	if err != ErrInvalidEncryptionKey {
		t.Errorf("expected ErrInvalidEncryptionKey, got %v", err)
	}
}
