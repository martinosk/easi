import { Badge, Stack, Text, Title } from '@mantine/core';
import React from 'react';
import type { AcquiredEntity, OriginRelationship } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { hasLink } from '../../../utils/hateoas';
import { AuditHistorySection } from '../../audit';
import { OriginEntityActions, OriginEntityRelationshipsList } from './OriginEntityPanelChrome';

interface AcquiredEntityDetailsProps {
  entity: AcquiredEntity;
  relationships: OriginRelationship[];
  canRemoveFromView: boolean;
  onEdit: () => void;
  onRemoveFromView: () => void;
}

const formatDate = (dateString: string | undefined): string => {
  if (!dateString) return 'Not set';
  try {
    return new Date(dateString).toLocaleDateString();
  } catch {
    return dateString;
  }
};

type IntegrationStatusMeta = { label: string; color: string };

const INTEGRATION_STATUS_META: Record<string, IntegrationStatusMeta> = {
  NotStarted: { label: 'Not Started', color: 'gray' },
  InProgress: { label: 'In Progress', color: 'yellow' },
  Completed: { label: 'Completed', color: 'green' },
  OnHold: { label: 'On Hold', color: 'red' },
};

const getIntegrationStatusMeta = (status: string): IntegrationStatusMeta =>
  INTEGRATION_STATUS_META[status] ?? { label: status, color: 'gray' };

export const AcquiredEntityDetails: React.FC<AcquiredEntityDetailsProps> = ({
  entity,
  relationships,
  canRemoveFromView,
  onEdit,
  onRemoveFromView,
}) => {
  const canEdit = hasLink(entity, 'edit');
  const formattedCreatedAt = new Date(entity.createdAt).toLocaleString();
  const statusMeta = getIntegrationStatusMeta(entity.integrationStatus);

  return (
    <Stack gap="sm" p="md">
      <Title order={4}>Acquired Entity Details</Title>

      <OriginEntityActions
        canEdit={canEdit}
        canRemoveFromView={canRemoveFromView}
        onEdit={onEdit}
        onRemoveFromView={onRemoveFromView}
      />

      <DetailField label="Name">{entity.name}</DetailField>

      <DetailField label="Acquisition Date">{formatDate(entity.acquisitionDate)}</DetailField>

      <DetailField label="Integration Status">
        <Badge color={statusMeta.color} variant="dot" size="sm">
          {statusMeta.label}
        </Badge>
      </DetailField>

      {entity.notes && <DetailField label="Notes">{entity.notes}</DetailField>}

      <DetailField label="Created">
        <Text size="sm" c="dimmed">
          {formattedCreatedAt}
        </Text>
      </DetailField>

      <DetailField label="Type">Acquired Entity</DetailField>

      <OriginEntityRelationshipsList relationships={relationships} relationshipLabel="Acquired via" />

      <AuditHistorySection aggregateId={entity.id} />
    </Stack>
  );
};
