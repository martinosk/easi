interface StageFormOverlayProps {
  isEditing: boolean;
  formData: { name: string; description: string };
  onFormDataChange: (data: { name: string; description: string }) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

export function StageFormOverlay({ isEditing, formData, onFormDataChange, onSubmit, onCancel }: StageFormOverlayProps) {
  return (
    <div className="vs-form-overlay" data-testid="stage-form">
      <div className="vs-form">
        <h3>{isEditing ? 'Edit Stage' : 'Add Stage'}</h3>
        <div className="vs-form-field">
          <label htmlFor="stage-name">Name</label>
          <input
            id="stage-name"
            type="text"
            value={formData.name}
            onChange={(e) => onFormDataChange({ ...formData, name: e.target.value })}
            placeholder="e.g. Discovery"
            maxLength={100}
            autoFocus
          />
        </div>
        <div className="vs-form-field">
          <label htmlFor="stage-description">Description</label>
          <textarea
            id="stage-description"
            value={formData.description}
            onChange={(e) => onFormDataChange({ ...formData, description: e.target.value })}
            placeholder="Optional description..."
            maxLength={500}
            rows={3}
          />
        </div>
        <div className="vs-form-actions">
          <button type="button" className="btn btn-secondary" onClick={onCancel}>
            Cancel
          </button>
          <button
            type="button"
            className="btn btn-primary"
            onClick={onSubmit}
            disabled={!formData.name.trim()}
          >
            {isEditing ? 'Save' : 'Add'}
          </button>
        </div>
      </div>
    </div>
  );
}
