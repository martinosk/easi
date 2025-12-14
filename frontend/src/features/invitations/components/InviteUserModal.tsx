import { useState, useEffect, useRef, type FormEvent } from 'react';
import type { UserRole } from '../../auth/types';
import type { CreateInvitationRequest } from '../types';

interface InviteUserModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (request: CreateInvitationRequest) => Promise<void>;
}

export function InviteUserModal({ isOpen, onClose, onSubmit }: InviteUserModalProps) {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<UserRole>('stakeholder');
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
      await onSubmit({ email, role });
      setEmail('');
      setRole('stakeholder');
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create invitation');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setEmail('');
    setRole('stakeholder');
    setError(null);
    onClose();
  };

  return (
    <dialog ref={dialogRef} className="dialog" onClose={handleCancel} data-testid="invite-user-modal">
      <div className="dialog-content">
        <h2 className="dialog-title">Invite User</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label" htmlFor="invite-email">
              Email <span className="required">*</span>
            </label>
            <input
              id="invite-email"
              type="email"
              className="form-input"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              disabled={isSubmitting}
              placeholder="user@company.com"
              data-testid="invite-email-input"
            />
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="invite-role">
              Role <span className="required">*</span>
            </label>
            <select
              id="invite-role"
              className="form-select"
              value={role}
              onChange={(e) => setRole(e.target.value as UserRole)}
              required
              disabled={isSubmitting}
              data-testid="invite-role-select"
            >
              <option value="stakeholder">Stakeholder</option>
              <option value="architect">Architect</option>
              <option value="admin">Admin</option>
            </select>
          </div>

          {error && (
            <div className="error-message" data-testid="invite-error-message">
              {error}
            </div>
          )}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleCancel}
              disabled={isSubmitting}
              data-testid="invite-cancel-btn"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isSubmitting}
              data-testid="invite-submit-btn"
            >
              {isSubmitting ? 'Creating...' : 'Create Invitation'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
}
