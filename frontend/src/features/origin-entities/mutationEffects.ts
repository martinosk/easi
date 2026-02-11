import {
  acquiredEntitiesQueryKeys,
  vendorsQueryKeys,
  internalTeamsQueryKeys,
  originRelationshipsQueryKeys,
} from './queryKeys';
import { componentsQueryKeys } from '../components/queryKeys';
import { layoutsQueryKeys } from '../canvas/queryKeys';
import { auditQueryKeys } from '../audit/queryKeys';
import { artifactCreatorsQueryKeys } from '../navigation/hooks/useArtifactCreators';

interface OriginEntityQueryKeys {
  all: readonly [string];
  lists: () => readonly string[];
  details: () => readonly string[];
  detail: (id: string) => readonly string[];
  relationships: (id: string) => readonly string[];
}

function createOriginEntityMutationEffects(entityQueryKeys: OriginEntityQueryKeys) {
  return {
    create: () => [
      entityQueryKeys.lists(),
      artifactCreatorsQueryKeys.all,
    ],

    update: (id: string) => [
      entityQueryKeys.lists(),
      entityQueryKeys.detail(id),
      auditQueryKeys.history(id),
    ],

    delete: (id: string) => [
      entityQueryKeys.lists(),
      entityQueryKeys.detail(id),
      componentsQueryKeys.lists(),
      componentsQueryKeys.details(),
      originRelationshipsQueryKeys.lists(),
      layoutsQueryKeys.all,
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

export const acquiredEntitiesMutationEffects = createOriginEntityMutationEffects(acquiredEntitiesQueryKeys);
export const vendorsMutationEffects = createOriginEntityMutationEffects(vendorsQueryKeys);
export const internalTeamsMutationEffects = createOriginEntityMutationEffects(internalTeamsQueryKeys);
