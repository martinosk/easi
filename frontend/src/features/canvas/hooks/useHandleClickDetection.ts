import { useEffect, useRef } from 'react';
import {
  findHandleElement,
  type HandleClickEvent,
  HANDLE_CLICK_THRESHOLD_PX,
  isClickGesture,
  readHandleSide,
  readNodeId,
} from '../utils/handleClick';

interface PendingClick {
  handle: HTMLElement;
  x: number;
  y: number;
}

export function useHandleClickDetection(
  rootRef: React.RefObject<HTMLElement | null> | null,
  onHandleClick: (event: HandleClickEvent) => void,
  thresholdPx: number = HANDLE_CLICK_THRESHOLD_PX,
): void {
  const onHandleClickRef = useRef(onHandleClick);
  useEffect(() => {
    onHandleClickRef.current = onHandleClick;
  });

  useEffect(() => {
    const root: EventTarget = rootRef?.current ?? document;
    if (rootRef && !rootRef.current) return;

    let pending: PendingClick | null = null;

    const onMouseDown = (e: MouseEvent) => {
      const handle = findHandleElement(e.target);
      pending = handle ? { handle, x: e.clientX, y: e.clientY } : null;
    };

    const onMouseUp = (e: MouseEvent) => {
      const start = pending;
      pending = null;
      if (!start) return;

      const upHandle = findHandleElement(e.target);
      if (upHandle !== start.handle) return;

      if (!isClickGesture({ x: start.x, y: start.y }, { x: e.clientX, y: e.clientY }, thresholdPx)) {
        return;
      }

      const side = readHandleSide(start.handle);
      const nodeId = readNodeId(start.handle);
      if (!side || !nodeId) return;

      onHandleClickRef.current({
        nodeId,
        side,
        clientX: e.clientX,
        clientY: e.clientY,
      });
    };

    root.addEventListener('mousedown', onMouseDown);
    root.addEventListener('mouseup', onMouseUp);
    return () => {
      root.removeEventListener('mousedown', onMouseDown);
      root.removeEventListener('mouseup', onMouseUp);
    };
  }, [rootRef, thresholdPx]);
}
