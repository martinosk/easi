import type { Component, HostingClassification } from '../../../api/types';
import { componentsApi } from '../api';
import { componentsMutationEffects } from '../mutationEffects';
import { useComponentMutation } from './useComponents';

export function useClassifyComponentHosting() {
  return useComponentMutation(
    ({ component, hosting }: { component: Component; hosting: HostingClassification }) =>
      componentsApi.classifyHosting(component, hosting),
    (_, { component }) => componentsMutationEffects.hosting(component.id),
    'Hosting classified',
    'Failed to classify hosting',
  );
}
