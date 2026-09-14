import type { AttachComponentRequest, Component, ComponentId } from '../../../api/types';
import { componentsApi } from '../api';
import { componentsMutationEffects } from '../mutationEffects';
import { useComponentMutation } from './useComponents';

export function useAttachComponent() {
  return useComponentMutation(
    ({ component, request }: { component: Component; request: AttachComponentRequest }) =>
      componentsApi.attachToParent(component, request),
    (_, { component, request }) => componentsMutationEffects.containment(component.id, request.parentId),
    'Application attached',
    'Failed to attach application',
  );
}

export function useAttachComponentById() {
  return useComponentMutation(
    async ({ partId, request }: { partId: ComponentId; request: AttachComponentRequest }) =>
      componentsApi.attachToParent(await componentsApi.getById(partId), request),
    (_, { partId, request }) => componentsMutationEffects.containment(partId, request.parentId),
    'Application attached',
    'Failed to attach application',
  );
}

export function useDetachComponent() {
  return useComponentMutation(
    (component: Component) => componentsApi.detach(component),
    (_, component) => componentsMutationEffects.containment(component.id, component.partOf?.id ?? component.id),
    'Application detached',
    'Failed to detach application',
  );
}
