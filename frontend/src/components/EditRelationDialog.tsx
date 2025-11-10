import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../store/appStore';
import type { Relation } from '../api/types';

interface EditRelationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  relation: Relation | null;
}

export const EditRelationDialog: React.FC<EditRelationDialogProps> = ({
  isOpen,
  onClose,
  relation,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateRelation = useAppStore((state) => state.updateRelation);

  useEffect(() => {
    if (relation) {
      setName(relation.name || '');
      setDescription(relation.description || '');
    }
  }, [relation]);

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
    setName('');
    setDescription('');
    setError(null);
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!relation) {
      setError('No relation selected');
      return;
    }

    setIsUpdating(true);

    try {
      await updateRelation(
        relation.id,
        name.trim() || undefined,
        description.trim() || undefined
      );
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update relation');
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <dialog ref={dialogRef} className="dialog" onClose={handleClose}>
      <div className="dialog-content">
        <h2 className="dialog-title">Edit Relation</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="relation-name" className="form-label">
              Name
            </label>
            <input
              id="relation-name"
              type="text"
              className="form-input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter relation name (optional)"
              disabled={isUpdating}
              autoFocus
            />
          </div>

          <div className="form-group">
            <label htmlFor="relation-description" className="form-label">
              Description
            </label>
            <textarea
              id="relation-description"
              className="form-textarea"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter relation description (optional)"
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
              {isUpdating ? 'Updating...' : 'Update Relation'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
};
