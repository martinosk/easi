import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { originEntitiesApi } from '../api/originEntitiesApi';
import { internalTeamsQueryKeys } from '../queryKeys';
import { invalidateFor } from '../../../lib/invalidateFor';
import { internalTeamsMutationEffects } from '../mutationEffects';
import type {
  InternalTeamId,
  CreateInternalTeamRequest,
  UpdateInternalTeamRequest,
  ComponentId,
} from '../../../api/types';
import toast from 'react-hot-toast';

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

interface MutationConfig<TArgs, TResult> {
  mutationFn: (args: TArgs) => Promise<TResult>;
  effects: (result: TResult, args: TArgs) => ReadonlyArray<readonly string[]>;
  successMessage: (result: TResult, args: TArgs) => string;
  errorMessage: string;
}

function useTeamMutation<TArgs, TResult>(config: MutationConfig<TArgs, TResult>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: config.mutationFn,
    onSuccess: (result, args) => {
      invalidateFor(queryClient, config.effects(result, args));
      toast.success(config.successMessage(result, args));
    },
    onError: () => toast.error(config.errorMessage),
  });
}

export function useCreateInternalTeam() {
  return useTeamMutation({
    mutationFn: (request: CreateInternalTeamRequest) => originEntitiesApi.internalTeams.create(request),
    effects: () => internalTeamsMutationEffects.create(),
    successMessage: (team) => `Internal team "${team.name}" created successfully`,
    errorMessage: 'Failed to create internal team',
  });
}

export function useUpdateInternalTeam() {
  return useTeamMutation({
    mutationFn: ({ id, request }: { id: InternalTeamId; request: UpdateInternalTeamRequest }) =>
      originEntitiesApi.internalTeams.update(id, request),
    effects: (_, { id }) => internalTeamsMutationEffects.update(id),
    successMessage: (team) => `Internal team "${team.name}" updated`,
    errorMessage: 'Failed to update internal team',
  });
}

export function useDeleteInternalTeam() {
  return useTeamMutation({
    mutationFn: ({ id }: { id: InternalTeamId; name: string }) => originEntitiesApi.internalTeams.delete(id),
    effects: (_, { id }) => internalTeamsMutationEffects.delete(id),
    successMessage: (_, { name }) => `Internal team "${name}" deleted`,
    errorMessage: 'Failed to delete internal team',
  });
}

export function useLinkComponentToInternalTeam() {
  return useTeamMutation({
    mutationFn: ({ componentId, teamId, notes }: { componentId: ComponentId; teamId: InternalTeamId; notes?: string }) =>
      originEntitiesApi.internalTeams.linkComponent(componentId, teamId, notes),
    effects: (_, { teamId, componentId }) => internalTeamsMutationEffects.linkComponent(teamId, componentId),
    successMessage: () => 'Component linked to internal team',
    errorMessage: 'Failed to link component to internal team',
  });
}

export function useUnlinkComponentFromInternalTeam() {
  return useTeamMutation({
    mutationFn: ({ componentId }: { teamId: InternalTeamId; componentId: ComponentId }) =>
      originEntitiesApi.internalTeams.unlinkComponent(componentId),
    effects: (_, { teamId, componentId }) => internalTeamsMutationEffects.unlinkComponent(teamId, componentId),
    successMessage: () => 'Component unlinked',
    errorMessage: 'Failed to unlink component',
  });
}
