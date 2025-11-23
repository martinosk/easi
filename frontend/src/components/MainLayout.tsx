import React from 'react';
import { Toolbar } from './Toolbar';
import { NavigationTree } from './NavigationTree';
import { ViewSelector } from './ViewSelector';
import { ComponentCanvas, type ComponentCanvasRef } from './ComponentCanvas';
import { ComponentDetails } from './ComponentDetails';
import { RelationDetails } from './RelationDetails';
import { CapabilityDetails } from './CapabilityDetails';
import { useAppStore } from '../store/appStore';

interface MainLayoutProps {
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  onAddComponent: () => void;
  onAddCapability: () => void;
  onConnect: (source: string, target: string) => void;
  onComponentDrop: (componentId: string, x: number, y: number) => Promise<void>;
  onComponentSelect: (componentId: string) => void;
  onViewSelect: (viewId: string) => Promise<void>;
  onEditComponent: () => void;
  onEditRelation: () => void;
  onRemoveFromView: () => void;
}

export const MainLayout: React.FC<MainLayoutProps> = ({
  canvasRef,
  selectedNodeId,
  selectedEdgeId,
  onAddComponent,
  onAddCapability,
  onConnect,
  onComponentDrop,
  onComponentSelect,
  onViewSelect,
  onEditComponent,
  onEditRelation,
  onRemoveFromView,
}) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const removeCapabilityFromCanvas = useAppStore((state) => state.removeCapabilityFromCanvas);

  const handleRemoveCapabilityFromCanvas = () => {
    if (selectedCapabilityId) {
      removeCapabilityFromCanvas(selectedCapabilityId);
    }
  };

  return (
    <>
      <Toolbar />

      <div className="main-content">
        <NavigationTree
          onComponentSelect={onComponentSelect}
          onViewSelect={onViewSelect}
          onAddComponent={onAddComponent}
          onAddCapability={onAddCapability}
        />

        <div className="canvas-section">
          <ViewSelector />
          <ComponentCanvas
            ref={canvasRef}
            onConnect={onConnect}
            onComponentDrop={onComponentDrop}
          />
        </div>

        {(selectedNodeId || selectedEdgeId || selectedCapabilityId) && (
          <div className="detail-section">
            {selectedNodeId && (
              <ComponentDetails
                onEdit={onEditComponent}
                onRemoveFromView={onRemoveFromView}
              />
            )}
            {selectedEdgeId && <RelationDetails onEdit={onEditRelation} />}
            {selectedCapabilityId && (
              <CapabilityDetails onRemoveFromCanvas={handleRemoveCapabilityFromCanvas} />
            )}
          </div>
        )}
      </div>
    </>
  );
};
