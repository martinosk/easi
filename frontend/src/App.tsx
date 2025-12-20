import { lazy, Suspense, useEffect, useRef, useState } from 'react';
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
import { ReleaseNotesBrowser } from './contexts/releases/components/ReleaseNotesBrowser';
import type { ComponentCanvasRef } from './features/canvas/components/ComponentCanvas';
import { useDialogManagement } from './hooks/useDialogManagement';
import { useViewOperations } from './hooks/useViewOperations';
import { useCanvasNavigation } from './hooks/useCanvasNavigation';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useReleaseNotes } from './hooks/useReleaseNotes';
import { useRelations } from './features/relations/hooks/useRelations';
import { useComponents } from './features/components/hooks/useComponents';
import type { Release } from './api/types';

const BusinessDomainsRouter = lazy(() =>
  import('./features/business-domains').then(module => ({ default: module.BusinessDomainsRouter }))
);

const InvitationsPage = lazy(() =>
  import('./features/invitations').then(module => ({ default: module.InvitationsPage }))
);

type AppView = 'canvas' | 'business-domains' | 'invitations';

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

interface CanvasViewProps {
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  dialogActions: ReturnType<typeof useDialogManagement>['actions'];
  dialogState: ReturnType<typeof useDialogManagement>['state'];
  addComponentToView: ReturnType<typeof useViewOperations>['addComponentToView'];
  switchView: ReturnType<typeof useViewOperations>['switchView'];
  navigateToComponent: ReturnType<typeof useCanvasNavigation>['navigateToComponent'];
  navigateToCapability: ReturnType<typeof useCanvasNavigation>['navigateToCapability'];
  onRemoveFromView: () => void;
}

function CanvasView({
  canvasRef,
  selectedNodeId,
  selectedEdgeId,
  dialogActions,
  dialogState,
  addComponentToView,
  switchView,
  navigateToComponent,
  navigateToCapability,
  onRemoveFromView,
}: CanvasViewProps) {
  return (
    <>
      <DockviewLayout
        canvasRef={canvasRef}
        selectedNodeId={selectedNodeId}
        selectedEdgeId={selectedEdgeId}
        onAddComponent={dialogActions.openComponentDialog}
        onAddCapability={dialogActions.openCapabilityDialog}
        onConnect={dialogActions.openRelationDialog}
        onComponentDrop={(id, x, y) => addComponentToView(id as import('./api/types').ComponentId, x, y)}
        onComponentSelect={navigateToComponent}
        onCapabilitySelect={navigateToCapability}
        onViewSelect={(id) => switchView(id as import('./api/types').ViewId)}
        onEditComponent={dialogActions.openEditComponentDialog}
        onEditRelation={dialogActions.openEditRelationDialog}
        onEditCapability={dialogActions.openEditCapabilityDialog}
        onRemoveFromView={onRemoveFromView}
      />
      <DialogManager
        componentDialog={dialogState.componentDialog}
        relationDialog={dialogState.relationDialog}
        editComponentDialog={dialogState.editComponentDialog}
        editRelationDialog={dialogState.editRelationDialog}
        capabilityDialog={dialogState.capabilityDialog}
        editCapabilityDialog={dialogState.editCapabilityDialog}
      />
    </>
  );
}

interface ReleaseNotesDisplayProps {
  showOverlay: boolean;
  release: Release | null;
  onDismiss: (mode: 'forever' | 'untilNext') => void;
  browserIsOpen: boolean;
  onBrowserClose: () => void;
}

function ReleaseNotesDisplay({ showOverlay, release, onDismiss, browserIsOpen, onBrowserClose }: ReleaseNotesDisplayProps) {
  const showReleaseOverlay = showOverlay && release !== null;
  return (
    <>
      {showReleaseOverlay && (
        <ReleaseNotesOverlay
          isOpen={showOverlay}
          release={release}
          onDismiss={onDismiss}
        />
      )}
      <ReleaseNotesBrowser isOpen={browserIsOpen} onClose={onBrowserClose} />
    </>
  );
}

interface AppProps {
  view: AppView;
}

function App({ view }: AppProps) {
  const canvasRef = useRef<ComponentCanvasRef>(null);

  const { authError } = useAuthErrorHandler();
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);

  const loadData = useAppStore((state) => state.loadData);
  const isLoading = useAppStore((state) => state.isLoading);
  const error = useAppStore((state) => state.error);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const { data: relations = [] } = useRelations();
  const { data: components = [] } = useComponents();

  const { showOverlay: showReleaseNotes, release, dismiss: dismissReleaseNotes } = useReleaseNotes();
  const { state: dialogState, actions: dialogActions } = useDialogManagement(selectedEdgeId, relations, components);
  const { removeComponentFromView, addComponentToView, switchView } = useViewOperations();
  const { navigateToComponent, navigateToCapability } = useCanvasNavigation(canvasRef);

  const handleRemoveFromView = () => {
    if (selectedNodeId) {
      removeComponentFromView(selectedNodeId);
    }
  };

  useKeyboardShortcuts({ onDelete: handleRemoveFromView });

  useEffect(() => {
    if (isAuthenticated) {
      loadData();
    }
  }, [loadData, isAuthenticated]);

  const hasNoData = components.length === 0;
  const isLoadingInitialData = isLoading && hasNoData;
  const showLoadingScreen = isLoadingInitialData;
  const showErrorScreen = error && hasNoData;

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

  if (showLoadingScreen) {
    return <AppLayout><LoadingScreen /></AppLayout>;
  }

  if (showErrorScreen) {
    return <AppLayout><ErrorScreen error={error} onRetry={loadData} /></AppLayout>;
  }

  const isCanvasView = view === 'canvas';
  const isInvitationsView = view === 'invitations';

  return (
    <AppLayout>
      <AppNavigation currentView={view} onOpenReleaseNotes={dialogState.releaseNotesBrowserDialog.onOpen} />
      {isCanvasView ? (
        <CanvasView
          canvasRef={canvasRef}
          selectedNodeId={selectedNodeId}
          selectedEdgeId={selectedEdgeId}
          dialogActions={dialogActions}
          dialogState={dialogState}
          addComponentToView={addComponentToView}
          switchView={switchView}
          navigateToComponent={navigateToComponent}
          navigateToCapability={navigateToCapability}
          onRemoveFromView={handleRemoveFromView}
        />
      ) : isInvitationsView ? (
        <ErrorBoundary
          fallback={(error, reset) => (
            <FeatureErrorFallback featureName="Invitations" error={error} onReset={reset} />
          )}
        >
          <Suspense fallback={<LoadingFallback message="Loading Invitations..." />}>
            <InvitationsPage />
          </Suspense>
        </ErrorBoundary>
      ) : (
        <ErrorBoundary
          fallback={(error, reset) => (
            <FeatureErrorFallback featureName="Business Domains" error={error} onReset={reset} />
          )}
        >
          <Suspense fallback={<LoadingFallback message="Loading Business Domains..." />}>
            <BusinessDomainsRouter />
          </Suspense>
        </ErrorBoundary>
      )}
      <ReleaseNotesDisplay
        showOverlay={showReleaseNotes}
        release={release}
        onDismiss={dismissReleaseNotes}
        browserIsOpen={dialogState.releaseNotesBrowserDialog.isOpen}
        onBrowserClose={dialogState.releaseNotesBrowserDialog.onClose}
      />
    </AppLayout>
  );
}

export default App;
