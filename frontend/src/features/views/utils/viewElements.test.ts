import { describe, expect, it } from 'vitest';
import type { CapabilityId, ComponentId, View, ViewId } from '../../../api/types';
import type { EntityRef } from '../../canvas/utils/dynamicMode';
import { entityIdSets, viewToEntityRefs } from './viewElements';

const baseView = (overrides: Partial<View> = {}): View => ({
  id: 'view-1' as ViewId,
  name: 'Test',
  isDefault: false,
  isPrivate: false,
  components: [],
  capabilities: [],
  originEntities: [],
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' } },
  ...overrides,
});

describe('viewToEntityRefs', () => {
  it('returns empty array when view is null', () => {
    expect(viewToEntityRefs(null)).toEqual([]);
  });

  it('flattens components, capabilities, and origin entities into refs', () => {
    const view = baseView({
      components: [
        { componentId: 'c1' as ComponentId, x: 0, y: 0 },
        { componentId: 'c2' as ComponentId, x: 10, y: 10 },
      ],
      capabilities: [{ capabilityId: 'cap1' as CapabilityId, x: 0, y: 0 }],
      originEntities: [{ originEntityId: 'oe1', x: 0, y: 0 }],
    });

    expect(viewToEntityRefs(view)).toEqual([
      { id: 'c1', type: 'component' },
      { id: 'c2', type: 'component' },
      { id: 'cap1', type: 'capability' },
      { id: 'oe1', type: 'originEntity' },
    ]);
  });

  it('handles missing capability/originEntity arrays defensively', () => {
    const view = baseView();
    // Force the schema-violating shape that older snapshots may produce.
    const malformed = { ...view, capabilities: undefined, originEntities: undefined } as unknown as View;

    expect(viewToEntityRefs(malformed)).toEqual([]);
  });
});

describe('entityIdSets', () => {
  it('returns empty sets for empty input', () => {
    const sets = entityIdSets([]);
    expect(sets.components.size).toBe(0);
    expect(sets.capabilities.size).toBe(0);
    expect(sets.originEntities.size).toBe(0);
  });

  it('groups refs by type', () => {
    const refs: EntityRef[] = [
      { id: 'c1', type: 'component' },
      { id: 'c2', type: 'component' },
      { id: 'cap1', type: 'capability' },
      { id: 'oe1', type: 'originEntity' },
      { id: 'oe2', type: 'originEntity' },
    ];

    const sets = entityIdSets(refs);

    expect([...sets.components]).toEqual(['c1', 'c2']);
    expect([...sets.capabilities]).toEqual(['cap1']);
    expect([...sets.originEntities]).toEqual(['oe1', 'oe2']);
  });

  it('deduplicates ids within a type', () => {
    const refs: EntityRef[] = [
      { id: 'c1', type: 'component' },
      { id: 'c1', type: 'component' },
    ];

    expect(entityIdSets(refs).components.size).toBe(1);
  });
});
