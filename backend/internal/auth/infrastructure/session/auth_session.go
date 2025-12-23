package session

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"easi/backend/internal/auth/domain/valueobjects"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

const tokenExpiryBuffer = 5 * time.Minute

type AuthSession struct {
	tenantID            string
	state               string
	nonce               string
	codeVerifier        string
	expectedEmailDomain string
	returnURL           string
	userID              uuid.UUID
	userEmail           string
	accessToken         string
	refreshToken        string
	tokenExpiry         time.Time
	authenticated       bool
}

type authSessionJSON struct {
	TenantID            string    `json:"tenantId"`
	State               string    `json:"state"`
	Nonce               string    `json:"nonce"`
	CodeVerifier        string    `json:"codeVerifier"`
	ExpectedEmailDomain string    `json:"expectedEmailDomain"`
	ReturnURL           string    `json:"returnUrl"`
	UserID              string    `json:"userId"`
	UserEmail           string    `json:"userEmail"`
	AccessToken         string    `json:"accessToken"`
	RefreshToken        string    `json:"refreshToken"`
	TokenExpiry         time.Time `json:"tokenExpiry"`
	Authenticated       bool      `json:"authenticated"`
}

func NewPreAuthSession(tenantID sharedvo.TenantID, expectedEmailDomain string, returnURL string) AuthSession {
	return AuthSession{
		tenantID:            tenantID.Value(),
		state:               valueobjects.NewAuthState().Value(),
		nonce:               valueobjects.NewNonce().Value(),
		codeVerifier:        oauth2.GenerateVerifier(),
		expectedEmailDomain: expectedEmailDomain,
		returnURL:           returnURL,
		authenticated:       false,
	}
}

func (s AuthSession) TenantID() string {
	return s.tenantID
}

func (s AuthSession) State() string {
	return s.state
}

func (s AuthSession) Nonce() string {
	return s.nonce
}

func (s AuthSession) CodeVerifier() string {
	return s.codeVerifier
}

func (s AuthSession) ExpectedEmailDomain() string {
	return s.expectedEmailDomain
}

func (s AuthSession) ReturnURL() string {
	return s.returnURL
}

func (s AuthSession) UserID() uuid.UUID {
	return s.userID
}

func (s AuthSession) UserEmail() string {
	return s.userEmail
}

func (s AuthSession) AccessToken() string {
	return s.accessToken
}

func (s AuthSession) RefreshToken() string {
	return s.refreshToken
}

func (s AuthSession) TokenExpiry() time.Time {
	return s.tokenExpiry
}

func (s AuthSession) IsAuthenticated() bool {
	return s.authenticated
}

func (s AuthSession) IsTokenExpired() bool {
	return time.Now().Add(tokenExpiryBuffer).After(s.tokenExpiry)
}

type TokenInfo struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

type UserInfo struct {
	ID    uuid.UUID
	Email string
}

func (s AuthSession) UpgradeToAuthenticated(user UserInfo, tokens TokenInfo) AuthSession {
	return s.toAuthenticatedSession(user, tokens)
}

func (s AuthSession) UpdateTokens(tokens TokenInfo) AuthSession {
	return s.toAuthenticatedSession(UserInfo{ID: s.userID, Email: s.userEmail}, tokens)
}

func (s AuthSession) toAuthenticatedSession(user UserInfo, tokens TokenInfo) AuthSession {
	return AuthSession{
		tenantID:            s.tenantID,
		state:               "",
		nonce:               "",
		codeVerifier:        "",
		expectedEmailDomain: "",
		returnURL:           "",
		userID:              user.ID,
		userEmail:           user.Email,
		accessToken:         tokens.AccessToken,
		refreshToken:        tokens.RefreshToken,
		tokenExpiry:         tokens.Expiry,
		authenticated:       true,
	}
}

func (s AuthSession) Marshal() ([]byte, error) {
	data := authSessionJSON{
		TenantID:            s.tenantID,
		State:               s.state,
		Nonce:               s.nonce,
		CodeVerifier:        s.codeVerifier,
		ExpectedEmailDomain: s.expectedEmailDomain,
		ReturnURL:           s.returnURL,
		UserID:              s.userID.String(),
		UserEmail:           s.userEmail,
		AccessToken:         s.accessToken,
		RefreshToken:        s.refreshToken,
		TokenExpiry:         s.tokenExpiry,
		Authenticated:       s.authenticated,
	}
	return json.Marshal(data)
}

func UnmarshalAuthSession(data []byte) (AuthSession, error) {
	var j authSessionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return AuthSession{}, err
	}

	var userID uuid.UUID
	if j.UserID != "" && j.UserID != "00000000-0000-0000-0000-000000000000" {
		var err error
		userID, err = uuid.Parse(j.UserID)
		if err != nil {
			return AuthSession{}, err
		}
	}

	return AuthSession{
		tenantID:            j.TenantID,
		state:               j.State,
		nonce:               j.Nonce,
		codeVerifier:        j.CodeVerifier,
		expectedEmailDomain: j.ExpectedEmailDomain,
		returnURL:           j.ReturnURL,
		userID:              userID,
		userEmail:           j.UserEmail,
		accessToken:         j.AccessToken,
		refreshToken:        j.RefreshToken,
		tokenExpiry:         j.TokenExpiry,
		authenticated:       j.Authenticated,
	}, nil
}
