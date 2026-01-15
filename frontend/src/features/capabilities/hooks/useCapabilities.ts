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
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateCapabilityRequest) => capabilitiesApi.create(request),
    onSuccess: (newCapability) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.create({
          parentId: newCapability.parentId ?? undefined,
          businessDomainId: undefined,
        })
      );
      toast.success(`Capability "${newCapability.name}" created`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create capability');
    },
  });
}

export function useUpdateCapability() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      capability,
      request,
    }: {
      capability: Capability;
      request: UpdateCapabilityRequest;
    }) => capabilitiesApi.update(capability, request),
    onSuccess: (updatedCapability) => {
      invalidateFor(queryClient, mutationEffects.capabilities.update(updatedCapability.id));
      toast.success(`Capability "${updatedCapability.name}" updated`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update capability');
    },
  });
}

export function useUpdateCapabilityMetadata() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      request,
    }: {
      id: CapabilityId;
      request: UpdateCapabilityMetadataRequest;
    }) => capabilitiesApi.updateMetadata(id, request),
    onSuccess: (updatedCapability) => {
      invalidateFor(queryClient, mutationEffects.capabilities.update(updatedCapability.id));
      toast.success('Capability metadata updated');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update capability metadata');
    },
  });
}

export function useAddCapabilityExpert() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, request }: { id: CapabilityId; request: AddCapabilityExpertRequest }) =>
      capabilitiesApi.addExpert(id, request),
    onSuccess: (_, { id }) => {
      invalidateFor(queryClient, mutationEffects.capabilities.addExpert(id));
      toast.success('Expert added');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add expert');
    },
  });
}

export function useCapabilityExpertRoles() {
  return useQuery({
    queryKey: queryKeys.capabilities.expertRoles(),
    queryFn: () => capabilitiesApi.getExpertRoles(),
  });
}

export function useRemoveCapabilityExpert() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      expert,
    }: {
      id: CapabilityId;
      expert: { name: string; role: string; contact: string };
    }) => capabilitiesApi.removeExpert(id, expert),
    onSuccess: (_, { id }) => {
      invalidateFor(queryClient, mutationEffects.capabilities.removeExpert(id));
      toast.success('Expert removed');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove expert');
    },
  });
}

export function useAddCapabilityTag() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, request }: { id: CapabilityId; request: AddCapabilityTagRequest }) =>
      capabilitiesApi.addTag(id, request),
    onSuccess: (_, { id }) => {
      invalidateFor(queryClient, mutationEffects.capabilities.addTag(id));
      toast.success('Tag added');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add tag');
    },
  });
}

export function useDeleteCapability() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (context: { capability: Capability; parentId?: string; domainId?: string }) =>
      capabilitiesApi.delete(context.capability),
    onSuccess: (_, context) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.delete({
          id: context.capability.id,
          parentId: context.parentId,
          domainId: context.domainId,
        })
      );
      toast.success('Capability deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete capability');
    },
  });
}

export function useChangeCapabilityParent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (context: { id: CapabilityId; oldParentId?: string; newParentId?: CapabilityId | null }) =>
      capabilitiesApi.changeParent(context.id, context.newParentId ?? null),
    onSuccess: (_, context) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.changeParent({
          id: context.id,
          oldParentId: context.oldParentId,
          newParentId: context.newParentId ?? undefined,
        })
      );
      toast.success('Capability parent updated');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to change parent');
    },
  });
}

export function useCreateCapabilityDependency() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateCapabilityDependencyRequest) =>
      capabilitiesApi.createDependency(request),
    onSuccess: (newDependency) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.addDependency(
          newDependency.sourceCapabilityId,
          newDependency.targetCapabilityId
        )
      );
      toast.success('Dependency created');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create dependency');
    },
  });
}

export function useDeleteCapabilityDependency() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (dependency: CapabilityDependency) => capabilitiesApi.deleteDependency(dependency),
    onSuccess: (_, dependency) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.removeDependency(
          dependency.sourceCapabilityId,
          dependency.targetCapabilityId
        )
      );
      toast.success('Dependency deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete dependency');
    },
  });
}

export function useLinkSystemToCapability() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      capabilityId,
      request,
    }: {
      capabilityId: CapabilityId;
      request: LinkSystemToCapabilityRequest;
    }) => capabilitiesApi.linkSystem(capabilityId, request),
    onSuccess: (_, { capabilityId, request }) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.linkSystem(capabilityId, request.componentId)
      );
      toast.success('System linked to capability');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to link system');
    },
  });
}

export function useUpdateRealization() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (context: {
      realization: CapabilityRealization;
      request: UpdateRealizationRequest;
    }) => capabilitiesApi.updateRealization(context.realization, context.request),
    onSuccess: (_, context) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.updateRealization(context.realization.capabilityId, context.realization.componentId)
      );
      toast.success('Realization updated');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update realization');
    },
  });
}

export function useDeleteRealization() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (realization: CapabilityRealization) => capabilitiesApi.deleteRealization(realization),
    onSuccess: (_, realization) => {
      invalidateFor(
        queryClient,
        mutationEffects.capabilities.deleteRealization(realization.capabilityId, realization.componentId)
      );
      toast.success('Realization deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete realization');
    },
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
