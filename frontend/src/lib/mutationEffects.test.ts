import { describe, it, expect } from 'vitest';
import { auditQueryKeys } from '../features/audit/queryKeys';
import { layoutsQueryKeys } from '../features/canvas/queryKeys';
import { capabilitiesQueryKeys } from '../features/capabilities/queryKeys';
import { capabilitiesMutationEffects } from '../features/capabilities/mutationEffects';
import { businessDomainsQueryKeys } from '../features/business-domains/queryKeys';
import { componentsQueryKeys } from '../features/components/queryKeys';
import {
  acquiredEntitiesQueryKeys,
  vendorsQueryKeys,
  internalTeamsQueryKeys,
  originRelationshipsQueryKeys,
} from '../features/origin-entities/queryKeys';
import {
  acquiredEntitiesMutationEffects,
  vendorsMutationEffects,
  internalTeamsMutationEffects,
} from '../features/origin-entities/mutationEffects';
import { editGrantsQueryKeys } from '../features/edit-grants/queryKeys';
import { editGrantsMutationEffects } from '../features/edit-grants/mutationEffects';
import { componentsMutationEffects } from '../features/components/mutationEffects';
import { artifactCreatorsQueryKeys } from '../features/navigation/hooks/useArtifactCreators';

describe('mutationEffects', () => {
  describe('capabilitiesMutationEffects.linkSystem', () => {
    it('should invalidate business domain details to refresh realizations in domain views', () => {
      const effects = capabilitiesMutationEffects.linkSystem({ capabilityId: 'cap-1', componentId: 'comp-1' });

      expect(effects).toContainEqual(capabilitiesQueryKeys.realizations('cap-1'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.byComponent('comp-1'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.realizationsByComponents());
      expect(effects).toContainEqual(businessDomainsQueryKeys.details());
    });
  });

  describe('capabilitiesMutationEffects.updateRealization', () => {
    it('should invalidate business domain details to refresh realizations in domain views', () => {
      const effects = capabilitiesMutationEffects.updateRealization({ capabilityId: 'cap-1', componentId: 'comp-1' });

      expect(effects).toContainEqual(capabilitiesQueryKeys.realizations('cap-1'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.byComponent('comp-1'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.realizationsByComponents());
      expect(effects).toContainEqual(businessDomainsQueryKeys.details());
    });
  });

  describe('capabilitiesMutationEffects.deleteRealization', () => {
    it('should invalidate business domain details to refresh realizations in domain views', () => {
      const effects = capabilitiesMutationEffects.deleteRealization({ capabilityId: 'cap-1', componentId: 'comp-1' });

      expect(effects).toContainEqual(capabilitiesQueryKeys.realizations('cap-1'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.byComponent('comp-1'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.realizationsByComponents());
      expect(effects).toContainEqual(businessDomainsQueryKeys.details());
    });

    it('should invalidate domain views so deleted components no longer appear as realizing systems', () => {
      const effects = capabilitiesMutationEffects.deleteRealization({ capabilityId: 'any-cap', componentId: 'deleted-comp' });

      const businessDomainDetailsKey = businessDomainsQueryKeys.details();
      expect(effects).toContainEqual(businessDomainDetailsKey);
    });
  });

  describe('acquiredEntities.linkComponent', () => {
    it('should invalidate component origins to show relationship in details panel', () => {
      const effects = acquiredEntitiesMutationEffects.linkComponent('entity-1', 'comp-1');

      expect(effects).toContainEqual(componentsQueryKeys.origins('comp-1'));
    });

    it('should invalidate origin relationships list for canvas edges', () => {
      const effects = acquiredEntitiesMutationEffects.linkComponent('entity-1', 'comp-1');

      expect(effects).toContainEqual(originRelationshipsQueryKeys.lists());
    });

    it('should invalidate acquired entity relationships', () => {
      const effects = acquiredEntitiesMutationEffects.linkComponent('entity-1', 'comp-1');

      expect(effects).toContainEqual(acquiredEntitiesQueryKeys.relationships('entity-1'));
      expect(effects).toContainEqual(acquiredEntitiesQueryKeys.detail('entity-1'));
    });
  });

  describe('vendors.linkComponent', () => {
    it('should invalidate component origins to show relationship in details panel', () => {
      const effects = vendorsMutationEffects.linkComponent('vendor-1', 'comp-1');

      expect(effects).toContainEqual(componentsQueryKeys.origins('comp-1'));
    });

    it('should invalidate origin relationships list for canvas edges', () => {
      const effects = vendorsMutationEffects.linkComponent('vendor-1', 'comp-1');

      expect(effects).toContainEqual(originRelationshipsQueryKeys.lists());
    });
  });

  describe('internalTeams.linkComponent', () => {
    it('should invalidate component origins to show relationship in details panel', () => {
      const effects = internalTeamsMutationEffects.linkComponent('team-1', 'comp-1');

      expect(effects).toContainEqual(componentsQueryKeys.origins('comp-1'));
    });

    it('should invalidate origin relationships list for canvas edges', () => {
      const effects = internalTeamsMutationEffects.linkComponent('team-1', 'comp-1');

      expect(effects).toContainEqual(originRelationshipsQueryKeys.lists());
    });
  });

  describe('acquiredEntities.delete', () => {
    it('should invalidate layouts cache to remove deleted entity from canvas', () => {
      const effects = acquiredEntitiesMutationEffects.delete('entity-1');

      expect(effects).toContainEqual(layoutsQueryKeys.all);
    });

    it('should invalidate entity lists and detail', () => {
      const effects = acquiredEntitiesMutationEffects.delete('entity-1');

      expect(effects).toContainEqual(acquiredEntitiesQueryKeys.lists());
      expect(effects).toContainEqual(acquiredEntitiesQueryKeys.detail('entity-1'));
    });
  });

  describe('vendors.delete', () => {
    it('should invalidate layouts cache to remove deleted vendor from canvas', () => {
      const effects = vendorsMutationEffects.delete('vendor-1');

      expect(effects).toContainEqual(layoutsQueryKeys.all);
    });

    it('should invalidate vendor lists and detail', () => {
      const effects = vendorsMutationEffects.delete('vendor-1');

      expect(effects).toContainEqual(vendorsQueryKeys.lists());
      expect(effects).toContainEqual(vendorsQueryKeys.detail('vendor-1'));
    });
  });

  describe('internalTeams.delete', () => {
    it('should invalidate layouts cache to remove deleted team from canvas', () => {
      const effects = internalTeamsMutationEffects.delete('team-1');

      expect(effects).toContainEqual(layoutsQueryKeys.all);
    });

    it('should invalidate team lists and detail', () => {
      const effects = internalTeamsMutationEffects.delete('team-1');

      expect(effects).toContainEqual(internalTeamsQueryKeys.lists());
      expect(effects).toContainEqual(internalTeamsQueryKeys.detail('team-1'));
    });
  });

  describe('capabilities.changeParent', () => {
    it('should invalidate moved capability realizations', () => {
      const effects = capabilitiesMutationEffects.changeParent({
        id: 'cap-child',
        oldParentId: 'cap-old-parent',
        newParentId: 'cap-new-parent',
      });

      expect(effects).toContainEqual(capabilitiesQueryKeys.realizations('cap-child'));
    });

    it('should invalidate new parent realizations to show inherited realizations from moved capability', () => {
      const effects = capabilitiesMutationEffects.changeParent({
        id: 'cap-child',
        newParentId: 'cap-new-parent',
      });

      expect(effects).toContainEqual(capabilitiesQueryKeys.realizations('cap-new-parent'));
    });

    it('should invalidate old parent realizations to remove inherited realizations', () => {
      const effects = capabilitiesMutationEffects.changeParent({
        id: 'cap-child',
        oldParentId: 'cap-old-parent',
      });

      expect(effects).toContainEqual(capabilitiesQueryKeys.realizations('cap-old-parent'));
    });

    it('should invalidate global realization indexes', () => {
      const effects = capabilitiesMutationEffects.changeParent({
        id: 'cap-child',
        newParentId: 'cap-new-parent',
      });

      expect(effects).toContainEqual(capabilitiesQueryKeys.realizationsByComponents());
      expect(effects).toContainEqual(businessDomainsQueryKeys.details());
    });

    it('should invalidate all affected caches when moving between parents', () => {
      const effects = capabilitiesMutationEffects.changeParent({
        id: 'cap-child',
        oldParentId: 'cap-old-parent',
        newParentId: 'cap-new-parent',
      });

      expect(effects).toContainEqual(capabilitiesQueryKeys.detail('cap-child'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.children('cap-old-parent'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.children('cap-new-parent'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.realizations('cap-child'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.realizations('cap-old-parent'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.realizations('cap-new-parent'));
      expect(effects).toContainEqual(capabilitiesQueryKeys.realizationsByComponents());
      expect(effects).toContainEqual(businessDomainsQueryKeys.details());
      expect(effects).toContainEqual(capabilitiesQueryKeys.lists());
      expect(effects).toContainEqual(auditQueryKeys.history('cap-child'));
    });
  });

  describe('artifact-creators invalidation on create', () => {
    it('should invalidate artifact-creators when a component is created', () => {
      const effects = componentsMutationEffects.create();
      expect(effects).toContainEqual(artifactCreatorsQueryKeys.all);
    });

    it('should invalidate artifact-creators when a capability is created', () => {
      const effects = capabilitiesMutationEffects.create({});
      expect(effects).toContainEqual(artifactCreatorsQueryKeys.all);
    });

    it('should invalidate artifact-creators when an acquired entity is created', () => {
      const effects = acquiredEntitiesMutationEffects.create();
      expect(effects).toContainEqual(artifactCreatorsQueryKeys.all);
    });

    it('should invalidate artifact-creators when a vendor is created', () => {
      const effects = vendorsMutationEffects.create();
      expect(effects).toContainEqual(artifactCreatorsQueryKeys.all);
    });

    it('should invalidate artifact-creators when an internal team is created', () => {
      const effects = internalTeamsMutationEffects.create();
      expect(effects).toContainEqual(artifactCreatorsQueryKeys.all);
    });
  });

  describe('editGrantsMutationEffects.create', () => {
    it('should invalidate the correct query keys', () => {
      const effects = editGrantsMutationEffects.create();

      expect(effects).toContainEqual(editGrantsQueryKeys.mine());
      expect(effects).toContainEqual(editGrantsQueryKeys.all);
    });
  });

  describe('editGrantsMutationEffects.revoke', () => {
    it('should invalidate the correct query keys', () => {
      const effects = editGrantsMutationEffects.revoke();

      expect(effects).toContainEqual(editGrantsQueryKeys.mine());
      expect(effects).toContainEqual(editGrantsQueryKeys.all);
    });
  });
});
