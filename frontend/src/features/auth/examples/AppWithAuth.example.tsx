import { useCallback, useLayoutEffect, useState } from 'react';
import App from '../../../App';
import { LoginPage } from '../pages/LoginPage';

export function AppWithAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);

  const checkAuthentication = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/auth/me', {
        credentials: 'include',
      });

      setIsAuthenticated(response.ok);
    } catch {
      setIsAuthenticated(false);
    }
  }, []);

  useLayoutEffect(() => {
    queueMicrotask(() => checkAuthentication());
  }, [checkAuthentication]);

  if (isAuthenticated === null) {
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <p>Checking authentication...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <LoginPage />;
  }

  return <App view="canvas" />;
}
