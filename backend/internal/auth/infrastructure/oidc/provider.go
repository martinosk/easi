package oidc

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"easi/backend/internal/shared/config"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	ErrNonceMismatch = errors.New("nonce mismatch")
	ErrInvalidToken  = errors.New("invalid token")
)

type urlRewriteTransport struct {
	base      http.RoundTripper
	fromURL   string
	toURL     string
}

func (t *urlRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqURL := req.URL.String()
	if strings.HasPrefix(reqURL, t.fromURL) {
		newURL := strings.Replace(reqURL, t.fromURL, t.toURL, 1)
		newReq, err := http.NewRequestWithContext(req.Context(), req.Method, newURL, req.Body)
		if err != nil {
			return nil, err
		}
		newReq.Header = req.Header
		req = newReq
	}
	return t.base.RoundTrip(req)
}

type TokenResult struct {
	AccessToken  string
	RefreshToken string
	TokenExpiry  time.Time
	Subject      string
	Email        string
	Name         string
}

type OIDCProvider struct {
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	httpClient   *http.Client
}

type ProviderConfig struct {
	DiscoveryURL string
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

var defaultScopes = []string{oidc.ScopeOpenID, "email", "profile", "offline_access"}

var ErrInsecureIssuerNotAllowed = errors.New("insecure issuer URL override requires AUTH_MODE=local_oidc or AUTH_MODE=bypass")

func NewOIDCProviderFromConfig(ctx context.Context, cfg ProviderConfig) (*OIDCProvider, error) {
	httpClient, ctx, err := setupHTTPClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(ctx, cfg.DiscoveryURL)
	if err != nil {
		return nil, err
	}

	endpoint := adjustEndpoint(provider.Endpoint(), cfg)
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = defaultScopes
	}

	return &OIDCProvider{
		provider: provider,
		oauth2Config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     endpoint,
			Scopes:       scopes,
		},
		verifier:   provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		httpClient: httpClient,
	}, nil
}

func setupHTTPClient(ctx context.Context, cfg ProviderConfig) (*http.Client, context.Context, error) {
	if !needsIssuerOverride(cfg) {
		return nil, ctx, nil
	}

	if !config.IsHTTPAllowed() {
		return nil, ctx, ErrInsecureIssuerNotAllowed
	}

	ctx = oidc.InsecureIssuerURLContext(ctx, cfg.IssuerURL)
	httpClient := &http.Client{
		Transport: &urlRewriteTransport{
			base:    http.DefaultTransport,
			fromURL: cfg.IssuerURL,
			toURL:   cfg.DiscoveryURL,
		},
	}
	return httpClient, context.WithValue(ctx, oauth2.HTTPClient, httpClient), nil
}

func needsIssuerOverride(cfg ProviderConfig) bool {
	return cfg.IssuerURL != "" && cfg.IssuerURL != cfg.DiscoveryURL
}

func adjustEndpoint(endpoint oauth2.Endpoint, cfg ProviderConfig) oauth2.Endpoint {
	if needsIssuerOverride(cfg) {
		endpoint.TokenURL = strings.Replace(endpoint.TokenURL, cfg.IssuerURL, cfg.DiscoveryURL, 1)
	}
	return endpoint
}

func (p *OIDCProvider) AuthCodeURL(state, nonce, codeVerifier string) string {
	return p.oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.S256ChallengeOption(codeVerifier),
	)
}

func (p *OIDCProvider) ExchangeCode(ctx context.Context, code, codeVerifier, expectedNonce string) (*TokenResult, error) {
	// Use custom HTTP client for URL rewriting if configured
	if p.httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, p.httpClient)
	}

	token, err := p.oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(codeVerifier))
	if err != nil {
		return nil, err
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	var claims struct {
		Nonce string `json:"nonce"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	if claims.Nonce != expectedNonce {
		return nil, ErrNonceMismatch
	}

	return &TokenResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
		Subject:      idToken.Subject,
		Email:        claims.Email,
		Name:         claims.Name,
	}, nil
}

func (p *OIDCProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error) {
	tokenSource := p.oauth2Config.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	token, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
	}, nil
}
