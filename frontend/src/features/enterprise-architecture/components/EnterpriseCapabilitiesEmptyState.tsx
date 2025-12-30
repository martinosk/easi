import React from 'react';

interface EnterpriseCapabilitiesEmptyStateProps {
  onCreateNew: () => void;
  canWrite?: boolean;
}

export const EnterpriseCapabilitiesEmptyState = React.memo<EnterpriseCapabilitiesEmptyStateProps>(({
  onCreateNew,
  canWrite = false
}) => {
  return (
    <div className="empty-state" data-testid="empty-state">
      <div className="empty-state-icon">
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" width="48" height="48">
          <path d="M19 3H5C3.89543 3 3 3.89543 3 5V19C3 20.1046 3.89543 21 5 21H19C20.1046 21 21 20.1046 21 19V5C21 3.89543 20.1046 3 19 3Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          <path d="M3 9H21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          <path d="M9 21V9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
      </div>
      <h3 className="empty-state-title">
        {canWrite ? 'No Enterprise Capabilities Yet' : 'No Enterprise Capabilities'}
      </h3>
      <p className="empty-state-description">
        {canWrite
          ? 'Enterprise capabilities help you group related domain capabilities across your organization. Create your first enterprise capability to start organizing your architecture.'
          : 'No enterprise capabilities have been created yet.'}
      </p>
      {canWrite && (
        <button
          type="button"
          className="btn btn-primary"
          onClick={onCreateNew}
          data-testid="create-first-capability-btn"
        >
          Create Enterprise Capability
        </button>
      )}
    </div>
  );
});

EnterpriseCapabilitiesEmptyState.displayName = 'EnterpriseCapabilitiesEmptyState';
