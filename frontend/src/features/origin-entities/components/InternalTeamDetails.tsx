import React from 'react';
import { DetailField } from '../../../components/shared/DetailField';
import { AuditHistorySection } from '../../audit';
import { hasLink } from '../../../utils/hateoas';
import type { InternalTeam, OriginRelationship } from '../../../api/types';

interface InternalTeamDetailsProps {
  team: InternalTeam;
  relationships: OriginRelationship[];
  onEdit: () => void;
  onDelete: () => void;
}

export const InternalTeamDetails: React.FC<InternalTeamDetailsProps> = ({
  team,
  relationships,
  onEdit,
  onDelete,
}) => {
  const canEdit = hasLink(team, 'edit');
  const canDelete = hasLink(team, 'delete');
  const formattedCreatedAt = new Date(team.createdAt).toLocaleString();

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Internal Team Details</h3>
      </div>

      <div className="detail-content">
        <div className="detail-actions">
          {canEdit && (
            <button className="btn btn-secondary btn-small" onClick={onEdit}>
              Edit
            </button>
          )}
          {canDelete && (
            <button className="btn btn-danger btn-small" onClick={onDelete}>
              Delete
            </button>
          )}
        </div>

        <DetailField label="Name">{team.name}</DetailField>

        {team.department && (
          <DetailField label="Department">{team.department}</DetailField>
        )}

        {team.contactPerson && (
          <DetailField label="Contact Person">{team.contactPerson}</DetailField>
        )}

        {team.notes && <DetailField label="Notes">{team.notes}</DetailField>}

        <DetailField label="Created">
          <span className="detail-date">{formattedCreatedAt}</span>
        </DetailField>

        <DetailField label="Type">Internal Team</DetailField>

        {relationships.length > 0 && (
          <DetailField label={`Applications (${relationships.length})`}>
            <ul className="realization-list">
              {relationships.map((rel) => (
                <li key={rel.id} className="realization-item">
                  <span className="realization-name">{rel.componentName}</span>
                  <span className="realization-level">Built by</span>
                </li>
              ))}
            </ul>
          </DetailField>
        )}

        <AuditHistorySection aggregateId={team.id} />
      </div>
    </div>
  );
};
