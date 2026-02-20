import { lazy, Suspense, useCallback, useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { useUserStore } from './store/userStore';
import { AppLayout } from './components/layout/AppLayout';
import { AppNavigation } from './components/layout/AppNavigation';
import { LoadingScreen } from './components/shared/LoadingScreen';
import { ErrorScreen } from './components/shared/ErrorScreen';
import { ErrorBoundary, FeatureErrorFallback } from './components/shared/ErrorBoundary';
import { LoadingFallback } from './components/shared/LoadingFallback';
import { ReleaseNotesOverlay } from './contexts/releases/components/ReleaseNotesOverlay';
import { ChatButton, useChatStore } from './features/chat';
import { useDialogContext } from './contexts/dialogs';
import { useAppInitialization } from './hooks/useAppInitialization';
import { useReleaseNotes } from './hooks/useReleaseNotes';
import type { Release } from './api/types';

const CanvasContainer = lazy(() => import('./features/canvas/CanvasContainer'));

const DialogManager = lazy(() =>
  import('./components/shared/DialogManager').then(module => ({ default: module.DialogManager }))
);

const ChatPanel = lazy(() =>
  import('./features/chat/components/ChatPanel').then(module => ({ default: module.ChatPanel }))
);

const BusinessDomainsRouter = lazy(() =>
  import('./features/business-domains').then(module => ({ default: module.BusinessDomainsRouter }))
);

const InvitationsPage = lazy(() =>
  import('./features/invitations').then(module => ({ default: module.InvitationsPage }))
);

const UsersPage = lazy(() =>
  import('./features/users').then(module => ({ default: module.UsersPage }))
);

const SettingsPage = lazy(() =>
  import('./features/settings').then(module => ({ default: module.SettingsPage }))
);

const ValueStreamsRouter = lazy(() =>
  import('./features/value-streams').then(module => ({ default: module.ValueStreamsRouter }))
);

const EnterpriseArchRouter = lazy(() =>
  import('./features/enterprise-architecture').then(module => ({ default: module.EnterpriseArchRouter }))
);

const MyEditAccessPage = lazy(() =>
  import('./features/edit-grants/pages/MyEditAccessPage')
);

type AppView = 'canvas' | 'business-domains' | 'value-streams' | 'invitations' | 'users' | 'settings' | 'enterprise-architecture' | 'my-edit-access';

function useAuthErrorHandler() {
  const [authError, setAuthError] = useState<string | null>(null);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const errorCode = params.get('auth_error');
    const errorMessage = params.get('auth_error_message');

    if (errorCode && errorMessage) {
      setAuthError(errorMessage);
      toast.error(errorMessage, { duration: 10000 });

      const url = new URL(window.location.href);
      url.searchParams.delete('auth_error');
      url.searchParams.delete('auth_error_message');
      window.history.replaceState({}, '', url.toString());
    }
  }, []);

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
  return (
    <ReleaseNotesOverlay
      isOpen={showOverlay}
      release={release}
      onDismiss={onDismiss}
    />
  );
}

function LazyFeatureView({ featureName, children }: { featureName: string; children: React.ReactNode }) {
  return (
    <ErrorBoundary
      fallback={(error, reset) => (
        <FeatureErrorFallback featureName={featureName} error={error} onReset={reset} />
      )}
    >
      <Suspense fallback={<LoadingFallback message={`Loading ${featureName}...`} />}>
        {children}
      </Suspense>
    </ErrorBoundary>
  );
}

function MainContent({ view }: { view: AppView }) {
  if (view === 'canvas') {
    return <LazyFeatureView featureName="Canvas"><CanvasContainer /></LazyFeatureView>;
  }
  if (view === 'invitations') {
    return <LazyFeatureView featureName="Invitations"><InvitationsPage /></LazyFeatureView>;
  }
  if (view === 'users') {
    return <LazyFeatureView featureName="Users"><UsersPage /></LazyFeatureView>;
  }
  if (view === 'settings') {
    return <LazyFeatureView featureName="Settings"><SettingsPage /></LazyFeatureView>;
  }
  if (view === 'value-streams') {
    return <LazyFeatureView featureName="Value Streams"><ValueStreamsRouter /></LazyFeatureView>;
  }
  if (view === 'enterprise-architecture') {
    return <LazyFeatureView featureName="Enterprise Architecture"><EnterpriseArchRouter /></LazyFeatureView>;
  }
  if (view === 'my-edit-access') {
    return <LazyFeatureView featureName="My Edit Access"><MyEditAccessPage /></LazyFeatureView>;
  }
  return <LazyFeatureView featureName="Business Domains"><BusinessDomainsRouter /></LazyFeatureView>;
}

interface AppProps {
  view: AppView;
}

function App({ view }: AppProps) {
  const { authError } = useAuthErrorHandler();
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const sessionLinks = useUserStore((state) => state.sessionLinks);
  const assistantAvailable = Boolean(sessionLinks?.['x-assistant']);
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
          onRetry={() => window.location.href = '/easi/login'}
          retryLabel="Back to Login"
        />
      </AppLayout>
    );
  }

  if (isLoading) {
    return <AppLayout><LoadingScreen /></AppLayout>;
  }

  if (error) {
    return <AppLayout><ErrorScreen error={error.message} onRetry={() => window.location.reload()} /></AppLayout>;
  }

  return (
    <AppLayout>
      <AppNavigation
        currentView={view}
        onOpenReleaseNotes={openReleaseNotesBrowser}
        chatButton={
          <ChatButton
            assistantAvailable={assistantAvailable}
            onClick={toggleChat}
            isActive={chatIsOpen}
          />
        }
      />
      <MainContent view={view} />
      <Suspense fallback={null}>
        <DialogManager />
      </Suspense>
      {chatIsOpen && (
        <Suspense fallback={null}>
          <ChatPanel isOpen={chatIsOpen} onClose={closeChat} />
        </Suspense>
      )}
      <ReleaseNotesDisplay
        showOverlay={showReleaseNotes}
        release={release}
        onDismiss={dismissReleaseNotes}
      />
    </AppLayout>
  );
}

export default App;
