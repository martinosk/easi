import type { Component } from '../../../api/types';
import type { ConnectionKind } from '../../../lib/schemas';
import { hasLink } from '../../../utils/hateoas';

export type ContainmentConnectionKind = Extract<ConnectionKind, 'composition' | 'aggregation'>;

export interface ConnectionKindOption {
  value: ConnectionKind;
  label: string;
}

const RELATION_KIND_OPTIONS: ConnectionKindOption[] = [
  { value: 'Triggers', label: 'Triggers' },
  { value: 'Serves', label: 'Serves' },
];

const CONTAINMENT_KIND_OPTIONS: ConnectionKindOption[] = [
  { value: 'composition', label: 'Part of (composition)' },
  { value: 'aggregation', label: 'Part of (aggregation)' },
];

export const isContainmentKind = (kind: ConnectionKind): kind is ContainmentConnectionKind =>
  kind === 'composition' || kind === 'aggregation';

export function canContain(source: Component | undefined, target: Component | undefined): boolean {
  return hasLink(source, 'x-attach-to') && target !== undefined && !target.partOf;
}

export function connectionKindOptions(containmentAllowed: boolean): ConnectionKindOption[] {
  return containmentAllowed ? [...RELATION_KIND_OPTIONS, ...CONTAINMENT_KIND_OPTIONS] : RELATION_KIND_OPTIONS;
}
