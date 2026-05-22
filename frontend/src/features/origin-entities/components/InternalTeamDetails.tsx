import { Stack, Text, Title } from '@mantine/core';
import React from 'react';
import type { InternalTeam, OriginRelationship } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { hasLink } from '../../../utils/hateoas';
import { AuditHistorySection } from '../../audit';
import { OriginEntityActions, OriginEntityRelationshipsList } from './OriginEntityPanelChrome';

interface InternalTeamDetailsProps {
  team: InternalTeam;
  relationships: OriginRelationship[];
  canRemoveFromView: boolean;
  onEdit: () => void;
  onRemoveFromView: () => void;
}

export const InternalTeamDetails: React.FC<InternalTeamDetailsProps> = ({
  team,
  relationships,
  canRemoveFromView,
  onEdit,
  onRemoveFromView,
}) => {
  const canEdit = hasLink(team, 'edit');
  const formattedCreatedAt = new Date(team.createdAt).toLocaleString();

  return (
    <Stack gap="sm" p="md">
      <Title order={4}>Internal Team Details</Title>

      <OriginEntityActions
        canEdit={canEdit}
        canRemoveFromView={canRemoveFromView}
        onEdit={onEdit}
        onRemoveFromView={onRemoveFromView}
      />

      <DetailField label="Name">{team.name}</DetailField>

      {team.department && <DetailField label="Department">{team.department}</DetailField>}

      {team.contactPerson && <DetailField label="Contact Person">{team.contactPerson}</DetailField>}

      {team.notes && <DetailField label="Notes">{team.notes}</DetailField>}

      <DetailField label="Created">
        <Text size="sm" c="dimmed">
          {formattedCreatedAt}
        </Text>
      </DetailField>

      <DetailField label="Type">Internal Team</DetailField>

      <OriginEntityRelationshipsList relationships={relationships} relationshipLabel="Built by" />

      <AuditHistorySection aggregateId={team.id} />
    </Stack>
  );
};
