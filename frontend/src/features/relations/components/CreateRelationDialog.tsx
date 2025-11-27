import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../../../store/appStore';

interface CreateRelationDialogProps {
  isOpen: boolean;
  onClose: () => void;
  sourceComponentId?: string;
  targetComponentId?: string;
}

export const CreateRelationDialog: React.FC<CreateRelationDialogProps> = ({
  isOpen,
  onClose,
  sourceComponentId: initialSource,
  targetComponentId: initialTarget,
}) => {
  const [sourceComponentId, setSourceComponentId] = useState(initialSource || '');
  const [targetComponentId, setTargetComponentId] = useState(initialTarget || '');
  const [relationType, setRelationType] = useState<'Triggers' | 'Serves'>('Triggers');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const components = useAppStore((state) => state.components);
  const createRelation = useAppStore((state) => state.createRelation);

  useEffect(() => {
    if (initialSource) setSourceComponentId(initialSource);
    if (initialTarget) setTargetComponentId(initialTarget);
  }, [initialSource, initialTarget]);

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
    setSourceComponentId('');
    setTargetComponentId('');
    setRelationType('Triggers');
    setName('');
    setDescription('');
    setError(null);
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!sourceComponentId || !targetComponentId) {
      setError('Both source and target components are required');
      return;
    }

    if (sourceComponentId === targetComponentId) {
      setError('Source and target components must be different');
      return;
    }

    setIsCreating(true);

    try {
      await createRelation({
        sourceComponentId: sourceComponentId as import('../../../api/types').ComponentId,
        targetComponentId: targetComponentId as import('../../../api/types').ComponentId,
        relationType,
        name: name.trim() || undefined,
        description: description.trim() || undefined,
      });
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create relation');
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <dialog ref={dialogRef} className="dialog" onClose={handleClose} data-testid="create-relation-dialog">
      <div className="dialog-content">
        <h2 className="dialog-title">Create Relation</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="source-component" className="form-label">
              Source Component <span className="required">*</span>
            </label>
            <select
              id="source-component"
              className="form-select"
              value={sourceComponentId}
              onChange={(e) => setSourceComponentId(e.target.value)}
              disabled={isCreating || !!initialSource}
              data-testid="relation-source-select"
            >
              <option value="">Select source component</option>
              {components.map((component) => (
                <option key={component.id} value={component.id}>
                  {component.name}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="target-component" className="form-label">
              Target Component <span className="required">*</span>
            </label>
            <select
              id="target-component"
              className="form-select"
              value={targetComponentId}
              onChange={(e) => setTargetComponentId(e.target.value)}
              disabled={isCreating || !!initialTarget}
              data-testid="relation-target-select"
            >
              <option value="">Select target component</option>
              {components.map((component) => (
                <option key={component.id} value={component.id}>
                  {component.name}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="relation-type" className="form-label">
              Relation Type <span className="required">*</span>
            </label>
            <select
              id="relation-type"
              className="form-select"
              value={relationType}
              onChange={(e) => setRelationType(e.target.value as 'Triggers' | 'Serves')}
              disabled={isCreating}
              data-testid="relation-type-select"
            >
              <option value="Triggers">Triggers</option>
              <option value="Serves">Serves</option>
            </select>
          </div>

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
              disabled={isCreating}
              data-testid="relation-name-input"
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
              rows={3}
              disabled={isCreating}
              data-testid="relation-description-input"
            />
          </div>

          {error && <div className="error-message" data-testid="create-relation-error">{error}</div>}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleClose}
              disabled={isCreating}
              data-testid="create-relation-cancel"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isCreating || !sourceComponentId || !targetComponentId}
              data-testid="create-relation-submit"
            >
              {isCreating ? 'Creating...' : 'Create Relation'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
};
