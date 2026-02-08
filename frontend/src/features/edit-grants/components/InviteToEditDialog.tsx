import { useState, useEffect, useRef, type FormEvent } from 'react';
import type { CreateEditGrantRequest, ArtifactType } from '../types';

interface InviteToEditDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (request: CreateEditGrantRequest) => Promise<void>;
  artifactType: ArtifactType;
  artifactId: string;
}

export function InviteToEditDialog({
  isOpen,
  onClose,
  onSubmit,
  artifactType,
  artifactId,
}: InviteToEditDialogProps) {
  const [granteeEmail, setGranteeEmail] = useState('');
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      await onSubmit({
        granteeEmail,
        artifactType,
        artifactId,
        reason: reason || undefined,
      });
      setGranteeEmail('');
      setReason('');
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to grant edit access');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setGranteeEmail('');
    setReason('');
    setError(null);
    onClose();
  };

  return (
    <dialog ref={dialogRef} className="dialog" onClose={handleCancel} data-testid="invite-to-edit-dialog">
      <div className="dialog-content">
        <h2 className="dialog-title">Invite to Edit</h2>
        <p className="dialog-description">
          Grant temporary edit access for this {artifactType} to a stakeholder.
        </p>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label" htmlFor="grantee-email">
              User Email <span className="required">*</span>
            </label>
            <input
              id="grantee-email"
              type="email"
              className="form-input"
              value={granteeEmail}
              onChange={(e) => setGranteeEmail(e.target.value)}
              required
              disabled={isSubmitting}
              placeholder="stakeholder@company.com"
              data-testid="grantee-email-input"
            />
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="grant-reason">
              Reason
            </label>
            <input
              id="grant-reason"
              type="text"
              className="form-input"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              disabled={isSubmitting}
              placeholder="Optional reason for granting access"
              data-testid="grant-reason-input"
            />
          </div>

          {error && (
            <div className="error-message" data-testid="grant-error-message">
              {error}
            </div>
          )}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleCancel}
              disabled={isSubmitting}
              data-testid="grant-cancel-btn"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isSubmitting}
              data-testid="grant-submit-btn"
            >
              {isSubmitting ? 'Granting...' : 'Grant Edit Access'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
}
