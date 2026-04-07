import React from 'react';
import {
  getEntityId,
  getEntityType,
  isOriginRelationshipEdge,
  isRealizationEdge,
  isRelationEdge,
  type NodeEntityType,
  toEdgeId,
  toNodeId,
} from '../../constants/entityIdentifiers';
import { CapabilityDetails } from '../../features/capabilities';
import { ComponentDetails } from '../../features/components';
import {
  AcquiredEntityDetailsPanel,
  InternalTeamDetailsPanel,
  OriginRelationshipDetails,
  VendorDetailsPanel,
} from '../../features/origin-entities';
import { RealizationDetails, RelationDetails } from '../../features/relations';

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
  const typedEdgeId = toEdgeId(edgeId);
  if (isRealizationEdge(typedEdgeId)) {
    return <RealizationDetails />;
  }
  if (isOriginRelationshipEdge(typedEdgeId)) {
    return <OriginRelationshipDetails />;
  }
  if (isRelationEdge(typedEdgeId)) {
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
        entityType={getEntityType(toNodeId(selectedNodeId))}
        entityId={getEntityId(toNodeId(selectedNodeId))}
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
    <div style={{ color: 'var(--color-gray-500)' }}>Select a component, relation, or capability to view details</div>
  );
};
