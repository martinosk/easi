import { useEffect, useRef, useState } from 'react';
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
import type { ComponentCanvasRef } from './features/canvas/components/ComponentCanvas';
import { useDialogManagement } from './hooks/useDialogManagement';
import { useViewOperations } from './hooks/useViewOperations';
import { useCanvasNavigation } from './hooks/useCanvasNavigation';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useReleaseNotes } from './hooks/useReleaseNotes';

function App() {
  const canvasRef = useRef<ComponentCanvasRef>(null);
  const [currentView, setCurrentView] = useState<'canvas' | 'business-domains'>('canvas');

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

  useKeyboardShortcuts({
    onDelete: () => selectedNodeId && removeComponentFromView(selectedNodeId),
  });

  useEffect(() => {
    loadData();
  }, [loadData]);

  const hasNoData = !useAppStore.getState().components.length;

  if (isLoading && hasNoData) {
    return (
      <AppLayout>
        <LoadingScreen />
      </AppLayout>
    );
  }

  if (error && hasNoData) {
    return (
      <AppLayout>
        <ErrorScreen error={error} onRetry={loadData} />
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <AppNavigation
        onViewChange={setCurrentView}
        onOpenReleaseNotes={dialogState.releaseNotesBrowserDialog.onOpen}
      />

      {currentView === 'canvas' ? (
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
            onRemoveFromView={() => selectedNodeId && removeComponentFromView(selectedNodeId)}
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
      ) : (
        <BusinessDomainsRouter />
      )}

      {showReleaseNotes && release && (
        <ReleaseNotesOverlay
          isOpen={showReleaseNotes}
          release={release}
          onDismiss={dismissReleaseNotes}
        />
      )}

      <ReleaseNotesBrowser
        isOpen={dialogState.releaseNotesBrowserDialog.isOpen}
        onClose={dialogState.releaseNotesBrowserDialog.onClose}
      />
    </AppLayout>
  );
}

export default App;
