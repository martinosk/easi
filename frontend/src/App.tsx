import { useEffect, useRef, useState, useCallback } from 'react';
import { useAppStore } from './store/appStore';
import { AppLayout } from './components/AppLayout';
import { LoadingScreen } from './components/LoadingScreen';
import { ErrorScreen } from './components/ErrorScreen';
import { MainLayout } from './components/MainLayout';
import { DialogManager } from './components/DialogManager';
import type { ComponentCanvasRef } from './components/ComponentCanvas';
import type { Capability } from './api/types';
import { useDialogState } from './hooks/useDialogState';
import { useRelationDialog } from './hooks/useRelationDialog';
import { useViewOperations } from './hooks/useViewOperations';
import { useCanvasNavigation } from './hooks/useCanvasNavigation';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';

function App() {
  const canvasRef = useRef<ComponentCanvasRef>(null);

  // Dialog state management
  const componentDialog = useDialogState();
  const editComponentDialog = useDialogState();
  const relationDialog = useRelationDialog();
  const editRelationDialog = useDialogState();
  const capabilityDialog = useDialogState();
  const editCapabilityDialogState = useDialogState();
  const [editCapabilityTarget, setEditCapabilityTarget] = useState<Capability | null>(null);

  const openEditCapabilityDialog = useCallback((capability: Capability) => {
    setEditCapabilityTarget(capability);
    editCapabilityDialogState.open();
  }, [editCapabilityDialogState]);

  const closeEditCapabilityDialog = useCallback(() => {
    editCapabilityDialogState.close();
    setEditCapabilityTarget(null);
  }, [editCapabilityDialogState]);

  const selectNode = useAppStore((state) => state.selectNode);

  const openEditComponentDialog = useCallback((componentId?: string) => {
    if (componentId) {
      selectNode(componentId);
    }
    editComponentDialog.open();
  }, [selectNode, editComponentDialog]);

  // Store selectors
  const loadData = useAppStore((state) => state.loadData);
  const isLoading = useAppStore((state) => state.isLoading);
  const error = useAppStore((state) => state.error);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const components = useAppStore((state) => state.components);
  const relations = useAppStore((state) => state.relations);

  // Custom hooks for operations
  const { removeComponentFromView, addComponentToView, switchView } = useViewOperations();
  const { navigateToComponent, navigateToCapability } = useCanvasNavigation(canvasRef);

  // Keyboard shortcuts
  useKeyboardShortcuts({
    onDelete: () => selectedNodeId && removeComponentFromView(selectedNodeId),
  });

  // Load data on mount
  useEffect(() => {
    loadData();
  }, [loadData]);

  // Derived state
  const selectedComponent = components.find((c) => c.id === selectedNodeId);
  const selectedRelation = relations.find((r) => r.id === selectedEdgeId);
  const hasNoData = !useAppStore.getState().components.length;

  // Loading state
  if (isLoading && hasNoData) {
    return (
      <AppLayout>
        <LoadingScreen />
      </AppLayout>
    );
  }

  // Error state
  if (error && hasNoData) {
    return (
      <AppLayout>
        <ErrorScreen error={error} onRetry={loadData} />
      </AppLayout>
    );
  }

  // Main application
  return (
    <AppLayout>
      <MainLayout
        canvasRef={canvasRef}
        selectedNodeId={selectedNodeId}
        selectedEdgeId={selectedEdgeId}
        onAddComponent={componentDialog.open}
        onAddCapability={capabilityDialog.open}
        onConnect={relationDialog.open}
        onComponentDrop={addComponentToView}
        onComponentSelect={navigateToComponent}
        onCapabilitySelect={navigateToCapability}
        onViewSelect={switchView}
        onEditComponent={openEditComponentDialog}
        onEditRelation={editRelationDialog.open}
        onEditCapability={openEditCapabilityDialog}
        onRemoveFromView={() =>
          selectedNodeId && removeComponentFromView(selectedNodeId)
        }
      />

      <DialogManager
        componentDialog={{
          isOpen: componentDialog.isOpen,
          onClose: componentDialog.close,
        }}
        relationDialog={{
          isOpen: relationDialog.isOpen,
          onClose: relationDialog.close,
          sourceComponentId: relationDialog.sourceId,
          targetComponentId: relationDialog.targetId,
        }}
        editComponentDialog={{
          isOpen: editComponentDialog.isOpen,
          onClose: editComponentDialog.close,
          component: selectedComponent || null,
        }}
        editRelationDialog={{
          isOpen: editRelationDialog.isOpen,
          onClose: editRelationDialog.close,
          relation: selectedRelation || null,
        }}
        capabilityDialog={{
          isOpen: capabilityDialog.isOpen,
          onClose: capabilityDialog.close,
        }}
        editCapabilityDialog={{
          isOpen: editCapabilityDialogState.isOpen,
          onClose: closeEditCapabilityDialog,
          capability: editCapabilityTarget,
        }}
      />
    </AppLayout>
  );
}

export default App;
