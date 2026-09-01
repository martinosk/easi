import { describe, expect, it } from 'vitest';
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

