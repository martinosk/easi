import type { HATEOASLinks, HttpMethod } from '../api/types';

export interface RelatedLink {
  href: string;
  methods: HttpMethod[];
  title: string;
  targetType: RelatedTargetType;
  relationType: string;
}

export type RelatedTargetType = 'component' | 'capability' | 'acquiredEntity' | 'vendor' | 'internalTeam';

export interface ResourceWithRelated {
  _links?: HATEOASLinks;
}

const X_RELATED = 'x-related';

export function getXRelated(resource: ResourceWithRelated | null | undefined): RelatedLink[] {
  const value = (resource?._links as Record<string, unknown> | undefined)?.[X_RELATED];
  return Array.isArray(value) ? (value as RelatedLink[]) : [];
}

export function getPostableRelated(resource: ResourceWithRelated | null | undefined): RelatedLink[] {
  return getXRelated(resource).filter((entry) => entry.methods.includes('POST'));
}

export interface RelationEndpoint {
  path: string;
  method: HttpMethod;
}

const RELATION_ENDPOINTS: Record<string, RelationEndpoint> = {
  'component-relation': { path: '/api/v1/relations', method: 'POST' },
  'capability-parent': { path: '/api/v1/capabilities/{id}/parent', method: 'PATCH' },
  'capability-realization': { path: '/api/v1/capabilities/{id}/systems', method: 'POST' },
  'origin-acquired-via': { path: '/api/v1/components/{id}/origin/acquired-via', method: 'PUT' },
  'origin-purchased-from': { path: '/api/v1/components/{id}/origin/purchased-from', method: 'PUT' },
  'origin-built-by': { path: '/api/v1/components/{id}/origin/built-by', method: 'PUT' },
};

export function resolveRelationEndpoint(relationType: string): RelationEndpoint | undefined {
  return RELATION_ENDPOINTS[relationType];
}
