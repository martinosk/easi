package middleware

import (
	"net"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"easi/backend/internal/shared/config"
)

func ClientIP() func(http.Handler) http.Handler {
	cfg := config.GetTrustedProxyConfig()

	if len(cfg.CIDRs) > 0 {
		return chimiddleware.ClientIPFromXFF(cfg.CIDRs...)
	}
	if cfg.Count > 0 {
		return chimiddleware.ClientIPFromXFFTrustedProxies(cfg.Count)
	}
	return chimiddleware.ClientIPFromRemoteAddr
}

func getClientIP(r *http.Request) string {
	if ip := chimiddleware.GetClientIP(r.Context()); ip != "" {
		return ip
	}
	return remoteAddrHost(r)
}

func remoteAddrHost(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
