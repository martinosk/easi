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
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

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
	Links map[string]string `json:"_links"`
}

// PostSessions godoc
// @Summary Initiate OIDC login
// @Description Initiates the OIDC authentication flow by resolving the tenant from the email domain and returning an authorization URL
// @Tags auth
// @Accept json
// @Produce json
// @Param request body PostSessionsRequest true "Login request with email"
// @Success 200 {object} PostSessionsResponse "Authorization URL for OIDC login"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid email format or unregistered domain"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Failure 503 {object} sharedAPI.ErrorResponse "Identity provider unavailable"
// @Router /auth/sessions [post]
func (h *AuthHandlers) PostSessions(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseLoginRequest(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	domain, tenantConfig, err := h.resolveEmailDomain(r.Context(), req.Email)
	if err != nil {
		h.handleDomainResolutionError(w, err)
		return
	}

	authURL, err := h.initiateOIDCFlow(r, domain, tenantConfig)
	if err != nil {
		h.handleOIDCFlowError(w, err)
		return
	}

	sharedAPI.RespondJSON(w, http.StatusOK, PostSessionsResponse{
		Links: map[string]string{"self": "/api/v1/auth/sessions", "authorize": authURL},
	})
}

func (h *AuthHandlers) parseLoginRequest(r *http.Request) (*PostSessionsRequest, error) {
	var req PostSessionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (h *AuthHandlers) resolveEmailDomain(ctx context.Context, email string) (string, *repositories.TenantOIDCConfig, error) {
	domain, err := extractEmailDomain(email)
	if err != nil {
		log.Printf("Invalid email format in login attempt: %v", err)
		return "", nil, fmt.Errorf("invalid email: %w", err)
	}

	tenantConfig, err := h.tenantRepo.GetByEmailDomain(ctx, domain)
	if err != nil {
		h.logDomainLookupError(domain, err)
		return "", nil, fmt.Errorf("domain lookup failed: %w", err)
	}

	return domain, tenantConfig, nil
}

func (h *AuthHandlers) logDomainLookupError(domain string, err error) {
	if errors.Is(err, repositories.ErrDomainNotFound) || errors.Is(err, repositories.ErrTenantInactive) {
		log.Printf("Login attempt for unregistered or inactive domain: %s", domain)
	} else {
		log.Printf("Unexpected error during domain lookup: %v", err)
	}
}

func (h *AuthHandlers) handleDomainResolutionError(w http.ResponseWriter, err error) {
	if errors.Is(err, repositories.ErrDomainNotFound) || errors.Is(err, repositories.ErrTenantInactive) {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Unable to process login request")
		return
	}
	sharedAPI.RespondError(w, http.StatusBadRequest, nil, "Unable to process login request")
}

type oidcFlowError struct {
	statusCode int
	message    string
	err        error
}

func (e *oidcFlowError) Error() string {
	return e.message
}

func (h *AuthHandlers) initiateOIDCFlow(r *http.Request, domain string, tenantConfig *repositories.TenantOIDCConfig) (string, error) {
	provider, err := h.createOIDCProvider(r.Context(), tenantConfig)
	if err != nil {
		return "", &oidcFlowError{http.StatusServiceUnavailable, "IdP unavailable", err}
	}

	tenantID, _ := sharedvo.NewTenantID(tenantConfig.TenantID)
	returnURL := h.validateReturnURL(r.Header.Get("Origin"))
	preAuth := session.NewPreAuthSession(tenantID, domain, returnURL)

	if err := h.sessionManager.StorePreAuthSession(r.Context(), preAuth); err != nil {
		return "", &oidcFlowError{http.StatusInternalServerError, "Failed to create session", err}
	}

	return provider.AuthCodeURL(preAuth.State(), preAuth.Nonce(), preAuth.CodeVerifier()), nil
}

func (h *AuthHandlers) handleOIDCFlowError(w http.ResponseWriter, err error) {
	if flowErr, ok := err.(*oidcFlowError); ok {
		sharedAPI.RespondError(w, flowErr.statusCode, flowErr.err, flowErr.message)
		return
	}
	sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to initiate login")
}

type callbackParams struct {
	code  string
	state string
}

// GetCallback godoc
// @Summary OIDC callback handler
// @Description Handles the OIDC callback after user authentication, exchanges the authorization code for tokens, and creates an authenticated session
// @Tags auth
// @Produce json
// @Param code query string true "Authorization code from OIDC provider"
// @Param state query string true "State parameter for CSRF protection"
// @Success 302 "Redirect to frontend application"
// @Failure 400 {object} sharedAPI.ErrorResponse "Missing code/state or invalid session"
// @Failure 401 {object} sharedAPI.ErrorResponse "Token validation failed"
// @Failure 403 {object} sharedAPI.ErrorResponse "Email domain mismatch"
// @Failure 500 {object} sharedAPI.ErrorResponse "Failed to create session"
// @Failure 502 {object} sharedAPI.ErrorResponse "Token exchange with IdP failed"
// @Router /auth/callback [get]
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
		log.Printf("[AUTH] createAuthenticatedSession failed for email=%s tenant=%s: %v", result.Email, preAuth.TenantID(), err)
		if strings.Contains(err.Error(), "no valid invitation") {
			redirectURL := h.buildErrorRedirectURL(preAuth.ReturnURL(), "no_invitation", "You need an invitation to access this application. Please contact your administrator.")
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
		if errors.Is(err, services.ErrUserDisabled) {
			redirectURL := h.buildErrorRedirectURL(preAuth.ReturnURL(), "account_disabled", "Your account has been disabled. Please contact your administrator.")
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
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
		loginResult, err := h.loginService.ProcessLogin(tenantCtx, result.Email, result.Name)
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

func (h *AuthHandlers) buildErrorRedirectURL(returnURL, errorCode, errorMessage string) string {
	baseURL := h.getSafeRedirectURL(returnURL)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := parsedURL.Query()
	q.Set("auth_error", errorCode)
	q.Set("auth_error_message", errorMessage)
	parsedURL.RawQuery = q.Encode()

	return parsedURL.String()
}
