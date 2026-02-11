import { useState, useCallback } from 'react';
import { hasLink, type ResourceWithLinks } from '../../../utils/hateoas';
import { InviteToEditDialog } from './InviteToEditDialog';
import { useCreateEditGrant } from '../hooks/useEditGrants';
import type { ArtifactType, CreateEditGrantRequest } from '../types';

interface InviteToEditButtonProps {
  resource: ResourceWithLinks;
  artifactType: ArtifactType;
  artifactId: string;
  className?: string;
}

export function InviteToEditButton({
  resource,
  artifactType,
  artifactId,
  className,
}: InviteToEditButtonProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const createGrant = useCreateEditGrant();

  const handleSubmit = useCallback(
    async (request: CreateEditGrantRequest) => {
      await createGrant.mutateAsync(request);
    },
    [createGrant]
  );

  if (!hasLink(resource, 'x-edit-grants')) {
    return null;
  }

  return (
    <>
      <button
        className={className ?? 'btn btn-secondary btn-sm'}
        onClick={() => setIsDialogOpen(true)}
        data-testid="invite-to-edit-btn"
      >
        Invite to Edit...
      </button>
      <InviteToEditDialog
        isOpen={isDialogOpen}
        onClose={() => setIsDialogOpen(false)}
        onSubmit={handleSubmit}
        artifactType={artifactType}
        artifactId={artifactId}
      />
    </>
  );
}
