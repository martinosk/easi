package config

import (
	"log"
	"net/netip"
	"os"
	"strconv"
	"strings"
)

type TrustedProxyConfig struct {
	CIDRs []string
	Count int
}

func GetTrustedProxyConfig() TrustedProxyConfig {
	return TrustedProxyConfig{
		CIDRs: parseTrustedProxyCIDRs(os.Getenv("TRUSTED_PROXY_CIDRS")),
		Count: parseTrustedProxyCount(os.Getenv("TRUSTED_PROXY_COUNT")),
	}
}

func parseTrustedProxyCIDRs(raw string) []string {
	var valid []string
	for entry := range strings.SplitSeq(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, err := netip.ParsePrefix(entry); err != nil {
			log.Printf("[CONFIG] Ignoring invalid TRUSTED_PROXY_CIDRS entry %q: %v", entry, err)
			continue
		}
		valid = append(valid, entry)
	}
	return valid
}

func parseTrustedProxyCount(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	count, err := strconv.Atoi(raw)
	if err != nil || count < 0 {
		log.Printf("[CONFIG] Ignoring invalid TRUSTED_PROXY_COUNT %q", raw)
		return 0
	}
	return count
}
