import type { ViewId } from '../../../api/types';
import type { EntityRef, EntityType } from './dynamicMode';

export type Operation = 'add' | 'remove' | 'position';

export interface DraftSaveApi {
  addComponent: (viewId: ViewId, componentId: string, x: number, y: number) => Promise<void>;
  addCapability: (viewId: ViewId, capabilityId: string, x: number, y: number) => Promise<void>;
  addOriginEntity: (viewId: ViewId, originEntityId: string, x: number, y: number) => Promise<void>;
  removeComponent: (viewId: ViewId, componentId: string) => Promise<void>;
  removeCapability: (viewId: ViewId, capabilityId: string) => Promise<void>;
  removeOriginEntity: (viewId: ViewId, originEntityId: string) => Promise<void>;
  updateComponentPosition: (viewId: ViewId, componentId: string, x: number, y: number) => Promise<void>;
  updateCapabilityPosition: (viewId: ViewId, capabilityId: string, x: number, y: number) => Promise<void>;
  updateOriginEntityPosition: (viewId: ViewId, originEntityId: string, x: number, y: number) => Promise<void>;
}

export interface PositionedEntity extends EntityRef {
  x: number;
  y: number;
}

export interface DraftSaveInput {
  viewId: ViewId;
  additions: PositionedEntity[];
  removals: EntityRef[];
  positionDeltas: PositionedEntity[];
}

export interface DraftSaveFailure {
  entity: EntityRef;
  operation: Operation;
  message: string;
}

export interface DraftSaveResult {
  successCount: number;
  failures: DraftSaveFailure[];
}

type AddFn = (id: string, x: number, y: number) => Promise<void>;
type RemoveFn = (id: string) => Promise<void>;
type PositionFn = AddFn;

export async function saveDraft(api: DraftSaveApi, input: DraftSaveInput): Promise<DraftSaveResult> {
  const { viewId } = input;
  const failures: DraftSaveFailure[] = [];
  let successCount = 0;

  const addByType: Record<EntityType, AddFn> = {
    component: (id, x, y) => api.addComponent(viewId, id, x, y),
    capability: (id, x, y) => api.addCapability(viewId, id, x, y),
    originEntity: (id, x, y) => api.addOriginEntity(viewId, id, x, y),
  };
  const removeByType: Record<EntityType, RemoveFn> = {
    component: (id) => api.removeComponent(viewId, id),
    capability: (id) => api.removeCapability(viewId, id),
    originEntity: (id) => api.removeOriginEntity(viewId, id),
  };
  const positionByType: Record<EntityType, PositionFn> = {
    component: (id, x, y) => api.updateComponentPosition(viewId, id, x, y),
    capability: (id, x, y) => api.updateCapabilityPosition(viewId, id, x, y),
    originEntity: (id, x, y) => api.updateOriginEntityPosition(viewId, id, x, y),
  };

  const run = async (entity: EntityRef, operation: Operation, fn: () => Promise<void>) => {
    try {
      await fn();
      successCount++;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      failures.push({ entity: { id: entity.id, type: entity.type }, operation, message });
    }
  };

  for (const e of input.removals) {
    await run(e, 'remove', () => removeByType[e.type](e.id));
  }

  const addedIds = new Set(input.additions.map((e) => e.id));
  for (const e of input.additions) {
    await run(e, 'add', () => addByType[e.type](e.id, e.x, e.y));
  }

  for (const e of input.positionDeltas) {
    if (addedIds.has(e.id)) continue;
    await run(e, 'position', () => positionByType[e.type](e.id, e.x, e.y));
  }

  return { successCount, failures };
}
