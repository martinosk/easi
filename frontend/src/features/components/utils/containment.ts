import type { Component, ContainmentKind } from '../../../api/types';

export const CONTAINMENT_KIND_LABELS: Record<ContainmentKind, string> = {
  composition: 'Composition',
  aggregation: 'Aggregation',
};

export const CONTAINMENT_KIND_DESCRIPTIONS: Record<ContainmentKind, string> = {
  composition: 'The part exists only within its parent and is deleted with it',
  aggregation: 'The part is autonomous and is released when its parent is deleted',
};

export interface ContainmentDeletionEffects {
  deleted: string[];
  released: string[];
}

export function describeContainmentDeletion(component: Pick<Component, 'parts'> | undefined): ContainmentDeletionEffects {
  const parts = component?.parts ?? [];
  return {
    deleted: parts.filter((part) => part.kind === 'composition').map((part) => part.name),
    released: parts.filter((part) => part.kind === 'aggregation').map((part) => part.name),
  };
}

function nameList(names: string[]): string {
  return names.map((name) => `"${name}"`).join(', ');
}

export function containmentDeletionMessage(baseMessage: string, component: Pick<Component, 'parts'> | undefined): string {
  const { deleted, released } = describeContainmentDeletion(component);
  const sentences = [baseMessage];
  if (deleted.length > 0) {
    sentences.push(`The composed ${deleted.length === 1 ? 'part' : 'parts'} ${nameList(deleted)} will be deleted with it.`);
  }
  if (released.length > 0) {
    sentences.push(
      `The aggregated ${released.length === 1 ? 'part' : 'parts'} ${nameList(released)} will be released as standalone.`,
    );
  }
  return sentences.join(' ');
}
