import { describe, expect, it } from 'vitest';
import type { ComponentId } from '../../../api/types';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { containmentDeletionMessage, describeContainmentDeletion } from './containment';

const parent = buildComponent({
  id: 'crm' as ComponentId,
  parts: [
    { id: 'quoting' as ComponentId, name: 'Quoting', kind: 'composition' },
    { id: 'billing' as ComponentId, name: 'Billing', kind: 'aggregation' },
    { id: 'orders' as ComponentId, name: 'Orders', kind: 'composition' },
  ],
});

describe('describeContainmentDeletion', () => {
  it('splits parts into deleted composed parts and released aggregated parts', () => {
    expect(describeContainmentDeletion(parent)).toEqual({ deleted: ['Quoting', 'Orders'], released: ['Billing'] });
  });

  it('reports nothing for a standalone or unknown component', () => {
    expect(describeContainmentDeletion(buildComponent())).toEqual({ deleted: [], released: [] });
    expect(describeContainmentDeletion(undefined)).toEqual({ deleted: [], released: [] });
  });
});

describe('containmentDeletionMessage', () => {
  it('names the composed parts that will be deleted and the aggregated parts that will be released', () => {
    expect(containmentDeletionMessage('Base.', parent)).toBe(
      'Base. The composed parts "Quoting", "Orders" will be deleted with it. The aggregated part "Billing" will be released as standalone.',
    );
  });

  it('leaves the base message untouched for a standalone component', () => {
    expect(containmentDeletionMessage('Base.', buildComponent())).toBe('Base.');
  });
});
