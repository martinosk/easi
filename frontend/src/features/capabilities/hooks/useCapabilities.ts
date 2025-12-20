import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { capabilitiesApi } from '../api';
import { queryKeys } from '../../../lib/queryClient';
import type {
  Capability,
  CapabilityId,
  ComponentId,
  RealizationId,
  CapabilityDependencyId,
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
      queryClient.setQueryData<Capability[]>(
        queryKeys.capabilities.lists(),
        (old) => (old ? [...old, newCapability] : [newCapability])
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
      id,
      request,
    }: {
      id: CapabilityId;
      request: UpdateCapabilityRequest;
    }) => capabilitiesApi.update(id, request),
    onSuccess: (updatedCapability) => {
      queryClient.setQueryData<Capability[]>(
        queryKeys.capabilities.lists(),
        (old) =>
          old?.map((c) =>
            c.id === updatedCapability.id ? updatedCapability : c
          ) ?? []
      );
      queryClient.setQueryData(
        queryKeys.capabilities.detail(updatedCapability.id),
        updatedCapability
      );
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
      queryClient.setQueryData<Capability[]>(
        queryKeys.capabilities.lists(),
        (old) =>
          old?.map((c) =>
            c.id === updatedCapability.id ? updatedCapability : c
          ) ?? []
      );
      queryClient.setQueryData(
        queryKeys.capabilities.detail(updatedCapability.id),
        updatedCapability
      );
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
    mutationFn: ({
      id,
      request,
    }: {
      id: CapabilityId;
      request: AddCapabilityExpertRequest;
    }) => capabilitiesApi.addExpert(id, request),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.detail(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.lists() });
      toast.success('Expert added');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to add expert');
    },
  });
}

export function useAddCapabilityTag() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      request,
    }: {
      id: CapabilityId;
      request: AddCapabilityTagRequest;
    }) => capabilitiesApi.addTag(id, request),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.detail(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.lists() });
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
    mutationFn: (id: CapabilityId) => capabilitiesApi.delete(id),
    onSuccess: (_, deletedId) => {
      queryClient.setQueryData<Capability[]>(
        queryKeys.capabilities.lists(),
        (old) => old?.filter((c) => c.id !== deletedId) ?? []
      );
      queryClient.removeQueries({
        queryKey: queryKeys.capabilities.detail(deletedId),
      });
      queryClient.invalidateQueries({ queryKey: queryKeys.businessDomains.all });
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
    mutationFn: ({
      id,
      parentId,
    }: {
      id: CapabilityId;
      parentId: CapabilityId | null;
    }) => capabilitiesApi.changeParent(id, parentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.businessDomains.all });
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
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.dependencies() });
      queryClient.invalidateQueries({
        queryKey: queryKeys.capabilities.outgoing(newDependency.sourceCapabilityId),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.capabilities.incoming(newDependency.targetCapabilityId),
      });
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
    mutationFn: (id: CapabilityDependencyId) => capabilitiesApi.deleteDependency(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.dependencies() });
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.all });
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
    onSuccess: (_, { capabilityId }) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.capabilities.realizations(capabilityId),
      });
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
    mutationFn: ({
      id,
      request,
    }: {
      id: RealizationId;
      request: UpdateRealizationRequest;
    }) => capabilitiesApi.updateRealization(id, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.all });
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
    mutationFn: (id: RealizationId) => capabilitiesApi.deleteRealization(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.capabilities.all });
      toast.success('Realization deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete realization');
    },
  });
}

export function useRealizationsForComponents(componentIds: ComponentId[]) {
  return useQuery({
    queryKey: ['realizations', 'byComponents', componentIds.sort().join(',')],
    queryFn: async () => {
      const results = await Promise.all(
        componentIds.map((id) => capabilitiesApi.getCapabilitiesByComponent(id))
      );
      return results.flat();
    },
    enabled: componentIds.length > 0,
  });
}
