import React from 'react';

interface EnterpriseArchHeaderProps {
  canWrite: boolean;
  onCreateNew: () => void;
  onManageLinks?: () => void;
}

export const EnterpriseArchHeader = React.memo<EnterpriseArchHeaderProps>(({ canWrite, onCreateNew, onManageLinks }) => {
  return (
    <div className="enterprise-arch-header">
      <div>
        <h1 className="enterprise-arch-title">Enterprise Capabilities</h1>
        <p className="enterprise-arch-subtitle">
          Group related domain capabilities across your organization into canonical enterprise-level capabilities.
        </p>
      </div>
      <div style={{ display: 'flex', gap: '0.75rem' }}>
        {onManageLinks && (
          <button
            type="button"
            className="btn btn-secondary"
            onClick={onManageLinks}
            data-testid="manage-links-btn"
          >
            <svg className="btn-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Manage Links
          </button>
        )}
        {canWrite && (
          <button
            type="button"
            className="btn btn-primary"
            onClick={onCreateNew}
            data-testid="create-capability-btn"
          >
            <svg className="btn-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Create Capability
          </button>
        )}
      </div>
    </div>
  );
});

EnterpriseArchHeader.displayName = 'EnterpriseArchHeader';
