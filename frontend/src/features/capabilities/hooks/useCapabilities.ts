import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { capabilitiesApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import { invalidateFor } from '../../../lib/invalidateFor';
import { mutationEffects } from '../../../lib/mutationEffects';
import type {
  Capability,
  CapabilityId,
  CapabilityDependency,
  CapabilityRealization,
  ComponentId,
  CreateCapabilityRequest,
  UpdateCapabilityRequest,
  UpdateCapabilityMetadataRequest,
  AddCapabilityExpertRequest,
  AddCapabilityTagRequest,
  CreateCapabilityDependencyRequest,
  LinkSystemToCapabilityRequest,
  UpdateRealizationRequest,
} from '../../../api/types';
import toast from 'react-hot-toast';

type InvalidationEffect = ReadonlyArray<readonly unknown[]>;

type MutationConfig<TData, TVariables> = {
  mutationFn: (variables: TVariables) => Promise<TData>;
  getEffects: (data: TData, variables: TVariables) => InvalidationEffect;
  successMessage: string | ((data: TData, variables: TVariables) => string);
  errorMessage: string;
};

function useCapabilityMutation<TData, TVariables>(config: MutationConfig<TData, TVariables>) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: config.mutationFn,
    onSuccess: (data, variables) => {
      invalidateFor(queryClient, config.getEffects(data, variables));
      const message = typeof config.successMessage === 'function'
        ? config.successMessage(data, variables)
        : config.successMessage;
      toast.success(message);
    },
    onError: (error: Error) => {
      toast.error(error.message || config.errorMessage);
    },
  });
}

type IdRequestPair<TRequest> = { id: CapabilityId; request: TRequest };

function createIdRequestMutation<TRequest, TData>(
  apiFn: (id: CapabilityId, request: TRequest) => Promise<TData>,
  getEffects: (id: CapabilityId) => InvalidationEffect,
  successMessage: string,
  errorMessage: string
) {
  return () =>
    useCapabilityMutation<TData, IdRequestPair<TRequest>>({
      mutationFn: ({ id, request }) => apiFn(id, request),
      getEffects: (_, { id }) => getEffects(id),
      successMessage,
      errorMessage,
    });
}

export function useCapabilities() {
  return useQuery({
    queryKey: queryKeys.capabilities.lists(),
    queryFn: () => capabilitiesApi.getAll(),
  });
}

export function useCapability(id: CapabilityId | undefined) {
  return useQuery({
    queryKey: queryKeys.capabilities.detail(id!),
    queryFn: () => capabilitiesApi.getById(id!),
    enabled: !!id,
  });
}

export function useCapabilityChildren(id: CapabilityId | undefined) {
  return useQuery({
    queryKey: queryKeys.capabilities.children(id!),
    queryFn: () => capabilitiesApi.getChildren(id!),
    enabled: !!id,
  });
}

export function useCapabilityDependencies() {
  return useQuery({
    queryKey: queryKeys.capabilities.dependencies(),
    queryFn: () => capabilitiesApi.getAllDependencies(),
  });
}

export function useOutgoingDependencies(capabilityId: CapabilityId | undefined) {
  return useQuery({
    queryKey: queryKeys.capabilities.outgoing(capabilityId!),
    queryFn: () => capabilitiesApi.getOutgoingDependencies(capabilityId!),
    enabled: !!capabilityId,
  });
}

export function useIncomingDependencies(capabilityId: CapabilityId | undefined) {
  return useQuery({
    queryKey: queryKeys.capabilities.incoming(capabilityId!),
    queryFn: () => capabilitiesApi.getIncomingDependencies(capabilityId!),
    enabled: !!capabilityId,
  });
}

export function useCapabilityRealizations(capabilityId: CapabilityId | undefined) {
  return useQuery({
    queryKey: queryKeys.capabilities.realizations(capabilityId!),
    queryFn: () => capabilitiesApi.getSystemsByCapability(capabilityId!),
    enabled: !!capabilityId,
  });
}

export function useCapabilitiesByComponent(componentId: ComponentId | undefined) {
  return useQuery({
    queryKey: queryKeys.capabilities.byComponent(componentId!),
    queryFn: () => capabilitiesApi.getCapabilitiesByComponent(componentId!),
    enabled: !!componentId,
  });
}

export function useCreateCapability() {
  return useCapabilityMutation({
    mutationFn: (request: CreateCapabilityRequest) => capabilitiesApi.create(request),
    getEffects: (newCapability) =>
      mutationEffects.capabilities.create({
        parentId: newCapability.parentId ?? undefined,
        businessDomainId: undefined,
      }),
    successMessage: (cap) => `Capability "${cap.name}" created`,
    errorMessage: 'Failed to create capability',
  });
}

export function useUpdateCapability() {
  return useCapabilityMutation({
    mutationFn: ({ capability, request }: { capability: Capability; request: UpdateCapabilityRequest }) =>
      capabilitiesApi.update(capability, request),
    getEffects: (updatedCapability) => mutationEffects.capabilities.update(updatedCapability.id),
    successMessage: (cap) => `Capability "${cap.name}" updated`,
    errorMessage: 'Failed to update capability',
  });
}

