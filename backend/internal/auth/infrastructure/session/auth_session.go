package session

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"easi/backend/internal/auth/domain/valueobjects"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

type AuthSession struct {
	tenantID      string
	state         string
	nonce         string
	codeVerifier  string
	returnURL     string
	userID        uuid.UUID
	accessToken   string
	refreshToken  string
	tokenExpiry   time.Time
	authenticated bool
}

type authSessionJSON struct {
	TenantID      string    `json:"tenantId"`
	State         string    `json:"state"`
	Nonce         string    `json:"nonce"`
	CodeVerifier  string    `json:"codeVerifier"`
	ReturnURL     string    `json:"returnUrl"`
	UserID        string    `json:"userId"`
	AccessToken   string    `json:"accessToken"`
	RefreshToken  string    `json:"refreshToken"`
	TokenExpiry   time.Time `json:"tokenExpiry"`
	Authenticated bool      `json:"authenticated"`
}

func NewPreAuthSession(tenantID sharedvo.TenantID, returnURL string) AuthSession {
	return AuthSession{
		tenantID:      tenantID.Value(),
		state:         valueobjects.NewAuthState().Value(),
		nonce:         valueobjects.NewNonce().Value(),
		codeVerifier:  oauth2.GenerateVerifier(),
		returnURL:     returnURL,
		authenticated: false,
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

func (s AuthSession) ReturnURL() string {
	return s.returnURL
}

func (s AuthSession) UserID() uuid.UUID {
	return s.userID
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
	return time.Now().After(s.tokenExpiry)
}

func (s AuthSession) UpgradeToAuthenticated(
	userID uuid.UUID,
	accessToken string,
	refreshToken string,
	tokenExpiry time.Time,
) AuthSession {
	return AuthSession{
		tenantID:      s.tenantID,
		state:         s.state,
		nonce:         s.nonce,
		codeVerifier:  s.codeVerifier,
		returnURL:     s.returnURL,
		userID:        userID,
		accessToken:   accessToken,
		refreshToken:  refreshToken,
		tokenExpiry:   tokenExpiry,
		authenticated: true,
	}
}

func (s AuthSession) UpdateTokens(
	accessToken string,
	refreshToken string,
	tokenExpiry time.Time,
) AuthSession {
	return AuthSession{
		tenantID:      s.tenantID,
		state:         s.state,
		nonce:         s.nonce,
		codeVerifier:  s.codeVerifier,
		returnURL:     s.returnURL,
		userID:        s.userID,
		accessToken:   accessToken,
		refreshToken:  refreshToken,
		tokenExpiry:   tokenExpiry,
		authenticated: s.authenticated,
	}
}

func (s AuthSession) Marshal() ([]byte, error) {
	data := authSessionJSON{
		TenantID:      s.tenantID,
		State:         s.state,
		Nonce:         s.nonce,
		CodeVerifier:  s.codeVerifier,
		ReturnURL:     s.returnURL,
		UserID:        s.userID.String(),
		AccessToken:   s.accessToken,
		RefreshToken:  s.refreshToken,
		TokenExpiry:   s.tokenExpiry,
		Authenticated: s.authenticated,
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
		tenantID:      j.TenantID,
		state:         j.State,
		nonce:         j.Nonce,
		codeVerifier:  j.CodeVerifier,
		returnURL:     j.ReturnURL,
		userID:        userID,
		accessToken:   j.AccessToken,
		refreshToken:  j.RefreshToken,
		tokenExpiry:   j.TokenExpiry,
		authenticated: j.Authenticated,
	}, nil
}
