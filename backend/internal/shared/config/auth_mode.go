package config

import (
	"os"
	"strings"
	"sync"
)

type AuthMode string

const (
	AuthModeProduction AuthMode = "production"
	AuthModeLocalOIDC  AuthMode = "local_oidc"
	AuthModeBypass     AuthMode = "bypass"
)

var (
	currentAuthMode AuthMode
	authModeOnce    sync.Once
)

func GetAuthMode() AuthMode {
	authModeOnce.Do(func() {
		mode := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_MODE")))
		switch mode {
		case "local_oidc":
			currentAuthMode = AuthModeLocalOIDC
		case "bypass":
			currentAuthMode = AuthModeBypass
		default:
			currentAuthMode = AuthModeProduction
		}
	})
	return currentAuthMode
}

func IsHTTPAllowed() bool {
	mode := GetAuthMode()
	return mode == AuthModeLocalOIDC || mode == AuthModeBypass
}

func IsAuthBypassed() bool {
	return GetAuthMode() == AuthModeBypass
}

func IsProduction() bool {
	return GetAuthMode() == AuthModeProduction
}
