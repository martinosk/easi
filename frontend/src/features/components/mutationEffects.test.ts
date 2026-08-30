import { describe, expect, it } from 'vitest';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { componentsMutationEffects } from './mutationEffects';
import { componentsQueryKeys } from './queryKeys';

describe('componentsMutationEffects one-pager freshness', () => {
  const cases = [
    ['update', () => componentsMutationEffects.update('comp-1')],
    ['addExpert', () => componentsMutationEffects.addExpert('comp-1')],
    ['removeExpert', () => componentsMutationEffects.removeExpert('comp-1')],
  ] as const;

  it.each(cases)('%s invalidates the composed one-pager for the application', (_, effectsFor) => {
    expect(effectsFor()).toContainEqual(onePagersQueryKeys.onePager('application', 'comp-1'));
  });

  it.each(cases)('%s invalidates the one-pager quality lists', (_, effectsFor) => {
    expect(effectsFor()).toContainEqual(onePagerQualityQueryKeys.lists());
  });

  it('update still invalidates the component detail', () => {
    expect(componentsMutationEffects.update('comp-1')).toContainEqual(componentsQueryKeys.detail('comp-1'));
  });

  it('update invalidates the completeness for applications', () => {
    expect(componentsMutationEffects.update('comp-1')).toContainEqual(
      onePagersQueryKeys.completenessForSubjectType('application'),
    );
  });
});
