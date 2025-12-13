package session

import (
	"context"
	"errors"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

const (
	sessionKey = "auth_session"
)

type SessionManager struct {
	scs *scs.SessionManager
}

func NewSessionManager(scsManager *scs.SessionManager) *SessionManager {
	return &SessionManager{scs: scsManager}
}

func (m *SessionManager) LoadAndSave(next http.Handler) http.Handler {
	return m.scs.LoadAndSave(next)
}

func (m *SessionManager) StorePreAuthSession(ctx context.Context, session AuthSession) error {
	data, err := session.Marshal()
	if err != nil {
		return err
	}
	m.scs.Put(ctx, sessionKey, data)
	return nil
}

func (m *SessionManager) LoadPreAuthSession(ctx context.Context) (AuthSession, error) {
	data := m.scs.GetBytes(ctx, sessionKey)
	if data == nil {
		return AuthSession{}, ErrSessionNotFound
	}
	return UnmarshalAuthSession(data)
}

func (m *SessionManager) StoreAuthenticatedSession(ctx context.Context, session AuthSession) error {
	data, err := session.Marshal()
	if err != nil {
		return err
	}
	m.scs.Put(ctx, sessionKey, data)
	return nil
}

func (m *SessionManager) LoadAuthenticatedSession(ctx context.Context) (AuthSession, error) {
	data := m.scs.GetBytes(ctx, sessionKey)
	if data == nil {
		return AuthSession{}, ErrSessionNotFound
	}
	session, err := UnmarshalAuthSession(data)
	if err != nil {
		return AuthSession{}, err
	}
	if !session.IsAuthenticated() {
		return AuthSession{}, ErrSessionNotFound
	}
	return session, nil
}

func (m *SessionManager) ClearSession(ctx context.Context) error {
	return m.scs.Destroy(ctx)
}

func (m *SessionManager) RenewToken(ctx context.Context) error {
	return m.scs.RenewToken(ctx)
}
