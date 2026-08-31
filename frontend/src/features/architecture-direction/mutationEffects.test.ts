import { describe, expect, it } from 'vitest';
import { compositionSummariesQueryKeys, enterpriseCapabilitiesQueryKeys } from '../enterprise-architecture/queryKeys';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { directionMutationEffects } from './mutationEffects';

const DIRECTION_MUTATIONS = [
  'capture',
  'addSource',
  'removeSource',
  'update',
  'propose',
  'agree',
  'reject',
  'revert',
] as const;

describe('directionMutationEffects', () => {
  it.each(DIRECTION_MUTATIONS)('invalidates the composition summaries after %s', (mutation) => {
    const effects = directionMutationEffects[mutation]('ec-1');

    expect(effects).toContainEqual(compositionSummariesQueryKeys.lists());
  });

  it.each(DIRECTION_MUTATIONS)('%s invalidates the composed enterprise-capability one-pager', (mutation) => {
    expect(directionMutationEffects[mutation]('ec-1')).toContainEqual(
      onePagersQueryKeys.onePager('enterprise-capability', 'ec-1'),
    );
  });

  it.each(DIRECTION_MUTATIONS)('%s invalidates the one-pager quality lists', (mutation) => {
    expect(directionMutationEffects[mutation]('ec-1')).toContainEqual(onePagerQualityQueryKeys.lists());
  });

  it('capture still invalidates the enterprise capability composition', () => {
    expect(directionMutationEffects.capture('ec-1')).toContainEqual(enterpriseCapabilitiesQueryKeys.composition('ec-1'));
  });
});
