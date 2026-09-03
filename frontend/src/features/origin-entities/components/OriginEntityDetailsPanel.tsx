import type React from 'react';
import type { AcquiredEntityId, InternalTeamId, OriginRelationshipType, VendorId } from '../../../api/types';
import { DetailPanelFailure, DetailPanelLoading } from '../../../components/shared/DetailPanelStatus';
import type { OriginEntityType } from '../../../constants/entityIdentifiers';
import { useAcquiredEntity } from '../hooks/useAcquiredEntities';
import { useInternalTeam } from '../hooks/useInternalTeams';
import { useOriginRelationshipsQuery } from '../hooks/useOriginRelationships';
import { useVendor } from '../hooks/useVendors';
import { type OriginEntity, OriginEntityDetailsContent } from './OriginEntityDetailsContent';

const RELATIONSHIP_TYPES: Record<OriginEntityType, OriginRelationshipType> = {
  acquired: 'AcquiredVia',
  vendor: 'PurchasedFrom',
  team: 'BuiltBy',
};

interface OriginEntityResult {
  entity: OriginEntity | undefined;
  isLoading: boolean;
  error: Error | null;
}

function useOriginEntity(entityType: OriginEntityType, entityId: string): OriginEntityResult {
  const acquiredQuery = useAcquiredEntity(entityType === 'acquired' ? (entityId as AcquiredEntityId) : undefined);
  const vendorQuery = useVendor(entityType === 'vendor' ? (entityId as VendorId) : undefined);
  const teamQuery = useInternalTeam(entityType === 'team' ? (entityId as InternalTeamId) : undefined);
  const query = { acquired: acquiredQuery, vendor: vendorQuery, team: teamQuery }[entityType];
  return { entity: query.data, isLoading: query.isLoading, error: query.error };
}

export interface OriginEntityDetailsPanelProps {
  entityType: OriginEntityType;
  entityId: string;
  viewMembership?: React.ReactNode;
}

export const OriginEntityDetailsPanel: React.FC<OriginEntityDetailsPanelProps> = ({
  entityType,
  entityId,
  viewMembership,
}) => {
  const { entity, isLoading, error } = useOriginEntity(entityType, entityId);
  const { data: allRelationships = [] } = useOriginRelationshipsQuery();

  if (isLoading) return <DetailPanelLoading />;
  if (error || !entity) return <DetailPanelFailure message="Failed to load entity" />;

  const relationships = allRelationships.filter(
    (relationship) =>
      relationship.relationshipType === RELATIONSHIP_TYPES[entityType] && relationship.originEntityId === entityId,
  );

  return (
    <OriginEntityDetailsContent
      entityType={entityType}
      entity={entity}
      relationships={relationships}
      viewMembership={viewMembership}
    />
  );
};
