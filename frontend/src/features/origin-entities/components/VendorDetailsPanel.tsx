import React from 'react';
import { OriginEntityDetailsPanel } from './OriginEntityDetailsPanel';

interface VendorDetailsPanelProps {
  entityId: string;
}

export const VendorDetailsPanel: React.FC<VendorDetailsPanelProps> = ({
  entityId,
}) => {
  return <OriginEntityDetailsPanel entityType="vendor" entityId={entityId} />;
};
