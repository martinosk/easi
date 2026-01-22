import React from 'react';
import { Toolbar } from './Toolbar';
import { NavigationTree } from '../../features/navigation';
import { ViewSelector } from '../../features/views';
import { ComponentCanvas, type ComponentCanvasRef } from '../../features/canvas';
import { useAppStore } from '../../store/appStore';
import { useRemoveCapabilityFromView } from '../../features/views/hooks/useViews';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import { DetailContentRenderer } from '../shared/DetailContentRenderer';
import type { Capability } from '../../api/types';

interface MainLayoutProps {
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  onAddComponent?: () => void;
  onAddCapability?: () => void;
  canCreateView?: boolean;
  canCreateOriginEntity?: boolean;
  onConnect: (source: string, target: string) => void;
  onComponentDrop: (componentId: string, x: number, y: number) => Promise<void>;
  onComponentSelect: (componentId: string) => void;
  onCapabilitySelect: (capabilityId: string) => void;
  onOriginEntitySelect: (nodeId: string) => void;
  onViewSelect: (viewId: string) => Promise<void>;
  onEditComponent: (componentId?: string) => void;
  onEditRelation: () => void;
  onEditCapability: (capability: Capability) => void;
  onRemoveFromView: () => void;
}

interface DetailSectionProps {
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  selectedCapabilityId: string | null;
  onEditComponent: (componentId?: string) => void;
  onEditRelation: () => void;
  onRemoveFromView: () => void;
  onRemoveCapabilityFromView: () => void;
}

const DetailSection: React.FC<DetailSectionProps> = (props) => {
  const hasSelection = props.selectedNodeId || props.selectedEdgeId || props.selectedCapabilityId;
  if (!hasSelection) return null;

  return (
    <div className="detail-section">
      <DetailContentRenderer {...props} />
    </div>
  );
};

export const MainLayout: React.FC<MainLayoutProps> = ({
  canvasRef,
  selectedNodeId,
  selectedEdgeId,
  onAddComponent,
  onAddCapability,
  canCreateView = true,
  canCreateOriginEntity = false,
  onConnect,
  onComponentDrop,
  onComponentSelect,
  onCapabilitySelect,
  onOriginEntitySelect,
  onViewSelect,
  onEditComponent,
  onEditRelation,
  onEditCapability,
  onRemoveFromView,
}) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const { currentViewId } = useCurrentView();
  const removeCapabilityFromViewMutation = useRemoveCapabilityFromView();

  const handleRemoveCapabilityFromView = () => {
    if (selectedCapabilityId && currentViewId) {
      removeCapabilityFromViewMutation.mutate({
        viewId: currentViewId,
        capabilityId: selectedCapabilityId,
      });
    }
  };

  return (
    <>
      <Toolbar />

      <div className="main-content">
        <NavigationTree
          onComponentSelect={onComponentSelect}
          onCapabilitySelect={onCapabilitySelect}
          onOriginEntitySelect={onOriginEntitySelect}
          onViewSelect={onViewSelect}
          onAddComponent={onAddComponent}
          onAddCapability={onAddCapability}
          onEditCapability={onEditCapability}
          onEditComponent={onEditComponent}
          canCreateView={canCreateView}
          canCreateOriginEntity={canCreateOriginEntity}
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
