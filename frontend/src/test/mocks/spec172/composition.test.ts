import { describe, expect, it } from 'vitest';
import type { ActiveDirection, StubCapability } from './composition';
import { evaluateEligibility, resolveComposition } from './composition';

function tree(): StubCapability[] {
  return [
    { id: 'cap-cim', name: 'Customer Identity Mgmt', level: 'L1', parentId: null, businessDomainId: 'bd-c', businessDomainName: 'Customer' },
    { id: 'cap-consent', name: 'Customer Consent', level: 'L2', parentId: 'cap-cim', businessDomainId: 'bd-c', businessDomainName: 'Customer' },
    { id: 'cap-fraud', name: 'Customer Fraud Prevention', level: 'L2', parentId: 'cap-cim', businessDomainId: 'bd-c', businessDomainName: 'Customer' },
    { id: 'cap-charge', name: 'Chargeback Handling', level: 'L3', parentId: 'cap-fraud', businessDomainId: 'bd-c', businessDomainName: 'Customer' },
    { id: 'cap-cred', name: 'Credential Issuance', level: 'L4', parentId: null, businessDomainId: 'bd-a', businessDomainName: 'Access' },
  ];
}

describe('resolveComposition', () => {
  it('returns nothing when the EC has no active direction', () => {
    const result = resolveComposition('ec-crm', [], tree());
    expect(result).toEqual([]);
  });

  it('includes a single source with no descendants (count 1)', () => {
    const directions: ActiveDirection[] = [{ ecId: 'ec-x', ecName: 'X', sourceCapabilityIds: ['cap-consent'] }];
    const result = resolveComposition('ec-x', directions, tree());
    expect(result).toHaveLength(1);
    expect(result[0]).toMatchObject({ capabilityId: 'cap-consent', role: 'source' });
  });

  it('includes a source and all its descendants as implicit', () => {
    const directions: ActiveDirection[] = [{ ecId: 'ec-x', ecName: 'X', sourceCapabilityIds: ['cap-cim'] }];
    const byId = Object.fromEntries(resolveComposition('ec-x', directions, tree()).map((i) => [i.capabilityId, i]));
    expect(byId['cap-cim'].role).toBe('source');
    expect(byId['cap-consent'].role).toBe('implicit');
    expect(byId['cap-fraud'].role).toBe('implicit');
    expect(byId['cap-charge'].role).toBe('implicit');
  });

  it('accepts sources at any level L1-L4 with no ancestor/descendant relationship', () => {
    const directions: ActiveDirection[] = [
      { ecId: 'ec-x', ecName: 'X', sourceCapabilityIds: ['cap-consent', 'cap-charge', 'cap-cred'] },
    ];
    const result = resolveComposition('ec-x', directions, tree());
    const ids = result.map((i) => i.capabilityId).sort();
    expect(ids).toEqual(['cap-charge', 'cap-consent', 'cap-cred']);
    expect(result.every((i) => i.role === 'source')).toBe(true);
  });

  it('carves out a descendant that is the explicit source of another active direction (R2)', () => {
    const directions: ActiveDirection[] = [
      { ecId: 'ec-crm', ecName: 'CRM', sourceCapabilityIds: ['cap-cim'] },
      { ecId: 'ec-tp', ecName: 'Take Payment', sourceCapabilityIds: ['cap-fraud'] },
    ];
    const byId = Object.fromEntries(resolveComposition('ec-crm', directions, tree()).map((i) => [i.capabilityId, i]));
    expect(byId['cap-cim'].role).toBe('source');
    expect(byId['cap-consent'].role).toBe('implicit');
    expect(byId['cap-fraud']).toMatchObject({
      role: 'carved-out',
      carvedOutBy: { enterpriseCapabilityId: 'ec-tp', enterpriseCapabilityName: 'Take Payment' },
    });
  });

  it('a carve-out carries its entire subtree (the carved subtree is not listed under the ancestor EC) (R2)', () => {
    const directions: ActiveDirection[] = [
      { ecId: 'ec-crm', ecName: 'CRM', sourceCapabilityIds: ['cap-cim'] },
      { ecId: 'ec-tp', ecName: 'Take Payment', sourceCapabilityIds: ['cap-fraud'] },
    ];
    const ids = resolveComposition('ec-crm', directions, tree()).map((i) => i.capabilityId);
    expect(ids).not.toContain('cap-charge');
  });

  it('a more-specific source carves a deeper node out of a carve-out (R2)', () => {
    const tpAndDisputesDirections: ActiveDirection[] = [
      { ecId: 'ec-tp', ecName: 'Take Payment', sourceCapabilityIds: ['cap-fraud'] },
      { ecId: 'ec-disp', ecName: 'Disputes', sourceCapabilityIds: ['cap-charge'] },
    ];
    const byId = Object.fromEntries(
      resolveComposition('ec-tp', tpAndDisputesDirections, tree()).map((i) => [i.capabilityId, i]),
    );
    expect(byId['cap-fraud'].role).toBe('source');
    expect(byId['cap-charge']).toMatchObject({
      role: 'carved-out',
      carvedOutBy: { enterpriseCapabilityId: 'ec-disp', enterpriseCapabilityName: 'Disputes' },
    });
  });

  it('a source remains owned even when an ancestor is sourced by another EC (most-specific-wins)', () => {
    const directions: ActiveDirection[] = [
      { ecId: 'ec-crm', ecName: 'CRM', sourceCapabilityIds: ['cap-cim'] },
      { ecId: 'ec-tp', ecName: 'Take Payment', sourceCapabilityIds: ['cap-fraud'] },
    ];
    const tpIds = resolveComposition('ec-tp', directions, tree()).map((i) => i.capabilityId).sort();
    expect(tpIds).toEqual(['cap-charge', 'cap-fraud']);
  });
});

describe('evaluateEligibility (R1 same-node exclusivity)', () => {
  it('is eligible when no other active direction sources the capability', () => {
    const result = evaluateEligibility('cap-consent', 'ec-x', []);
    expect(result.eligible).toBe(true);
  });

  it('is ineligible when the exact node is sourced by another EC, naming the conflicting EC', () => {
    const directions: ActiveDirection[] = [
      { ecId: 'ec-tp', ecName: 'Take Payment', sourceCapabilityIds: ['cap-fraud'] },
    ];
    const result = evaluateEligibility('cap-fraud', 'ec-crm', directions);
    expect(result.eligible).toBe(false);
    expect(result.conflictingEnterpriseCapability).toEqual({ id: 'ec-tp', name: 'Take Payment' });
  });

  it('is eligible when the capability is sourced by the target EC itself', () => {
    const directions: ActiveDirection[] = [
      { ecId: 'ec-crm', ecName: 'CRM', sourceCapabilityIds: ['cap-fraud'] },
    ];
    expect(evaluateEligibility('cap-fraud', 'ec-crm', directions).eligible).toBe(true);
  });

  it('does not flag an ancestor/descendant relationship as ineligible (only the exact node)', () => {
    const directions: ActiveDirection[] = [
      { ecId: 'ec-tp', ecName: 'Take Payment', sourceCapabilityIds: ['cap-fraud'] },
    ];
    expect(evaluateEligibility('cap-cim', 'ec-crm', directions).eligible).toBe(true);
  });
});
