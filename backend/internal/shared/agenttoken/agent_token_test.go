package agenttoken

import (
	"testing"
	"time"
)

func setTestSecret(t *testing.T) {
	t.Helper()
	t.Setenv("AGENT_TOKEN_SECRET", "test-secret-key-for-agent-tokens")
}

func TestMintAndVerify(t *testing.T) {
	setTestSecret(t)

	token, err := Mint("user-123", "tenant-abc", DefaultTTL)
	if err != nil {
		t.Fatalf("Mint failed: %v", err)
	}

	claims, err := Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("expected userId user-123, got %s", claims.UserID)
	}
	if claims.TenantID != "tenant-abc" {
		t.Errorf("expected tenantId tenant-abc, got %s", claims.TenantID)
	}
	if claims.Source != "agent" {
		t.Errorf("expected source agent, got %s", claims.Source)
	}
}

func TestExpiredToken(t *testing.T) {
	setTestSecret(t)

	token, err := Mint("user-123", "tenant-abc", -1*time.Minute)
	if err != nil {
		t.Fatalf("Mint failed: %v", err)
	}

	_, err = Verify(token)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestTamperedSignature(t *testing.T) {
	setTestSecret(t)

	token, err := Mint("user-123", "tenant-abc", DefaultTTL)
	if err != nil {
		t.Fatalf("Mint failed: %v", err)
	}

	tampered := token + "x"
	_, err = Verify(tampered)
	if err != ErrInvalidSig {
		t.Errorf("expected ErrInvalidSig, got %v", err)
	}
}

func TestInvalidFormat(t *testing.T) {
	setTestSecret(t)

	_, err := Verify("no-dot-separator")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestTokenValidAtExactExpiry(t *testing.T) {
	setTestSecret(t)

	token, err := Mint("user-123", "tenant-abc", 0)
	if err != nil {
		t.Fatalf("Mint failed: %v", err)
	}

	claims, err := Verify(token)
	if err != nil {
		t.Fatalf("token should be valid at exact expiry second, got %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected userId user-123, got %s", claims.UserID)
	}
}

func TestMissingSecret(t *testing.T) {
	t.Setenv("AGENT_TOKEN_SECRET", "")

	_, err := Mint("user", "tenant", DefaultTTL)
	if err != ErrMissingSecret {
		t.Errorf("expected ErrMissingSecret, got %v", err)
	}
}
