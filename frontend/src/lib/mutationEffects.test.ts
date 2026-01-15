import { describe, it, expect } from 'vitest';
import { mutationEffects } from './mutationEffects';
import { queryKeys } from './queryClient';

describe('mutationEffects', () => {
  describe('capabilities.linkSystem', () => {
    it('should invalidate business domain details to refresh realizations in domain views', () => {
      const effects = mutationEffects.capabilities.linkSystem('cap-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.capabilities.realizations('cap-1'));
      expect(effects).toContainEqual(queryKeys.capabilities.byComponent('comp-1'));
      expect(effects).toContainEqual(queryKeys.capabilities.realizationsByComponents());
      expect(effects).toContainEqual(queryKeys.businessDomains.details());
    });
  });

  describe('capabilities.updateRealization', () => {
    it('should invalidate business domain details to refresh realizations in domain views', () => {
      const effects = mutationEffects.capabilities.updateRealization('cap-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.capabilities.realizations('cap-1'));
      expect(effects).toContainEqual(queryKeys.capabilities.byComponent('comp-1'));
      expect(effects).toContainEqual(queryKeys.capabilities.realizationsByComponents());
      expect(effects).toContainEqual(queryKeys.businessDomains.details());
    });
  });

  describe('capabilities.deleteRealization', () => {
    it('should invalidate business domain details to refresh realizations in domain views', () => {
      const effects = mutationEffects.capabilities.deleteRealization('cap-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.capabilities.realizations('cap-1'));
      expect(effects).toContainEqual(queryKeys.capabilities.byComponent('comp-1'));
      expect(effects).toContainEqual(queryKeys.capabilities.realizationsByComponents());
      expect(effects).toContainEqual(queryKeys.businessDomains.details());
    });

    it('should invalidate domain views so deleted components no longer appear as realizing systems', () => {
      const effects = mutationEffects.capabilities.deleteRealization('any-cap', 'deleted-comp');

      const businessDomainDetailsKey = queryKeys.businessDomains.details();
      expect(effects).toContainEqual(businessDomainDetailsKey);
    });
  });
});
