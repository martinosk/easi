import { useEffect, useState } from 'react';
import { Toaster } from 'react-hot-toast';
import { useAppStore } from './store/appStore';
import { Toolbar } from './components/Toolbar';
import { ComponentCanvas } from './components/ComponentCanvas';
import { CreateComponentDialog } from './components/CreateComponentDialog';
import { CreateRelationDialog } from './components/CreateRelationDialog';
import { EditComponentDialog } from './components/EditComponentDialog';
import { EditRelationDialog } from './components/EditRelationDialog';
import { ComponentDetails } from './components/ComponentDetails';
import { RelationDetails } from './components/RelationDetails';

function App() {
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
  const components = useAppStore((state) => state.components);
  const relations = useAppStore((state) => state.relations);

  useEffect(() => {
    loadData();
  }, [loadData]);

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
        <div className="canvas-section">
          <ComponentCanvas onConnect={handleConnect} />
        </div>

        {(selectedNodeId || selectedEdgeId) && (
          <div className="detail-section">
            {selectedNodeId && <ComponentDetails onEdit={handleEditComponent} />}
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
