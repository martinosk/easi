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
}

const isRealizationEdge = (edgeId: string): boolean => edgeId.startsWith('realization-');
const isParentEdge = (edgeId: string): boolean => edgeId.startsWith('parent-');
const isRelationEdge = (edgeId: string): boolean => !isRealizationEdge(edgeId) && !isParentEdge(edgeId);

interface DetailSectionProps {
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  selectedCapabilityId: string | null;
  onEditComponent: (componentId?: string) => void;
  onEditRelation: () => void;
  onRemoveFromView: () => void;
  onRemoveCapabilityFromView: () => void;
}

const renderDetailContent = (
  selectedNodeId: string | null,
  selectedEdgeId: string | null,
  selectedCapabilityId: string | null,
  onEditComponent: (componentId?: string) => void,
  onEditRelation: () => void,
  onRemoveFromView: () => void,
  onRemoveCapabilityFromView: () => void
): React.ReactNode => {
  if (selectedNodeId) {
    return <ComponentDetails onEdit={onEditComponent} onRemoveFromView={onRemoveFromView} />;
  }
  if (selectedEdgeId && isRealizationEdge(selectedEdgeId)) {
    return <RealizationDetails />;
  }
  if (selectedEdgeId && isRelationEdge(selectedEdgeId)) {
    return <RelationDetails onEdit={onEditRelation} />;
  }
  if (selectedCapabilityId) {
    return <CapabilityDetails onRemoveFromView={onRemoveCapabilityFromView} />;
  }
  return null;
};

const DetailSection: React.FC<DetailSectionProps> = ({
  selectedNodeId,
  selectedEdgeId,
  selectedCapabilityId,
  onEditComponent,
  onEditRelation,
  onRemoveFromView,
  onRemoveCapabilityFromView,
}) => {
  const hasSelection = selectedNodeId || selectedEdgeId || selectedCapabilityId;
  if (!hasSelection) return null;

  return (
    <div className="detail-section">
      {renderDetailContent(
        selectedNodeId,
        selectedEdgeId,
        selectedCapabilityId,
        onEditComponent,
        onEditRelation,
        onRemoveFromView,
        onRemoveCapabilityFromView
      )}
    </div>
  );
};

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

        <DetailSection
          selectedNodeId={selectedNodeId}
          selectedEdgeId={selectedEdgeId}
          selectedCapabilityId={selectedCapabilityId}
          onEditComponent={onEditComponent}
          onEditRelation={onEditRelation}
          onRemoveFromView={onRemoveFromView}
          onRemoveCapabilityFromView={handleRemoveCapabilityFromView}
        />
      </div>
    </>
  );
};
