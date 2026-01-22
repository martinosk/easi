import React, { useState } from 'react';
import { useAcquiredEntity, useDeleteAcquiredEntity } from '../hooks/useAcquiredEntities';
import { useVendor, useDeleteVendor } from '../hooks/useVendors';
import { useInternalTeam, useDeleteInternalTeam } from '../hooks/useInternalTeams';
import { useOriginRelationshipsQuery } from '../hooks/useOriginRelationships';
import { AcquiredEntityDetails } from './AcquiredEntityDetails';
import { VendorDetails } from './VendorDetails';
import { InternalTeamDetails } from './InternalTeamDetails';
import { EditAcquiredEntityDialog } from './EditAcquiredEntityDialog';
import { EditVendorDialog } from './EditVendorDialog';
import { EditInternalTeamDialog } from './EditInternalTeamDialog';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { type OriginEntityType } from '../../../constants/entityIdentifiers';
import type {
  AcquiredEntityId,
  VendorId,
  InternalTeamId,
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

interface EntityConfig {
  relationshipType: OriginRelationshipType;
  deleteDialogTitle: string;
  deleteDialogMessageTemplate: (name: string) => string;
}

const ENTITY_CONFIGS: Record<OriginEntityType, EntityConfig> = {
  acquired: {
    relationshipType: 'AcquiredVia',
    deleteDialogTitle: 'Delete Acquired Entity',
    deleteDialogMessageTemplate: (name) =>
      `Are you sure you want to delete "${name}"? This will also delete all relationships to this entity.`,
  },
  vendor: {
    relationshipType: 'PurchasedFrom',
    deleteDialogTitle: 'Delete Vendor',
    deleteDialogMessageTemplate: (name) =>
      `Are you sure you want to delete "${name}"? This will also delete all relationships to this vendor.`,
  },
  team: {
    relationshipType: 'BuiltBy',
    deleteDialogTitle: 'Delete Internal Team',
    deleteDialogMessageTemplate: (name) =>
      `Are you sure you want to delete "${name}"? This will also delete all relationships to this team.`,
  },
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

interface UseOriginEntityDeleteResult {
  deleteEntity: (id: string, name: string) => Promise<void>;
  isPending: boolean;
}

function useOriginEntityDelete(entityType: OriginEntityType): UseOriginEntityDeleteResult {
  const deleteAcquired = useDeleteAcquiredEntity();
  const deleteVendor = useDeleteVendor();
  const deleteTeam = useDeleteInternalTeam();

  const deleteEntity = async (id: string, name: string) => {
    switch (entityType) {
      case 'acquired':
        await deleteAcquired.mutateAsync({ id: id as AcquiredEntityId, name });
        break;
      case 'vendor':
        await deleteVendor.mutateAsync({ id: id as VendorId, name });
        break;
      case 'team':
        await deleteTeam.mutateAsync({ id: id as InternalTeamId, name });
        break;
    }
  };

  const isPending =
    entityType === 'acquired' ? deleteAcquired.isPending :
    entityType === 'vendor' ? deleteVendor.isPending :
    deleteTeam.isPending;

  return { deleteEntity, isPending };
}

interface OriginEntityDetailsPanelProps {
  entityType: OriginEntityType;
  entityId: string;
}

export const OriginEntityDetailsPanel: React.FC<OriginEntityDetailsPanelProps> = ({
  entityType,
  entityId,
}) => {
  const config = ENTITY_CONFIGS[entityType];
  const { entity, isLoading, error } = useOriginEntity(entityType, entityId);
  const { data: allRelationships = [] } = useOriginRelationshipsQuery();
  const { deleteEntity, isPending: isDeleting } = useOriginEntityDelete(entityType);

  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

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
      rel.relationshipType === config.relationshipType && rel.originEntityId === entityId
  );

  const handleEdit = () => setShowEditDialog(true);
  const handleDelete = () => setShowDeleteConfirm(true);

  const confirmDelete = async () => {
    await deleteEntity(entityId, entity.name);
    setShowDeleteConfirm(false);
  };

  return (
    <>
      <EntityDetailsContent
        entityType={entityType}
        entity={entity}
        relationships={relationships}
        onEdit={handleEdit}
        onDelete={handleDelete}
      />

      <EntityEditDialog
        entityType={entityType}
        entity={entity}
        isOpen={showEditDialog}
        onClose={() => setShowEditDialog(false)}
      />

      {showDeleteConfirm && (
        <ConfirmationDialog
          title={config.deleteDialogTitle}
          message={config.deleteDialogMessageTemplate(entity.name)}
          confirmText="Delete"
          onConfirm={confirmDelete}
          onCancel={() => setShowDeleteConfirm(false)}
          isLoading={isDeleting}
        />
      )}
    </>
  );
};

interface EntityDetailsContentProps {
  entityType: OriginEntityType;
  entity: AcquiredEntity | Vendor | InternalTeam;
  relationships: OriginRelationship[];
  onEdit: () => void;
  onDelete: () => void;
}

const EntityDetailsContent: React.FC<EntityDetailsContentProps> = ({
  entityType,
  entity,
  relationships,
  onEdit,
  onDelete,
}) => {
  switch (entityType) {
    case 'acquired':
      return (
        <AcquiredEntityDetails
          entity={entity as AcquiredEntity}
          relationships={relationships}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      );
    case 'vendor':
      return (
        <VendorDetails
          vendor={entity as Vendor}
          relationships={relationships}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      );
    case 'team':
      return (
        <InternalTeamDetails
          team={entity as InternalTeam}
          relationships={relationships}
          onEdit={onEdit}
          onDelete={onDelete}
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
