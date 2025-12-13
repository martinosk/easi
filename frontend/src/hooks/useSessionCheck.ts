import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useUserStore } from '../store/userStore';

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
    if (!isLoading && !isAuthenticated && !location.pathname.endsWith('/login')) {
      const returnUrl = encodeURIComponent(location.pathname + location.search);
      navigate(`/login?returnUrl=${returnUrl}`, { replace: true });
    }
  }, [isLoading, isAuthenticated, location, navigate]);

  return { isLoading, isAuthenticated };
}
