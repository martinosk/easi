import { Stack, Text, Title } from '@mantine/core';
import React from 'react';
import { DetailField } from '../../../components/shared/DetailField';
import type { RealizationData } from '../hooks/useRealizationDetails';
import { InheritedRealizationInfo } from './InheritedRealizationInfo';
import { OriginBadge } from './OriginBadge';
import { RealizationActions } from './RealizationActions';
import { RealizationLevelBadge } from './RealizationLevelBadge';

interface RealizationDetailsContentProps {
  data: RealizationData;
  onEditClick: () => void;
}

export const RealizationDetailsContent: React.FC<RealizationDetailsContentProps> = ({ data, onEditClick }) => {
  const { realization, capability, component, formattedDate, isInherited } = data;
  const canEdit = !isInherited && realization._links?.edit !== undefined;

  return (
    <Stack gap="sm" p="md">
      <Title order={4}>Realization Details</Title>

      <RealizationActions canEdit={canEdit} onEditClick={onEditClick} />
      <DetailField label="Capability">{capability?.name || 'Unknown'}</DetailField>
      <DetailField label="Application">{component?.name || 'Unknown'}</DetailField>
      <RealizationLevelBadge level={realization.realizationLevel} />
      <OriginBadge origin={realization.origin} isInherited={isInherited} />
      {realization.notes && <DetailField label="Notes">{realization.notes}</DetailField>}
      <DetailField label="Linked">
        <Text size="sm" c="dimmed">
          {formattedDate}
        </Text>
      </DetailField>
      <DetailField label="ID">
        <Text size="xs" ff="monospace" c="gray.5" style={{ wordBreak: 'break-all' }}>
          {realization.id}
        </Text>
      </DetailField>
      <InheritedRealizationInfo isInherited={isInherited} />
    </Stack>
  );
};
