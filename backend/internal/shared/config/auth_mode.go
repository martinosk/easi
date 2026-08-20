package config

import (
	"fmt"
	"log"
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

func resolveAuthMode(raw string) (AuthMode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "local_oidc":
		return AuthModeLocalOIDC, nil
	case "bypass":
		if !bypassBuildEnabled {
			return "", fmt.Errorf("AUTH_MODE=bypass requires a binary built with -tags devauth")
		}
		return AuthModeBypass, nil
	default:
		return AuthModeProduction, nil
	}
}

func GetAuthMode() AuthMode {
	authModeOnce.Do(func() {
		mode, err := resolveAuthMode(os.Getenv("AUTH_MODE"))
		if err != nil {
			log.Fatalf("refusing to start: %v", err)
		}
		currentAuthMode = mode
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
