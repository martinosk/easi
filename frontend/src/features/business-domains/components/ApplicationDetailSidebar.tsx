import type { Component, ComponentId } from '../../../api/types';
import { useComponentDetails } from '../hooks/useComponentDetails';

interface ApplicationDetailSidebarProps {
  componentId: ComponentId | null;
  onClose: () => void;
}

export function ApplicationDetailSidebar({ componentId, onClose }: ApplicationDetailSidebarProps) {
  const { component, isLoading, error } = useComponentDetails(componentId);

  if (!componentId) return null;

  return (
    <aside
      style={{
        width: '300px',
        borderLeft: '1px solid #e5e7eb',
        padding: '1rem',
        overflow: 'auto',
        backgroundColor: 'white',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h3>Application Details</h3>
        <button
          type="button"
          onClick={onClose}
          style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.5rem' }}
        >
          &times;
        </button>
      </div>

      {isLoading && <div style={{ marginTop: '1rem', color: '#6b7280' }}>Loading...</div>}

      {error && (
        <div style={{ marginTop: '1rem', color: '#dc2626' }}>Failed to load application details</div>
      )}

      {component && <ApplicationContent component={component} />}
    </aside>
  );
}

function ApplicationContent({ component }: { component: Component }) {
  return (
    <div style={{ marginTop: '1rem' }}>
      <p>
        <strong>Name:</strong> {component.name}
      </p>
      {component.description && (
        <p>
          <strong>Description:</strong> {component.description}
        </p>
      )}
      <p>
        <strong>Created:</strong> {new Date(component.createdAt).toLocaleDateString()}
      </p>
      {component._links.reference && (
        <p style={{ marginTop: '1rem' }}>
          <a
            href={component._links.reference}
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: '#3b82f6' }}
          >
            Reference Documentation
          </a>
        </p>
      )}
    </div>
  );
}
