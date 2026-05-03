import { describe, expect, it } from 'vitest';
import type { HATEOASLinks } from '../api/types';
import { getPostableRelated, getXRelated, type RelatedLink } from './xRelated';

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
  }) as unknown as HATEOASLinks;

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
