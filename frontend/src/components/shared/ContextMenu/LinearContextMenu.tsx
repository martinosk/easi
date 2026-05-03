import { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import './ContextMenu.css';
import type { ContextMenuItem } from './types';
import { useContextMenuController } from './useContextMenuController';

interface LinearContextMenuProps {
  x: number;
  y: number;
  items: ContextMenuItem[];
  onClose: () => void;
}

const VIEWPORT_PADDING = 10;

export const LinearContextMenu = ({ x, y, items, onClose }: LinearContextMenuProps) => {
  const ref = useContextMenuController<HTMLDivElement>(onClose);
  const [pos, setPos] = useState({ x, y });

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    let nx = x;
    let ny = y;
    if (x + rect.width > window.innerWidth) nx = window.innerWidth - rect.width - VIEWPORT_PADDING;
    if (y + rect.height > window.innerHeight) ny = window.innerHeight - rect.height - VIEWPORT_PADDING;
    setPos({ x: nx, y: ny });
  }, [x, y, ref]);

  const handleClick = (item: ContextMenuItem) => {
    if (item.disabled) return;
    item.onClick();
    onClose();
  };

  const itemClassName = (item: ContextMenuItem): string => {
    const parts = ['ctx-menu__item'];
    if (item.isDanger) parts.push('ctx-menu__item--danger');
    if (item.disabled) parts.push('ctx-menu__item--disabled');
    return parts.join(' ');
  };

  const node = (
    <div
      ref={ref}
      className="ctx-menu ctx-menu--linear"
      role="menu"
      aria-label="Context menu"
      style={{ left: pos.x, top: pos.y }}
    >
      {items.map((item, index) => (
        <button
          key={index}
          type="button"
          className={itemClassName(item)}
          onClick={() => handleClick(item)}
          role="menuitem"
          aria-label={item.ariaLabel ?? item.label}
          disabled={item.disabled}
          aria-disabled={item.disabled}
        >
          {item.icon != null && <span className="ctx-menu__icon">{item.icon}</span>}
          <span className="ctx-menu__text">
            <span className="ctx-menu__label">{item.label}</span>
            {item.description && <span className="ctx-menu__desc">{item.description}</span>}
          </span>
        </button>
      ))}
    </div>
  );

  return createPortal(node, document.body);
};
