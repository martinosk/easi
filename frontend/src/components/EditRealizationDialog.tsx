import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../store/appStore';
import type { CapabilityRealization, RealizationLevel } from '../api/types';

interface EditRealizationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  realization: CapabilityRealization | null;
}

export const EditRealizationDialog: React.FC<EditRealizationDialogProps> = ({
  isOpen,
  onClose,
  realization,
}) => {
  const [realizationLevel, setRealizationLevel] = useState<RealizationLevel>('Full');
  const [notes, setNotes] = useState('');
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateRealization = useAppStore((state) => state.updateRealization);
  const capabilities = useAppStore((state) => state.capabilities);
  const components = useAppStore((state) => state.components);

  const capability = realization
    ? capabilities.find((c) => c.id === realization.capabilityId)
    : null;
  const component = realization
    ? components.find((c) => c.id === realization.componentId)
    : null;

  useEffect(() => {
    if (realization) {
      setRealizationLevel(realization.realizationLevel);
      setNotes(realization.notes || '');
    }
  }, [realization]);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  const handleClose = () => {
    setRealizationLevel('Full');
    setNotes('');
    setError(null);
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!realization) {
      setError('No realization selected');
      return;
    }

    setIsUpdating(true);

    try {
      await updateRealization(realization.id, {
        realizationLevel,
        notes: notes.trim() || undefined,
      });
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update realization');
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <dialog ref={dialogRef} className="dialog" onClose={handleClose}>
      <div className="dialog-content">
        <h2 className="dialog-title">Edit Realization</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Capability</label>
            <div className="form-static">{capability?.name || 'Unknown'}</div>
          </div>

          <div className="form-group">
            <label className="form-label">Application</label>
            <div className="form-static">{component?.name || 'Unknown'}</div>
          </div>

          <div className="form-group">
            <label htmlFor="realization-level" className="form-label">
              Realization Level
            </label>
            <select
              id="realization-level"
              className="form-select"
              value={realizationLevel}
              onChange={(e) => setRealizationLevel(e.target.value as RealizationLevel)}
              disabled={isUpdating}
            >
              <option value="Full">Full (100%)</option>
              <option value="Partial">Partial</option>
              <option value="Planned">Planned</option>
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="realization-notes" className="form-label">
              Notes
            </label>
            <textarea
              id="realization-notes"
              className="form-textarea"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Enter notes (optional)"
              rows={4}
              disabled={isUpdating}
            />
          </div>

          {error && <div className="error-message">{error}</div>}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleClose}
              disabled={isUpdating}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isUpdating}
            >
              {isUpdating ? 'Updating...' : 'Update Realization'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
};
