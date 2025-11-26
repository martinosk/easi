import React from 'react';
import { Toolbar } from './Toolbar';
import { NavigationTree } from '../../features/navigation';
import { ViewSelector } from '../../features/views';
import { ComponentCanvas, type ComponentCanvasRef } from '../../features/canvas';
import { ComponentDetails } from '../../features/components';
import { RelationDetails, RealizationDetails } from '../../features/relations';
import { CapabilityDetails } from '../../features/capabilities';
import { useAppStore } from '../../store/appStore';
import type { Capability } from '../../api/types';

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
  onOpenReleaseNotes?: () => void;
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
  onOpenReleaseNotes,
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
      <Toolbar onOpenReleaseNotes={onOpenReleaseNotes} />

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
