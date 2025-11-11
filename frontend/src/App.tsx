import { useEffect, useState, useRef, useCallback } from 'react';
import { Toaster } from 'react-hot-toast';
import { useAppStore } from './store/appStore';
import { Toolbar } from './components/Toolbar';
import { ComponentCanvas, type ComponentCanvasRef } from './components/ComponentCanvas';
import { CreateComponentDialog } from './components/CreateComponentDialog';
import { CreateRelationDialog } from './components/CreateRelationDialog';
import { EditComponentDialog } from './components/EditComponentDialog';
import { EditRelationDialog } from './components/EditRelationDialog';
import { ComponentDetails } from './components/ComponentDetails';
import { RelationDetails } from './components/RelationDetails';
import { NavigationTree } from './components/NavigationTree';
import { ViewSelector } from './components/ViewSelector';
import apiClient from './api/client';

function App() {
  const canvasRef = useRef<ComponentCanvasRef>(null);
  const [isComponentDialogOpen, setIsComponentDialogOpen] = useState(false);
  const [isEditComponentDialogOpen, setIsEditComponentDialogOpen] = useState(false);
  const [isRelationDialogOpen, setIsRelationDialogOpen] = useState(false);
  const [isEditRelationDialogOpen, setIsEditRelationDialogOpen] = useState(false);
  const [relationSource, setRelationSource] = useState<string | undefined>();
  const [relationTarget, setRelationTarget] = useState<string | undefined>();

  const loadData = useAppStore((state) => state.loadData);
  const isLoading = useAppStore((state) => state.isLoading);
  const error = useAppStore((state) => state.error);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const selectNode = useAppStore((state) => state.selectNode);
  const components = useAppStore((state) => state.components);
  const relations = useAppStore((state) => state.relations);
  const currentView = useAppStore((state) => state.currentView);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleRemoveFromView = useCallback(async () => {
    if (!currentView || !selectedNodeId) return;

    try {
      await apiClient.removeComponentFromView(currentView.id, selectedNodeId);
      // Reload the view to get updated component list
      const updatedView = await apiClient.getViewById(currentView.id);
      useAppStore.setState({ currentView: updatedView });
      // Clear selection since component is no longer in view
      useAppStore.getState().clearSelection();
    } catch (error) {
      console.error('Failed to remove component from view:', error);
    }
  }, [currentView, selectedNodeId]);

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Delete key to remove component from view
      if (event.key === 'Delete' && selectedNodeId && currentView) {
        const isInCurrentView = currentView.components.some(
          (vc) => vc.componentId === selectedNodeId
        );
        if (isInCurrentView) {
          handleRemoveFromView();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedNodeId, currentView, handleRemoveFromView]);

  const handleAddComponent = () => {
    setIsComponentDialogOpen(true);
  };

  const handleFitView = () => {
    // This will be implemented via React Flow imperative handle if needed
    // For now, we'll just log it
    console.log('Fit view requested');
  };

  const handleConnect = (source: string, target: string) => {
    setRelationSource(source);
    setRelationTarget(target);
    setIsRelationDialogOpen(true);
  };

  const handleCloseRelationDialog = () => {
    setIsRelationDialogOpen(false);
    setRelationSource(undefined);
    setRelationTarget(undefined);
  };

  const handleEditComponent = () => {
    setIsEditComponentDialogOpen(true);
  };

  const handleEditRelation = () => {
    setIsEditRelationDialogOpen(true);
  };

  const handleComponentSelect = (componentId: string) => {
    // Check if component is in current view
    const isInCurrentView = currentView?.components.some(
      vc => vc.componentId === componentId
    );

    if (isInCurrentView) {
      selectNode(componentId);
      // Pan canvas to show component
      canvasRef.current?.centerOnNode(componentId);
    }
  };

  const switchView = useAppStore((state) => state.switchView);

  const handleViewSelect = async (viewId: string) => {
    try {
      await switchView(viewId);
    } catch (error) {
      console.error('Failed to switch view:', error);
    }
  };

  const handleComponentDrop = async (componentId: string, x: number, y: number) => {
    if (!currentView) return;

    try {
      await apiClient.addComponentToView(currentView.id, {
        componentId,
        x,
        y,
      });
      // Reload the view to get updated component list
      const updatedView = await apiClient.getViewById(currentView.id);
      useAppStore.setState({ currentView: updatedView });
    } catch (error) {
      console.error('Failed to add component to view:', error);
    }
  };

  const selectedComponent = components.find((c) => c.id === selectedNodeId);
  const selectedRelation = relations.find((r) => r.id === selectedEdgeId);

  if (isLoading && !useAppStore.getState().components.length) {
    return (
      <div className="app-container">
        <div className="loading-container">
          <div className="loading-spinner"></div>
          <p>Loading component modeler...</p>
        </div>
        <Toaster position="top-right" />
      </div>
    );
  }

  if (error && !useAppStore.getState().components.length) {
    return (
      <div className="app-container">
        <div className="error-container">
          <h2>Error Loading Data</h2>
          <p>{error}</p>
          <button className="btn btn-primary" onClick={loadData}>
            Retry
          </button>
        </div>
        <Toaster position="top-right" />
      </div>
    );
  }

  return (
    <div className="app-container">
      <Toolbar onAddComponent={handleAddComponent} onFitView={handleFitView} />

      <div className="main-content">
        <NavigationTree
          onComponentSelect={handleComponentSelect}
          onViewSelect={handleViewSelect}
        />

        <div className="canvas-section">
          <ViewSelector />
          <ComponentCanvas ref={canvasRef} onConnect={handleConnect} onComponentDrop={handleComponentDrop} />
        </div>

        {(selectedNodeId || selectedEdgeId) && (
          <div className="detail-section">
            {selectedNodeId && <ComponentDetails onEdit={handleEditComponent} onRemoveFromView={handleRemoveFromView} />}
            {selectedEdgeId && <RelationDetails onEdit={handleEditRelation} />}
          </div>
        )}
      </div>

      <CreateComponentDialog
        isOpen={isComponentDialogOpen}
        onClose={() => setIsComponentDialogOpen(false)}
      />

      <CreateRelationDialog
        isOpen={isRelationDialogOpen}
        onClose={handleCloseRelationDialog}
        sourceComponentId={relationSource}
        targetComponentId={relationTarget}
      />

      <EditComponentDialog
        isOpen={isEditComponentDialogOpen}
        onClose={() => setIsEditComponentDialogOpen(false)}
        component={selectedComponent || null}
      />

      <EditRelationDialog
        isOpen={isEditRelationDialogOpen}
        onClose={() => setIsEditRelationDialogOpen(false)}
        relation={selectedRelation || null}
      />

      <Toaster
        position="top-right"
        toastOptions={{
          duration: 3000,
          style: {
            background: '#363636',
            color: '#fff',
          },
          success: {
            iconTheme: {
              primary: '#10b981',
              secondary: '#fff',
            },
          },
          error: {
            iconTheme: {
              primary: '#ef4444',
              secondary: '#fff',
            },
          },
        }}
      />
    </div>
  );
}

export default App;
