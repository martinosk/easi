package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"

	"easi/backend/internal/auth/application/services"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"

	"easi/backend/internal/auth/infrastructure/oidc"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"
)

type TenantOIDCRepository interface {
	GetByEmailDomain(ctx context.Context, domain string) (*repositories.TenantOIDCConfig, error)
	GetByTenantID(ctx context.Context, tenantID string) (*repositories.TenantOIDCConfig, error)
}

type AuthHandlersConfig struct {
	ClientSecret   string
	RedirectURL    string
	AllowedOrigins []string
}

type AuthHandlers struct {
	sessionManager *session.SessionManager
	tenantRepo     TenantOIDCRepository
	config         AuthHandlersConfig
	loginService   *services.LoginService
}

func NewAuthHandlers(
	sessionManager *session.SessionManager,
	tenantRepo TenantOIDCRepository,
	config AuthHandlersConfig,
) *AuthHandlers {
	return &AuthHandlers{
		sessionManager: sessionManager,
		tenantRepo:     tenantRepo,
		config:         config,
	}
}

func (h *AuthHandlers) WithLoginService(loginService *services.LoginService) *AuthHandlers {
	h.loginService = loginService
	return h
}

type PostSessionsRequest struct {
	Email string `json:"email"`
}

type PostSessionsResponse struct {
	AuthorizationURL string            `json:"authorizationUrl"`
	Links            map[string]string `json:"_links"`
}

func (h *AuthHandlers) PostSessions(w http.ResponseWriter, r *http.Request) {
	var req PostSessionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	domain, err := extractEmailDomain(req.Email)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid email format")
		return
	}

	tenantConfig, err := h.tenantRepo.GetByEmailDomain(r.Context(), domain)
	if err != nil {
		if errors.Is(err, repositories.ErrDomainNotFound) || errors.Is(err, repositories.ErrTenantInactive) {
			log.Printf("Login attempt for unregistered or inactive domain: %s", domain)
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "Login failed")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Login failed")
		return
	}

	provider, err := h.createOIDCProvider(r.Context(), tenantConfig)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusServiceUnavailable, err, "IdP unavailable")
		return
	}

	tenantID, _ := sharedvo.NewTenantID(tenantConfig.TenantID)

	returnURL := h.validateReturnURL(r.Header.Get("Origin"))
	preAuth := session.NewPreAuthSession(tenantID, domain, returnURL)

	if err := h.sessionManager.StorePreAuthSession(r.Context(), preAuth); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create session")
		return
	}

	authURL := provider.AuthCodeURL(preAuth.State(), preAuth.Nonce(), preAuth.CodeVerifier())

	response := PostSessionsResponse{
		AuthorizationURL: authURL,
		Links: map[string]string{
			"authorize": authURL,
		},
	}

	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

type callbackParams struct {
	code  string
	state string
}

func (h *AuthHandlers) GetCallback(w http.ResponseWriter, r *http.Request) {
	params, err := h.extractCallbackParams(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, err.Error())
		return
	}

	preAuth, err := h.validateCallbackState(r.Context(), params.state)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, err.Error())
		return
	}

	result, err := h.exchangeCodeForTokens(r.Context(), preAuth, params.code)
	if err != nil {
		h.handleTokenExchangeError(w, err)
		return
	}

	if err := h.validateEmailDomain(result.Email, preAuth.ExpectedEmailDomain()); err != nil {
		sharedAPI.RespondError(w, http.StatusForbidden, err, "Authentication failed")
		return
	}

	if err := h.createAuthenticatedSession(r.Context(), preAuth, result); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create session")
		return
	}

	http.Redirect(w, r, h.getSafeRedirectURL(preAuth.ReturnURL()), http.StatusFound)
}

func (h *AuthHandlers) extractCallbackParams(r *http.Request) (callbackParams, error) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		errorDesc := r.URL.Query().Get("error_description")
		if errorDesc == "" {
			errorDesc = "Missing authorization code or state"
		}
		return callbackParams{}, errors.New(errorDesc)
	}
	return callbackParams{code: code, state: state}, nil
}

func (h *AuthHandlers) validateCallbackState(ctx context.Context, state string) (session.AuthSession, error) {
	preAuth, err := h.sessionManager.LoadPreAuthSession(ctx)
	if err != nil {
		return session.AuthSession{}, errors.New("Invalid session")
	}

	if subtle.ConstantTimeCompare([]byte(preAuth.State()), []byte(state)) != 1 {
		return session.AuthSession{}, errors.New("Invalid state parameter")
	}
	return preAuth, nil
}

