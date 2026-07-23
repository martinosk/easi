import type { Position, View } from '../../../api/types';

function collect(target: Record<string, Position>, id: string, x: number | undefined, y: number | undefined): void {
  if (Number.isFinite(x) && Number.isFinite(y)) {
    target[id] = { x: x as number, y: y as number };
  }
}

export function viewPositions(view: View): Record<string, Position> {
  const positions: Record<string, Position> = {};
  for (const component of view.components ?? []) {
    collect(positions, component.componentId, component.x, component.y);
  }
  for (const capability of view.capabilities ?? []) {
    collect(positions, capability.capabilityId, capability.x, capability.y);
  }
  for (const originEntity of view.originEntities ?? []) {
    collect(positions, originEntity.originEntityId, originEntity.x, originEntity.y);
  }
  return positions;
}
