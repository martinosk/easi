import type { Capability } from '../../../api/types';

interface CapabilityDetailSidebarProps {
  capability: Capability | null;
  onClose: () => void;
}

export function CapabilityDetailSidebar({ capability, onClose }: CapabilityDetailSidebarProps) {
  if (!capability) return null;

  return (
    <aside style={{ width: '300px', borderLeft: '1px solid #e5e7eb', padding: '1rem', overflow: 'auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h3>Capability Details</h3>
        <button
          type="button"
          onClick={onClose}
          style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.5rem' }}
        >
          &times;
        </button>
      </div>
      <div style={{ marginTop: '1rem' }}>
        <p><strong>Name:</strong> {capability.name}</p>
        <p><strong>Level:</strong> {capability.level}</p>
        {capability.description && (
          <p><strong>Description:</strong> {capability.description}</p>
        )}
      </div>
    </aside>
  );
}
