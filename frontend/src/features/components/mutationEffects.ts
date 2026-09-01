import { auditQueryKeys } from '../audit/queryKeys';
import { businessDomainsQueryKeys } from '../business-domains/queryKeys';
import { strategicFitAnalysisQueryKeys } from '../strategic-fit/queryKeys';
import { artifactCreatorsQueryKeys } from '../navigation/hooks/useArtifactCreators';
import { onePagerQualityQueryKeys } from '../one-pager-quality/queryKeys';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { componentsQueryKeys, fitScoresQueryKeys } from './queryKeys';

function onePagerFreshness(componentId: string) {
  return [onePagersQueryKeys.onePager('application', componentId), onePagerQualityQueryKeys.lists()];
}

export const componentsMutationEffects = {
  create: () => [
    componentsQueryKeys.lists(),
    artifactCreatorsQueryKeys.all,
    onePagersQueryKeys.completenessForSubjectType('application'),
  ],

  update: (componentId: string) => [
    componentsQueryKeys.lists(),
    componentsQueryKeys.detail(componentId),
    businessDomainsQueryKeys.all,
    auditQueryKeys.history(componentId),
    ...onePagerFreshness(componentId),
    onePagersQueryKeys.completenessForSubjectType('application'),
  ],

  delete: (componentId: string) => [
    componentsQueryKeys.lists(),
    componentsQueryKeys.detail(componentId),
    onePagersQueryKeys.completenessForSubjectType('application'),
  ],

  addExpert: (componentId: string) => [
    componentsQueryKeys.detail(componentId),
    componentsQueryKeys.lists(),
    componentsQueryKeys.expertRoles(),
    auditQueryKeys.history(componentId),
    ...onePagerFreshness(componentId),
  ],

  removeExpert: (componentId: string) => [
    componentsQueryKeys.detail(componentId),
    componentsQueryKeys.lists(),
    componentsQueryKeys.expertRoles(),
    auditQueryKeys.history(componentId),
    ...onePagerFreshness(componentId),
  ],
};

export const fitScoresMutationEffects = {
  set: (componentId: string) => [fitScoresQueryKeys.byComponent(componentId), strategicFitAnalysisQueryKeys.all],

  delete: (componentId: string) => [fitScoresQueryKeys.byComponent(componentId), strategicFitAnalysisQueryKeys.all],
};
