import { describe, expect, it } from 'vitest';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersMutationEffects } from './mutationEffects';
import { onePagersQueryKeys } from './queryKeys';

describe('onePagersMutationEffects', () => {
  describe('configuration', () => {
    it('invalidates the configuration and the subject-type-level one-pager views', () => {
      const effects = onePagersMutationEffects.configuration('vendor');

      expect(effects).toContainEqual(onePagersQueryKeys.configuration('vendor'));
      expect(effects).toContainEqual(onePagersQueryKeys.viewsForSubjectType('vendor'));
    });

    it('invalidates the one-pager quality lists', () => {
      const effects = onePagersMutationEffects.configuration('vendor');

      expect(effects).toContainEqual(onePagerQualityQueryKeys.lists());
    });
  });

  describe('facts', () => {
    it('invalidates the facts and the composed one-pager view for the subject', () => {
      const effects = onePagersMutationEffects.facts('vendor', 'vendor-1');

      expect(effects).toContainEqual(onePagersQueryKeys.factsForSubject('vendor', 'vendor-1'));
      expect(effects).toContainEqual(onePagersQueryKeys.onePager('vendor', 'vendor-1'));
    });

    it('invalidates the one-pager quality lists', () => {
      const effects = onePagersMutationEffects.facts('vendor', 'vendor-1');

      expect(effects).toContainEqual(onePagerQualityQueryKeys.lists());
    });
  });
});

describe('onePagersQueryKeys', () => {
  it('nests the one-pager key under the subject-type-level views key', () => {
    expect(onePagersQueryKeys.onePager('vendor', 'vendor-1')).toEqual([
      ...onePagersQueryKeys.viewsForSubjectType('vendor'),
      'vendor-1',
    ]);
  });

  it('nests viewsForSubjectType under views', () => {
    expect(onePagersQueryKeys.viewsForSubjectType('vendor')).toEqual([...onePagersQueryKeys.views(), 'vendor']);
  });
});
