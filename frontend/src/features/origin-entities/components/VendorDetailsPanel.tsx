import type React from 'react';
import { OriginEntityDetailsPanel } from './OriginEntityDetailsPanel';

interface VendorDetailsPanelProps {
  entityId: string;
  viewMembership?: React.ReactNode;
}

export const VendorDetailsPanel: React.FC<VendorDetailsPanelProps> = ({ entityId, viewMembership }) => (
  <OriginEntityDetailsPanel entityType="vendor" entityId={entityId} viewMembership={viewMembership} />
);
