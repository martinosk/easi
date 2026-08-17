package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTrustedProxyConfig_DefaultsToNoTrustedProxies(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "")
	t.Setenv("TRUSTED_PROXY_COUNT", "")

	cfg := GetTrustedProxyConfig()

	assert.Empty(t, cfg.CIDRs)
	assert.Zero(t, cfg.Count)
}

func TestGetTrustedProxyConfig_ParsesCIDRList(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8, 2600:9000::/28")
	t.Setenv("TRUSTED_PROXY_COUNT", "")

	cfg := GetTrustedProxyConfig()

	assert.Equal(t, []string{"10.0.0.0/8", "2600:9000::/28"}, cfg.CIDRs)
}

func TestGetTrustedProxyConfig_DropsInvalidCIDRs(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8,not-a-cidr,192.168.1.1")
	t.Setenv("TRUSTED_PROXY_COUNT", "")

	cfg := GetTrustedProxyConfig()

	assert.Equal(t, []string{"10.0.0.0/8"}, cfg.CIDRs)
}

func TestGetTrustedProxyConfig_ParsesCount(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "")
	t.Setenv("TRUSTED_PROXY_COUNT", "1")

	cfg := GetTrustedProxyConfig()

	assert.Equal(t, 1, cfg.Count)
}

func TestGetTrustedProxyConfig_RejectsNonNumericCount(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "")
	t.Setenv("TRUSTED_PROXY_COUNT", "many")

	cfg := GetTrustedProxyConfig()

	assert.Zero(t, cfg.Count)
}

func TestGetTrustedProxyConfig_RejectsNegativeCount(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "")
	t.Setenv("TRUSTED_PROXY_COUNT", "-2")

	cfg := GetTrustedProxyConfig()

	assert.Zero(t, cfg.Count)
}
