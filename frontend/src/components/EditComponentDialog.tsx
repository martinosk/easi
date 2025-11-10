import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../store/appStore';
import type { Component } from '../api/types';

interface EditComponentDialogProps {
  isOpen: boolean;
  onClose: () => void;
  component: Component | null;
}

export const EditComponentDialog: React.FC<EditComponentDialogProps> = ({
  isOpen,
  onClose,
  component,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateComponent = useAppStore((state) => state.updateComponent);

  useEffect(() => {
    if (component) {
      setName(component.name);
      setDescription(component.description || '');
    }
  }, [component]);

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

    if (!name.trim()) {
      setError('Component name is required');
      return;
    }

    if (!component) {
      setError('No component selected');
      return;
    }

    setIsUpdating(true);

    try {
      await updateComponent(component.id, name.trim(), description.trim() || undefined);
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update component');
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <dialog ref={dialogRef} className="dialog" onClose={handleClose}>
      <div className="dialog-content">
        <h2 className="dialog-title">Edit Component</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="component-name" className="form-label">
              Name <span className="required">*</span>
            </label>
            <input
              id="component-name"
              type="text"
              className="form-input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter component name"
              disabled={isUpdating}
              autoFocus
            />
          </div>

          <div className="form-group">
            <label htmlFor="component-description" className="form-label">
              Description
            </label>
            <textarea
              id="component-description"
              className="form-textarea"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter component description (optional)"
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
              disabled={isUpdating || !name.trim()}
            >
              {isUpdating ? 'Updating...' : 'Update Component'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
};