export const useUpdateCapabilityMetadata = createIdRequestMutation<UpdateCapabilityMetadataRequest, Capability>(
  capabilitiesApi.updateMetadata,
  (id) => mutationEffects.capabilities.update(id),
  'Capability metadata updated',
  'Failed to update capability metadata'
);

export const useAddCapabilityExpert = createIdRequestMutation<AddCapabilityExpertRequest, void>(
  capabilitiesApi.addExpert,
  (id) => mutationEffects.capabilities.addExpert(id),
  'Expert added',
  'Failed to add expert'
);

export function useCapabilityExpertRoles() {
  return useQuery({
    queryKey: queryKeys.capabilities.expertRoles(),
    queryFn: () => capabilitiesApi.getExpertRoles(),
  });
}

export function useRemoveCapabilityExpert() {
  return useCapabilityMutation({
    mutationFn: ({ id, expert }: { id: CapabilityId; expert: { name: string; role: string; contact: string } }) =>
      capabilitiesApi.removeExpert(id, expert),
    getEffects: (_, { id }) => mutationEffects.capabilities.removeExpert(id),
    successMessage: 'Expert removed',
    errorMessage: 'Failed to remove expert',
  });
}

export const useAddCapabilityTag = createIdRequestMutation<AddCapabilityTagRequest, void>(
  capabilitiesApi.addTag,
  (id) => mutationEffects.capabilities.addTag(id),
  'Tag added',
  'Failed to add tag'
);

export function useDeleteCapability() {
  return useCapabilityMutation({
    mutationFn: (context: { capability: Capability; parentId?: string; domainId?: string }) =>
      capabilitiesApi.delete(context.capability),
    getEffects: (_, context) =>
      mutationEffects.capabilities.delete({
        id: context.capability.id,
        parentId: context.parentId,
        domainId: context.domainId,
      }),
    successMessage: 'Capability deleted',
    errorMessage: 'Failed to delete capability',
  });
}

export function useChangeCapabilityParent() {
  return useCapabilityMutation({
    mutationFn: (context: { id: CapabilityId; oldParentId?: string; newParentId?: CapabilityId | null }) =>
      capabilitiesApi.changeParent(context.id, context.newParentId ?? null),
    getEffects: (_, context) =>
      mutationEffects.capabilities.changeParent({
        id: context.id,
        oldParentId: context.oldParentId,
        newParentId: context.newParentId ?? undefined,
      }),
    successMessage: 'Capability parent updated',
    errorMessage: 'Failed to change parent',
  });
}

export function useCreateCapabilityDependency() {
  return useCapabilityMutation({
    mutationFn: (request: CreateCapabilityDependencyRequest) => capabilitiesApi.createDependency(request),
    getEffects: (newDependency) =>
      mutationEffects.capabilities.addDependency(newDependency.sourceCapabilityId, newDependency.targetCapabilityId),
    successMessage: 'Dependency created',
    errorMessage: 'Failed to create dependency',
  });
}

export function useDeleteCapabilityDependency() {
  return useCapabilityMutation({
    mutationFn: (dependency: CapabilityDependency) => capabilitiesApi.deleteDependency(dependency),
    getEffects: (_, dependency) =>
      mutationEffects.capabilities.removeDependency(dependency.sourceCapabilityId, dependency.targetCapabilityId),
    successMessage: 'Dependency deleted',
    errorMessage: 'Failed to delete dependency',
  });
}

export function useLinkSystemToCapability() {
  return useCapabilityMutation({
    mutationFn: ({ capabilityId, request }: { capabilityId: CapabilityId; request: LinkSystemToCapabilityRequest }) =>
      capabilitiesApi.linkSystem(capabilityId, request),
    getEffects: (_, { capabilityId, request }) =>
      mutationEffects.capabilities.linkSystem(capabilityId, request.componentId),
    successMessage: 'System linked to capability',
    errorMessage: 'Failed to link system',
  });
}

export function useUpdateRealization() {
  return useCapabilityMutation({
    mutationFn: (context: { realization: CapabilityRealization; request: UpdateRealizationRequest }) =>
      capabilitiesApi.updateRealization(context.realization, context.request),
    getEffects: (_, context) =>
      mutationEffects.capabilities.updateRealization(context.realization.capabilityId, context.realization.componentId),
    successMessage: 'Realization updated',
    errorMessage: 'Failed to update realization',
  });
}

export function useDeleteRealization() {
  return useCapabilityMutation({
    mutationFn: (realization: CapabilityRealization) => capabilitiesApi.deleteRealization(realization),
    getEffects: (_, realization) =>
      mutationEffects.capabilities.deleteRealization(realization.capabilityId, realization.componentId),
    successMessage: 'Realization deleted',
    errorMessage: 'Failed to delete realization',
  });
}

export function useRealizationsForComponents(componentIds: ComponentId[]) {
  return useQuery({
    queryKey: queryKeys.capabilities.realizationsByComponents(componentIds),
    queryFn: async () => {
      const results = await Promise.all(
        componentIds.map((id) => capabilitiesApi.getCapabilitiesByComponent(id))
      );
      return results.flat();
    },
    enabled: componentIds.length > 0,
  });
}
