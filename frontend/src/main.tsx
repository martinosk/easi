/* eslint-disable react-refresh/only-export-components */

import { Button, Center, Group, MantineProvider, Stack, Text, Title } from '@mantine/core';
import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { StrictMode, useEffect } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import 'dockview/dist/styles/dockview.css';
import '@mantine/core/styles.css';
import './index.css';
import App from './App.tsx';
import { ErrorBoundary } from './components/shared/ErrorBoundary.tsx';
import { DialogProvider } from './contexts/dialogs';
import { LoginPage } from './features/auth/pages/LoginPage.tsx';
import { queryClient } from './lib/queryClient';
import { ROUTES } from './routes/routePaths.ts';
import { ProtectedRoute } from './routes/routes.tsx';
import { useUserStore } from './store/userStore.ts';
import { theme } from './theme/mantine';

const basename = import.meta.env.BASE_URL.replace(/\/$/, '') || '';

function RootErrorFallback({ error, onReset }: { error: Error; onReset: () => void }) {
  return (
    <MantineProvider theme={theme} defaultColorScheme="light">
      <Center mih="100vh" p="lg">
        <Stack align="center" gap="md" maw={520}>
          <Title order={3} c="red">
            Application Error
          </Title>
          <Text size="sm" c="dimmed" ta="center">
            {error.message}
          </Text>
          <Group gap="sm">
            <Button onClick={onReset} color="red">
              Try again
            </Button>
            <Button variant="default" onClick={() => (window.location.href = `${basename}/`)}>
              Go to home
            </Button>
          </Group>
        </Stack>
      </Center>
    </MantineProvider>
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
    <ErrorBoundary fallback={(error, reset) => <RootErrorFallback error={error} onReset={reset} />}>
      <QueryClientProvider client={queryClient}>
        <MantineProvider theme={theme} defaultColorScheme="light">
          <BrowserRouter basename={basename}>
            <SessionInitializer>
              <DialogProvider>
                <Routes>
                  <Route path={ROUTES.LOGIN} element={<LoginPage />} />
                  <Route element={<ProtectedRoute />}>
                    <Route path={ROUTES.HOME} element={<App view="canvas" />} />
                    <Route path={ROUTES.CANVAS} element={<Navigate to={ROUTES.HOME} replace />} />
                    <Route path={ROUTES.BUSINESS_DOMAINS} element={<App view="business-domains" />} />
                    <Route path={ROUTES.BUSINESS_DOMAIN_DETAIL} element={<App view="business-domains" />} />
                    <Route path={`${ROUTES.VALUE_STREAMS}/*`} element={<App view="value-streams" />} />
                    <Route path={ROUTES.ENTERPRISE_ARCHITECTURE} element={<App view="enterprise-architecture" />} />
                    <Route path={ROUTES.INVITATIONS} element={<App view="invitations" />} />
                    <Route path={ROUTES.USERS} element={<App view="users" />} />
                    <Route path="/settings/*" element={<App view="settings" />} />
                    <Route path={ROUTES.MY_EDIT_ACCESS} element={<App view="my-edit-access" />} />
                  </Route>
                  <Route path="*" element={<Navigate to={ROUTES.HOME} replace />} />
                </Routes>
              </DialogProvider>
            </SessionInitializer>
          </BrowserRouter>
        </MantineProvider>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </ErrorBoundary>
  </StrictMode>,
);
