import type { Capability } from '../../../api/types';

export interface ReassignConfirmDialogProps {
  isOpen: boolean;
  capability: Capability | null;
  newParent: Capability | null;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading?: boolean;
}

export function ReassignConfirmDialog({
  isOpen,
  capability,
  newParent,
  onConfirm,
  onCancel,
  isLoading = false,
}: ReassignConfirmDialogProps) {
  if (!isOpen || !capability || !newParent) {
    return null;
  }

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1000,
      }}
      onClick={onCancel}
    >
      <div
        style={{
          backgroundColor: 'white',
          borderRadius: '0.5rem',
          padding: '1.5rem',
          maxWidth: '400px',
          width: '90%',
          boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1)',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <h2 style={{ margin: 0, marginBottom: '1rem', fontSize: '1.25rem' }}>
          Reassign Capability
        </h2>

        <p style={{ color: '#4b5563', marginBottom: '1.5rem' }}>
          Are you sure you want to move <strong>{capability.name}</strong> under{' '}
          <strong>{newParent.name}</strong>?
        </p>

        <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'flex-end' }}>
          <button
            type="button"
            onClick={onCancel}
            disabled={isLoading}
            style={{
              padding: '0.5rem 1rem',
              borderRadius: '0.375rem',
              border: '1px solid #d1d5db',
              backgroundColor: 'white',
              cursor: isLoading ? 'not-allowed' : 'pointer',
              fontWeight: 500,
            }}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isLoading}
            style={{
              padding: '0.5rem 1rem',
              borderRadius: '0.375rem',
              border: 'none',
              backgroundColor: '#3b82f6',
              color: 'white',
              cursor: isLoading ? 'not-allowed' : 'pointer',
              fontWeight: 500,
              opacity: isLoading ? 0.7 : 1,
            }}
          >
            {isLoading ? 'Moving...' : 'Confirm'}
          </button>
        </div>
      </div>
    </div>
  );
}
