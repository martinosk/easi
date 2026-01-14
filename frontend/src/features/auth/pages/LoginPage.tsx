import { useState, useEffect, useMemo, type FormEvent, type FC } from 'react';
import { useSearchParams } from 'react-router-dom';
import { authApi } from '../api/authApi';
import { resetLoginRedirectFlag } from '../../../api';

function getReturnUrlFromParams(searchParams: URLSearchParams): string | undefined {
  return searchParams.get('returnUrl') ?? undefined;
}

function isExternalHttps(url: URL): boolean {
  return url.protocol === 'https:' && url.origin !== window.location.origin;
}

function isDevLocalhost(url: URL): boolean {
  if (!import.meta.env.DEV) return false;
  if (url.protocol !== 'http:') return false;
  return url.hostname === 'localhost' || url.hostname === '127.0.0.1';
}

function isAllowedAuthorizeUrl(url: URL): boolean {
  return isExternalHttps(url) || isDevLocalhost(url);
}

function sanitizeAuthorizeUrl(untrustedUrl: string): string | null {
  let parsed: URL;
  try {
    parsed = new URL(untrustedUrl);
  } catch {
    return null;
  }
  if (isAllowedAuthorizeUrl(parsed)) {
    return parsed.href;
  }
  return null;
}

export const LoginPage: FC = () => {
  const [searchParams] = useSearchParams();
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const returnUrl = useMemo(() => getReturnUrlFromParams(searchParams), [searchParams]);

  useEffect(() => {
    resetLoginRedirectFlag();
  }, []);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!email.trim()) {
      setError('Email is required');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await authApi.initiateLogin(email, returnUrl);
      const sanitizedUrl = sanitizeAuthorizeUrl(response._links.authorize);
      if (sanitizedUrl === null) {
        setLoading(false);
        setError('Invalid authorization URL received');
        return;
      }
      window.location.href = sanitizedUrl;
    } catch (err) {
      setLoading(false);
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An unexpected error occurred');
      }
    }
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <div className="login-header">
          <h1>Welcome to EASI</h1>
          <p>Enterprise Architecture - Simple</p>
        </div>

        <form onSubmit={handleSubmit} className="login-form">
          <div className="form-group">
            <label htmlFor="email" className="form-label">
              Email Address
            </label>
            <input
              id="email"
              type="email"
              className="form-input"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="john@acme.com"
              disabled={loading}
              autoFocus
            />
          </div>

          {error && (
            <div className="error-message">
              {error}
            </div>
          )}

          <button
            type="submit"
            className="btn btn-primary login-submit-btn"
            disabled={loading}
          >
            {loading ? (
              <>
                <div className="loading-spinner-small"></div>
                <span>Redirecting...</span>
              </>
            ) : (
              'Continue with SSO'
            )}
          </button>
        </form>
      </div>
    </div>
  );
};
