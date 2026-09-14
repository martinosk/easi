import { useEffect } from 'react';
import { LinearContextMenu } from './LinearContextMenu';
import { RadialContextMenu } from './RadialContextMenu';
import type { ContextMenuItem, ContextMenuVariant } from './types';

const RADIAL_MAX_ITEMS = 6;
const MENU_ROOT_SELECTOR = '[data-testid="context-menu"]';

export interface ContextMenuProps {
  x: number;
  y: number;
  items: ContextMenuItem[];
  onClose: () => void;
  variant?: ContextMenuVariant;
  title?: string;
}

function useCloseOnOutsidePointerDown(onClose: () => void) {
  useEffect(() => {
    const onPointerDown = (event: PointerEvent) => {
      const target = event.target instanceof Element ? event.target : null;
      if (!target?.closest(MENU_ROOT_SELECTOR)) onClose();
    };
    document.addEventListener('pointerdown', onPointerDown, true);
    return () => document.removeEventListener('pointerdown', onPointerDown, true);
  }, [onClose]);
}

export const ContextMenu = ({ x, y, items, onClose, variant = 'auto', title }: ContextMenuProps) => {
  useCloseOnOutsidePointerDown(onClose);
  if (items.length === 0) return null;

  const useRadial =
    variant === 'radial' || (variant === 'auto' && items.length > 0 && items.length <= RADIAL_MAX_ITEMS);

  if (useRadial) {
    return <RadialContextMenu x={x} y={y} items={items} title={title} onClose={onClose} />;
  }

  return <LinearContextMenu x={x} y={y} items={items} onClose={onClose} />;
};
