import type { Node } from '@xyflow/react';

export type HandlePair = { sourceHandle: string; targetHandle: string };

const HANDLE_PAIRS: HandlePair[] = [
  { sourceHandle: 'right', targetHandle: 'left' },
  { sourceHandle: 'bottom', targetHandle: 'top' },
  { sourceHandle: 'left', targetHandle: 'right' },
  { sourceHandle: 'top', targetHandle: 'bottom' },
];

const DEFAULT_HANDLES: HandlePair = { sourceHandle: 'top', targetHandle: 'top' };

export const getNodeCenter = (node: Node): { x: number; y: number } => ({
  x: node.position.x + (node.width || 150) / 2,
  y: node.position.y + (node.height || 100) / 2,
});

export const angleToHandleIndex = (angleDegrees: number): number => {
  const normalized = ((angleDegrees % 360) + 360) % 360;
  if (normalized < 45 || normalized >= 315) return 0;
  if (normalized < 135) return 1;
  if (normalized < 225) return 2;
  return 3;
};

export const getBestHandles = (sourceNode: Node | undefined, targetNode: Node | undefined): HandlePair => {
  if (!sourceNode || !targetNode) return DEFAULT_HANDLES;

  const source = getNodeCenter(sourceNode);
  const target = getNodeCenter(targetNode);
  const angle = Math.atan2(target.y - source.y, target.x - source.x) * (180 / Math.PI);

  return HANDLE_PAIRS[angleToHandleIndex(angle)];
};
