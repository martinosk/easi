import React from 'react';
import { DetailField } from '../../../components/shared/DetailField';
import { RealizationActions } from './RealizationActions';
import { RealizationLevelBadge } from './RealizationLevelBadge';
import { OriginBadge } from './OriginBadge';
import { InheritedRealizationInfo } from './InheritedRealizationInfo';
import type { RealizationData } from '../hooks/useRealizationDetails';

interface RealizationDetailsContentProps {
  data: RealizationData;
  onEditClick: () => void;
}

export const RealizationDetailsContent: React.FC<RealizationDetailsContentProps> = ({ data, onEditClick }) => {
  const { realization, capability, component, formattedDate, isInherited } = data;
  const canEdit = !isInherited && realization._links?.edit !== undefined;

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Realization Details</h3>
      </div>

      <div className="detail-content">
        <RealizationActions canEdit={canEdit} onEditClick={onEditClick} />
        <DetailField label="Capability">{capability?.name || 'Unknown'}</DetailField>
        <DetailField label="Application">{component?.name || 'Unknown'}</DetailField>
        <RealizationLevelBadge level={realization.realizationLevel} />
        <OriginBadge origin={realization.origin} isInherited={isInherited} />
        {realization.notes && <DetailField label="Notes">{realization.notes}</DetailField>}
        <DetailField label="Linked">
          <span className="detail-date">{formattedDate}</span>
        </DetailField>
        <DetailField label="ID">
          <span className="detail-id">{realization.id}</span>
        </DetailField>
        <InheritedRealizationInfo isInherited={isInherited} />
      </div>
    </div>
  );
};
