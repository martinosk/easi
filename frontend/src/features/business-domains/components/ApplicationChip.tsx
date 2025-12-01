import type { CapabilityRealization, ComponentId } from '../../../api/types';

export interface ApplicationChipProps {
  realization: CapabilityRealization;
  onClick: (componentId: ComponentId) => void;
}

const REALIZATION_LEVEL_STYLES = {
  Full: {
    border: '2px solid #22c55e',
  },
  Partial: {
    border: '2px dashed #eab308',
  },
  Planned: {
    border: '2px dotted #94a3b8',
  },
} as const;

const ORIGIN_STYLES = {
  Direct: {
    backgroundColor: '#e2e8f0',
  },
  Inherited: {
    backgroundColor: '#f1f5f9',
  },
} as const;

export function ApplicationChip({ realization, onClick }: ApplicationChipProps) {
  const componentName = realization.componentName || realization.componentId;
  const isInherited = realization.origin === 'Inherited';

  const tooltipText = isInherited && realization.sourceCapabilityName
    ? `${componentName} (inherited from ${realization.sourceCapabilityName})`
    : componentName;

  return (
    <button
      type="button"
      onClick={() => onClick(realization.componentId)}
      title={tooltipText}
      style={{
        ...REALIZATION_LEVEL_STYLES[realization.realizationLevel],
        ...ORIGIN_STYLES[realization.origin],
        padding: '0.25rem 0.5rem',
        borderRadius: '0.375rem',
        fontSize: '0.75rem',
        fontWeight: 500,
        cursor: 'pointer',
        display: 'inline-flex',
        alignItems: 'center',
        gap: '0.25rem',
        maxWidth: '150px',
        whiteSpace: 'nowrap',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        transition: 'transform 0.1s ease, box-shadow 0.1s ease',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.transform = 'scale(1.05)';
        e.currentTarget.style.boxShadow = '0 2px 4px rgba(0, 0, 0, 0.1)';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.transform = 'scale(1)';
        e.currentTarget.style.boxShadow = 'none';
      }}
    >
      {isInherited && (
        <span
          style={{
            fontSize: '0.625rem',
            opacity: 0.7,
          }}
        >
          ↓
        </span>
      )}
      <span
        style={{
          overflow: 'hidden',
          textOverflow: 'ellipsis',
        }}
      >
        {componentName}
      </span>
    </button>
  );
}
