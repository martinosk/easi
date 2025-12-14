import { useEffect, useRef, useState } from 'react';
import toast from 'react-hot-toast';
import { useAppStore } from './store/appStore';
import { AppLayout } from './components/layout/AppLayout';
import { AppNavigation } from './components/layout/AppNavigation';
import { LoadingScreen } from './components/shared/LoadingScreen';
import { ErrorScreen } from './components/shared/ErrorScreen';
import { MainLayout } from './components/layout/MainLayout';
import { DialogManager } from './components/shared/DialogManager';
import { ReleaseNotesOverlay } from './contexts/releases/components/ReleaseNotesOverlay';
import { ReleaseNotesBrowser } from './contexts/releases/components/ReleaseNotesBrowser';
import { BusinessDomainsRouter } from './features/business-domains';
import { InvitationsPage } from './features/invitations';
import type { ComponentCanvasRef } from './features/canvas/components/ComponentCanvas';
import { useDialogManagement } from './hooks/useDialogManagement';
import { useViewOperations } from './hooks/useViewOperations';
import { useCanvasNavigation } from './hooks/useCanvasNavigation';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useReleaseNotes } from './hooks/useReleaseNotes';
import { useSessionCheck } from './hooks/useSessionCheck';
import type { Release } from './api/types';

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
      <MainLayout
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

function App() {
  const canvasRef = useRef<ComponentCanvasRef>(null);
  const [currentView, setCurrentView] = useState<'canvas' | 'business-domains' | 'invitations'>('canvas');

  const { authError } = useAuthErrorHandler();
  const { isLoading: isSessionLoading, isAuthenticated } = useSessionCheck();

  const loadData = useAppStore((state) => state.loadData);
  const isLoading = useAppStore((state) => state.isLoading);
  const error = useAppStore((state) => state.error);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const relations = useAppStore((state) => state.relations);

  const { showOverlay: showReleaseNotes, release, dismiss: dismissReleaseNotes } = useReleaseNotes();
  const { state: dialogState, actions: dialogActions } = useDialogManagement(selectedEdgeId, relations);
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

  const hasNoData = !useAppStore.getState().components.length;
  const isInitializing = isSessionLoading || !isAuthenticated;
  const isLoadingInitialData = isLoading && hasNoData;
  const showLoadingScreen = isInitializing || isLoadingInitialData;
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

  const isCanvasView = currentView === 'canvas';
  const isInvitationsView = currentView === 'invitations';

  return (
    <AppLayout>
      <AppNavigation onViewChange={setCurrentView} onOpenReleaseNotes={dialogState.releaseNotesBrowserDialog.onOpen} />
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
        <InvitationsPage />
      ) : (
        <BusinessDomainsRouter />
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
