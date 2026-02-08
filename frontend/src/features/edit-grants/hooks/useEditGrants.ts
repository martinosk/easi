import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { editGrantApi } from '../api/editGrantApi';
import { editGrantsQueryKeys } from '../queryKeys';
import { editGrantsMutationEffects } from '../mutationEffects';
import { invalidateFor } from '../../../lib/invalidateFor';
import type { EditGrant, CreateEditGrantRequest, ArtifactType } from '../types';
import toast from 'react-hot-toast';

export function useMyEditGrants() {
  return useQuery<EditGrant[]>({
    queryKey: editGrantsQueryKeys.mine(),
    queryFn: () => editGrantApi.getMyGrants(),
  });
}

export function useEditGrantsForArtifact(artifactType: ArtifactType | undefined, artifactId: string | undefined) {
  return useQuery<EditGrant[]>({
    queryKey: editGrantsQueryKeys.forArtifact(artifactType!, artifactId!),
    queryFn: () => editGrantApi.getForArtifact(artifactType!, artifactId!),
    enabled: !!artifactType && !!artifactId,
  });
}

export function useCreateEditGrant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateEditGrantRequest) => editGrantApi.create(request),
    onSuccess: (grant) => {
      invalidateFor(queryClient, editGrantsMutationEffects.create());
      toast.success(`Edit access granted to ${grant.granteeEmail}`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to grant edit access');
    },
  });
}

export function useRevokeEditGrant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => editGrantApi.revoke(id),
    onSuccess: () => {
      invalidateFor(queryClient, editGrantsMutationEffects.revoke());
      toast.success('Edit access revoked');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to revoke edit access');
    },
  });
}
