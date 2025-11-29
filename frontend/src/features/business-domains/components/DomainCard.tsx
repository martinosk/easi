import type { BusinessDomain } from '../../../api/types';

interface DomainCardProps {
  domain: BusinessDomain;
  onVisualize: (domain: BusinessDomain) => void;
  onContextMenu: (e: React.MouseEvent, domain: BusinessDomain) => void;
  isSelected?: boolean;
}

export function DomainCard({ domain, onVisualize, onContextMenu, isSelected }: DomainCardProps) {
  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onContextMenu(e, domain);
  };

  return (
    <button
      type="button"
      className={`domain-card-button${isSelected ? ' domain-card-button--selected' : ''}`}
      onClick={() => onVisualize(domain)}
      onContextMenu={handleContextMenu}
      data-testid={`domain-card-${domain.id}`}
      style={{
        display: 'block',
        width: '100%',
        textAlign: 'left',
        padding: '0.75rem',
        marginBottom: '0.5rem',
        border: isSelected ? '2px solid #3b82f6' : '1px solid #e5e7eb',
        borderRadius: '0.5rem',
        backgroundColor: isSelected ? '#eff6ff' : '#fff',
        cursor: 'pointer',
        transition: 'all 0.15s ease',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.25rem' }}>
        <h3 style={{ margin: 0, fontSize: '1rem', fontWeight: 600, color: '#111827' }}>{domain.name}</h3>
        <span style={{ fontSize: '0.75rem', color: '#6b7280', whiteSpace: 'nowrap', marginLeft: '0.5rem' }}>
          {domain.capabilityCount} {domain.capabilityCount === 1 ? 'capability' : 'capabilities'}
        </span>
      </div>

      <p style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', color: '#6b7280' }}>
        {domain.description || 'No description'}
      </p>

      <span style={{ fontSize: '0.75rem', color: '#9ca3af' }}>
        Created: {new Date(domain.createdAt).toLocaleDateString()}
      </span>
    </button>
  );
}
