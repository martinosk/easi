import { Badge, Divider, Group, Stack, Text } from '@mantine/core';
import React from 'react';
import type { OriginRelationship } from '../../../api/types';

interface OriginEntityRelationshipsListProps {
  relationships: OriginRelationship[];
  relationshipLabel: string;
}

export const OriginEntityRelationshipsList: React.FC<OriginEntityRelationshipsListProps> = ({
  relationships,
  relationshipLabel,
}) => {
  if (relationships.length === 0) {
    return (
      <Text size="sm" c="dimmed" fs="italic">
        no related application
      </Text>
    );
  }

  return (
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
  );
};
