import React from 'react';

interface PanelVisibility {
  domains: boolean;
  explorer: boolean;
  details: boolean;
}

interface DockviewToolbarProps {
  panelVisibility: PanelVisibility;
  onTogglePanel: (panelId: 'domains' | 'explorer' | 'details') => void;
}

const toolbarStyle: React.CSSProperties = {
  height: '32px',
  flexShrink: 0,
  display: 'flex',
  alignItems: 'center',
  gap: '8px',
  padding: '0 12px',
  backgroundColor: 'var(--color-gray-50)',
  borderBottom: '1px solid var(--color-gray-200)',
  fontSize: '13px',
};

const buttonStyle: React.CSSProperties = {
  padding: '4px 12px',
  border: '1px solid var(--color-gray-300)',
  borderRadius: '4px',
  backgroundColor: 'white',
  cursor: 'pointer',
  fontSize: '13px',
};

export const DockviewToolbar: React.FC<DockviewToolbarProps> = ({ panelVisibility, onTogglePanel }) => {
  return (
    <div style={toolbarStyle}>
      <span style={{ color: 'var(--color-gray-600)', fontWeight: 500 }}>View:</span>
      <button onClick={() => onTogglePanel('domains')} style={buttonStyle}>
        {panelVisibility.domains ? '☑' : '☐'} Business Domains
      </button>
      <button onClick={() => onTogglePanel('explorer')} style={buttonStyle}>
        {panelVisibility.explorer ? '☑' : '☐'} Capability Explorer
      </button>
      <button onClick={() => onTogglePanel('details')} style={buttonStyle}>
        {panelVisibility.details ? '☑' : '☐'} Details
      </button>
    </div>
  );
};
