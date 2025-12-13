package oidc

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

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

func NewOIDCProvider(ctx context.Context, discoveryURL, clientID, clientSecret, redirectURL string) (*OIDCProvider, error) {
	return NewOIDCProviderWithIssuer(ctx, discoveryURL, "", clientID, clientSecret, redirectURL)
}

func NewOIDCProviderWithIssuer(ctx context.Context, discoveryURL, issuerURL, clientID, clientSecret, redirectURL string) (*OIDCProvider, error) {
	var httpClient *http.Client

	// When discovery URL differs from issuer URL (e.g., Docker internal vs external),
	// create a custom HTTP client that rewrites URLs from issuer to discovery URL
	if issuerURL != "" && issuerURL != discoveryURL {
		ctx = oidc.InsecureIssuerURLContext(ctx, issuerURL)
		httpClient = &http.Client{
			Transport: &urlRewriteTransport{
				base:    http.DefaultTransport,
				fromURL: issuerURL,
				toURL:   discoveryURL,
			},
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	}

	provider, err := oidc.NewProvider(ctx, discoveryURL)
	if err != nil {
		return nil, err
	}

	endpoint := provider.Endpoint()

	// Also rewrite the token endpoint URL
	if issuerURL != "" && issuerURL != discoveryURL {
		endpoint.TokenURL = strings.Replace(endpoint.TokenURL, issuerURL, discoveryURL, 1)
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile", "offline_access"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCProvider{
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
		httpClient:   httpClient,
	}, nil
}

func NewOIDCProviderWithScopes(ctx context.Context, discoveryURL, clientID, clientSecret, redirectURL string, scopes []string) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, discoveryURL)
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCProvider{
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
	}, nil
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
