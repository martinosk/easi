import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { originEntitiesApi } from '../api/originEntitiesApi';
import { internalTeamsQueryKeys } from '../queryKeys';
import { invalidateFor } from '../../../lib/invalidateFor';
import { internalTeamsMutationEffects } from '../mutationEffects';
import type {
  InternalTeam,
  InternalTeamId,
  CreateInternalTeamRequest,
  UpdateInternalTeamRequest,
  ComponentId,
} from '../../../api/types';
import toast from 'react-hot-toast';

export interface UseInternalTeamsResult {
  internalTeams: InternalTeam[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  createTeam: (request: CreateInternalTeamRequest) => Promise<InternalTeam>;
  updateTeam: (id: InternalTeamId, request: UpdateInternalTeamRequest) => Promise<InternalTeam>;
  deleteTeam: (id: InternalTeamId, name: string) => Promise<void>;
}

export function useInternalTeams(): UseInternalTeamsResult {
  const query = useInternalTeamsQuery();
  const createMutation = useCreateInternalTeam();
  const updateMutation = useUpdateInternalTeam();
  const deleteMutation = useDeleteInternalTeam();

  const createTeam = useCallback(
    async (request: CreateInternalTeamRequest): Promise<InternalTeam> => {
      return createMutation.mutateAsync(request);
    },
    [createMutation]
  );

  const updateTeam = useCallback(
    async (id: InternalTeamId, request: UpdateInternalTeamRequest): Promise<InternalTeam> => {
      return updateMutation.mutateAsync({ id, request });
    },
    [updateMutation]
  );

  const deleteTeam = useCallback(
    async (id: InternalTeamId, name: string): Promise<void> => {
      await deleteMutation.mutateAsync({ id, name });
    },
    [deleteMutation]
  );

  const refetch = useCallback(async () => {
    await query.refetch();
  }, [query]);

  return {
    internalTeams: query.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    refetch,
    createTeam,
    updateTeam,
    deleteTeam,
  };
}

export function useInternalTeamsQuery() {
  return useQuery({
    queryKey: internalTeamsQueryKeys.lists(),
    queryFn: () => originEntitiesApi.internalTeams.getAll(),
  });
}

export function useInternalTeam(id: InternalTeamId | undefined) {
  return useQuery({
    queryKey: internalTeamsQueryKeys.detail(id!),
    queryFn: () => originEntitiesApi.internalTeams.getById(id!),
    enabled: !!id,
  });
}

export function useCreateInternalTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateInternalTeamRequest) =>
      originEntitiesApi.internalTeams.create(request),
    onSuccess: (newTeam) => {
      invalidateFor(queryClient, internalTeamsMutationEffects.create());
      toast.success(`Internal team "${newTeam.name}" created successfully`);
    },
    onError: () => {
      toast.error('Failed to create internal team');
    },
  });
}

export function useUpdateInternalTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, request }: { id: InternalTeamId; request: UpdateInternalTeamRequest }) =>
      originEntitiesApi.internalTeams.update(id, request),
    onSuccess: (updatedTeam, { id }) => {
      invalidateFor(queryClient, internalTeamsMutationEffects.update(id));
      toast.success(`Internal team "${updatedTeam.name}" updated`);
    },
    onError: () => {
      toast.error('Failed to update internal team');
    },
  });
}

export function useDeleteInternalTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }: { id: InternalTeamId; name: string }) =>
      originEntitiesApi.internalTeams.delete(id),
    onSuccess: (_, { id, name }) => {
      invalidateFor(queryClient, internalTeamsMutationEffects.delete(id));
      toast.success(`Internal team "${name}" deleted`);
    },
    onError: () => {
      toast.error('Failed to delete internal team');
    },
  });
}

export function useLinkComponentToInternalTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      componentId,
      teamId,
      notes,
    }: {
      componentId: ComponentId;
      teamId: InternalTeamId;
      notes?: string;
    }) => originEntitiesApi.internalTeams.linkComponent(componentId, teamId, notes),
    onSuccess: (_, { teamId, componentId }) => {
      invalidateFor(queryClient, internalTeamsMutationEffects.linkComponent(teamId, componentId));
      toast.success('Component linked to internal team');
    },
    onError: () => {
      toast.error('Failed to link component to internal team');
    },
  });
}

export function useUnlinkComponentFromInternalTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ componentId }: { teamId: InternalTeamId; componentId: ComponentId }) =>
      originEntitiesApi.internalTeams.unlinkComponent(componentId),
    onSuccess: (_, { teamId, componentId }) => {
      invalidateFor(queryClient, internalTeamsMutationEffects.unlinkComponent(teamId, componentId));
      toast.success('Component unlinked');
    },
    onError: () => {
      toast.error('Failed to unlink component');
    },
  });
}
