import { Badge } from '@mantine/core';
import { useEditGrantsForArtifact } from '../hooks/useEditGrants';
import type { ArtifactType } from '../types';

interface EditGrantBadgeProps {
  artifactType: ArtifactType;
  artifactId: string;
}

export function EditGrantBadge({ artifactType, artifactId }: EditGrantBadgeProps) {
  const { data: grants } = useEditGrantsForArtifact(artifactType, artifactId);

  const activeCount = grants?.filter((g) => g.status === 'active').length ?? 0;

  if (activeCount === 0) return null;

  return (
    <Badge color="blue" variant="light" data-testid="edit-grant-badge">
      {activeCount} active {activeCount === 1 ? 'grant' : 'grants'}
    </Badge>
  );
}
