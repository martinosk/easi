import React from 'react';
import type {
  AcquiredEntityId,
  InternalTeamId,
  OriginRelationship,
  OriginRelationshipType,
  VendorId,
} from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { EDGE_PREFIXES, ORIGIN_RELATIONSHIP_LABELS } from '../../../constants/entityIdentifiers';
import { useAppStore } from '../../../store/appStore';
import { hasLink } from '../../../utils/hateoas';
import { useUnlinkComponentFromAcquiredEntity } from '../hooks/useAcquiredEntities';
import { useUnlinkComponentFromInternalTeam } from '../hooks/useInternalTeams';
import { useOriginRelationshipsQuery } from '../hooks/useOriginRelationships';
import { useUnlinkComponentFromVendor } from '../hooks/useVendors';

const ORIGIN_EDGE_PREFIX = EDGE_PREFIXES.origin;

const isOriginEdge = (edgeId: string | null): boolean => edgeId !== null && edgeId.startsWith(ORIGIN_EDGE_PREFIX);

const extractRelationshipId = (edgeId: string): string => edgeId.replace(ORIGIN_EDGE_PREFIX, '');

interface RelationshipData {
  relationship: OriginRelationship;
  formattedDate: string;
  typeLabel: string;
}

const getRelationshipData = (
  selectedEdgeId: string | null,
  relationships: OriginRelationship[],
): RelationshipData | null => {
  if (!isOriginEdge(selectedEdgeId)) {
    return null;
  }

  const relationshipId = extractRelationshipId(selectedEdgeId!);
  const relationship = relationships.find((r) => r.id === relationshipId);

  if (!relationship) {
    return null;
  }

  const formattedDate = new Date(relationship.createdAt).toLocaleString();
  const typeLabel = ORIGIN_RELATIONSHIP_LABELS[relationship.relationshipType];

  return { relationship, formattedDate, typeLabel };
};

interface UnlinkFunctions {
  unlinkFromAcquired: ReturnType<typeof useUnlinkComponentFromAcquiredEntity>;
  unlinkFromVendor: ReturnType<typeof useUnlinkComponentFromVendor>;
  unlinkFromTeam: ReturnType<typeof useUnlinkComponentFromInternalTeam>;
}

const handleUnlink = async (relationship: OriginRelationship, unlinkFunctions: UnlinkFunctions): Promise<void> => {
  const { unlinkFromAcquired, unlinkFromVendor, unlinkFromTeam } = unlinkFunctions;

  switch (relationship.relationshipType) {
    case 'AcquiredVia':
      await unlinkFromAcquired.mutateAsync({
        entityId: relationship.originEntityId as AcquiredEntityId,
        componentId: relationship.componentId,
      });
      break;
    case 'PurchasedFrom':
      await unlinkFromVendor.mutateAsync({
        vendorId: relationship.originEntityId as VendorId,
        componentId: relationship.componentId,
      });
      break;
    case 'BuiltBy':
      await unlinkFromTeam.mutateAsync({
        teamId: relationship.originEntityId as InternalTeamId,
        componentId: relationship.componentId,
      });
      break;
  }
};

const TYPE_ICON_MAP: Record<OriginRelationshipType, string> = {
  AcquiredVia: '🏢',
  PurchasedFrom: '🏪',
  BuiltBy: '👥',
};

export const OriginRelationshipDetails: React.FC = () => {
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const { data: relationships = [] } = useOriginRelationshipsQuery();

  const unlinkFromAcquired = useUnlinkComponentFromAcquiredEntity();
  const unlinkFromVendor = useUnlinkComponentFromVendor();
  const unlinkFromTeam = useUnlinkComponentFromInternalTeam();

  const data = getRelationshipData(selectedEdgeId, relationships);

  if (!data) {
    return null;
  }

  const { relationship, formattedDate, typeLabel } = data;
  const canDelete = hasLink({ _links: relationship._links }, 'delete');
  const isPending = unlinkFromAcquired.isPending || unlinkFromVendor.isPending || unlinkFromTeam.isPending;
  const icon = TYPE_ICON_MAP[relationship.relationshipType];

  const onDelete = async () => {
    await handleUnlink(relationship, { unlinkFromAcquired, unlinkFromVendor, unlinkFromTeam });
  };

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Origin Relationship</h3>
      </div>

      <div className="detail-content">
        {canDelete && (
          <div className="detail-actions">
            <button className="btn btn-danger btn-small" onClick={onDelete} disabled={isPending}>
              {isPending ? 'Deleting...' : 'Delete'}
            </button>
          </div>
        )}

        <DetailField label="Relationship Type">
          <span className="origin-relationship-type">
            {icon} {typeLabel}
          </span>
        </DetailField>
        <DetailField label="Origin Entity">{relationship.originEntityName}</DetailField>
        <DetailField label="Application">{relationship.componentName}</DetailField>
        {relationship.notes && <DetailField label="Notes">{relationship.notes}</DetailField>}
        <DetailField label="Created">
          <span className="detail-date">{formattedDate}</span>
        </DetailField>
        <DetailField label="ID">
          <span className="detail-id">{relationship.id}</span>
        </DetailField>
      </div>
    </div>
  );
};
