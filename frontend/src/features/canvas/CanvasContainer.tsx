import { useMemo, useRef } from 'react';
import type { ComponentId, ViewId } from '../../api/types';
import { DockviewLayout } from '../../components/layout/DockviewLayout';
import { useKeyboardShortcuts } from '../../hooks/useKeyboardShortcuts';
import { useAppStore } from '../../store/appStore';
import { useUserStore } from '../../store/userStore';
import { useComponents } from '../components/hooks/useComponents';
import { useRelations } from '../relations/hooks/useRelations';
import { useViewOperations } from '../views/hooks/useViewOperations';
import type { ComponentCanvasRef } from './components/ComponentCanvas';
import { useCanvasDialogs } from './hooks/useCanvasDialogs';
import { useCanvasNavigation } from './hooks/useCanvasNavigation';

export default function CanvasContainer() {
  const canvasRef = useRef<ComponentCanvasRef>(null);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const hasPermission = useUserStore((state) => state.hasPermission);

  const permissions = useMemo(
    () => ({
      canCreateComponent: hasPermission('components:write'),
      canCreateCapability: hasPermission('capabilities:write'),
      canCreateView: hasPermission('views:write'),
      canCreateOriginEntity: hasPermission('components:write'),
    }),
    [hasPermission],
  );

  const { data: relations = [] } = useRelations();
  const { data: components = [] } = useComponents();
  const dialogActions = useCanvasDialogs(selectedEdgeId, relations, components);
  const { removeComponentFromView, addComponentToView, switchView } = useViewOperations();
  const { navigateToComponent, navigateToCapability, navigateToOriginEntity } = useCanvasNavigation(canvasRef);

  const handleRemoveFromView = () => {
    if (selectedNodeId) {
      removeComponentFromView(selectedNodeId);
    }
  };

  useKeyboardShortcuts({ onDelete: handleRemoveFromView });

  return (
    <DockviewLayout
      canvasRef={canvasRef}
      selectedNodeId={selectedNodeId}
      selectedEdgeId={selectedEdgeId}
      onAddComponent={permissions.canCreateComponent ? dialogActions.openComponentDialog : undefined}
      onAddCapability={permissions.canCreateCapability ? dialogActions.openCapabilityDialog : undefined}
      canCreateView={permissions.canCreateView}
      canCreateOriginEntity={permissions.canCreateOriginEntity}
      onConnect={dialogActions.openRelationDialog}
      onComponentDrop={(id, x, y) => addComponentToView(id as ComponentId, x, y)}
      onComponentSelect={navigateToComponent}
      onCapabilitySelect={navigateToCapability}
      onOriginEntitySelect={navigateToOriginEntity}
      onViewSelect={async (id) => switchView(id as ViewId)}
      onEditComponent={dialogActions.openEditComponentDialog}
      onEditRelation={dialogActions.openEditRelationDialog}
      onEditCapability={dialogActions.openEditCapabilityDialog}
      onRemoveFromView={handleRemoveFromView}
    />
  );
}
