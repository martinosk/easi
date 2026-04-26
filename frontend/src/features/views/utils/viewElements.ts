import type { View } from '../../../api/types';
import type { EntityRef } from '../../canvas/utils/dynamicMode';

export interface EntityIdSets {
  components: Set<string>;
  capabilities: Set<string>;
  originEntities: Set<string>;
}

export function viewToEntityRefs(view: View | null): EntityRef[] {
  if (!view) return [];
  return [
    ...view.components.map((c) => ({ id: c.componentId, type: 'component' as const })),
    ...(view.capabilities ?? []).map((c) => ({ id: c.capabilityId, type: 'capability' as const })),
    ...(view.originEntities ?? []).map((oe) => ({ id: oe.originEntityId, type: 'originEntity' as const })),
  ];
}

export function entityIdSets(refs: Iterable<EntityRef>): EntityIdSets {
  const components = new Set<string>();
  const capabilities = new Set<string>();
  const originEntities = new Set<string>();
  for (const ref of refs) {
    if (ref.type === 'component') components.add(ref.id);
    else if (ref.type === 'capability') capabilities.add(ref.id);
    else originEntities.add(ref.id);
  }
  return { components, capabilities, originEntities };
}
