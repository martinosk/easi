import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { httpClient } from '../../../api/core/httpClient';
import { DetailField } from '../../../components/shared/DetailField';
import type { ComponentId, OriginRelationship, CollectionResponse, OriginRelationshipType } from '../../../api/types';

interface ComponentOriginsSectionProps {
  componentId: ComponentId;
}

interface ComponentOriginsResponse extends CollectionResponse<OriginRelationship> {}

const fetchComponentOrigins = async (componentId: ComponentId): Promise<OriginRelationship[]> => {
  const response = await httpClient.get<ComponentOriginsResponse>(
    `/api/v1/components/${componentId}/origins`
  );
  return response.data.data;
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
    queryKey: ['components', componentId, 'origins'],
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
