import { useEffect, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import './ContextMenu.css';
import { DotIcon } from './icons';
import { clampCenter, PETAL_HALF, placePetals, type Point, radiusFor } from './radialGeometry';
import { nextFocusForKey } from './radialFocus';
import type { ContextMenuItem } from './types';
import { useContextMenuController } from './useContextMenuController';

interface RadialContextMenuProps {
  x: number;
  y: number;
  items: ContextMenuItem[];
  title?: string;
  onClose: () => void;
}

const STAGGER_MS = 30;

function pickHubDesc(focused: ContextMenuItem | null, itemCount: number): string | null {
  if (focused?.description) return focused.description;
  if (focused) return null;
  if (itemCount === 0) return null;
  return `${itemCount} ${itemCount === 1 ? 'action' : 'actions'}`;
}

function hubLabelClass(focused: ContextMenuItem | null): string {
  const parts = ['ctx-menu__hub-label'];
  if (focused) parts.push('ctx-menu__hub-label--active');
  if (focused?.isDanger) parts.push('ctx-menu__hub-label--danger');
  return parts.join(' ');
}

function petalClass(item: ContextMenuItem, isFocused: boolean): string {
  const parts = ['ctx-menu__petal'];
  if (item.isDanger) parts.push('ctx-menu__petal--danger');
  if (item.disabled) parts.push('ctx-menu__petal--disabled');
  if (isFocused) parts.push('ctx-menu__petal--focus');
  return parts.join(' ');
}

function viewportSize() {
  return { width: window.innerWidth, height: window.innerHeight };
}

interface HubProps {
  focused: ContextMenuItem | null;
  title: string | undefined;
  itemCount: number;
}

const Hub = ({ focused, title, itemCount }: HubProps) => {
  const label = focused?.label ?? title ?? 'Actions';
  const desc = pickHubDesc(focused, itemCount);
  return (
    <div className="ctx-menu__hub" aria-live="polite">
      <span className={hubLabelClass(focused)}>{label}</span>
      {desc && <span className="ctx-menu__hub-desc">{desc}</span>}
    </div>
  );
};

interface PetalProps {
  item: ContextMenuItem;
  index: number;
  position: Point;
  isFocused: boolean;
  buttonRef: (el: HTMLButtonElement | null) => void;
  onActivate: () => void;
  onFocus: () => void;
  onMouseEnter: () => void;
  onMouseLeave: () => void;
}

const Petal = ({
  item,
  index,
  position,
  isFocused,
  buttonRef,
  onActivate,
  onFocus,
  onMouseEnter,
  onMouseLeave,
}: PetalProps) => (
  <button
    ref={buttonRef}
    type="button"
    role="menuitem"
    aria-label={item.ariaLabel ?? item.label}
    className={petalClass(item, isFocused)}
    tabIndex={isFocused ? 0 : -1}
    disabled={item.disabled}
    aria-disabled={item.disabled}
    onMouseEnter={onMouseEnter}
    onMouseLeave={onMouseLeave}
    onFocus={onFocus}
    onClick={onActivate}
    style={{
      transform: `translate(${position.x - PETAL_HALF}px, ${position.y - PETAL_HALF}px)`,
      animationDelay: `${index * STAGGER_MS}ms`,
    }}
  >
    <span className="ctx-menu__petal-icon">{item.icon ?? <DotIcon />}</span>
    <span className="ctx-menu__sr">{item.label}</span>
  </button>
);

export const RadialContextMenu = ({ x, y, items, title, onClose }: RadialContextMenuProps) => {
  const ref = useContextMenuController<HTMLDivElement>(onClose);
  const buttonRefs = useRef<(HTMLButtonElement | null)[]>([]);
  const radius = radiusFor(items.length);
  const petals = useMemo(() => placePetals(items.length, radius), [items.length, radius]);
  const [focusIdx, setFocusIdx] = useState<number | null>(null);
  const [center, setCenter] = useState<Point>(() => clampCenter({ x, y }, radius, viewportSize()));

  useEffect(() => {
    setCenter(clampCenter({ x, y }, radius, viewportSize()));
  }, [x, y, radius]);

  useEffect(() => {
    ref.current?.focus({ preventScroll: true });
  }, [ref]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (items.length === 0) return;
    const next = nextFocusForKey(
      { key: e.key, shiftKey: e.shiftKey },
      { current: focusIdx, count: items.length },
    );
    if (next === null) return;
    e.preventDefault();
    setFocusIdx(next);
    buttonRefs.current[next]?.focus();
  };

  const focused = focusIdx != null ? items[focusIdx] : null;

  const node = (
    <div
      ref={ref}
      className="ctx-menu ctx-menu--radial"
      role="menu"
      aria-label={title ?? 'Context menu'}
      tabIndex={-1}
      onKeyDown={handleKeyDown}
      style={{ left: center.x, top: center.y }}
    >
      <Hub focused={focused} title={title} itemCount={items.length} />
      {items.map((item, i) => (
        <Petal
          key={i}
          item={item}
          index={i}
          position={petals[i]}
          isFocused={focusIdx === i}
          buttonRef={(el) => {
            buttonRefs.current[i] = el;
          }}
          onActivate={() => {
            if (item.disabled) return;
            item.onClick();
            onClose();
          }}
          onFocus={() => setFocusIdx(i)}
          onMouseEnter={() => setFocusIdx(i)}
          onMouseLeave={() => setFocusIdx(null)}
        />
      ))}
    </div>
  );

  return createPortal(node, document.body);
};
