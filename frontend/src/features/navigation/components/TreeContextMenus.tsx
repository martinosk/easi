import React from 'react';
import { ContextMenu } from '../../../components/shared/ContextMenu';
import type { ViewContextMenuState, ComponentContextMenuState, CapabilityContextMenuState } from '../types';
import type { OriginEntityContextMenuState } from '../hooks/useTreeContextMenus';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';

interface TreeContextMenusProps {
  viewContextMenu: ViewContextMenuState | null;
  componentContextMenu: ComponentContextMenuState | null;
  capabilityContextMenu: CapabilityContextMenuState | null;
  originEntityContextMenu: OriginEntityContextMenuState | null;
  getViewContextMenuItems: (menu: ViewContextMenuState) => ContextMenuItem[];
  getComponentContextMenuItems: (menu: ComponentContextMenuState) => ContextMenuItem[];
  getCapabilityContextMenuItems: (menu: CapabilityContextMenuState) => ContextMenuItem[];
  getOriginEntityContextMenuItems: (menu: OriginEntityContextMenuState) => ContextMenuItem[];
  setViewContextMenu: (menu: ViewContextMenuState | null) => void;
  setComponentContextMenu: (menu: ComponentContextMenuState | null) => void;
  setCapabilityContextMenu: (menu: CapabilityContextMenuState | null) => void;
  setOriginEntityContextMenu: (menu: OriginEntityContextMenuState | null) => void;
}

export const TreeContextMenus: React.FC<TreeContextMenusProps> = ({
  viewContextMenu,
  componentContextMenu,
  capabilityContextMenu,
  originEntityContextMenu,
  getViewContextMenuItems,
  getComponentContextMenuItems,
  getCapabilityContextMenuItems,
  getOriginEntityContextMenuItems,
  setViewContextMenu,
  setComponentContextMenu,
  setCapabilityContextMenu,
  setOriginEntityContextMenu,
}) => (
  <>
    {viewContextMenu && (
      <ContextMenu
        x={viewContextMenu.x}
        y={viewContextMenu.y}
        items={getViewContextMenuItems(viewContextMenu)}
        onClose={() => setViewContextMenu(null)}
      />
    )}
    {componentContextMenu && (
      <ContextMenu
        x={componentContextMenu.x}
        y={componentContextMenu.y}
        items={getComponentContextMenuItems(componentContextMenu)}
        onClose={() => setComponentContextMenu(null)}
      />
    )}
    {capabilityContextMenu && (
      <ContextMenu
        x={capabilityContextMenu.x}
        y={capabilityContextMenu.y}
        items={getCapabilityContextMenuItems(capabilityContextMenu)}
        onClose={() => setCapabilityContextMenu(null)}
      />
    )}
    {originEntityContextMenu && (
      <ContextMenu
        x={originEntityContextMenu.x}
        y={originEntityContextMenu.y}
        items={getOriginEntityContextMenuItems(originEntityContextMenu)}
        onClose={() => setOriginEntityContextMenu(null)}
      />
    )}
  </>
);
