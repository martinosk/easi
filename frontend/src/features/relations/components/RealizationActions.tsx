import React from 'react';

interface RealizationActionsProps {
  canEdit: boolean;
  onEditClick: () => void;
}

export const RealizationActions: React.FC<RealizationActionsProps> = ({ canEdit, onEditClick }) => {
  if (!canEdit) return null;

  return (
    <div className="detail-actions">
      <button className="btn btn-secondary btn-small" onClick={onEditClick}>
        Edit
      </button>
    </div>
  );
};
