import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { LoadingFallback } from '../components/shared/LoadingFallback';
import { useUserStore } from '../store/userStore';

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
