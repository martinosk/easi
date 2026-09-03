import type React from 'react';
import { OriginEntityDetailsPanel } from './OriginEntityDetailsPanel';

interface AcquiredEntityDetailsPanelProps {
  entityId: string;
  viewMembership?: React.ReactNode;
}

export const AcquiredEntityDetailsPanel: React.FC<AcquiredEntityDetailsPanelProps> = ({ entityId, viewMembership }) => (
  <OriginEntityDetailsPanel entityType="acquired" entityId={entityId} viewMembership={viewMembership} />
);
