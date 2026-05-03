import { useEffect, useRef } from 'react';
import {
  findHandleElement,
  type HandleClickEvent,
  HANDLE_CLICK_THRESHOLD_PX,
  type HandleSide,
  isClickGesture,
  readHandleSide,
  readNodeId,
} from '../utils/handleClick';

interface PendingClick {
  nodeId: string;
  side: HandleSide;
  x: number;
  y: number;
}

const PRIMARY_MOUSE_BUTTON = 0;

function readPendingClick(target: EventTarget | null, x: number, y: number): PendingClick | null {
  const handle = findHandleElement(target);
  if (!handle) return null;
  const side = readHandleSide(handle);
  const nodeId = readNodeId(handle);
  if (!side || !nodeId) return null;
  return { nodeId, side, x, y };
}

function matchesPending(target: EventTarget | null, start: PendingClick): boolean {
  const handle = findHandleElement(target);
  if (!handle) return false;
  return readNodeId(handle) === start.nodeId && readHandleSide(handle) === start.side;
}

export function useHandleClickDetection(
  onHandleClick: (event: HandleClickEvent) => void,
  thresholdPx: number = HANDLE_CLICK_THRESHOLD_PX,
): void {
  const onHandleClickRef = useRef(onHandleClick);
  useEffect(() => {
    onHandleClickRef.current = onHandleClick;
  });

  useEffect(() => {
    let pending: PendingClick | null = null;

    const onMouseDown = (e: MouseEvent) => {
      if (e.button !== PRIMARY_MOUSE_BUTTON) {
        pending = null;
        return;
      }
      pending = readPendingClick(e.target, e.clientX, e.clientY);
    };

    const onMouseUp = (e: MouseEvent) => {
      const start = pending;
      pending = null;
      if (!start) return;
      if (e.button !== PRIMARY_MOUSE_BUTTON) return;
      if (!matchesPending(e.target, start)) return;
      if (!isClickGesture({ x: start.x, y: start.y }, { x: e.clientX, y: e.clientY }, thresholdPx)) {
        return;
      }

      onHandleClickRef.current({
        nodeId: start.nodeId,
        side: start.side,
        clientX: e.clientX,
        clientY: e.clientY,
      });
    };

    document.addEventListener('mousedown', onMouseDown);
    document.addEventListener('mouseup', onMouseUp);
    return () => {
      document.removeEventListener('mousedown', onMouseDown);
      document.removeEventListener('mouseup', onMouseUp);
    };
  }, [thresholdPx]);
}
