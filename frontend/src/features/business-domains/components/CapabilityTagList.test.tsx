import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CapabilityTagList } from './CapabilityTagList';
import type { Capability } from '../../../api/types';

describe('CapabilityTagList', () => {
  const mockCapabilities: Capability[] = [
    {
      id: 'cap-1' as any,
      name: 'Customer Onboarding',
      description: 'Process for onboarding customers',
      level: 'L1',
      createdAt: '2025-01-15T10:00:00Z',
      _links: {
        self: { href: '/api/v1/capabilities/cap-1' },
        dissociate: { href: '/api/v1/business-domains/domain-1/capabilities/cap-1' },
      },
    },
    {
      id: 'cap-2' as any,
      name: 'Customer Support',
      description: 'Support for existing customers',
      level: 'L1',
      createdAt: '2025-01-15T10:00:00Z',
      _links: {
        self: { href: '/api/v1/capabilities/cap-2' },
        dissociate: { href: '/api/v1/business-domains/domain-1/capabilities/cap-2' },
      },
    },
  ];

  it('renders empty state when no capabilities', () => {
    const onRemove = vi.fn();

    render(<CapabilityTagList capabilities={[]} onRemove={onRemove} />);

    expect(screen.getByTestId('capabilities-empty-state')).toBeInTheDocument();
    expect(screen.getByText('No capabilities associated yet.')).toBeInTheDocument();
  });

  it('renders capability tags', () => {
    const onRemove = vi.fn();

    render(<CapabilityTagList capabilities={mockCapabilities} onRemove={onRemove} />);

    expect(screen.getByTestId('capability-tag-list')).toBeInTheDocument();
    expect(screen.getByText('Customer Onboarding')).toBeInTheDocument();
    expect(screen.getByText('Customer Support')).toBeInTheDocument();
  });

  it('shows remove button for each capability with dissociate link', () => {
    const onRemove = vi.fn();

    render(<CapabilityTagList capabilities={mockCapabilities} onRemove={onRemove} />);

    expect(screen.getByTestId('capability-remove-cap-1')).toBeInTheDocument();
    expect(screen.getByTestId('capability-remove-cap-2')).toBeInTheDocument();
  });

  it('hides remove button when dissociate link is missing', () => {
    const capabilities: Capability[] = [
      {
        ...mockCapabilities[0],
        _links: {
          self: { href: '/api/v1/capabilities/cap-1' },
        },
      },
    ];
    const onRemove = vi.fn();

    render(<CapabilityTagList capabilities={capabilities} onRemove={onRemove} />);

    expect(screen.queryByTestId('capability-remove-cap-1')).not.toBeInTheDocument();
  });

  it('shows confirmation dialog when remove button is clicked', () => {
    const onRemove = vi.fn();

    render(<CapabilityTagList capabilities={mockCapabilities} onRemove={onRemove} />);

    const removeButton = screen.getByTestId('capability-remove-cap-1');
    fireEvent.click(removeButton);

    expect(screen.getByText('Remove Capability')).toBeInTheDocument();
    expect(screen.getByText('Are you sure you want to remove "Customer Onboarding" from this domain?')).toBeInTheDocument();
  });

  it('calls onRemove when confirmation is confirmed', () => {
    const onRemove = vi.fn();

    render(<CapabilityTagList capabilities={mockCapabilities} onRemove={onRemove} />);

    const removeButton = screen.getByTestId('capability-remove-cap-1');
    fireEvent.click(removeButton);

    const confirmButton = screen.getByText('Remove');
    fireEvent.click(confirmButton);

    expect(onRemove).toHaveBeenCalledWith(mockCapabilities[0]);
  });

  it('does not call onRemove when confirmation is cancelled', () => {
    const onRemove = vi.fn();

    render(<CapabilityTagList capabilities={mockCapabilities} onRemove={onRemove} />);

    const removeButton = screen.getByTestId('capability-remove-cap-1');
    fireEvent.click(removeButton);

    const cancelButton = screen.getByText('Cancel');
    fireEvent.click(cancelButton);

    expect(onRemove).not.toHaveBeenCalled();
  });

  it('closes confirmation dialog after confirming', () => {
    const onRemove = vi.fn();

    render(<CapabilityTagList capabilities={mockCapabilities} onRemove={onRemove} />);

    const removeButton = screen.getByTestId('capability-remove-cap-1');
    fireEvent.click(removeButton);

    expect(screen.getByText('Remove Capability')).toBeInTheDocument();

    const confirmButton = screen.getByText('Remove');
    fireEvent.click(confirmButton);

    expect(screen.queryByText('Remove Capability')).not.toBeInTheDocument();
  });
});
