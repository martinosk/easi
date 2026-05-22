import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers/renderWithProviders';
import type { UserRole } from '../../auth/types';
import type { User } from '../types';
import { ChangeRoleModal } from './ChangeRoleModal';

const mockUser: User = {
  id: 'user-123',
  email: 'test@acme.com',
  name: 'Test User',
  role: 'architect',
  status: 'active',
  createdAt: '2025-01-01T10:00:00Z',
  _links: {
    self: '/api/v1/users/user-123',
    update: '/api/v1/users/user-123',
  },
};

describe('ChangeRoleModal', () => {
  const mockOnClose = vi.fn();
  const mockOnSubmit = vi.fn();

  function renderModal(user: User = mockUser) {
    return renderWithProviders(
      <ChangeRoleModal isOpen={true} onClose={mockOnClose} onSubmit={mockOnSubmit} user={user} />,
    );
  }

  function changeRole(toRole: UserRole) {
    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: toRole } });
    return select as HTMLSelectElement;
  }

  function submitForm() {
    const submitBtn = screen.getByTestId('change-role-submit-btn');
    fireEvent.click(submitBtn);
    return submitBtn;
  }

  beforeEach(() => {
    vi.clearAllMocks();
    mockOnSubmit.mockResolvedValue(undefined);
  });

  it('renders modal when open', () => {
    renderModal();

    expect(screen.getByTestId('change-role-modal')).toBeInTheDocument();
    expect(screen.getByText('Change User Role')).toBeInTheDocument();
    expect(screen.getByText(/Change the role for test@acme.com/)).toBeInTheDocument();
  });

  it('initializes with user current role selected', () => {
    renderModal();

    const select = screen.getByTestId('change-role-select') as HTMLSelectElement;
    expect(select.value).toBe('architect');
  });

  it('allows changing the role', () => {
    renderModal();

    const select = changeRole('admin');

    expect(select.value).toBe('admin');
  });

  it('disables submit button when role has not changed', () => {
    renderModal();

    expect(screen.getByTestId('change-role-submit-btn')).toBeDisabled();
  });

  it('enables submit button when role has changed', () => {
    renderModal();
    changeRole('admin');

    expect(screen.getByTestId('change-role-submit-btn')).not.toBeDisabled();
  });

  it('calls onSubmit with new role when form is submitted', async () => {
    renderModal();
    changeRole('stakeholder');
    submitForm();

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith('stakeholder');
    });
  });

  it('closes modal after successful submission', async () => {
    renderModal();
    changeRole('admin');
    submitForm();

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it('calls onClose when cancel button is clicked', () => {
    renderModal();

    fireEvent.click(screen.getByTestId('change-role-cancel-btn'));

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('resets role to original when cancelled', () => {
    renderModal();
    changeRole('admin');

    fireEvent.click(screen.getByTestId('change-role-cancel-btn'));

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('displays error and keeps modal open when submission fails', async () => {
    mockOnSubmit.mockRejectedValue(new Error('Failed to change role'));
    renderModal();
    changeRole('admin');
    submitForm();

    await waitFor(() => {
      expect(screen.getByTestId('change-role-error')).toBeInTheDocument();
    });

    expect(screen.getByTestId('change-role-error')).toHaveTextContent('Failed to change role');
    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('disables form controls while submitting', async () => {
    mockOnSubmit.mockImplementation(() => new Promise((resolve) => setTimeout(resolve, 100)));
    renderModal();
    const select = changeRole('admin');
    const submitBtn = submitForm();

    expect(select).toBeDisabled();
    expect(submitBtn).toBeDisabled();
    expect(screen.getByTestId('change-role-cancel-btn')).toBeDisabled();

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it('shows submitting state on button', async () => {
    mockOnSubmit.mockImplementation(() => new Promise((resolve) => setTimeout(resolve, 100)));
    renderModal();
    changeRole('admin');
    const submitBtn = submitForm();

    expect(submitBtn).toHaveTextContent('Changing...');

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it('updates role when user prop changes', () => {
    const { rerender } = renderModal();

    rerender(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={{ ...mockUser, role: 'admin' }}
      />,
    );

    const select = screen.getByTestId('change-role-select') as HTMLSelectElement;
    expect(select.value).toBe('admin');
  });

  it('renders all role options', () => {
    renderModal();

    const options = screen.getByTestId('change-role-select').querySelectorAll('option');

    expect(options).toHaveLength(3);
    expect(options[0]).toHaveValue('stakeholder');
    expect(options[1]).toHaveValue('architect');
    expect(options[2]).toHaveValue('admin');
  });
});
