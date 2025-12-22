import { StrictMode, useEffect } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { MantineProvider } from '@mantine/core'
import { QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import 'dockview/dist/styles/dockview.css'
import '@mantine/core/styles.css'
import './index.css'
import App from './App.tsx'
import { LoginPage } from './features/auth/pages/LoginPage.tsx'
import { ErrorBoundary } from './components/shared/ErrorBoundary.tsx'
import { ProtectedRoute, ROUTES } from './routes/routes.tsx'
import { useUserStore } from './store/userStore.ts'
import { theme } from './theme/mantine'
import { queryClient } from './lib/queryClient'

const basename = import.meta.env.BASE_URL.replace(/\/$/, '') || '';

function RootErrorFallback({ error, onReset }: { error: Error; onReset: () => void }) {
  return (
    <div className="error-boundary-fallback" style={{ minHeight: '100vh' }}>
      <div className="error-boundary-content">
        <svg
          className="error-boundary-icon"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M12 9V13M12 17H12.01M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
        <h3 className="error-boundary-title">Application Error</h3>
        <p className="error-boundary-message">{error.message}</p>
        <div className="error-boundary-actions">
          <button
            type="button"
            className="error-boundary-button"
            onClick={onReset}
          >
            Try again
          </button>
          <button
            type="button"
            className="error-boundary-button error-boundary-button-secondary"
            onClick={() => window.location.href = basename + '/'}
          >
            Go to home
          </button>
        </div>
      </div>
    </div>
  );
}

function SessionInitializer({ children }: { children: React.ReactNode }) {
  const loadSession = useUserStore((state) => state.loadSession);

  useEffect(() => {
    loadSession();
  }, [loadSession]);

  return <>{children}</>;
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary
      fallback={(error, reset) => <RootErrorFallback error={error} onReset={reset} />}
    >
      <QueryClientProvider client={queryClient}>
        <MantineProvider theme={theme} defaultColorScheme="light">
          <BrowserRouter basename={basename}>
            <SessionInitializer>
              <Routes>
                <Route path={ROUTES.LOGIN} element={<LoginPage />} />
                <Route element={<ProtectedRoute />}>
                  <Route path={ROUTES.HOME} element={<App view="canvas" />} />
                  <Route path={ROUTES.CANVAS} element={<Navigate to={ROUTES.HOME} replace />} />
                  <Route path={ROUTES.BUSINESS_DOMAINS} element={<App view="business-domains" />} />
                  <Route path={ROUTES.BUSINESS_DOMAIN_DETAIL} element={<App view="business-domains" />} />
                  <Route path={ROUTES.INVITATIONS} element={<App view="invitations" />} />
                  <Route path={ROUTES.USERS} element={<App view="users" />} />
                </Route>
                <Route path="*" element={<Navigate to={ROUTES.HOME} replace />} />
              </Routes>
            </SessionInitializer>
          </BrowserRouter>
        </MantineProvider>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </ErrorBoundary>
  </StrictMode>,
)
