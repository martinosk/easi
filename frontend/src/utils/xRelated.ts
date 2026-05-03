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
