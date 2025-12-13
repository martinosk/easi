package valueobjects

import (
	"easi/backend/internal/shared/domain"
	"errors"
	"net/url"
	"strings"
)

var (
	ErrOIDCDiscoveryURLEmpty    = errors.New("OIDC discovery URL cannot be empty")
	ErrOIDCDiscoveryURLInvalid  = errors.New("OIDC discovery URL is not a valid URL")
	ErrOIDCDiscoveryURLNotHTTPS = errors.New("OIDC discovery URL must use HTTPS")
	ErrOIDCClientIDEmpty        = errors.New("OIDC client ID cannot be empty")
	ErrOIDCAuthMethodInvalid    = errors.New("OIDC auth method must be 'client_secret' or 'private_key_jwt'")
)

const defaultScopes = "openid email profile"

type OIDCAuthMethod string

const (
	OIDCAuthMethodClientSecret  OIDCAuthMethod = "client_secret"
	OIDCAuthMethodPrivateKeyJWT OIDCAuthMethod = "private_key_jwt"
)

func (m OIDCAuthMethod) IsValid() bool {
	return m == OIDCAuthMethodClientSecret || m == OIDCAuthMethodPrivateKeyJWT
}

type OIDCConfig struct {
	discoveryURL string
	clientID     string
	authMethod   OIDCAuthMethod
	scopes       string
}

func NewOIDCConfig(discoveryURL, clientID string, authMethod OIDCAuthMethod, scopes string) (OIDCConfig, error) {
	validatedURL, err := validateDiscoveryURL(discoveryURL)
	if err != nil {
		return OIDCConfig{}, err
	}

	validatedClientID, err := validateClientID(clientID)
	if err != nil {
		return OIDCConfig{}, err
	}

	if !authMethod.IsValid() {
		return OIDCConfig{}, ErrOIDCAuthMethodInvalid
	}

	return OIDCConfig{
		discoveryURL: validatedURL,
		clientID:     validatedClientID,
		authMethod:   authMethod,
		scopes:       normalizeScopes(scopes),
	}, nil
}

func validateDiscoveryURL(discoveryURL string) (string, error) {
	discoveryURL = strings.TrimSpace(discoveryURL)
	if discoveryURL == "" {
		return "", ErrOIDCDiscoveryURLEmpty
	}

	parsedURL, err := url.Parse(discoveryURL)
	if err != nil || parsedURL.Host == "" {
		return "", ErrOIDCDiscoveryURLInvalid
	}

	if parsedURL.Scheme != "https" && !isLocalhost(parsedURL.Host) {
		return "", ErrOIDCDiscoveryURLNotHTTPS
	}

	return discoveryURL, nil
}

func validateClientID(clientID string) (string, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "", ErrOIDCClientIDEmpty
	}
	return clientID, nil
}

func normalizeScopes(scopes string) string {
	scopes = strings.TrimSpace(scopes)
	if scopes == "" {
		return defaultScopes
	}
	return scopes
}

func (c OIDCConfig) DiscoveryURL() string {
	return c.discoveryURL
}

func (c OIDCConfig) ClientID() string {
	return c.clientID
}

func (c OIDCConfig) AuthMethod() OIDCAuthMethod {
	return c.authMethod
}

func (c OIDCConfig) Scopes() string {
	return c.scopes
}

func (c OIDCConfig) Equals(other domain.ValueObject) bool {
	if otherConfig, ok := other.(OIDCConfig); ok {
		return c.discoveryURL == otherConfig.discoveryURL &&
			c.clientID == otherConfig.clientID &&
			c.authMethod == otherConfig.authMethod &&
			c.scopes == otherConfig.scopes
	}
	return false
}

func isLocalhost(host string) bool {
	host = strings.Split(host, ":")[0]
	return host == "localhost" || host == "127.0.0.1"
}
