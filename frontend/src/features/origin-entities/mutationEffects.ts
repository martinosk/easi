import { auditQueryKeys } from '../audit/queryKeys';
import { componentsQueryKeys } from '../components/queryKeys';
import { artifactCreatorsQueryKeys } from '../navigation/hooks/useArtifactCreators';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import type { OnePagerSubjectType } from '../one-pagers/types';
import { viewsQueryKeys } from '../views/queryKeys';
import {
  acquiredEntitiesQueryKeys,
  internalTeamsQueryKeys,
  originRelationshipsQueryKeys,
  vendorsQueryKeys,
} from './queryKeys';

interface OriginEntityQueryKeys {
  all: readonly [string];
  lists: () => readonly string[];
  details: () => readonly string[];
  detail: (id: string) => readonly string[];
  relationships: (id: string) => readonly string[];
}

function createOriginEntityMutationEffects(entityQueryKeys: OriginEntityQueryKeys, subjectType: OnePagerSubjectType) {
  return {
    create: () => [
      entityQueryKeys.lists(),
      artifactCreatorsQueryKeys.all,
      onePagersQueryKeys.completenessForSubjectType(subjectType),
    ],

    update: (id: string) => [
      entityQueryKeys.lists(),
      entityQueryKeys.detail(id),
      auditQueryKeys.history(id),
      onePagersQueryKeys.onePager(subjectType, id),
      onePagerQualityQueryKeys.lists(),
    ],

    delete: (id: string) => [
      entityQueryKeys.lists(),
      entityQueryKeys.detail(id),
      componentsQueryKeys.lists(),
      componentsQueryKeys.details(),
      originRelationshipsQueryKeys.lists(),
      viewsQueryKeys.all,
      onePagersQueryKeys.completenessForSubjectType(subjectType),
    ],

    linkComponent: (entityId: string, componentId: string) => [
      entityQueryKeys.relationships(entityId),
      entityQueryKeys.detail(entityId),
      entityQueryKeys.lists(),
      originRelationshipsQueryKeys.lists(),
      componentsQueryKeys.detail(componentId),
      componentsQueryKeys.origins(componentId),
      componentsQueryKeys.lists(),
    ],

    unlinkComponent: (entityId: string, componentId: string) => [
      entityQueryKeys.relationships(entityId),
      entityQueryKeys.detail(entityId),
      entityQueryKeys.lists(),
      originRelationshipsQueryKeys.lists(),
      componentsQueryKeys.detail(componentId),
      componentsQueryKeys.origins(componentId),
      componentsQueryKeys.lists(),
    ],
  };
}

export const acquiredEntitiesMutationEffects = createOriginEntityMutationEffects(
  acquiredEntitiesQueryKeys,
  'acquired-entity',
);
export const vendorsMutationEffects = createOriginEntityMutationEffects(vendorsQueryKeys, 'vendor');
export const internalTeamsMutationEffects = createOriginEntityMutationEffects(internalTeamsQueryKeys, 'internal-team');