func (h *AuthHandlers) exchangeCodeForTokens(ctx context.Context, preAuth session.AuthSession, code string) (*oidc.TokenResult, error) {
	tenantConfig, err := h.tenantRepo.GetByTenantID(ctx, preAuth.TenantID())
	if err != nil {
		return nil, fmt.Errorf("tenant config: %w", err)
	}

	provider, err := h.createOIDCProvider(ctx, tenantConfig)
	if err != nil {
		return nil, fmt.Errorf("provider: %w", err)
	}

	return provider.ExchangeCode(ctx, code, preAuth.CodeVerifier(), preAuth.Nonce())
}

func (h *AuthHandlers) handleTokenExchangeError(w http.ResponseWriter, err error) {
	log.Printf("Token exchange failed: %v", err)
	if errors.Is(err, oidc.ErrNonceMismatch) {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Token validation failed")
		return
	}
	sharedAPI.RespondError(w, http.StatusBadGateway, err, "Token exchange failed")
}

func (h *AuthHandlers) validateEmailDomain(email, expectedDomain string) error {
	authenticatedDomain, err := extractEmailDomain(email)
	if err != nil {
		log.Printf("Email domain extraction failed: %v", err)
		return errors.New("email domain mismatch")
	}
	if authenticatedDomain != expectedDomain {
		log.Printf("Email domain mismatch: expected %s, got %s", expectedDomain, authenticatedDomain)
		return errors.New("email domain mismatch")
	}
	return nil
}

func (h *AuthHandlers) createAuthenticatedSession(ctx context.Context, preAuth session.AuthSession, result *oidc.TokenResult) error {
	if err := h.sessionManager.RenewToken(ctx); err != nil {
		return err
	}

	tenantID, err := sharedvo.NewTenantID(preAuth.TenantID())
	if err != nil {
		return err
	}
	tenantCtx := sharedctx.WithTenant(ctx, tenantID)

	var userInfo session.UserInfo

	if h.loginService != nil {
		loginResult, err := h.loginService.ProcessLogin(tenantCtx, result.Email)
		if err != nil {
			if errors.Is(err, services.ErrNoValidInvitation) {
				return fmt.Errorf("no valid invitation for email %s", result.Email)
			}
			return err
		}
		userInfo = session.UserInfo{ID: loginResult.UserID, Email: loginResult.Email}
	} else {
		userInfo = session.UserInfo{ID: [16]byte{}, Email: result.Email}
	}

	authenticatedSession := preAuth.UpgradeToAuthenticated(
		userInfo,
		session.TokenInfo{AccessToken: result.AccessToken, RefreshToken: result.RefreshToken, Expiry: result.TokenExpiry},
	)

	return h.sessionManager.StoreAuthenticatedSession(ctx, authenticatedSession)
}

func (h *AuthHandlers) createOIDCProvider(ctx context.Context, tenantConfig *repositories.TenantOIDCConfig) (*oidc.OIDCProvider, error) {
	return oidc.NewOIDCProviderFromConfig(ctx, oidc.ProviderConfig{
		DiscoveryURL: tenantConfig.DiscoveryURL,
		IssuerURL:    tenantConfig.IssuerURL,
		ClientID:     tenantConfig.ClientID,
		ClientSecret: h.config.ClientSecret,
		RedirectURL:  h.config.RedirectURL,
	})
}

func extractEmailDomain(email string) (string, error) {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", errors.New("invalid email format")
	}
	parts := strings.Split(addr.Address, "@")
	if !isValidEmailParts(parts) {
		return "", errors.New("invalid email format")
	}
	return strings.ToLower(parts[1]), nil
}

func isValidEmailParts(parts []string) bool {
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func (h *AuthHandlers) validateReturnURL(origin string) string {
	if origin == "" {
		return ""
	}
	if !h.isAllowedOrigin(origin) {
		return ""
	}
	return origin + "/easi/"
}

func (h *AuthHandlers) isAllowedOrigin(targetURL string) bool {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	targetHost := strings.ToLower(parsedURL.Host)
	return h.matchesAllowedHost(targetHost)
}

func (h *AuthHandlers) matchesAllowedHost(targetHost string) bool {
	for _, allowed := range h.config.AllowedOrigins {
		allowedURL, err := url.Parse(allowed)
		if err != nil {
			continue
		}
		if strings.ToLower(allowedURL.Host) == targetHost {
			return true
		}
	}
	return false
}

func (h *AuthHandlers) getSafeRedirectURL(returnURL string) string {
	if returnURL != "" && h.isAllowedOrigin(returnURL) {
		return returnURL
	}
	return h.getDefaultRedirectURL()
}

func (h *AuthHandlers) getDefaultRedirectURL() string {
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		return frontendURL
	}
	return "/easi/"
}
