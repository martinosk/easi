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
const VALID_TARGET_TYPES: ReadonlySet<RelatedTargetType> = new Set([
  'component',
  'capability',
  'acquiredEntity',
  'vendor',
  'internalTeam',
]);
const REQUIRED_STRING_FIELDS = ['href', 'title', 'relationType'] as const;

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === 'string');
}

function isValidTargetType(value: unknown): value is RelatedTargetType {
  return typeof value === 'string' && (VALID_TARGET_TYPES as Set<string>).has(value);
}

function isRelatedLink(value: unknown): value is RelatedLink {
  if (!value || typeof value !== 'object') return false;
  const entry = value as Record<string, unknown>;
  if (!REQUIRED_STRING_FIELDS.every((field) => typeof entry[field] === 'string')) return false;
  if (!isStringArray(entry.methods)) return false;
  return isValidTargetType(entry.targetType);
}

export function getXRelated(resource: ResourceWithRelated | null | undefined): RelatedLink[] {
  const value = (resource?._links as Record<string, unknown> | undefined)?.[X_RELATED];
  if (!Array.isArray(value)) return [];
  return value.filter(isRelatedLink);
}

export function getPostableRelated(resource: ResourceWithRelated | null | undefined): RelatedLink[] {
  return getXRelated(resource).filter((entry) => entry.methods.includes('POST'));
}
