package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"

	sharedAPI "easi/backend/internal/shared/api"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"

	"easi/backend/internal/auth/infrastructure/oidc"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"
)

type TenantOIDCRepository interface {
	GetByEmailDomain(ctx context.Context, domain string) (*repositories.TenantOIDCConfig, error)
	GetByTenantID(ctx context.Context, tenantID string) (*repositories.TenantOIDCConfig, error)
}

type AuthHandlers struct {
	sessionManager *session.SessionManager
	tenantRepo     TenantOIDCRepository
	clientSecret   string
	redirectURL    string
	allowedOrigins []string
}

func NewAuthHandlers(
	sessionManager *session.SessionManager,
	tenantRepo TenantOIDCRepository,
	clientSecret string,
	redirectURL string,
	allowedOrigins []string,
) *AuthHandlers {
	return &AuthHandlers{
		sessionManager: sessionManager,
		tenantRepo:     tenantRepo,
		clientSecret:   clientSecret,
		redirectURL:    redirectURL,
		allowedOrigins: allowedOrigins,
	}
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

	provider, err := oidc.NewOIDCProviderWithIssuer(
		r.Context(),
		tenantConfig.DiscoveryURL,
		tenantConfig.IssuerURL,
		tenantConfig.ClientID,
		h.clientSecret,
		h.redirectURL,
	)
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

func (h *AuthHandlers) GetCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		errorDesc := r.URL.Query().Get("error_description")
		if errorDesc == "" {
			errorDesc = "Missing authorization code or state"
		}
		sharedAPI.RespondError(w, http.StatusBadRequest, errors.New(errorDesc), errorDesc)
		return
	}

	preAuth, err := h.sessionManager.LoadPreAuthSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid session")
		return
	}

	if subtle.ConstantTimeCompare([]byte(preAuth.State()), []byte(state)) != 1 {
		sharedAPI.RespondError(w, http.StatusBadRequest, errors.New("state mismatch"), "Invalid state parameter")
		return
	}

	tenantConfig, err := h.tenantRepo.GetByTenantID(r.Context(), preAuth.TenantID())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to load tenant configuration")
		return
	}

	provider, err := oidc.NewOIDCProviderWithIssuer(
		r.Context(),
		tenantConfig.DiscoveryURL,
		tenantConfig.IssuerURL,
		tenantConfig.ClientID,
		h.clientSecret,
		h.redirectURL,
	)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadGateway, err, "IdP unavailable")
		return
	}

	result, err := provider.ExchangeCode(r.Context(), code, preAuth.CodeVerifier(), preAuth.Nonce())
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		if errors.Is(err, oidc.ErrNonceMismatch) {
			sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Token validation failed")
			return
		}
		sharedAPI.RespondError(w, http.StatusBadGateway, err, "Token exchange failed")
		return
	}

	authenticatedEmailDomain, err := extractEmailDomain(result.Email)
	if err != nil || authenticatedEmailDomain != preAuth.ExpectedEmailDomain() {
		log.Printf("Email domain mismatch: expected %s, got %s", preAuth.ExpectedEmailDomain(), authenticatedEmailDomain)
		sharedAPI.RespondError(w, http.StatusForbidden, errors.New("email domain mismatch"), "Authentication failed")
		return
	}

	if err := h.sessionManager.RenewToken(r.Context()); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create session")
		return
	}

	authenticatedSession := preAuth.UpgradeToAuthenticated(
		uuid.Nil,
		result.Email,
		result.AccessToken,
		result.RefreshToken,
		result.TokenExpiry,
	)

	if err := h.sessionManager.StoreAuthenticatedSession(r.Context(), authenticatedSession); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to store session")
		return
	}

	redirectURL := h.getSafeRedirectURL(preAuth.ReturnURL())
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func extractEmailDomain(email string) (string, error) {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", errors.New("invalid email format")
	}
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", errors.New("invalid email format")
	}
	return strings.ToLower(parts[1]), nil
}

func (h *AuthHandlers) validateReturnURL(origin string) string {
	if origin == "" {
		return ""
	}

	parsedURL, err := url.Parse(origin)
	if err != nil {
		return ""
	}

	originHost := strings.ToLower(parsedURL.Host)
	for _, allowed := range h.allowedOrigins {
		allowedURL, err := url.Parse(allowed)
		if err != nil {
			continue
		}
		if strings.ToLower(allowedURL.Host) == originHost {
			return origin + "/easi/"
		}
	}

	return ""
}

func (h *AuthHandlers) getSafeRedirectURL(returnURL string) string {
	if returnURL != "" {
		parsedURL, err := url.Parse(returnURL)
		if err == nil {
			returnHost := strings.ToLower(parsedURL.Host)
			for _, allowed := range h.allowedOrigins {
				allowedURL, err := url.Parse(allowed)
				if err != nil {
					continue
				}
				if strings.ToLower(allowedURL.Host) == returnHost {
					return returnURL
				}
			}
		}
	}

	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		return frontendURL
	}

	return "/easi/"
}
