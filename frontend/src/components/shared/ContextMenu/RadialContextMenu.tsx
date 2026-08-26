import { Portal, UnstyledButton, VisuallyHidden } from '@mantine/core';
import { useClickOutside, useMergedRef } from '@mantine/hooks';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { DotIcon } from './icons';
import classes from './RadialContextMenu.module.css';
import { nextFocusForKey } from './radialFocus';
import { clampCenter, PETAL_HALF, placePetals, type Point, radiusFor } from './radialGeometry';
import type { ContextMenuItem } from './types';

interface RadialContextMenuProps {
  x: number;
  y: number;
  items: ContextMenuItem[];
  title?: string;
  onClose: () => void;
}

const STAGGER_MS = 30;

function joinClasses(...parts: (string | false | undefined)[]): string {
  return parts.filter(Boolean).join(' ');
}

function pickHubDesc(focused: ContextMenuItem | null, itemCount: number): string | null {
  if (focused?.description) return focused.description;
  if (focused) return null;
  if (itemCount === 0) return null;
  return `${itemCount} ${itemCount === 1 ? 'action' : 'actions'}`;
}

function hubLabelClass(focused: ContextMenuItem | null): string {
  return joinClasses(
    classes.hubLabel,
    focused != null && classes.hubLabelActive,
    focused?.isDanger && classes.hubLabelDanger,
  );
}

function petalClass(item: ContextMenuItem, isFocused: boolean): string {
  return joinClasses(
    classes.petal,
    item.isDanger && classes.petalDanger,
    item.disabled && classes.petalDisabled,
    isFocused && classes.petalFocus,
  );
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
    <div className={classes.hub} aria-live="polite">
      <span className={hubLabelClass(focused)}>{label}</span>
      {desc && <span className={classes.hubDesc}>{desc}</span>}
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
  <UnstyledButton
    ref={buttonRef}
    component="button"
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
    <span className={classes.petalIcon}>{item.icon ?? <DotIcon />}</span>
    <VisuallyHidden>{item.label}</VisuallyHidden>
  </UnstyledButton>
);

export const RadialContextMenu = ({ x, y, items, title, onClose }: RadialContextMenuProps) => {
  const clickOutsideRef = useClickOutside<HTMLDivElement>(onClose);
  const focusOnMount = useCallback((el: HTMLDivElement | null) => el?.focus({ preventScroll: true }), []);
  const rootRef = useMergedRef(clickOutsideRef, focusOnMount);
  const buttonRefs = useRef<(HTMLButtonElement | null)[]>([]);
  const radius = radiusFor(items.length);
  const petals = useMemo(() => placePetals(items.length, radius), [items.length, radius]);
  const [focusIdx, setFocusIdx] = useState<number | null>(null);
  const [center, setCenter] = useState<Point>(() => clampCenter({ x, y }, radius, viewportSize()));

  useEffect(() => {
    setCenter(clampCenter({ x, y }, radius, viewportSize()));
  }, [x, y, radius]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
      return;
    }
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

  return (
    <Portal>
      <div
        ref={rootRef}
        className={classes.root}
        role="menu"
        aria-label={title ?? 'Context menu'}
        data-testid="context-menu"
        data-variant="radial"
        tabIndex={-1}
        onKeyDown={handleKeyDown}
        style={{ left: center.x, top: center.y }}
      >
        <Hub focused={focused} title={title} itemCount={items.length} />
        {items.map((item, i) => (
          <Petal
            key={item.label}
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
    </Portal>
  );
};
