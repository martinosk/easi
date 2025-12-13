package secrets

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrSecretNotFound = errors.New("secret not found")

type SecretProvider interface {
	GetClientSecret(ctx context.Context, tenantID string) (string, error)
	GetPrivateKey(ctx context.Context, tenantID string) ([]byte, error)
	GetCertificate(ctx context.Context, tenantID string) ([]byte, error)
	IsProvisioned(ctx context.Context, tenantID string) bool
}

type FileSecretProvider struct {
	basePath string
}

func NewFileSecretProvider(basePath string) *FileSecretProvider {
	return &FileSecretProvider{basePath: basePath}
}

func (p *FileSecretProvider) GetClientSecret(ctx context.Context, tenantID string) (string, error) {
	path := filepath.Join(p.basePath, tenantID, "client-secret")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrSecretNotFound
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (p *FileSecretProvider) GetPrivateKey(ctx context.Context, tenantID string) ([]byte, error) {
	path := filepath.Join(p.basePath, tenantID, "private-key")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSecretNotFound
		}
		return nil, err
	}
	return data, nil
}

func (p *FileSecretProvider) GetCertificate(ctx context.Context, tenantID string) ([]byte, error) {
	path := filepath.Join(p.basePath, tenantID, "certificate")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSecretNotFound
		}
		return nil, err
	}
	return data, nil
}

func (p *FileSecretProvider) IsProvisioned(ctx context.Context, tenantID string) bool {
	clientSecretPath := filepath.Join(p.basePath, tenantID, "client-secret")
	privateKeyPath := filepath.Join(p.basePath, tenantID, "private-key")

	if _, err := os.Stat(clientSecretPath); err == nil {
		return true
	}
	if _, err := os.Stat(privateKeyPath); err == nil {
		return true
	}
	return false
}
