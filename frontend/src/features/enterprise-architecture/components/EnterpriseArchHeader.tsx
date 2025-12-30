import React from 'react';

interface EnterpriseArchHeaderProps {
  canWrite: boolean;
  onCreateNew: () => void;
}

export const EnterpriseArchHeader = React.memo<EnterpriseArchHeaderProps>(({ canWrite, onCreateNew }) => {
  return (
    <div className="enterprise-arch-header">
      <div>
        <h1 className="enterprise-arch-title">Enterprise Capabilities</h1>
        <p className="enterprise-arch-subtitle">
          Group related domain capabilities across your organization into canonical enterprise-level capabilities.
        </p>
      </div>
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
  );
});

EnterpriseArchHeader.displayName = 'EnterpriseArchHeader';
