-- Migration: SCS HTTP Sessions Table
-- Spec: 066_SingleTenantLogin
-- Description: Creates table required by SCS (alexedwards/scs) for HTTP session storage

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions (expiry);
