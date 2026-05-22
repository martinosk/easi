import { Button, type ButtonProps } from '@mantine/core';
import { useCallback, useState } from 'react';
import { hasLink, type ResourceWithLinks } from '../../../utils/hateoas';
import { useCreateEditGrant } from '../hooks/useEditGrants';
import type { ArtifactType, CreateEditGrantRequest } from '../types';
import { InviteToEditDialog } from './InviteToEditDialog';

interface InviteToEditButtonProps {
  resource: ResourceWithLinks;
  artifactType: ArtifactType;
  artifactId: string;
  size?: ButtonProps['size'];
}

export function InviteToEditButton({ resource, artifactType, artifactId, size = 'xs' }: InviteToEditButtonProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const createGrant = useCreateEditGrant();

  const handleSubmit = useCallback(
    async (request: CreateEditGrantRequest) => {
      await createGrant.mutateAsync(request);
    },
    [createGrant],
  );

  if (!hasLink(resource, 'x-edit-grants')) {
    return null;
  }

  return (
    <>
      <Button
        variant="default"
        size={size}
        onClick={() => setIsDialogOpen(true)}
        data-testid="invite-to-edit-btn"
      >
        Invite to Edit...
      </Button>
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
