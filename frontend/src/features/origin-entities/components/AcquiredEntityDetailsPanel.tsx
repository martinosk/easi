import React from 'react';
import { OriginEntityDetailsPanel } from './OriginEntityDetailsPanel';

interface AcquiredEntityDetailsPanelProps {
  entityId: string;
}

export const AcquiredEntityDetailsPanel: React.FC<AcquiredEntityDetailsPanelProps> = ({
  entityId,
}) => {
  return <OriginEntityDetailsPanel entityType="acquired" entityId={entityId} />;
};
