import React from 'react';
import { Toolbar } from './Toolbar';
import { NavigationTree } from './NavigationTree';
import { ViewSelector } from './ViewSelector';
import { ComponentCanvas, type ComponentCanvasRef } from './ComponentCanvas';
import { ComponentDetails } from './ComponentDetails';
import { RelationDetails } from './RelationDetails';
import { RealizationDetails } from './RealizationDetails';
import { CapabilityDetails } from './CapabilityDetails';
import { useAppStore } from '../store/appStore';
import type { Capability } from '../api/types';

interface MainLayoutProps {
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  onAddComponent: () => void;
  onAddCapability: () => void;
  onConnect: (source: string, target: string) => void;
  onComponentDrop: (componentId: string, x: number, y: number) => Promise<void>;
  onComponentSelect: (componentId: string) => void;
  onCapabilitySelect: (capabilityId: string) => void;
  onViewSelect: (viewId: string) => Promise<void>;
  onEditComponent: (componentId?: string) => void;
  onEditRelation: () => void;
  onEditCapability: (capability: Capability) => void;
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
  onCapabilitySelect,
  onViewSelect,
  onEditComponent,
  onEditRelation,
  onEditCapability,
  onRemoveFromView,
}) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const removeCapabilityFromCanvas = useAppStore((state) => state.removeCapabilityFromCanvas);

  const handleRemoveCapabilityFromView = () => {
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
          onCapabilitySelect={onCapabilitySelect}
          onViewSelect={onViewSelect}
          onAddComponent={onAddComponent}
          onAddCapability={onAddCapability}
          onEditCapability={onEditCapability}
          onEditComponent={onEditComponent}
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
            {selectedEdgeId && selectedEdgeId.startsWith('realization-') && (
              <RealizationDetails />
            )}
            {selectedEdgeId && !selectedEdgeId.startsWith('realization-') && !selectedEdgeId.startsWith('parent-') && (
              <RelationDetails onEdit={onEditRelation} />
            )}
            {selectedCapabilityId && (
              <CapabilityDetails onRemoveFromView={handleRemoveCapabilityFromView} />
            )}
          </div>
        )}
      </div>
    </>
  );
};
