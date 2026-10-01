import { Navigate, Outlet, Route, Routes, useLocation } from 'react-router-dom';
import App from '../App';
import { LoginPage } from '../features/auth/pages/LoginPage';
import { deepLinkParams } from '../lib/deepLinks/registry';
import { ROUTES } from './routePaths';
import { ProtectedRoute } from './routes';

function LegacyViewLinkRedirect() {
  const { pathname, search, hash } = useLocation();
  const isLegacyViewLink = pathname === ROUTES.HOME && new URLSearchParams(search).has(deepLinkParams.VIEW.param);
  if (isLegacyViewLink) {
    return <Navigate to={{ pathname: ROUTES.CANVAS, search, hash }} replace />;
  }
  return <Outlet />;
}

export function AppRoutes() {
  return (
    <Routes>
      <Route path={ROUTES.LOGIN} element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<LegacyViewLinkRedirect />}>
          <Route path={ROUTES.HOME} element={<App view="home" />} />
          <Route path={ROUTES.CANVAS} element={<App view="canvas" />} />
          <Route path={ROUTES.BUSINESS_DOMAINS} element={<App view="business-domains" />} />
          <Route path={ROUTES.BUSINESS_DOMAIN_DETAIL} element={<App view="business-domains" />} />
          <Route path={`${ROUTES.VALUE_STREAMS}/*`} element={<App view="value-streams" />} />
          <Route path={ROUTES.STRATEGIC_FIT} element={<App view="strategic-fit" />} />
          <Route path={ROUTES.INVITATIONS} element={<App view="invitations" />} />
          <Route path={ROUTES.USERS} element={<App view="users" />} />
          <Route path={`${ROUTES.SETTINGS}/*`} element={<App view="settings" />} />
          <Route path={ROUTES.MY_EDIT_ACCESS} element={<App view="my-edit-access" />} />
          <Route path={`${ROUTES.ONE_PAGERS}/*`} element={<App view="one-pagers" />} />
          <Route path={ROUTES.ONE_PAGER_QUALITY} element={<App view="one-pager-quality" />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to={ROUTES.HOME} replace />} />
    </Routes>
  );
}
