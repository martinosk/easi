import type { Position } from '../../../api/types';
import type { HandleSide } from './handleClick';

const NODE_WIDTH_PX = 220;
const NODE_HEIGHT_PX = 120;
const GAP_PX = 60;

const X_OFFSET = NODE_WIDTH_PX + GAP_PX;
const Y_OFFSET = NODE_HEIGHT_PX + GAP_PX;

const OFFSETS: Record<HandleSide, Position> = {
  top: { x: 0, y: -Y_OFFSET },
  right: { x: X_OFFSET, y: 0 },
  bottom: { x: 0, y: Y_OFFSET },
  left: { x: -X_OFFSET, y: 0 },
};

export function computeOffsetPosition(source: Position, side: HandleSide): Position {
  const offset = OFFSETS[side];
  return { x: source.x + offset.x, y: source.y + offset.y };
}
