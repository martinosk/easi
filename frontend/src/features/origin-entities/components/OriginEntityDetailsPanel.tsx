import React, { useState } from 'react';
import { useAcquiredEntity } from '../hooks/useAcquiredEntities';
import { useVendor } from '../hooks/useVendors';
import { useInternalTeam } from '../hooks/useInternalTeams';
import { useOriginRelationshipsQuery } from '../hooks/useOriginRelationships';
import { useRemoveOriginEntityFromView } from '../../views/hooks/useViews';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { AcquiredEntityDetails } from './AcquiredEntityDetails';
import { VendorDetails } from './VendorDetails';
import { InternalTeamDetails } from './InternalTeamDetails';
import { EditAcquiredEntityDialog } from './EditAcquiredEntityDialog';
import { EditVendorDialog } from './EditVendorDialog';
import { EditInternalTeamDialog } from './EditInternalTeamDialog';
import { type OriginEntityType } from '../../../constants/entityIdentifiers';
import type {
  AcquiredEntity,
  Vendor,
  InternalTeam,
  OriginRelationship,
  OriginRelationshipType,
} from '../../../api/types';

type EntityMap = {
  acquired: AcquiredEntity;
  vendor: Vendor;
  team: InternalTeam;
};

const RELATIONSHIP_TYPES: Record<OriginEntityType, OriginRelationshipType> = {
  acquired: 'AcquiredVia',
  vendor: 'PurchasedFrom',
  team: 'BuiltBy',
};

interface UseOriginEntityResult<T extends OriginEntityType> {
  entity: EntityMap[T] | undefined;
  isLoading: boolean;
  error: Error | null;
}

function useOriginEntity<T extends OriginEntityType>(
  entityType: T,
  entityId: string
): UseOriginEntityResult<T> {
  const acquiredQuery = useAcquiredEntity(
    entityType === 'acquired' ? (entityId as AcquiredEntityId) : undefined
  );
  const vendorQuery = useVendor(
    entityType === 'vendor' ? (entityId as VendorId) : undefined
  );
  const teamQuery = useInternalTeam(
    entityType === 'team' ? (entityId as InternalTeamId) : undefined
  );

  if (entityType === 'acquired') {
    return {
      entity: acquiredQuery.data as EntityMap[T] | undefined,
      isLoading: acquiredQuery.isLoading,
      error: acquiredQuery.error,
    };
  }
  if (entityType === 'vendor') {
    return {
      entity: vendorQuery.data as EntityMap[T] | undefined,
      isLoading: vendorQuery.isLoading,
      error: vendorQuery.error,
    };
  }
  return {
    entity: teamQuery.data as EntityMap[T] | undefined,
    isLoading: teamQuery.isLoading,
    error: teamQuery.error,
  };
}

interface OriginEntityDetailsPanelProps {
  entityType: OriginEntityType;
  entityId: string;
}

export const OriginEntityDetailsPanel: React.FC<OriginEntityDetailsPanelProps> = ({
  entityType,
  entityId,
}) => {
  const { entity, isLoading, error } = useOriginEntity(entityType, entityId);
  const { data: allRelationships = [] } = useOriginRelationshipsQuery();
  const { currentView } = useCurrentView();
  const removeFromViewMutation = useRemoveOriginEntityFromView();

  const [showEditDialog, setShowEditDialog] = useState(false);

  if (isLoading) {
    return (
      <div className="detail-panel">
        <div className="detail-loading">Loading...</div>
      </div>
    );
  }

  if (error || !entity) {
    return (
      <div className="detail-panel">
        <div className="detail-error">Failed to load entity</div>
      </div>
    );
  }

  const relationships = allRelationships.filter(
    (rel) =>
      rel.relationshipType === RELATIONSHIP_TYPES[entityType] && rel.originEntityId === entityId
  );

  const entityInView = currentView?.originEntities.find(
    (oe) => oe.originEntityId === entityId
  );
  const canRemoveFromView = entityInView?._links?.['x-remove'] !== undefined;

  const handleEdit = () => setShowEditDialog(true);
  const handleRemoveFromView = () => {
    if (currentView) {
      removeFromViewMutation.mutate({ viewId: currentView.id, originEntityId: entityId });
    }
  };

  return (
    <>
      <EntityDetailsContent
        entityType={entityType}
        entity={entity}
        relationships={relationships}
        canRemoveFromView={canRemoveFromView}
        onEdit={handleEdit}
        onRemoveFromView={handleRemoveFromView}
      />

      <EntityEditDialog
        entityType={entityType}
        entity={entity}
        isOpen={showEditDialog}
        onClose={() => setShowEditDialog(false)}
      />
    </>
  );
};

interface EntityDetailsContentProps {
  entityType: OriginEntityType;
  entity: AcquiredEntity | Vendor | InternalTeam;
  relationships: OriginRelationship[];
  canRemoveFromView: boolean;
  onEdit: () => void;
  onRemoveFromView: () => void;
}

const EntityDetailsContent: React.FC<EntityDetailsContentProps> = ({
  entityType,
  entity,
  relationships,
  canRemoveFromView,
  onEdit,
  onRemoveFromView,
}) => {
  switch (entityType) {
    case 'acquired':
      return (
        <AcquiredEntityDetails
          entity={entity as AcquiredEntity}
          relationships={relationships}
          canRemoveFromView={canRemoveFromView}
          onEdit={onEdit}
          onRemoveFromView={onRemoveFromView}
        />
      );
    case 'vendor':
      return (
        <VendorDetails
          vendor={entity as Vendor}
          relationships={relationships}
          canRemoveFromView={canRemoveFromView}
          onEdit={onEdit}
          onRemoveFromView={onRemoveFromView}
        />
      );
    case 'team':
      return (
        <InternalTeamDetails
          team={entity as InternalTeam}
          relationships={relationships}
          canRemoveFromView={canRemoveFromView}
          onEdit={onEdit}
          onRemoveFromView={onRemoveFromView}
        />
      );
  }
};

interface EntityEditDialogProps {
  entityType: OriginEntityType;
  entity: AcquiredEntity | Vendor | InternalTeam;
  isOpen: boolean;
  onClose: () => void;
}

const EntityEditDialog: React.FC<EntityEditDialogProps> = ({
  entityType,
  entity,
  isOpen,
  onClose,
}) => {
  switch (entityType) {
    case 'acquired':
      return (
        <EditAcquiredEntityDialog
          isOpen={isOpen}
          onClose={onClose}
          entity={entity as AcquiredEntity}
        />
      );
    case 'vendor':
      return (
        <EditVendorDialog isOpen={isOpen} onClose={onClose} vendor={entity as Vendor} />
      );
    case 'team':
      return (
        <EditInternalTeamDialog
          isOpen={isOpen}
          onClose={onClose}
          team={entity as InternalTeam}
        />
      );
  }
};
