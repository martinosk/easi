import { useQuery } from '@tanstack/react-query';
import type { AllOriginRelationshipsResponse, OriginRelationship } from '../../../api/types';
import { toComponentId, toOriginRelationshipId } from '../../../api/types';
import { originEntitiesApi } from '../api/originEntitiesApi';
import { originRelationshipsQueryKeys } from '../queryKeys';

function transformToOriginRelationships(response: AllOriginRelationshipsResponse): OriginRelationship[] {
  const relationships: OriginRelationship[] = [];

  for (const rel of response.acquiredVia) {
    relationships.push({
      id: toOriginRelationshipId(rel.id),
      componentId: toComponentId(rel.componentId),
      componentName: rel.componentName,
      relationshipType: 'AcquiredVia',
      originEntityId: rel.acquiredEntityId,
      originEntityName: rel.acquiredEntityName,
      notes: rel.notes,
      createdAt: rel.createdAt,
      _links: rel._links,
    });
  }

  for (const rel of response.purchasedFrom) {
    relationships.push({
      id: toOriginRelationshipId(rel.id),
      componentId: toComponentId(rel.componentId),
      componentName: rel.componentName,
      relationshipType: 'PurchasedFrom',
      originEntityId: rel.vendorId,
      originEntityName: rel.vendorName,
      notes: rel.notes,
      createdAt: rel.createdAt,
      _links: rel._links,
    });
  }

  for (const rel of response.builtBy) {
    relationships.push({
      id: toOriginRelationshipId(rel.id),
      componentId: toComponentId(rel.componentId),
      componentName: rel.componentName,
      relationshipType: 'BuiltBy',
      originEntityId: rel.internalTeamId,
      originEntityName: rel.internalTeamName,
      notes: rel.notes,
      createdAt: rel.createdAt,
      _links: rel._links,
    });
  }

  return relationships;
}

export function useOriginRelationshipsQuery() {
  return useQuery({
    queryKey: originRelationshipsQueryKeys.lists(),
    queryFn: async () => {
      const response = await originEntitiesApi.getAllOriginRelationships();
      return transformToOriginRelationships(response);
    },
  });
}
