import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { invalidateFor } from '../../../lib/invalidateFor';
import { editGrantApi } from '../api/editGrantApi';
import { editGrantsMutationEffects } from '../mutationEffects';
import { editGrantsQueryKeys } from '../queryKeys';
import type { ArtifactType, CreateEditGrantRequest, EditGrant } from '../types';

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
      if (grant.invitationCreated) {
        toast.success(`Edit access granted. An invitation to join EASI was also created for ${grant.granteeEmail}.`);
      } else {
        toast.success(`Edit access granted to ${grant.granteeEmail}`);
      }
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
