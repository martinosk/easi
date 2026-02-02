import type { View, Component, Capability, ViewId, ComponentId } from '../../api/types';
import type { TreeSelectedItem } from './hooks/useTreeMultiSelect';

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
  onOriginEntitySelect?: (nodeId: string) => void;
  canCreateView?: boolean;
  canCreateOriginEntity?: boolean;
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

export interface TreeMultiSelectProps {
  isMultiSelected: (id: string) => boolean;
  handleItemClick: (item: TreeSelectedItem, sectionId: string, visibleItems: TreeSelectedItem[], event: React.MouseEvent) => 'multi' | 'single';
  handleContextMenu: (event: React.MouseEvent, itemId: string, selectedItems: TreeSelectedItem[]) => boolean;
  handleDragStart: (event: React.DragEvent, itemId: string) => boolean;
  selectedItems: TreeSelectedItem[];
}
