import { useEditGrantsForArtifact } from '../hooks/useEditGrants';
import type { ArtifactType } from '../types';

interface EditGrantBadgeProps {
  artifactType: ArtifactType;
  artifactId: string;
}

export function EditGrantBadge({ artifactType, artifactId }: EditGrantBadgeProps) {
  const { data: grants } = useEditGrantsForArtifact(artifactType, artifactId);

  const activeCount = grants?.filter(g => g.status === 'active').length ?? 0;

  if (activeCount === 0) return null;

  return (
    <span className="badge badge-info" data-testid="edit-grant-badge">
      {activeCount} active {activeCount === 1 ? 'grant' : 'grants'}
    </span>
  );
}
