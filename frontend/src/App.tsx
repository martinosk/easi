import { type ComponentType, lazy, Suspense, useCallback, useLayoutEffect, useState } from 'react';
import toast from 'react-hot-toast';
import type { Release } from './api/types';
import classes from './App.module.css';
import { AppLayout } from './components/layout/AppLayout';
import { AppNavigation } from './components/layout/AppNavigation';
import { ErrorBoundary, FeatureErrorFallback } from './components/shared/ErrorBoundary';
import { ErrorScreen } from './components/shared/ErrorScreen';
import { LoadingFallback } from './components/shared/LoadingFallback';
import { LoadingScreen } from './components/shared/LoadingScreen';
import { useDialogContext } from './contexts/dialogs';
import { ReleaseNotesOverlay } from './contexts/releases/components/ReleaseNotesOverlay';
import { ChatButton, useAssistantAvailability, useChatStore } from './features/chat';
import { useAppInitialization } from './hooks/useAppInitialization';
import { useUnloadGuard } from './hooks/useUnloadGuard';
import { useReleaseNotes } from './contexts/releases/store/useReleaseNotes';
import type { AppView } from './routes/routePaths';
import { useUserStore } from './store/userStore';

const CanvasContainer = lazy(() => import('./features/canvas/CanvasContainer'));

const DialogManager = lazy(() =>
  import('./components/shared/DialogManager').then((module) => ({ default: module.DialogManager })),
);

const ChatPanel = lazy(() =>
  import('./features/chat/components/ChatPanel').then((module) => ({ default: module.ChatPanel })),
);

const BusinessDomainsRouter = lazy(() =>
  import('./features/business-domains').then((module) => ({ default: module.BusinessDomainsRouter })),
);

const InvitationsPage = lazy(() =>
  import('./features/invitations').then((module) => ({ default: module.InvitationsPage })),
);

const UsersPage = lazy(() => import('./features/users').then((module) => ({ default: module.UsersPage })));

const SettingsPage = lazy(() => import('./features/settings').then((module) => ({ default: module.SettingsPage })));

const ValueStreamsRouter = lazy(() =>
  import('./features/value-streams').then((module) => ({ default: module.ValueStreamsRouter })),
);

const StrategicFitPage = lazy(() =>
  import('./features/strategic-fit').then((module) => ({ default: module.StrategicFitPage })),
);

const MyEditAccessPage = lazy(() => import('./features/edit-grants/pages/MyEditAccessPage'));

const OnePagersRouter = lazy(() =>
  import('./features/one-pagers').then((module) => ({ default: module.OnePagersRouter })),
);

const OnePagerQualityPage = lazy(() =>
  import('./features/one-pager-quality').then((module) => ({ default: module.OnePagerQualityPage })),
);

function useAuthErrorHandler() {
  const [authError, setAuthError] = useState<string | null>(null);

  useLayoutEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const errorCode = params.get('auth_error');
    const errorMessage = params.get('auth_error_message');

    if (errorCode && errorMessage) {
      if (authError !== errorMessage) queueMicrotask(() => setAuthError(errorMessage));
      toast.error(errorMessage, { duration: 10000 });

      const url = new URL(window.location.href);
      url.searchParams.delete('auth_error');
      url.searchParams.delete('auth_error_message');
      window.history.replaceState({}, '', url.toString());
    }
  }, [authError]);

  return { authError, clearAuthError: () => setAuthError(null) };
}

interface ReleaseNotesDisplayProps {
  showOverlay: boolean;
  release: Release | null;
  onDismiss: (mode: 'forever' | 'untilNext') => void;
}

function ReleaseNotesDisplay({ showOverlay, release, onDismiss }: ReleaseNotesDisplayProps) {
  const showReleaseOverlay = showOverlay && release !== null;
  if (!showReleaseOverlay) return null;
  return <ReleaseNotesOverlay isOpen={showOverlay} release={release} onDismiss={onDismiss} />;
}

function LazyFeatureView({ featureName, children }: { featureName: string; children: React.ReactNode }) {
  return (
    <ErrorBoundary
      fallback={(error, reset) => <FeatureErrorFallback featureName={featureName} error={error} onReset={reset} />}
    >
      <Suspense fallback={<LoadingFallback message={`Loading ${featureName}...`} />}>{children}</Suspense>
    </ErrorBoundary>
  );
}

const mainViews: Record<AppView, { featureName: string; Component: ComponentType }> = {
  canvas: { featureName: 'Canvas', Component: CanvasContainer },
  'business-domains': { featureName: 'Business Domains', Component: BusinessDomainsRouter },
  'value-streams': { featureName: 'Value Streams', Component: ValueStreamsRouter },
  invitations: { featureName: 'Invitations', Component: InvitationsPage },
  users: { featureName: 'Users', Component: UsersPage },
  settings: { featureName: 'Settings', Component: SettingsPage },
  'strategic-fit': { featureName: 'Strategic Fit', Component: StrategicFitPage },
  'my-edit-access': { featureName: 'My Edit Access', Component: MyEditAccessPage },
  'one-pagers': { featureName: 'One-Pagers', Component: OnePagersRouter },
  'one-pager-quality': { featureName: 'One-Pager Quality', Component: OnePagerQualityPage },
};

function MainContent({ view }: { view: AppView }) {
  const { featureName, Component } = mainViews[view] ?? mainViews['business-domains'];
  return (
    <main className={classes.mainRegion} data-testid="main-region">
      <LazyFeatureView featureName={featureName}>
        <Component />
      </LazyFeatureView>
    </main>
  );
}

interface AppProps {
  view: AppView;
}

function App({ view }: AppProps) {
  useUnloadGuard();
  const { authError } = useAuthErrorHandler();
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const { assistantAvailable, assistantWriteAvailable } = useAssistantAvailability();
  const chatIsOpen = useChatStore((state) => state.isOpen);
  const toggleChat = useChatStore((state) => state.togglePanel);
  const closeChat = useChatStore((state) => state.closePanel);
  const { openDialog } = useDialogContext();

  const { isLoading, error } = useAppInitialization();
  const { showOverlay: showReleaseNotes, release, dismiss: dismissReleaseNotes } = useReleaseNotes();

  const openReleaseNotesBrowser = useCallback(() => {
    openDialog('release-notes-browser');
  }, [openDialog]);

  if (authError && !isAuthenticated) {
    return (
      <AppLayout>
        <ErrorScreen
          title="Access Denied"
          error={authError}
          onRetry={() => (window.location.href = '/easi/login')}
          retryLabel="Back to Login"
        />
      </AppLayout>
    );
  }

  if (isLoading) {
    return (
      <AppLayout>
        <LoadingScreen />
      </AppLayout>
    );
  }

  if (error) {
    return (
      <AppLayout>
        <ErrorScreen error={error.message} onRetry={() => window.location.reload()} />
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <AppNavigation
        currentView={view}
        onOpenReleaseNotes={openReleaseNotesBrowser}
        chatButton={<ChatButton assistantAvailable={assistantAvailable} onClick={toggleChat} isActive={chatIsOpen} />}
      />
      <MainContent view={view} />
      <Suspense fallback={null}>
        <DialogManager />
      </Suspense>
      {chatIsOpen && (
        <Suspense fallback={null}>
          <ChatPanel isOpen={chatIsOpen} onClose={closeChat} writeAvailable={assistantWriteAvailable} />
        </Suspense>
      )}
      <ReleaseNotesDisplay showOverlay={showReleaseNotes} release={release} onDismiss={dismissReleaseNotes} />
    </AppLayout>
  );
}

export default App;
