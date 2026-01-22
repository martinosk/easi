import type { Capability, ViewId } from '../../api/types';
import type { ComponentCanvasRef } from '../../features/canvas/components/ComponentCanvas';

export interface NavigationActions {
  navigateToComponent: (componentId: string) => void;
  navigateToCapability: (capabilityId: string) => void;
  navigateToOriginEntity: (nodeId: string) => void;
  switchView: (viewId: ViewId) => Promise<void>;
}

export interface DialogActions {
  addComponent: () => void;
  addCapability: () => void;
  editComponent: (componentId?: string) => void;
  editCapability: (capability: Capability) => void;
}

export interface ViewActions {
  removeFromView: () => void;
}

export interface Permissions {
  canCreateView: boolean;
  canCreateOriginEntity: boolean;
  canCreateComponent: boolean;
  canCreateCapability: boolean;
}

export interface NavigationContextValue {
  navigationActions: NavigationActions;
  dialogActions: DialogActions;
  viewActions: ViewActions;
  permissions: Permissions;
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
}
