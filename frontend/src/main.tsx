import { Button, Center, Group, MantineProvider, Stack, Text, Title } from '@mantine/core';
import { QueryClientProvider } from '@tanstack/react-query';
import { type ComponentProps, type ComponentType, lazy, StrictMode, Suspense, useEffect } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import '@fontsource-variable/inter/index.css';
import '@fontsource-variable/schibsted-grotesk/index.css';
import '@fontsource-variable/spline-sans-mono/index.css';
import '@mantine/core/styles.css';
import './theme/tokens.css';
import './theme/skins.css';
import './index.css';
import { ErrorBoundary } from './components/shared/ErrorBoundary.tsx';
import { DialogProvider } from './contexts/dialogs';
import { queryClient } from './lib/queryClient';
import { AppRoutes } from './routes/AppRoutes.tsx';
import { useUserStore } from './store/userStore.ts';
import { theme } from './theme/mantine';
import { initSkin } from './theme/skin';

type DevtoolsProps = ComponentProps<typeof import('@tanstack/react-query-devtools')['ReactQueryDevtools']>;

const ReactQueryDevtools: ComponentType<DevtoolsProps> = import.meta.env.PROD
  ? () => null
  : (lazy(() =>
      import('@tanstack/react-query-devtools').then((m) => ({
        default: m.ReactQueryDevtools,
      })),
    ) as ComponentType<DevtoolsProps>);

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

async function enableMocking(): Promise<void> {
  if (!import.meta.env.DEV || import.meta.env.VITE_USE_MOCK_API !== 'true') return;
  const [{ worker }, { seedDevData }] = await Promise.all([
    import('./test/mocks/browser'),
    import('./test/mocks/seedDevData'),
  ]);
  seedDevData();
  await worker.start({ onUnhandledRequest: 'bypass' });
}

function renderApp() {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <ErrorBoundary fallback={(error, reset) => <RootErrorFallback error={error} onReset={reset} />}>
        <QueryClientProvider client={queryClient}>
          <MantineProvider theme={theme} defaultColorScheme="light">
            <BrowserRouter basename={basename}>
              <SessionInitializer>
                <DialogProvider>
                  <AppRoutes />
                </DialogProvider>
              </SessionInitializer>
            </BrowserRouter>
          </MantineProvider>
          <Suspense fallback={null}>
            <ReactQueryDevtools initialIsOpen={false} />
          </Suspense>
        </QueryClientProvider>
      </ErrorBoundary>
    </StrictMode>,
  );
}

initSkin();
enableMocking().then(renderApp);
