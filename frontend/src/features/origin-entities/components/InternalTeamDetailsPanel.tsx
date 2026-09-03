import type React from 'react';
import { OriginEntityDetailsPanel } from './OriginEntityDetailsPanel';

interface InternalTeamDetailsPanelProps {
  entityId: string;
  viewMembership?: React.ReactNode;
}

export const InternalTeamDetailsPanel: React.FC<InternalTeamDetailsPanelProps> = ({ entityId, viewMembership }) => (
  <OriginEntityDetailsPanel entityType="team" entityId={entityId} viewMembership={viewMembership} />
);
