import { componentsQueryKeys, fitScoresQueryKeys } from './queryKeys';
import { businessDomainsQueryKeys } from '../business-domains/queryKeys';
import { strategicFitAnalysisQueryKeys } from '../enterprise-architecture/queryKeys';
import { auditQueryKeys } from '../audit/queryKeys';
import { artifactCreatorsQueryKeys } from '../navigation/hooks/useArtifactCreators';

export const componentsMutationEffects = {
  create: () => [
    componentsQueryKeys.lists(),
    artifactCreatorsQueryKeys.all,
  ],

  update: (componentId: string) => [
    componentsQueryKeys.lists(),
    componentsQueryKeys.detail(componentId),
    businessDomainsQueryKeys.all,
    auditQueryKeys.history(componentId),
  ],

  delete: (componentId: string) => [
    componentsQueryKeys.lists(),
    componentsQueryKeys.detail(componentId),
  ],

  addExpert: (componentId: string) => [
    componentsQueryKeys.detail(componentId),
    componentsQueryKeys.lists(),
    componentsQueryKeys.expertRoles(),
    auditQueryKeys.history(componentId),
  ],

  removeExpert: (componentId: string) => [
    componentsQueryKeys.detail(componentId),
    componentsQueryKeys.lists(),
    componentsQueryKeys.expertRoles(),
    auditQueryKeys.history(componentId),
  ],
};

export const fitScoresMutationEffects = {
  set: (componentId: string) => [
    fitScoresQueryKeys.byComponent(componentId),
    strategicFitAnalysisQueryKeys.all,
  ],

  delete: (componentId: string) => [
    fitScoresQueryKeys.byComponent(componentId),
    strategicFitAnalysisQueryKeys.all,
  ],
};
