import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ReassignConfirmDialog } from './ReassignConfirmDialog';
import type { Capability } from '../../../api/types';

describe('ReassignConfirmDialog', () => {
  const createCapability = (id: string, name: string, level: 'L1' | 'L2' | 'L3' | 'L4'): Capability => ({
    id: id as any,
    name,
    level,
    createdAt: '2024-01-01',
    _links: { self: { href: `/api/v1/capabilities/${id}` } },
  });

  const childCapability = createCapability('cap-child', 'Accounting', 'L2');
  const newParentCapability = createCapability('cap-parent', 'Finance', 'L1');

  it('should render when open', () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();

    render(
      <ReassignConfirmDialog
        isOpen={true}
        capability={childCapability}
        newParent={newParentCapability}
        onConfirm={onConfirm}
        onCancel={onCancel}
      />
    );

    expect(screen.getByText('Reassign Capability')).toBeInTheDocument();
    expect(screen.getByText(/Accounting/)).toBeInTheDocument();
    expect(screen.getByText(/Finance/)).toBeInTheDocument();
  });

  it('should not render when closed', () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();

    render(
      <ReassignConfirmDialog
        isOpen={false}
        capability={childCapability}
        newParent={newParentCapability}
        onConfirm={onConfirm}
        onCancel={onCancel}
      />
    );

    expect(screen.queryByText('Reassign Capability')).not.toBeInTheDocument();
  });

  it('should call onConfirm when confirm button clicked', () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();

    render(
      <ReassignConfirmDialog
        isOpen={true}
        capability={childCapability}
        newParent={newParentCapability}
        onConfirm={onConfirm}
        onCancel={onCancel}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: /confirm/i }));

    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  it('should call onCancel when cancel button clicked', () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();

    render(
      <ReassignConfirmDialog
        isOpen={true}
        capability={childCapability}
        newParent={newParentCapability}
        onConfirm={onConfirm}
        onCancel={onCancel}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));

    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it('should show loading state when isLoading is true', () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();

    render(
      <ReassignConfirmDialog
        isOpen={true}
        capability={childCapability}
        newParent={newParentCapability}
        onConfirm={onConfirm}
        onCancel={onCancel}
        isLoading={true}
      />
    );

    expect(screen.getByRole('button', { name: /moving/i })).toBeDisabled();
  });
});
