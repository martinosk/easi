import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DomainCard } from './DomainCard';
import type { BusinessDomain } from '../../../api/types';

describe('DomainCard', () => {
  const mockDomain: BusinessDomain = {
    id: 'domain-1' as any,
    name: 'Customer Experience',
    description: 'All customer-facing capabilities',
    capabilityCount: 3,
    createdAt: '2025-01-15T10:00:00Z',
    updatedAt: '2025-01-16T12:00:00Z',
    _links: {
      self: { href: '/api/v1/business-domains/domain-1' },
      update: { href: '/api/v1/business-domains/domain-1' },
      capabilities: { href: '/api/v1/business-domains/domain-1/capabilities' },
    },
  };

  it('renders domain information correctly', () => {
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    const onView = vi.fn();

    render(<DomainCard domain={mockDomain} onEdit={onEdit} onDelete={onDelete} onView={onView} />);

    expect(screen.getByText('Customer Experience')).toBeInTheDocument();
    expect(screen.getByText('All customer-facing capabilities')).toBeInTheDocument();
    expect(screen.getByText('3 capabilities')).toBeInTheDocument();
  });

  it('shows singular capability text when count is 1', () => {
    const domain: BusinessDomain = { ...mockDomain, capabilityCount: 1 };
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    const onView = vi.fn();

    render(<DomainCard domain={domain} onEdit={onEdit} onDelete={onDelete} onView={onView} />);

    expect(screen.getByText('1 capability')).toBeInTheDocument();
  });

  it('shows Manage button that calls onView', () => {
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    const onView = vi.fn();

    render(<DomainCard domain={mockDomain} onEdit={onEdit} onDelete={onDelete} onView={onView} />);

    const manageButton = screen.getByTestId('domain-view-domain-1');
    fireEvent.click(manageButton);

    expect(onView).toHaveBeenCalledWith(mockDomain);
  });

  it('shows Edit button when update link is present', () => {
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    const onView = vi.fn();

    render(<DomainCard domain={mockDomain} onEdit={onEdit} onDelete={onDelete} onView={onView} />);

    const editButton = screen.getByTestId('domain-edit-domain-1');
    fireEvent.click(editButton);

    expect(onEdit).toHaveBeenCalledWith(mockDomain);
  });

  it('hides Delete button when domain has capabilities', () => {
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    const onView = vi.fn();

    render(<DomainCard domain={mockDomain} onEdit={onEdit} onDelete={onDelete} onView={onView} />);

    expect(screen.queryByTestId('domain-delete-domain-1')).not.toBeInTheDocument();
  });

  it('shows Delete button when domain has no capabilities and delete link exists', () => {
    const domain: BusinessDomain = {
      ...mockDomain,
      capabilityCount: 0,
      _links: {
        ...mockDomain._links,
        delete: { href: '/api/v1/business-domains/domain-1' },
      },
    };
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    const onView = vi.fn();

    render(<DomainCard domain={domain} onEdit={onEdit} onDelete={onDelete} onView={onView} />);

    const deleteButton = screen.getByTestId('domain-delete-domain-1');
    expect(deleteButton).toBeInTheDocument();

    fireEvent.click(deleteButton);
    expect(onDelete).toHaveBeenCalledWith(domain);
  });

  it('displays "No description" when description is empty', () => {
    const domain: BusinessDomain = { ...mockDomain, description: '' };
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    const onView = vi.fn();

    render(<DomainCard domain={domain} onEdit={onEdit} onDelete={onDelete} onView={onView} />);

    expect(screen.getByText('No description')).toBeInTheDocument();
  });

  it('hides Edit button when update link is not present', () => {
    const domain: BusinessDomain = {
      ...mockDomain,
      _links: {
        self: { href: '/api/v1/business-domains/domain-1' },
      },
    };
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    const onView = vi.fn();

    render(<DomainCard domain={domain} onEdit={onEdit} onDelete={onDelete} onView={onView} />);

    expect(screen.queryByTestId('domain-edit-domain-1')).not.toBeInTheDocument();
  });
});
