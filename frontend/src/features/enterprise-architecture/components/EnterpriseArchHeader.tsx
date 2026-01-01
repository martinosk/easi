import React from 'react';

interface EnterpriseArchHeaderProps {
  canWrite: boolean;
  onCreateNew: () => void;
  isDockPanelOpen: boolean;
  onToggleDockPanel: () => void;
  activeTab?: string;
  showTabActions?: boolean;
}

export const EnterpriseArchHeader = React.memo<EnterpriseArchHeaderProps>(({
  canWrite,
  onCreateNew,
  isDockPanelOpen,
  onToggleDockPanel,
  showTabActions = true,
}) => {
  return (
    <div className="enterprise-arch-header">
      <div>
        <h1 className="enterprise-arch-title">Enterprise Architecture</h1>
        <p className="enterprise-arch-subtitle">
          Manage enterprise capabilities, analyze maturity gaps, and discover unlinked domain capabilities.
        </p>
      </div>
      {showTabActions && (
        <div style={{ display: 'flex', gap: '0.75rem' }}>
          <button
            type="button"
            className={`btn ${isDockPanelOpen ? 'btn-primary' : 'btn-secondary'}`}
            onClick={onToggleDockPanel}
            data-testid="toggle-dock-panel-btn"
            aria-pressed={isDockPanelOpen}
          >
            <svg className="btn-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            {isDockPanelOpen ? 'Hide Linking Panel' : 'Link Capabilities'}
          </button>
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
      )}
    </div>
  );
});

EnterpriseArchHeader.displayName = 'EnterpriseArchHeader';
