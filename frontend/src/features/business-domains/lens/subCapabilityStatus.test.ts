import { describe, expect, it } from 'vitest';
import type { CapabilityRealization, RealizationLevel } from '../../../api/types';
import { toComponentId } from '../../../api/types';
import { buildCapabilityJourney, buildCapabilityRealization } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import {
  buildSubCapabilityBreakdown,
  deriveSubCapabilityAppLabel,
  deriveSubCapabilityStatus,
  type JourneyApps,
} from './subCapabilityStatus';

const APPS: JourneyApps = { fromComponentIds: new Set(['seabook']), toComponentId: 'phoenix' };

function realization(componentId: string, level: RealizationLevel, extra: Partial<CapabilityRealization> = {}) {
  return buildCapabilityRealization({
    componentId: toComponentId(componentId),
    componentName: componentId === 'phoenix' ? 'Phoenix' : componentId === 'seabook' ? 'Seabook' : componentId,
    realizationLevel: level,
    origin: 'Direct',
    ...extra,
  });
}

describe('deriveSubCapabilityStatus', () => {
  it('is done when it realises the to-app at Full with no from-app', () => {
    expect(deriveSubCapabilityStatus([realization('phoenix', 'Full')], APPS)).toBe('done');
  });

  it('is in flight when it realises the to-app at Full alongside a from-app', () => {
    expect(deriveSubCapabilityStatus([realization('phoenix', 'Full'), realization('seabook', 'Full')], APPS)).toBe(
      'in-flight',
    );
  });

  it('is in flight when it realises the to-app only at Planned', () => {
    expect(deriveSubCapabilityStatus([realization('phoenix', 'Planned')], APPS)).toBe('in-flight');
  });

  it('is in flight when it realises the to-app only at Partial', () => {
    expect(deriveSubCapabilityStatus([realization('phoenix', 'Partial')], APPS)).toBe('in-flight');
  });

  it('is in flight when it realises the to-app at Partial alongside a from-app', () => {
    expect(deriveSubCapabilityStatus([realization('phoenix', 'Partial'), realization('seabook', 'Full')], APPS)).toBe(
      'in-flight',
    );
  });

  it('is not started when it realises only from-apps', () => {
    expect(deriveSubCapabilityStatus([realization('seabook', 'Full')], APPS)).toBe('not-started');
  });

  it('is omitted (null) when it realises neither the from- nor the to-app', () => {
    expect(deriveSubCapabilityStatus([realization('other', 'Full')], APPS)).toBeNull();
  });

  it('is omitted (null) when it has no realisations at all', () => {
    expect(deriveSubCapabilityStatus([], APPS)).toBeNull();
  });

  it('ignores inherited realisations of the to-app', () => {
    const inheritedTo = realization('phoenix', 'Full', { origin: 'Inherited' });
    expect(deriveSubCapabilityStatus([inheritedTo], APPS)).toBeNull();
  });

  it('ignores inherited realisations when deriving from a direct from-app', () => {
    const inheritedTo = realization('phoenix', 'Full', { origin: 'Inherited' });
    expect(deriveSubCapabilityStatus([inheritedTo, realization('seabook', 'Full')], APPS)).toBe('not-started');
  });
});

describe('deriveSubCapabilityAppLabel', () => {
  it('labels a done row with the to-app name', () => {
    expect(deriveSubCapabilityAppLabel([realization('phoenix', 'Full')], APPS, 'done')).toBe('Phoenix');
  });

  it('labels a not-started row with the from-app name', () => {
    expect(deriveSubCapabilityAppLabel([realization('seabook', 'Full')], APPS, 'not-started')).toBe('Seabook');
  });

  it('labels an in-flight row realising both apps as "from → to"', () => {
    expect(
      deriveSubCapabilityAppLabel([realization('seabook', 'Full'), realization('phoenix', 'Full')], APPS, 'in-flight'),
    ).toBe('Seabook → Phoenix');
  });

  it('labels an in-flight row realising only the to-app with the to-app name', () => {
    expect(deriveSubCapabilityAppLabel([realization('phoenix', 'Planned')], APPS, 'in-flight')).toBe('Phoenix');
  });
});

describe('buildSubCapabilityBreakdown', () => {
  it('derives a row per descendant realising a journey app, omitting others, walking deeply', () => {
    const [node] = buildCapabilityTree(
      [
        cap('bm', 'Booking management', 'L2'),
        cap('bm-q', 'Quotation', 'L3', 'bm'),
        cap('bm-bc', 'Booking capture', 'L3', 'bm'),
        cap('bm-ac', 'Amendments', 'L3', 'bm'),
        cap('bm-x', 'Unrelated', 'L3', 'bm'),
      ],
      { orphanRoots: 'any-level' },
    );
    const journey = buildCapabilityJourney({
      fromApplications: [{ componentId: 'seabook', componentName: 'Seabook', stale: false }],
      toApplication: { componentId: 'phoenix', componentName: 'Phoenix', stale: false },
    });
    const realizationsByCapability: Record<string, CapabilityRealization[]> = {
      'bm-q': [realization('phoenix', 'Full')],
      'bm-bc': [realization('seabook', 'Full'), realization('phoenix', 'Planned')],
      'bm-ac': [realization('seabook', 'Full')],
      'bm-x': [realization('other', 'Full')],
    };

    const rows = buildSubCapabilityBreakdown(node, journey, (id) => realizationsByCapability[id] ?? []);

    expect(rows).toEqual([
      { capability: expect.objectContaining({ id: 'bm-q' }), status: 'done', appLabel: 'Phoenix' },
      { capability: expect.objectContaining({ id: 'bm-bc' }), status: 'in-flight', appLabel: 'Seabook → Phoenix' },
      { capability: expect.objectContaining({ id: 'bm-ac' }), status: 'not-started', appLabel: 'Seabook' },
    ]);
  });
});
