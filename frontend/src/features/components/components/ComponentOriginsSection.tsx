import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { httpClient } from '../../../api/core/httpClient';
import { DetailField } from '../../../components/shared/DetailField';
import { queryKeys } from '../../../lib/queryClient';
import type { ComponentId, OriginRelationshipType, HATEOASLinks } from '../../../api/types';

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
  const response = await httpClient.get<ComponentOriginsResponse>(
    `/api/v1/components/${componentId}/origins`
  );
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
    AcquiredVia: '#8b5cf6',
    PurchasedFrom: '#ec4899',
    BuiltBy: '#14b8a6',
  };
  return colors[type] || '#6b7280';
};

export const ComponentOriginsSection: React.FC<ComponentOriginsSectionProps> = ({
  componentId,
}) => {
  const { data: origins = [], isLoading } = useQuery({
    queryKey: queryKeys.components.origins(componentId),
    queryFn: () => fetchComponentOrigins(componentId),
    enabled: !!componentId,
  });

  if (isLoading) {
    return (
      <DetailField label="Origins">
        <span className="detail-loading">Loading...</span>
      </DetailField>
    );
  }

  if (origins.length === 0) {
    return null;
  }

  return (
    <DetailField label="Origins">
      <ul className="realization-list" style={{ listStyle: 'none', padding: 0, margin: 0 }}>
        {origins.map((origin) => (
          <li
            key={origin.id}
            className="realization-item"
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '8px 0',
              borderBottom: '1px solid #e5e7eb',
            }}
          >
            <span style={{ fontSize: '16px' }}>
              {getRelationshipTypeIcon(origin.relationshipType)}
            </span>
            <div style={{ flex: 1 }}>
              <div style={{ fontWeight: 500 }}>{origin.originEntityName}</div>
              <div
                style={{
                  fontSize: '12px',
                  color: getRelationshipTypeColor(origin.relationshipType),
                }}
              >
                {getRelationshipTypeLabel(origin.relationshipType)}
              </div>
            </div>
          </li>
        ))}
      </ul>
    </DetailField>
  );
};
