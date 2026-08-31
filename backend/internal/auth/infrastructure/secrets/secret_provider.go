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

func (p *FileSecretProvider) readSecretFile(tenantID, filename string) ([]byte, error) {
	path := filepath.Join(p.basePath, tenantID, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSecretNotFound
		}
		return nil, err
	}
	return data, nil
}

func (p *FileSecretProvider) GetClientSecret(_ context.Context, tenantID string) (string, error) {
	data, err := p.readSecretFile(tenantID, "client-secret")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (p *FileSecretProvider) GetPrivateKey(_ context.Context, tenantID string) ([]byte, error) {
	return p.readSecretFile(tenantID, "private-key")
}

func (p *FileSecretProvider) GetCertificate(_ context.Context, tenantID string) ([]byte, error) {
	return p.readSecretFile(tenantID, "certificate")
}

func (p *FileSecretProvider) IsProvisioned(_ context.Context, tenantID string) bool {
	for _, filename := range []string{"client-secret", "private-key"} {
		if _, err := os.Stat(filepath.Join(p.basePath, tenantID, filename)); err == nil {
			return true
		}
	}
	return false
}

type EnvSecretProvider struct {
	envVar string
}

func NewEnvSecretProvider(envVar string) *EnvSecretProvider {
	return &EnvSecretProvider{envVar: envVar}
}

func (p *EnvSecretProvider) GetClientSecret(_ context.Context, _ string) (string, error) {
	secret := os.Getenv(p.envVar)
	if secret == "" {
		return "", ErrSecretNotFound
	}
	return secret, nil
}

func (p *EnvSecretProvider) GetPrivateKey(_ context.Context, _ string) ([]byte, error) {
	return nil, ErrSecretNotFound
}

func (p *EnvSecretProvider) GetCertificate(_ context.Context, _ string) ([]byte, error) {
	return nil, ErrSecretNotFound
}

func (p *EnvSecretProvider) IsProvisioned(_ context.Context, _ string) bool {
	return os.Getenv(p.envVar) != ""
}
