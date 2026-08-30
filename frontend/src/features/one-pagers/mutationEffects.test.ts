import { describe, expect, it } from 'vitest';
import { capabilitiesMutationEffects } from '../capabilities/mutationEffects';
import { componentsMutationEffects } from '../components/mutationEffects';
import { enterpriseCapabilitiesMutationEffects } from '../enterprise-architecture/mutationEffects';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import {
  acquiredEntitiesMutationEffects,
  internalTeamsMutationEffects,
  vendorsMutationEffects,
} from '../origin-entities/mutationEffects';
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

    it('invalidates the completeness collection for the subject type', () => {
      const effects = onePagersMutationEffects.configuration('vendor');

      expect(effects).toContainEqual(onePagersQueryKeys.completenessForSubjectType('vendor'));
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

    it('invalidates the completeness collection for the subject type', () => {
      const effects = onePagersMutationEffects.facts('vendor', 'vendor-1');

      expect(effects).toContainEqual(onePagersQueryKeys.completenessForSubjectType('vendor'));
    });
  });
});

describe('subject mutation effects', () => {
  it('invalidates application completeness when an application is created or deleted', () => {
    expect(componentsMutationEffects.create()).toContainEqual(
      onePagersQueryKeys.completenessForSubjectType('application'),
    );
    expect(componentsMutationEffects.delete('comp-1')).toContainEqual(
      onePagersQueryKeys.completenessForSubjectType('application'),
    );
  });

  it('invalidates capability completeness when a capability is created or deleted', () => {
    expect(capabilitiesMutationEffects.create({})).toContainEqual(
      onePagersQueryKeys.completenessForSubjectType('capability'),
    );
    expect(capabilitiesMutationEffects.delete({ id: 'cap-1' })).toContainEqual(
      onePagersQueryKeys.completenessForSubjectType('capability'),
    );
    expect(capabilitiesMutationEffects.cascadeDelete({ id: 'cap-1', deleteApplications: false })).toContainEqual(
      onePagersQueryKeys.completenessForSubjectType('capability'),
    );
  });

  it('invalidates enterprise-capability completeness when an enterprise capability is created or deleted', () => {
    expect(enterpriseCapabilitiesMutationEffects.create()).toContainEqual(
      onePagersQueryKeys.completenessForSubjectType('enterprise-capability'),
    );
    expect(enterpriseCapabilitiesMutationEffects.delete('ec-1')).toContainEqual(
      onePagersQueryKeys.completenessForSubjectType('enterprise-capability'),
    );
  });

  it.each([
    ['vendor', vendorsMutationEffects],
    ['acquired-entity', acquiredEntitiesMutationEffects],
    ['internal-team', internalTeamsMutationEffects],
  ] as const)('invalidates %s completeness when an origin entity is created or deleted', (subjectType, effects) => {
    expect(effects.create()).toContainEqual(onePagersQueryKeys.completenessForSubjectType(subjectType));
    expect(effects.delete('entity-1')).toContainEqual(onePagersQueryKeys.completenessForSubjectType(subjectType));
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
