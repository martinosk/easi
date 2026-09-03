import { Badge, Divider, Group, Stack, Text } from '@mantine/core';
import React from 'react';
import type { OriginRelationship } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';

interface OriginEntityRelationshipsListProps {
  relationships: OriginRelationship[];
  relationshipLabel: string;
}

export const OriginEntityRelationshipsList: React.FC<OriginEntityRelationshipsListProps> = ({
  relationships,
  relationshipLabel,
}) => {
  if (relationships.length === 0) return null;

  return (
    <DetailField label={`Applications (${relationships.length})`}>
      <Stack gap={0}>
        {relationships.map((rel, index) => (
          <React.Fragment key={rel.id}>
            {index > 0 && <Divider />}
            <Group justify="space-between" py="sm" wrap="nowrap">
              <Text size="sm">{rel.componentName}</Text>
              <Badge color="green" variant="filled" size="sm">
                {relationshipLabel}
              </Badge>
            </Group>
          </React.Fragment>
        ))}
      </Stack>
    </DetailField>
  );
};
