import { useState, useEffect, useRef, type FormEvent } from 'react';
import type { UserRole } from '../../auth/types';
import type { User } from '../types';

interface ChangeRoleModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (newRole: UserRole) => Promise<void>;
  user: User;
}

export function ChangeRoleModal({ isOpen, onClose, onSubmit, user }: ChangeRoleModalProps) {
  const [role, setRole] = useState<UserRole>(user.role);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    setRole(user.role);
  }, [user]);

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
      await onSubmit(role);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to change role');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setRole(user.role);
    setError(null);
    onClose();
  };

  return (
    <dialog ref={dialogRef} className="dialog" onClose={handleCancel} data-testid="change-role-modal">
      <div className="dialog-content">
        <h2 className="dialog-title">Change User Role</h2>
        <p className="dialog-subtitle">Change the role for {user.email}</p>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label" htmlFor="change-role">
              Role <span className="required">*</span>
            </label>
            <select
              id="change-role"
              className="form-select"
              value={role}
              onChange={(e) => setRole(e.target.value as UserRole)}
              required
              disabled={isSubmitting}
              data-testid="change-role-select"
            >
              <option value="stakeholder">Stakeholder</option>
              <option value="architect">Architect</option>
              <option value="admin">Admin</option>
            </select>
          </div>

          {error && (
            <div className="error-message" data-testid="change-role-error">
              {error}
            </div>
          )}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleCancel}
              disabled={isSubmitting}
              data-testid="change-role-cancel-btn"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isSubmitting || role === user.role}
              data-testid="change-role-submit-btn"
            >
              {isSubmitting ? 'Changing...' : 'Change Role'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
}
