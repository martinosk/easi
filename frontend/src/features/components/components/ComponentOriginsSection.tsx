import { Box, Divider, Group, Stack, Text } from '@mantine/core';
import { useQuery } from '@tanstack/react-query';
import React from 'react';
import { httpClient } from '../../../api/core/httpClient';
import type { ComponentId, HATEOASLinks, OriginRelationshipType } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { componentsQueryKeys } from '../queryKeys';

interface ComponentOriginsSectionProps {
  componentId: ComponentId;
}

interface AcquiredViaRelationshipDTO {
  id: string;
  acquiredEntityId: string;
  acquiredEntityName: string;
  componentId: string;
  componentName: string;
  notes?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

interface PurchasedFromRelationshipDTO {
  id: string;
  vendorId: string;
  vendorName: string;
  componentId: string;
  componentName: string;
  notes?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

interface BuiltByRelationshipDTO {
  id: string;
  internalTeamId: string;
  internalTeamName: string;
  componentId: string;
  componentName: string;
  notes?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

interface ComponentOriginsResponse {
  componentId: string;
  acquiredVia: AcquiredViaRelationshipDTO[];
  purchasedFrom: PurchasedFromRelationshipDTO[];
  builtBy: BuiltByRelationshipDTO[];
  _links: HATEOASLinks;
}

interface OriginItem {
  id: string;
  originEntityName: string;
  relationshipType: OriginRelationshipType;
}

const transformToOriginItems = (response: ComponentOriginsResponse): OriginItem[] => {
  const items: OriginItem[] = [];

  for (const rel of response.acquiredVia) {
    items.push({
      id: rel.id,
      originEntityName: rel.acquiredEntityName,
      relationshipType: 'AcquiredVia',
    });
  }

  for (const rel of response.purchasedFrom) {
    items.push({
      id: rel.id,
      originEntityName: rel.vendorName,
      relationshipType: 'PurchasedFrom',
    });
  }

  for (const rel of response.builtBy) {
    items.push({
      id: rel.id,
      originEntityName: rel.internalTeamName,
      relationshipType: 'BuiltBy',
    });
  }

  return items;
};

const fetchComponentOrigins = async (componentId: ComponentId): Promise<OriginItem[]> => {
  const response = await httpClient.get<ComponentOriginsResponse>(`/api/v1/components/${componentId}/origins`);
  return transformToOriginItems(response.data);
};

const getRelationshipTypeLabel = (type: OriginRelationshipType): string => {
  const labels: Record<OriginRelationshipType, string> = {
    AcquiredVia: 'Acquired via',
    PurchasedFrom: 'Purchased from',
    BuiltBy: 'Built by',
  };
  return labels[type] || type;
};

const getRelationshipTypeIcon = (type: OriginRelationshipType): string => {
  const icons: Record<OriginRelationshipType, string> = {
    AcquiredVia: '🏢',
    PurchasedFrom: '🏪',
    BuiltBy: '👥',
  };
  return icons[type] || '•';
};

const getRelationshipTypeColor = (type: OriginRelationshipType): string => {
  const colors: Record<OriginRelationshipType, string> = {
    AcquiredVia: 'violet',
    PurchasedFrom: 'pink',
    BuiltBy: 'teal',
  };
  return colors[type] || 'gray';
};

export const ComponentOriginsSection: React.FC<ComponentOriginsSectionProps> = ({ componentId }) => {
  const { data: origins = [], isLoading } = useQuery({
    queryKey: componentsQueryKeys.origins(componentId),
    queryFn: () => fetchComponentOrigins(componentId),
    enabled: !!componentId,
  });

  if (isLoading) {
    return (
      <DetailField label="Origins">
        <Text size="sm" c="dimmed">
          Loading...
        </Text>
      </DetailField>
    );
  }

  if (origins.length === 0) {
    return null;
  }

  return (
    <DetailField label="Origins">
      <Stack gap={0}>
        {origins.map((origin, index) => (
          <React.Fragment key={origin.id}>
            {index > 0 && <Divider />}
            <Group gap="sm" py="xs" wrap="nowrap">
              <Text size="md">{getRelationshipTypeIcon(origin.relationshipType)}</Text>
              <Box flex={1}>
                <Text size="sm" fw={500}>
                  {origin.originEntityName}
                </Text>
                <Text size="xs" c={getRelationshipTypeColor(origin.relationshipType)}>
                  {getRelationshipTypeLabel(origin.relationshipType)}
                </Text>
              </Box>
            </Group>
          </React.Fragment>
        ))}
      </Stack>
    </DetailField>
  );
};
