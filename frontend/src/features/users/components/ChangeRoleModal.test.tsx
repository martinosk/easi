import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ChangeRoleModal } from './ChangeRoleModal';
import type { User } from '../types';

const mockUser: User = {
  id: 'user-123',
  email: 'test@acme.com',
  name: 'Test User',
  role: 'architect',
  status: 'active',
  createdAt: '2025-01-01T10:00:00Z',
  _links: {
    self: '/api/v1/users/user-123',
    changeRole: '/api/v1/users/user-123/change-role',
  },
};

describe('ChangeRoleModal', () => {
  const mockOnClose = vi.fn();
  const mockOnSubmit = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockOnSubmit.mockResolvedValue(undefined);
  });

  it('renders modal when open', () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    expect(screen.getByTestId('change-role-modal')).toBeInTheDocument();
    expect(screen.getByText('Change User Role')).toBeInTheDocument();
    expect(screen.getByText(/Change the role for test@acme.com/)).toBeInTheDocument();
  });

  it('initializes with user current role selected', () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select') as HTMLSelectElement;
    expect(select.value).toBe('architect');
  });

  it('allows changing the role', () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: 'admin' } });

    expect((select as HTMLSelectElement).value).toBe('admin');
  });

  it('disables submit button when role has not changed', () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const submitBtn = screen.getByTestId('change-role-submit-btn');
    expect(submitBtn).toBeDisabled();
  });

  it('enables submit button when role has changed', () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: 'admin' } });

    const submitBtn = screen.getByTestId('change-role-submit-btn');
    expect(submitBtn).not.toBeDisabled();
  });

  it('calls onSubmit with new role when form is submitted', async () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: 'stakeholder' } });

    const submitBtn = screen.getByTestId('change-role-submit-btn');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith('stakeholder');
    });
  });

  it('closes modal after successful submission', async () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: 'admin' } });

    const submitBtn = screen.getByTestId('change-role-submit-btn');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it('calls onClose when cancel button is clicked', () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const cancelBtn = screen.getByTestId('change-role-cancel-btn');
    fireEvent.click(cancelBtn);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('resets role to original when cancelled', () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: 'admin' } });

    const cancelBtn = screen.getByTestId('change-role-cancel-btn');
    fireEvent.click(cancelBtn);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('displays error message when submission fails', async () => {
    mockOnSubmit.mockRejectedValue(new Error('Failed to change role'));

    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: 'admin' } });

    const submitBtn = screen.getByTestId('change-role-submit-btn');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByTestId('change-role-error')).toBeInTheDocument();
    });

    expect(screen.getByTestId('change-role-error')).toHaveTextContent('Failed to change role');
  });

  it('does not close modal when submission fails', async () => {
    mockOnSubmit.mockRejectedValue(new Error('Failed to change role'));

    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: 'admin' } });

    const submitBtn = screen.getByTestId('change-role-submit-btn');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByTestId('change-role-error')).toBeInTheDocument();
    });

    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('disables form controls while submitting', async () => {
    mockOnSubmit.mockImplementation(() => new Promise((resolve) => setTimeout(resolve, 100)));

    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: 'admin' } });

    const submitBtn = screen.getByTestId('change-role-submit-btn');
    fireEvent.click(submitBtn);

    expect(select).toBeDisabled();
    expect(submitBtn).toBeDisabled();
    expect(screen.getByTestId('change-role-cancel-btn')).toBeDisabled();

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it('shows submitting state on button', async () => {
    mockOnSubmit.mockImplementation(() => new Promise((resolve) => setTimeout(resolve, 100)));

    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    fireEvent.change(select, { target: { value: 'admin' } });

    const submitBtn = screen.getByTestId('change-role-submit-btn');
    fireEvent.click(submitBtn);

    expect(submitBtn).toHaveTextContent('Changing...');

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it('updates role when user prop changes', () => {
    const { rerender } = render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const updatedUser: User = {
      ...mockUser,
      role: 'admin',
    };

    rerender(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={updatedUser}
      />
    );

    const select = screen.getByTestId('change-role-select') as HTMLSelectElement;
    expect(select.value).toBe('admin');
  });

  it('renders all role options', () => {
    render(
      <ChangeRoleModal
        isOpen={true}
        onClose={mockOnClose}
        onSubmit={mockOnSubmit}
        user={mockUser}
      />
    );

    const select = screen.getByTestId('change-role-select');
    const options = select.querySelectorAll('option');

    expect(options).toHaveLength(3);
    expect(options[0]).toHaveValue('stakeholder');
    expect(options[1]).toHaveValue('architect');
    expect(options[2]).toHaveValue('admin');
  });
});
