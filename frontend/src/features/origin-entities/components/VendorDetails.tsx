import { Stack, Text, Title } from '@mantine/core';
import React from 'react';
import type { OriginRelationship, Vendor } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { hasLink } from '../../../utils/hateoas';
import { AuditHistorySection } from '../../audit';
import { OriginEntityActions, OriginEntityRelationshipsList } from './OriginEntityPanelChrome';

interface VendorDetailsProps {
  vendor: Vendor;
  relationships: OriginRelationship[];
  canRemoveFromView: boolean;
  onEdit: () => void;
  onRemoveFromView: () => void;
}

export const VendorDetails: React.FC<VendorDetailsProps> = ({
  vendor,
  relationships,
  canRemoveFromView,
  onEdit,
  onRemoveFromView,
}) => {
  const canEdit = hasLink(vendor, 'edit');
  const formattedCreatedAt = new Date(vendor.createdAt).toLocaleString();

  return (
    <Stack gap="sm" p="md">
      <Title order={4}>Vendor Details</Title>

      <OriginEntityActions
        canEdit={canEdit}
        canRemoveFromView={canRemoveFromView}
        onEdit={onEdit}
        onRemoveFromView={onRemoveFromView}
      />

      <DetailField label="Name">{vendor.name}</DetailField>

      {vendor.implementationPartner && (
        <DetailField label="Implementation Partner">{vendor.implementationPartner}</DetailField>
      )}

      {vendor.notes && <DetailField label="Notes">{vendor.notes}</DetailField>}

      <DetailField label="Created">
        <Text size="sm" c="dimmed">
          {formattedCreatedAt}
        </Text>
      </DetailField>

      <DetailField label="Type">Vendor</DetailField>

      <OriginEntityRelationshipsList relationships={relationships} relationshipLabel="Purchased from" />

      <AuditHistorySection aggregateId={vendor.id} />
    </Stack>
  );
};
