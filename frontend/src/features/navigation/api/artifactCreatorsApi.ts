import { httpClient } from '../../../api/core/httpClient';
import type { ArtifactCreator } from '../utils/filterByCreator';

interface ArtifactCreatorsApiResponse {
  data: ArtifactCreator[];
  _links: Record<string, { href: string; method: string }>;
}

export const artifactCreatorsApi = {
  async getAll(): Promise<ArtifactCreator[]> {
    const response = await httpClient.get<ArtifactCreatorsApiResponse>('/api/v1/artifact-creators');
    return response.data.data;
  },
};
