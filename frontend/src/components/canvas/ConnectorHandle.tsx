import { Handle, type HandleProps } from '@xyflow/react';
import React, { useCallback, useRef } from 'react';

const CLICK_TIME_THRESHOLD_MS = 200;
const CLICK_DISTANCE_THRESHOLD_PX = 5;

export interface ConnectorClickInfo {
  nodeId: string;
  handlePosition: string;
}

export interface ConnectorHandleProps extends HandleProps {
  nodeId: string;
  onConnectorClick?: (info: ConnectorClickInfo) => void;
}

export const ConnectorHandle: React.FC<ConnectorHandleProps> = ({ nodeId, onConnectorClick, ...handleProps }) => {
  const mouseDownRef = useRef<{ x: number; y: number; time: number } | null>(null);

  const onMouseDown = useCallback((e: React.MouseEvent) => {
    mouseDownRef.current = { x: e.clientX, y: e.clientY, time: Date.now() };
  }, []);

  const onMouseUp = useCallback(
    (e: React.MouseEvent) => {
      const start = mouseDownRef.current;
      mouseDownRef.current = null;
      if (!start || !onConnectorClick) return;

      const elapsed = Date.now() - start.time;
      const dx = e.clientX - start.x;
      const dy = e.clientY - start.y;
      const distance = Math.sqrt(dx * dx + dy * dy);

      if (elapsed < CLICK_TIME_THRESHOLD_MS && distance < CLICK_DISTANCE_THRESHOLD_PX) {
        e.stopPropagation();
        onConnectorClick({ nodeId, handlePosition: handleProps.id ?? handleProps.position });
      }
    },
    [onConnectorClick, nodeId, handleProps.id, handleProps.position],
  );

  return (
    <Handle
      {...handleProps}
      onMouseDown={onMouseDown}
      onMouseUp={onMouseUp}
    />
  );
};
