import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useUserStore } from '../store/userStore';
import { LoadingFallback } from '../components/shared/LoadingFallback';

export function ProtectedRoute() {
  const location = useLocation();
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const isLoading = useUserStore((state) => state.isLoading);

  if (isLoading) {
    return <LoadingFallback message="Checking session..." />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <Outlet />;
}

export const ROUTES = {
  HOME: '/',
  CANVAS: '/canvas',
  BUSINESS_DOMAINS: '/business-domains',
  BUSINESS_DOMAIN_DETAIL: '/business-domains/:domainId',
  INVITATIONS: '/invitations',
  LOGIN: '/login',
} as const;
