import { describe, expect, it } from 'vitest';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import type { OnePagerSubjectType } from '../one-pagers/types';
import {
  acquiredEntitiesMutationEffects,
  internalTeamsMutationEffects,
  vendorsMutationEffects,
} from './mutationEffects';
import { acquiredEntitiesQueryKeys, internalTeamsQueryKeys, vendorsQueryKeys } from './queryKeys';

describe('origin entity mutation effects one-pager freshness', () => {
  const cases: [OnePagerSubjectType, (id: string) => ReadonlyArray<readonly unknown[]>][] = [
    ['acquired-entity', acquiredEntitiesMutationEffects.update],
    ['vendor', vendorsMutationEffects.update],
    ['internal-team', internalTeamsMutationEffects.update],
  ];

  it.each(cases)('update invalidates the composed one-pager for the %s', (subjectType, update) => {
    expect(update('entity-1')).toContainEqual(onePagersQueryKeys.onePager(subjectType, 'entity-1'));
  });

  it.each(cases)('update for %s invalidates the one-pager quality lists', (_, update) => {
    expect(update('entity-1')).toContainEqual(onePagerQualityQueryKeys.lists());
  });

  it('update still invalidates the entity detail for each origin entity kind', () => {
    expect(acquiredEntitiesMutationEffects.update('entity-1')).toContainEqual(
      acquiredEntitiesQueryKeys.detail('entity-1'),
    );
    expect(vendorsMutationEffects.update('entity-1')).toContainEqual(vendorsQueryKeys.detail('entity-1'));
    expect(internalTeamsMutationEffects.update('entity-1')).toContainEqual(internalTeamsQueryKeys.detail('entity-1'));
  });

  it.each(cases)('update for %s invalidates the completeness for the subject type', (subjectType, update) => {
    expect(update('entity-1')).toContainEqual(onePagersQueryKeys.completenessForSubjectType(subjectType));
  });
});
