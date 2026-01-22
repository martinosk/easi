import React from 'react';
import { OriginEntityDetailsPanel } from './OriginEntityDetailsPanel';

interface InternalTeamDetailsPanelProps {
  entityId: string;
}

export const InternalTeamDetailsPanel: React.FC<InternalTeamDetailsPanelProps> = ({
  entityId,
}) => {
  return <OriginEntityDetailsPanel entityType="team" entityId={entityId} />;
};
