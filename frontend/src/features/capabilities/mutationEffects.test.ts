import { describe, expect, it } from 'vitest';
import { compositionSummariesQueryKeys } from '../enterprise-architecture/queryKeys';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { capabilitiesMutationEffects } from './mutationEffects';
import { capabilitiesQueryKeys } from './queryKeys';

describe('capabilitiesMutationEffects one-pager freshness', () => {
  const cases = [
    ['update', () => capabilitiesMutationEffects.update('cap-1')],
    ['addExpert', () => capabilitiesMutationEffects.addExpert('cap-1')],
    ['removeExpert', () => capabilitiesMutationEffects.removeExpert('cap-1')],
  ] as const;

  it.each(cases)('%s invalidates the composed one-pager for the capability', (_, effectsFor) => {
    expect(effectsFor()).toContainEqual(onePagersQueryKeys.onePager('capability', 'cap-1'));
  });

  it.each(cases)('%s invalidates the one-pager quality lists', (_, effectsFor) => {
    expect(effectsFor()).toContainEqual(onePagerQualityQueryKeys.lists());
  });

  it('update still invalidates the capability detail', () => {
    expect(capabilitiesMutationEffects.update('cap-1')).toContainEqual(capabilitiesQueryKeys.detail('cap-1'));
  });

  it('update invalidates the completeness for capabilities', () => {
    expect(capabilitiesMutationEffects.update('cap-1')).toContainEqual(
      onePagersQueryKeys.completenessForSubjectType('capability'),
    );
  });
});

describe('capabilitiesMutationEffects composition summaries freshness', () => {
  it('assignToDomain invalidates the composition summaries', () => {
    expect(capabilitiesMutationEffects.assignToDomain({ capabilityId: 'cap-1', domainId: 'dom-1' })).toContainEqual(
      compositionSummariesQueryKeys.lists(),
    );
  });

  it('unassignFromDomain invalidates the composition summaries', () => {
    expect(
      capabilitiesMutationEffects.unassignFromDomain({ capabilityId: 'cap-1', domainId: 'dom-1' }),
    ).toContainEqual(compositionSummariesQueryKeys.lists());
  });

  it('changeParent invalidates the composition summaries', () => {
    expect(capabilitiesMutationEffects.changeParent({ id: 'cap-1' })).toContainEqual(
      compositionSummariesQueryKeys.lists(),
    );
  });

  it('delete invalidates the composition summaries', () => {
    expect(capabilitiesMutationEffects.delete({ id: 'cap-1' })).toContainEqual(compositionSummariesQueryKeys.lists());
  });

  it('cascadeDelete invalidates the composition summaries', () => {
    expect(
      capabilitiesMutationEffects.cascadeDelete({ id: 'cap-1', deleteApplications: false }),
    ).toContainEqual(compositionSummariesQueryKeys.lists());
  });
});
