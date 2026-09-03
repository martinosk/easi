import { Text } from '@mantine/core';
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
  OriginEntityViewMembershipSection,
  OriginRelationshipDetails,
  VendorDetailsPanel,
} from '../../features/origin-entities';
import { RealizationDetails, RelationDetails } from '../../features/relations';

export interface DetailContentRendererProps {
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  selectedCapabilityId: string | null;
  onRemoveFromView: () => void;
  onRemoveCapabilityFromView: () => void;
}

interface NodeDetailProps {
  entityType: NodeEntityType;
  entityId: string;
  onRemoveFromView: () => void;
  onRemoveCapabilityFromView: () => void;
}

const NodeDetail: React.FC<NodeDetailProps> = ({
  entityType,
  entityId,
  onRemoveFromView,
  onRemoveCapabilityFromView,
}) => {
  const viewMembership = <OriginEntityViewMembershipSection entityId={entityId} />;
  switch (entityType) {
    case 'acquired':
      return <AcquiredEntityDetailsPanel entityId={entityId} viewMembership={viewMembership} />;
    case 'vendor':
      return <VendorDetailsPanel entityId={entityId} viewMembership={viewMembership} />;
    case 'team':
      return <InternalTeamDetailsPanel entityId={entityId} viewMembership={viewMembership} />;
    case 'capability':
      return <CapabilityDetails onRemoveFromView={onRemoveCapabilityFromView} />;
    default:
      return <ComponentDetails onRemoveFromView={onRemoveFromView} />;
  }
};

interface EdgeDetailProps {
  edgeId: string;
}

const EdgeDetail: React.FC<EdgeDetailProps> = ({ edgeId }) => {
  const typedEdgeId = toEdgeId(edgeId);
  if (isRealizationEdge(typedEdgeId)) {
    return <RealizationDetails />;
  }
  if (isOriginRelationshipEdge(typedEdgeId)) {
    return <OriginRelationshipDetails />;
  }
  if (isRelationEdge(typedEdgeId)) {
    return <RelationDetails />;
  }
  return null;
};

export const DetailContentRenderer: React.FC<DetailContentRendererProps> = ({
  selectedNodeId,
  selectedEdgeId,
  selectedCapabilityId,
  onRemoveFromView,
  onRemoveCapabilityFromView,
}) => {
  if (selectedNodeId) {
    return (
      <NodeDetail
        entityType={getEntityType(toNodeId(selectedNodeId))}
        entityId={getEntityId(toNodeId(selectedNodeId))}
        onRemoveFromView={onRemoveFromView}
        onRemoveCapabilityFromView={onRemoveCapabilityFromView}
      />
    );
  }

  if (selectedEdgeId) {
    return <EdgeDetail edgeId={selectedEdgeId} />;
  }

  if (selectedCapabilityId) {
    return <CapabilityDetails onRemoveFromView={onRemoveCapabilityFromView} />;
  }

  return null;
};

export const DetailContentRendererWithPlaceholder: React.FC<DetailContentRendererProps> = (props) => {
  const content = DetailContentRenderer(props);
  if (content) return content;

  return <Text c="dimmed">Select a component, relation, or capability to view details</Text>;
};
