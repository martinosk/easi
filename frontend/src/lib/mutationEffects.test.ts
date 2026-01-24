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

  describe('acquiredEntities.linkComponent', () => {
    it('should invalidate component origins to show relationship in details panel', () => {
      const effects = mutationEffects.acquiredEntities.linkComponent('entity-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.components.origins('comp-1'));
    });

    it('should invalidate origin relationships list for canvas edges', () => {
      const effects = mutationEffects.acquiredEntities.linkComponent('entity-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.originRelationships.lists());
    });

    it('should invalidate acquired entity relationships', () => {
      const effects = mutationEffects.acquiredEntities.linkComponent('entity-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.acquiredEntities.relationships('entity-1'));
      expect(effects).toContainEqual(queryKeys.acquiredEntities.detail('entity-1'));
    });
  });

  describe('vendors.linkComponent', () => {
    it('should invalidate component origins to show relationship in details panel', () => {
      const effects = mutationEffects.vendors.linkComponent('vendor-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.components.origins('comp-1'));
    });

    it('should invalidate origin relationships list for canvas edges', () => {
      const effects = mutationEffects.vendors.linkComponent('vendor-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.originRelationships.lists());
    });
  });

  describe('internalTeams.linkComponent', () => {
    it('should invalidate component origins to show relationship in details panel', () => {
      const effects = mutationEffects.internalTeams.linkComponent('team-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.components.origins('comp-1'));
    });

    it('should invalidate origin relationships list for canvas edges', () => {
      const effects = mutationEffects.internalTeams.linkComponent('team-1', 'comp-1');

      expect(effects).toContainEqual(queryKeys.originRelationships.lists());
    });
  });

  describe('acquiredEntities.delete', () => {
    it('should invalidate layouts cache to remove deleted entity from canvas', () => {
      const effects = mutationEffects.acquiredEntities.delete('entity-1');

      expect(effects).toContainEqual(queryKeys.layouts.all);
    });

    it('should invalidate entity lists and detail', () => {
      const effects = mutationEffects.acquiredEntities.delete('entity-1');

      expect(effects).toContainEqual(queryKeys.acquiredEntities.lists());
      expect(effects).toContainEqual(queryKeys.acquiredEntities.detail('entity-1'));
    });
  });

  describe('vendors.delete', () => {
    it('should invalidate layouts cache to remove deleted vendor from canvas', () => {
      const effects = mutationEffects.vendors.delete('vendor-1');

      expect(effects).toContainEqual(queryKeys.layouts.all);
    });

    it('should invalidate vendor lists and detail', () => {
      const effects = mutationEffects.vendors.delete('vendor-1');

      expect(effects).toContainEqual(queryKeys.vendors.lists());
      expect(effects).toContainEqual(queryKeys.vendors.detail('vendor-1'));
    });
  });

  describe('internalTeams.delete', () => {
    it('should invalidate layouts cache to remove deleted team from canvas', () => {
      const effects = mutationEffects.internalTeams.delete('team-1');

      expect(effects).toContainEqual(queryKeys.layouts.all);
    });

    it('should invalidate team lists and detail', () => {
      const effects = mutationEffects.internalTeams.delete('team-1');

      expect(effects).toContainEqual(queryKeys.internalTeams.lists());
      expect(effects).toContainEqual(queryKeys.internalTeams.detail('team-1'));
    });
  });

  describe('capabilities.changeParent', () => {
    it('should invalidate moved capability realizations', () => {
      const effects = mutationEffects.capabilities.changeParent({
        id: 'cap-child',
        oldParentId: 'cap-old-parent',
        newParentId: 'cap-new-parent',
      });

      expect(effects).toContainEqual(queryKeys.capabilities.realizations('cap-child'));
    });

    it('should invalidate new parent realizations to show inherited realizations from moved capability', () => {
      const effects = mutationEffects.capabilities.changeParent({
        id: 'cap-child',
        newParentId: 'cap-new-parent',
      });

      expect(effects).toContainEqual(queryKeys.capabilities.realizations('cap-new-parent'));
    });

    it('should invalidate old parent realizations to remove inherited realizations', () => {
      const effects = mutationEffects.capabilities.changeParent({
        id: 'cap-child',
        oldParentId: 'cap-old-parent',
      });

      expect(effects).toContainEqual(queryKeys.capabilities.realizations('cap-old-parent'));
    });

    it('should invalidate global realization indexes', () => {
      const effects = mutationEffects.capabilities.changeParent({
        id: 'cap-child',
        newParentId: 'cap-new-parent',
      });

      expect(effects).toContainEqual(queryKeys.capabilities.realizationsByComponents());
      expect(effects).toContainEqual(queryKeys.businessDomains.details());
    });

    it('should invalidate all affected caches when moving between parents', () => {
      const effects = mutationEffects.capabilities.changeParent({
        id: 'cap-child',
        oldParentId: 'cap-old-parent',
        newParentId: 'cap-new-parent',
      });

      expect(effects).toContainEqual(queryKeys.capabilities.detail('cap-child'));
      expect(effects).toContainEqual(queryKeys.capabilities.children('cap-old-parent'));
      expect(effects).toContainEqual(queryKeys.capabilities.children('cap-new-parent'));
      expect(effects).toContainEqual(queryKeys.capabilities.realizations('cap-child'));
      expect(effects).toContainEqual(queryKeys.capabilities.realizations('cap-old-parent'));
      expect(effects).toContainEqual(queryKeys.capabilities.realizations('cap-new-parent'));
      expect(effects).toContainEqual(queryKeys.capabilities.realizationsByComponents());
      expect(effects).toContainEqual(queryKeys.businessDomains.details());
      expect(effects).toContainEqual(queryKeys.capabilities.lists());
      expect(effects).toContainEqual(queryKeys.audit.history('cap-child'));
    });
  });
});
