export type HandleSide = 'top' | 'right' | 'bottom' | 'left';

export interface HandleClickEvent {
  nodeId: string;
  side: HandleSide;
  clientX: number;
  clientY: number;
}

export interface PointerCoords {
  x: number;
  y: number;
}

export const HANDLE_CLICK_THRESHOLD_PX = 5;

const HANDLE_CLASS = 'react-flow__handle';
const NODE_ID_ATTR_ON_HANDLE = 'data-nodeid';
const NODE_ID_ATTR_ON_PARENT = 'data-id';
const HANDLE_POS_ATTR = 'data-handlepos';
const VALID_SIDES: ReadonlySet<HandleSide> = new Set(['top', 'right', 'bottom', 'left']);

export function isClickGesture(
  down: PointerCoords,
  up: PointerCoords,
  thresholdPx: number = HANDLE_CLICK_THRESHOLD_PX,
): boolean {
  const dx = up.x - down.x;
  const dy = up.y - down.y;
  return Math.sqrt(dx * dx + dy * dy) <= thresholdPx;
}

export function findHandleElement(target: EventTarget | null): HTMLElement | null {
  if (!(target instanceof Element)) return null;
  const handle = target.closest(`.${HANDLE_CLASS}`);
  return handle instanceof HTMLElement ? handle : null;
}

export function readHandleSide(handle: HTMLElement): HandleSide | null {
  const value = handle.getAttribute(HANDLE_POS_ATTR);
  if (value && (VALID_SIDES as Set<string>).has(value)) {
    return value as HandleSide;
  }
  return null;
}

export function readNodeId(handle: HTMLElement): string | null {
  const onHandle = handle.getAttribute(NODE_ID_ATTR_ON_HANDLE);
  if (onHandle) return onHandle;
  const parent = handle.parentElement?.closest(`[${NODE_ID_ATTR_ON_PARENT}]`);
  return parent?.getAttribute(NODE_ID_ATTR_ON_PARENT) ?? null;
}
