import { describe, expect, it } from 'vitest';
import { enterpriseCapabilitiesQueryKeys } from '../enterprise-architecture/queryKeys';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { directionMutationEffects } from './mutationEffects';

describe('directionMutationEffects one-pager freshness', () => {
  const mutations = Object.entries(directionMutationEffects);

  it.each(mutations)('%s invalidates the composed enterprise-capability one-pager', (_, effectsFor) => {
    expect(effectsFor('ec-1')).toContainEqual(onePagersQueryKeys.onePager('enterprise-capability', 'ec-1'));
  });

  it.each(mutations)('%s invalidates the one-pager quality lists', (_, effectsFor) => {
    expect(effectsFor('ec-1')).toContainEqual(onePagerQualityQueryKeys.lists());
  });

  it('composition effects still invalidate the enterprise capability composition', () => {
    expect(directionMutationEffects.capture('ec-1')).toContainEqual(enterpriseCapabilitiesQueryKeys.composition('ec-1'));
  });
});
