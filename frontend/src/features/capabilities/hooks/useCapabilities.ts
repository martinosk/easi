import { useQuery, useMutation, useQueryClient, QueryClient } from '@tanstack/react-query';
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

type MutationConfig<TData, TVariables> = {
  mutationFn: (variables: TVariables) => Promise<TData>;
  onSuccessEffect: (queryClient: QueryClient, data: TData, variables: TVariables) => void;
  successMessage: string | ((data: TData, variables: TVariables) => string);
  errorMessage: string;
};

function useCapabilityMutation<TData, TVariables>(config: MutationConfig<TData, TVariables>) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: config.mutationFn,
    onSuccess: (data, variables) => {
      config.onSuccessEffect(queryClient, data, variables);
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
    onSuccessEffect: (queryClient, newCapability) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.create({
          parentId: newCapability.parentId ?? undefined,
          businessDomainId: undefined,
        })
      );
    },
    successMessage: (cap) => `Capability "${cap.name}" created`,
    errorMessage: 'Failed to create capability',
  });
}

export function useUpdateCapability() {
  return useCapabilityMutation({
    mutationFn: ({ capability, request }: { capability: Capability; request: UpdateCapabilityRequest }) =>
      capabilitiesApi.update(capability, request),
    onSuccessEffect: (queryClient, updatedCapability) => {
      invalidateFor(queryClient, mutationEffects.capabilities.update(updatedCapability.id));
    },
    successMessage: (cap) => `Capability "${cap.name}" updated`,
    errorMessage: 'Failed to update capability',
  });
}

export function useUpdateCapabilityMetadata() {
  return useCapabilityMutation({
    mutationFn: ({ id, request }: { id: CapabilityId; request: UpdateCapabilityMetadataRequest }) =>
      capabilitiesApi.updateMetadata(id, request),
    onSuccessEffect: (queryClient, updatedCapability) => {
      invalidateFor(queryClient, mutationEffects.capabilities.update(updatedCapability.id));
    },
    successMessage: 'Capability metadata updated',
    errorMessage: 'Failed to update capability metadata',
  });
}

export function useAddCapabilityExpert() {
  return useCapabilityMutation({
    mutationFn: ({ id, request }: { id: CapabilityId; request: AddCapabilityExpertRequest }) =>
      capabilitiesApi.addExpert(id, request),
    onSuccessEffect: (queryClient, _, { id }) => {
      invalidateFor(queryClient, mutationEffects.capabilities.addExpert(id));
    },
    successMessage: 'Expert added',
    errorMessage: 'Failed to add expert',
  });
}

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
    onSuccessEffect: (queryClient, _, { id }) => {
      invalidateFor(queryClient, mutationEffects.capabilities.removeExpert(id));
    },
    successMessage: 'Expert removed',
    errorMessage: 'Failed to remove expert',
  });
}

export function useAddCapabilityTag() {
  return useCapabilityMutation({
    mutationFn: ({ id, request }: { id: CapabilityId; request: AddCapabilityTagRequest }) =>
      capabilitiesApi.addTag(id, request),
    onSuccessEffect: (queryClient, _, { id }) => {
      invalidateFor(queryClient, mutationEffects.capabilities.addTag(id));
    },
    successMessage: 'Tag added',
    errorMessage: 'Failed to add tag',
  });
}

export function useDeleteCapability() {
  return useCapabilityMutation({
    mutationFn: (context: { capability: Capability; parentId?: string; domainId?: string }) =>
      capabilitiesApi.delete(context.capability),
    onSuccessEffect: (queryClient, _, context) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.delete({
          id: context.capability.id,
          parentId: context.parentId,
          domainId: context.domainId,
        })
      );
    },
    successMessage: 'Capability deleted',
    errorMessage: 'Failed to delete capability',
  });
}

export function useChangeCapabilityParent() {
  return useCapabilityMutation({
    mutationFn: (context: { id: CapabilityId; oldParentId?: string; newParentId?: CapabilityId | null }) =>
      capabilitiesApi.changeParent(context.id, context.newParentId ?? null),
    onSuccessEffect: (queryClient, _, context) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.changeParent({
          id: context.id,
          oldParentId: context.oldParentId,
          newParentId: context.newParentId ?? undefined,
        })
      );
    },
    successMessage: 'Capability parent updated',
    errorMessage: 'Failed to change parent',
  });
}

export function useCreateCapabilityDependency() {
  return useCapabilityMutation({
    mutationFn: (request: CreateCapabilityDependencyRequest) => capabilitiesApi.createDependency(request),
    onSuccessEffect: (queryClient, newDependency) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.addDependency(
          newDependency.sourceCapabilityId,
          newDependency.targetCapabilityId
        )
      );
    },
    successMessage: 'Dependency created',
    errorMessage: 'Failed to create dependency',
  });
}

export function useDeleteCapabilityDependency() {
  return useCapabilityMutation({
    mutationFn: (dependency: CapabilityDependency) => capabilitiesApi.deleteDependency(dependency),
    onSuccessEffect: (queryClient, _, dependency) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.removeDependency(
          dependency.sourceCapabilityId,
          dependency.targetCapabilityId
        )
      );
    },
    successMessage: 'Dependency deleted',
    errorMessage: 'Failed to delete dependency',
  });
}

export function useLinkSystemToCapability() {
  return useCapabilityMutation({
    mutationFn: ({ capabilityId, request }: { capabilityId: CapabilityId; request: LinkSystemToCapabilityRequest }) =>
      capabilitiesApi.linkSystem(capabilityId, request),
    onSuccessEffect: (queryClient, _, { capabilityId, request }) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.linkSystem(capabilityId, request.componentId)
      );
    },
    successMessage: 'System linked to capability',
    errorMessage: 'Failed to link system',
  });
}

export function useUpdateRealization() {
  return useCapabilityMutation({
    mutationFn: (context: { realization: CapabilityRealization; request: UpdateRealizationRequest }) =>
      capabilitiesApi.updateRealization(context.realization, context.request),
    onSuccessEffect: (queryClient, _, context) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.updateRealization(context.realization.capabilityId, context.realization.componentId)
      );
    },
    successMessage: 'Realization updated',
    errorMessage: 'Failed to update realization',
  });
}

export function useDeleteRealization() {
  return useCapabilityMutation({
    mutationFn: (realization: CapabilityRealization) => capabilitiesApi.deleteRealization(realization),
    onSuccessEffect: (queryClient, _, realization) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.deleteRealization(realization.capabilityId, realization.componentId)
      );
    },
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
