import { describe, expect, it } from 'vitest';
import type { HATEOASLinks } from '../api/types';
import {
  getPostableRelated,
  getXRelated,
  type RelatedLink,
  resolveRelationEndpoint,
} from './xRelated';

const link = (overrides: Partial<RelatedLink> = {}): RelatedLink => ({
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (related)',
  targetType: 'component',
  relationType: 'component-relation',
  ...overrides,
});

const linksWith = (related: RelatedLink[]): HATEOASLinks =>
  ({
    self: { href: '/api/v1/components/c1', method: 'GET' },
    'x-related': related,
  }) as HATEOASLinks;

describe('getXRelated', () => {
  it('returns the array under x-related when present', () => {
    const entry = link();
    const links = linksWith([entry]);
    expect(getXRelated({ _links: links })).toEqual([entry]);
  });

  it('returns an empty array when x-related is missing', () => {
    expect(getXRelated({ _links: { self: { href: '/x', method: 'GET' } } as HATEOASLinks })).toEqual([]);
  });

  it('returns an empty array when _links is missing', () => {
    expect(getXRelated({})).toEqual([]);
  });

  it('returns an empty array when resource is null or undefined', () => {
    expect(getXRelated(null)).toEqual([]);
    expect(getXRelated(undefined)).toEqual([]);
  });

  it('returns an empty array when x-related is not an array', () => {
    const malformed = { _links: { 'x-related': { href: 'x', method: 'GET' } } as unknown as HATEOASLinks };
    expect(getXRelated(malformed)).toEqual([]);
  });
});

describe('getPostableRelated', () => {
  it('keeps entries whose methods include POST', () => {
    const entry = link({ methods: ['POST'] });
    expect(getPostableRelated({ _links: linksWith([entry]) })).toEqual([entry]);
  });

  it('keeps entries with mixed methods that include POST', () => {
    const entry = link({ methods: ['GET', 'POST'] });
    expect(getPostableRelated({ _links: linksWith([entry]) })).toEqual([entry]);
  });

  it('drops entries that only advertise GET', () => {
    const getOnly = link({ methods: ['GET'], relationType: 'capability-requires' });
    const postable = link({ methods: ['POST'] });
    const result = getPostableRelated({ _links: linksWith([getOnly, postable]) });
    expect(result).toEqual([postable]);
  });

  it('returns an empty array when nothing is POST-capable', () => {
    const getOnly = link({ methods: ['GET'] });
    expect(getPostableRelated({ _links: linksWith([getOnly]) })).toEqual([]);
  });

  it('returns an empty array when x-related is missing', () => {
    expect(getPostableRelated({})).toEqual([]);
  });
});

describe('resolveRelationEndpoint', () => {
  it.each([
    ['component-relation', { path: '/api/v1/relations', method: 'POST' }],
    ['capability-parent', { path: '/api/v1/capabilities/{id}/parent', method: 'PATCH' }],
    ['capability-realization', { path: '/api/v1/capabilities/{id}/systems', method: 'POST' }],
    ['origin-acquired-via', { path: '/api/v1/components/{id}/origin/acquired-via', method: 'PUT' }],
    ['origin-purchased-from', { path: '/api/v1/components/{id}/origin/purchased-from', method: 'PUT' }],
    ['origin-built-by', { path: '/api/v1/components/{id}/origin/built-by', method: 'PUT' }],
  ])('resolves %s to the canonical endpoint', (relationType, expected) => {
    expect(resolveRelationEndpoint(relationType)).toEqual(expected);
  });

  it('returns undefined for unknown relation types', () => {
    expect(resolveRelationEndpoint('totally-unknown')).toBeUndefined();
  });
});
