import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useUserStore } from '../store/userStore';

function shouldRedirectToLogin(isLoading: boolean, isAuthenticated: boolean, pathname: string): boolean {
  return !isLoading && !isAuthenticated && !pathname.endsWith('/login');
}

export function useSessionCheck() {
  const navigate = useNavigate();
  const location = useLocation();
  const loadSession = useUserStore((state) => state.loadSession);
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const isLoading = useUserStore((state) => state.isLoading);

  useEffect(() => {
    loadSession();
  }, [loadSession]);

  useEffect(() => {
    if (shouldRedirectToLogin(isLoading, isAuthenticated, location.pathname)) {
      const returnUrl = encodeURIComponent(location.pathname + location.search);
      navigate(`/login?returnUrl=${returnUrl}`, { replace: true });
    }
  }, [isLoading, isAuthenticated, location, navigate]);

  return { isLoading, isAuthenticated };
}
