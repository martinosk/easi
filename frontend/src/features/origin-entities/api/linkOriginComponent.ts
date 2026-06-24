import { httpClient } from '../../../api/core/httpClient';
import type { OriginRelationship } from '../../../api/types';

export async function linkOriginComponent(
  componentId: string,
  rel: string,
  body: Record<string, unknown>,
): Promise<OriginRelationship> {
  const response = await httpClient.put<OriginRelationship>(
    `/api/v1/components/${componentId}/origin/${rel}`,
    { ...body, componentId },
  );
  return response.data;
}
