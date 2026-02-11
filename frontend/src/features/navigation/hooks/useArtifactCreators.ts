import { useQuery } from '@tanstack/react-query';
import { artifactCreatorsApi } from '../api/artifactCreatorsApi';
import type { ArtifactCreator } from '../utils/filterByCreator';

export const artifactCreatorsQueryKeys = {
  all: ['artifact-creators'] as const,
};

export function useArtifactCreators() {
  return useQuery<ArtifactCreator[]>({
    queryKey: artifactCreatorsQueryKeys.all,
    queryFn: () => artifactCreatorsApi.getAll(),
    staleTime: 1000 * 60 * 5,
  });
}
