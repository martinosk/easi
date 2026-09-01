import { useQuery } from '@tanstack/react-query';
import type { Component, OwnerReferenceRequest } from '../../../api/types';
import { componentsApi } from '../api';
import { componentsMutationEffects } from '../mutationEffects';
import { componentsQueryKeys } from '../queryKeys';
import { useComponentMutation } from './useComponents';

export function useOwnershipStatistics() {
  return useQuery({
    queryKey: componentsQueryKeys.ownershipStatistics(),
    queryFn: () => componentsApi.getOwnershipStatistics(),
  });
}

export function useNominateComponentOwner() {
  return useComponentMutation(
    ({ component, request }: { component: Component; request: OwnerReferenceRequest }) =>
      componentsApi.nominateOwner(component, request),
    (_, { component }) => componentsMutationEffects.ownership(component.id),
    'Owner nominated',
    'Failed to nominate owner',
  );
}

export function useConfirmComponentOwnership() {
  return useComponentMutation(
    (component: Component) => componentsApi.confirmOwnership(component),
    (_, component) => componentsMutationEffects.ownership(component.id),
    'Ownership confirmed',
    'Failed to confirm ownership',
  );
}

export function useAssignComponentOwner() {
  return useComponentMutation(
    ({ component, request }: { component: Component; request: OwnerReferenceRequest }) =>
      componentsApi.assignOwner(component, request),
    (_, { component }) => componentsMutationEffects.ownership(component.id),
    'Owner assigned',
    'Failed to assign owner',
  );
}

export function useClearComponentOwnership() {
  return useComponentMutation(
    (component: Component) => componentsApi.clearOwnership(component),
    (_, component) => componentsMutationEffects.ownership(component.id),
    'Ownership cleared',
    'Failed to clear ownership',
  );
}
