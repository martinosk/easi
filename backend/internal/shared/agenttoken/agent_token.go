package agenttoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

const DefaultTTL = 5 * time.Minute

var (
	ErrMissingSecret = errors.New("AGENT_TOKEN_SECRET environment variable is not set")
	ErrInvalidToken  = errors.New("invalid agent token format")
	ErrTokenExpired  = errors.New("agent token has expired")
	ErrInvalidSig    = errors.New("agent token signature is invalid")
)

type AgentClaims struct {
	UserID   string `json:"userId"`
	TenantID string `json:"tenantId"`
	Source   string `json:"source"`
	Exp      int64  `json:"exp"`
}

func loadSecret() ([]byte, error) {
	secret := os.Getenv("AGENT_TOKEN_SECRET")
	if secret == "" {
		return nil, ErrMissingSecret
	}
	return []byte(secret), nil
}

func Mint(userID, tenantID string, ttl time.Duration) (string, error) {
	secret, err := loadSecret()
	if err != nil {
		return "", err
	}

	claims := AgentClaims{
		UserID:   userID,
		TenantID: tenantID,
		Source:   "agent",
		Exp:      time.Now().Add(ttl).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(payloadB64, secret)
	return payloadB64 + "." + sig, nil
}

func Verify(token string) (*AgentClaims, error) {
	secret, err := loadSecret()
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	payloadB64, sigB64 := parts[0], parts[1]

	expectedSig := sign(payloadB64, secret)
	if !hmac.Equal([]byte(sigB64), []byte(expectedSig)) {
		return nil, ErrInvalidSig
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims AgentClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > claims.Exp {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

func sign(data string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
