import type { View, Component, Capability, ViewId, ComponentId } from '../../api/types';

export interface CapabilityTreeNode {
  capability: Capability;
  children: CapabilityTreeNode[];
}

export interface NavigationTreeProps {
  onComponentSelect?: (componentId: string) => void;
  onViewSelect?: (viewId: string) => void;
  onAddComponent?: () => void;
  onCapabilitySelect?: (capabilityId: string) => void;
  onAddCapability?: () => void;
  onEditCapability?: (capability: Capability) => void;
  onEditComponent?: (componentId: string) => void;
  canCreateView?: boolean;
}

export interface ViewContextMenuState {
  x: number;
  y: number;
  view: View;
}

export interface ComponentContextMenuState {
  x: number;
  y: number;
  component: Component;
}

export interface CapabilityContextMenuState {
  x: number;
  y: number;
  capability: Capability;
}

export interface EditingState {
  viewId?: ViewId;
  componentId?: ComponentId;
  name: string;
}
