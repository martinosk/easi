import React from 'react';
import { ComponentDetails } from '../../features/components';
import { RelationDetails, RealizationDetails } from '../../features/relations';
import { CapabilityDetails } from '../../features/capabilities';
import {
  AcquiredEntityDetailsPanel,
  VendorDetailsPanel,
  InternalTeamDetailsPanel,
  OriginRelationshipDetails,
} from '../../features/origin-entities';
import {
  getEntityType,
  getEntityId,
  isRealizationEdge,
  isRelationEdge,
  isOriginRelationshipEdge,
  type NodeEntityType,
} from '../../constants/entityIdentifiers';

export interface DetailContentRendererProps {
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  selectedCapabilityId: string | null;
  onEditComponent: (componentId?: string) => void;
  onEditRelation: () => void;
  onRemoveFromView: () => void;
  onRemoveCapabilityFromView: () => void;
}

interface NodeDetailProps {
  entityType: NodeEntityType;
  entityId: string;
  onEditComponent: (componentId?: string) => void;
  onRemoveFromView: () => void;
  onRemoveCapabilityFromView: () => void;
}

const NodeDetail: React.FC<NodeDetailProps> = ({
  entityType,
  entityId,
  onEditComponent,
  onRemoveFromView,
  onRemoveCapabilityFromView,
}) => {
  switch (entityType) {
    case 'acquired':
      return <AcquiredEntityDetailsPanel entityId={entityId} />;
    case 'vendor':
      return <VendorDetailsPanel entityId={entityId} />;
    case 'team':
      return <InternalTeamDetailsPanel entityId={entityId} />;
    case 'capability':
      return <CapabilityDetails onRemoveFromView={onRemoveCapabilityFromView} />;
    default:
      return <ComponentDetails onEdit={onEditComponent} onRemoveFromView={onRemoveFromView} />;
  }
};

interface EdgeDetailProps {
  edgeId: string;
  onEditRelation: () => void;
}

const EdgeDetail: React.FC<EdgeDetailProps> = ({ edgeId, onEditRelation }) => {
  if (isRealizationEdge(edgeId)) {
    return <RealizationDetails />;
  }
  if (isOriginRelationshipEdge(edgeId)) {
    return <OriginRelationshipDetails />;
  }
  if (isRelationEdge(edgeId)) {
    return <RelationDetails onEdit={onEditRelation} />;
  }
  return null;
};

export const DetailContentRenderer: React.FC<DetailContentRendererProps> = ({
  selectedNodeId,
  selectedEdgeId,
  selectedCapabilityId,
  onEditComponent,
  onEditRelation,
  onRemoveFromView,
  onRemoveCapabilityFromView,
}) => {
  if (selectedNodeId) {
    return (
      <NodeDetail
        entityType={getEntityType(selectedNodeId)}
        entityId={getEntityId(selectedNodeId)}
        onEditComponent={onEditComponent}
        onRemoveFromView={onRemoveFromView}
        onRemoveCapabilityFromView={onRemoveCapabilityFromView}
      />
    );
  }

  if (selectedEdgeId) {
    return <EdgeDetail edgeId={selectedEdgeId} onEditRelation={onEditRelation} />;
  }

  if (selectedCapabilityId) {
    return <CapabilityDetails onRemoveFromView={onRemoveCapabilityFromView} />;
  }

  return null;
};

export const DetailContentRendererWithPlaceholder: React.FC<DetailContentRendererProps> = (props) => {
  const content = DetailContentRenderer(props);
  if (content) return content;

  return (
    <div style={{ color: 'var(--color-gray-500)' }}>
      Select a component, relation, or capability to view details
    </div>
  );
};
