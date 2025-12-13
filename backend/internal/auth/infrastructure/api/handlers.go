package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
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
}

func NewAuthHandlers(
	sessionManager *session.SessionManager,
	tenantRepo TenantOIDCRepository,
	clientSecret string,
	redirectURL string,
) *AuthHandlers {
	return &AuthHandlers{
		sessionManager: sessionManager,
		tenantRepo:     tenantRepo,
		clientSecret:   clientSecret,
		redirectURL:    redirectURL,
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
		if errors.Is(err, repositories.ErrDomainNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Domain not registered")
			return
		}
		if errors.Is(err, repositories.ErrTenantInactive) {
			sharedAPI.RespondError(w, http.StatusForbidden, err, "Tenant is inactive")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to lookup tenant")
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

	// Capture the origin to redirect back after auth
	returnURL := r.Header.Get("Origin")
	if returnURL != "" {
		returnURL = returnURL + "/easi/"
	}
	preAuth := session.NewPreAuthSession(tenantID, returnURL)

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

	if preAuth.State() != state {
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

	if err := h.sessionManager.RenewToken(r.Context()); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create session")
		return
	}

	authenticatedSession := preAuth.UpgradeToAuthenticated(
		uuid.Nil,
		result.AccessToken,
		result.RefreshToken,
		result.TokenExpiry,
	)

	if err := h.sessionManager.StoreAuthenticatedSession(r.Context(), authenticatedSession); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to store session")
		return
	}

	// Redirect to the URL captured during login initiation, or fallback to env/default
	redirectURL := preAuth.ReturnURL()
	if redirectURL == "" {
		redirectURL = os.Getenv("FRONTEND_URL")
	}
	if redirectURL == "" {
		redirectURL = "/easi/"
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func extractEmailDomain(email string) (string, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", errors.New("invalid email format")
	}
	return strings.ToLower(parts[1]), nil
}
