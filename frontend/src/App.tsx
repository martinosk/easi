import { lazy, Suspense, useEffect, useRef, useState, useMemo } from 'react';
import toast from 'react-hot-toast';
import { useAppStore } from './store/appStore';
import { useUserStore } from './store/userStore';
import { AppLayout } from './components/layout/AppLayout';
import { AppNavigation } from './components/layout/AppNavigation';
import { LoadingScreen } from './components/shared/LoadingScreen';
import { ErrorScreen } from './components/shared/ErrorScreen';
import { ErrorBoundary, FeatureErrorFallback } from './components/shared/ErrorBoundary';
import { LoadingFallback } from './components/shared/LoadingFallback';
import { DockviewLayout } from './components/layout/DockviewLayout';
import { DialogManager } from './components/shared/DialogManager';
import { ReleaseNotesOverlay } from './contexts/releases/components/ReleaseNotesOverlay';
import type { ComponentCanvasRef } from './features/canvas/components/ComponentCanvas';
import { useCanvasDialogs } from './features/canvas/hooks/useCanvasDialogs';
import { useViewOperations } from './features/views/hooks/useViewOperations';
import { useCanvasNavigation } from './features/canvas/hooks/useCanvasNavigation';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useReleaseNotes } from './hooks/useReleaseNotes';
import { useRelations } from './features/relations/hooks/useRelations';
import { useComponents } from './features/components/hooks/useComponents';
import { useAppInitialization } from './hooks/useAppInitialization';
import type { Release } from './api/types';

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

const EnterpriseArchRouter = lazy(() =>
  import('./features/enterprise-architecture').then(module => ({ default: module.EnterpriseArchRouter }))
);

const MyEditAccessPage = lazy(() =>
  import('./features/edit-grants/pages/MyEditAccessPage')
);

type AppView = 'canvas' | 'business-domains' | 'invitations' | 'users' | 'settings' | 'enterprise-architecture' | 'my-edit-access';

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

function useCanvasPermissions() {
  const hasPermission = useUserStore((state) => state.hasPermission);
  return useMemo(() => ({
    canCreateComponent: hasPermission('components:write'),
    canCreateCapability: hasPermission('capabilities:write'),
    canCreateView: hasPermission('views:write'),
    canCreateOriginEntity: hasPermission('components:write'),
  }), [hasPermission]);
}

interface CanvasViewProps {
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  dialogActions: ReturnType<typeof useCanvasDialogs>;
  addComponentToView: ReturnType<typeof useViewOperations>['addComponentToView'];
  switchView: ReturnType<typeof useViewOperations>['switchView'];
  navigateToComponent: ReturnType<typeof useCanvasNavigation>['navigateToComponent'];
  navigateToCapability: ReturnType<typeof useCanvasNavigation>['navigateToCapability'];
  navigateToOriginEntity: ReturnType<typeof useCanvasNavigation>['navigateToOriginEntity'];
  onRemoveFromView: () => void;
  permissions: ReturnType<typeof useCanvasPermissions>;
}

function CanvasView({
  canvasRef,
  selectedNodeId,
  selectedEdgeId,
  dialogActions,
  addComponentToView,
  switchView,
  navigateToComponent,
  navigateToCapability,
  navigateToOriginEntity,
  onRemoveFromView,
  permissions,
}: CanvasViewProps) {
  return (
    <DockviewLayout
      canvasRef={canvasRef}
      selectedNodeId={selectedNodeId}
      selectedEdgeId={selectedEdgeId}
      onAddComponent={permissions.canCreateComponent ? dialogActions.openComponentDialog : undefined}
      onAddCapability={permissions.canCreateCapability ? dialogActions.openCapabilityDialog : undefined}
      canCreateView={permissions.canCreateView}
      canCreateOriginEntity={permissions.canCreateOriginEntity}
      onConnect={dialogActions.openRelationDialog}
      onComponentDrop={(id, x, y) => addComponentToView(id as import('./api/types').ComponentId, x, y)}
      onComponentSelect={navigateToComponent}
      onCapabilitySelect={navigateToCapability}
      onOriginEntitySelect={navigateToOriginEntity}
      onViewSelect={async (id) => switchView(id as import('./api/types').ViewId)}
      onEditComponent={dialogActions.openEditComponentDialog}
      onEditRelation={dialogActions.openEditRelationDialog}
      onEditCapability={dialogActions.openEditCapabilityDialog}
      onRemoveFromView={onRemoveFromView}
    />
  );
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

interface MainContentProps {
  view: AppView;
  canvasViewProps: CanvasViewProps;
}

function MainContent({ view, canvasViewProps }: MainContentProps) {
  if (view === 'canvas') {
    return <CanvasView {...canvasViewProps} />;
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
  const canvasRef = useRef<ComponentCanvasRef>(null);

  const { authError } = useAuthErrorHandler();
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const permissions = useCanvasPermissions();

  const { isLoading, error } = useAppInitialization();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const { data: relations = [] } = useRelations();
  const { data: components = [] } = useComponents();

  const { showOverlay: showReleaseNotes, release, dismiss: dismissReleaseNotes } = useReleaseNotes();
  const dialogActions = useCanvasDialogs(selectedEdgeId, relations, components);
  const { removeComponentFromView, addComponentToView, switchView } = useViewOperations();
  const { navigateToComponent, navigateToCapability, navigateToOriginEntity } = useCanvasNavigation(canvasRef);

  const handleRemoveFromView = () => {
    if (selectedNodeId) {
      removeComponentFromView(selectedNodeId);
    }
  };

  useKeyboardShortcuts({ onDelete: handleRemoveFromView });

  const hasNoData = components.length === 0;

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

  if (isLoading && hasNoData) {
    return <AppLayout><LoadingScreen /></AppLayout>;
  }

  if (error && hasNoData) {
    return <AppLayout><ErrorScreen error={error.message} onRetry={() => window.location.reload()} /></AppLayout>;
  }

  const canvasViewProps: CanvasViewProps = {
    canvasRef,
    selectedNodeId,
    selectedEdgeId,
    dialogActions,
    addComponentToView,
    switchView,
    navigateToComponent,
    navigateToCapability,
    navigateToOriginEntity,
    onRemoveFromView: handleRemoveFromView,
    permissions,
  };

  return (
    <AppLayout>
      <AppNavigation currentView={view} onOpenReleaseNotes={dialogActions.openReleaseNotesBrowser} />
      <MainContent view={view} canvasViewProps={canvasViewProps} />
      <DialogManager />
      <ReleaseNotesDisplay
        showOverlay={showReleaseNotes}
        release={release}
        onDismiss={dismissReleaseNotes}
      />
    </AppLayout>
  );
}

export default App;
