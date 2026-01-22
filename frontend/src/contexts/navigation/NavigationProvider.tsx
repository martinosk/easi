import React, { useMemo, useRef, useCallback } from 'react';
import { NavigationContext } from './context';
import type { NavigationContextValue, NavigationActions, DialogActions, ViewActions, Permissions } from './types';
import type { ComponentCanvasRef } from '../../features/canvas/components/ComponentCanvas';
import type { ViewId } from '../../api/types';
import { useCanvasNavigation } from '../../features/canvas/hooks/useCanvasNavigation';
import { useViewOperations } from '../../features/views/hooks/useViewOperations';
import { useCanvasDialogs } from '../../features/canvas/hooks/useCanvasDialogs';
import { useUserStore } from '../../store/userStore';
import { useAppStore } from '../../store/appStore';
import { useRelations } from '../../features/relations/hooks/useRelations';
import { useComponents } from '../../features/components/hooks/useComponents';

interface NavigationProviderProps {
  children: React.ReactNode;
}

export const NavigationProvider: React.FC<NavigationProviderProps> = ({ children }) => {
  const canvasRef = useRef<ComponentCanvasRef>(null);

  const hasPermission = useUserStore((state) => state.hasPermission);
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const { data: relations = [] } = useRelations();
  const { data: components = [] } = useComponents();

  const { navigateToComponent, navigateToCapability, navigateToOriginEntity } = useCanvasNavigation(canvasRef);
  const { removeComponentFromView, switchView } = useViewOperations();
  const canvasDialogs = useCanvasDialogs(selectedEdgeId, relations, components);

  const permissions = useMemo<Permissions>(() => ({
    canCreateComponent: hasPermission('components:write'),
    canCreateCapability: hasPermission('capabilities:write'),
    canCreateView: hasPermission('views:write'),
    canCreateOriginEntity: hasPermission('components:write'),
  }), [hasPermission]);

  const navigationActions = useMemo<NavigationActions>(() => ({
    navigateToComponent,
    navigateToCapability,
    navigateToOriginEntity,
    switchView: async (viewId: ViewId) => { await switchView(viewId); },
  }), [navigateToComponent, navigateToCapability, navigateToOriginEntity, switchView]);

  const dialogActions = useMemo<DialogActions>(() => ({
    addComponent: permissions.canCreateComponent ? canvasDialogs.openComponentDialog : () => {},
    addCapability: permissions.canCreateCapability ? canvasDialogs.openCapabilityDialog : () => {},
    editComponent: canvasDialogs.openEditComponentDialog,
    editCapability: canvasDialogs.openEditCapabilityDialog,
  }), [permissions.canCreateComponent, permissions.canCreateCapability, canvasDialogs]);

  const handleRemoveFromView = useCallback(() => {
    if (selectedNodeId) {
      removeComponentFromView(selectedNodeId);
    }
  }, [selectedNodeId, removeComponentFromView]);

  const viewActions = useMemo<ViewActions>(() => ({
    removeFromView: handleRemoveFromView,
  }), [handleRemoveFromView]);

  const value = useMemo<NavigationContextValue>(() => ({
    navigationActions,
    dialogActions,
    viewActions,
    permissions,
    canvasRef,
  }), [navigationActions, dialogActions, viewActions, permissions]);

  return (
    <NavigationContext.Provider value={value}>
      {children}
    </NavigationContext.Provider>
  );
};
