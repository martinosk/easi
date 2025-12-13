import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { DomainFilter } from './DomainFilter';
import type { BusinessDomain } from '../../../api/types';

describe('DomainFilter', () => {
  const mockDomains: BusinessDomain[] = [
    {
      id: 'domain-1' as any,
      name: 'Customer Management',
      description: 'Handles customer data',
      capabilityCount: 5,
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/business-domains/domain-1' } },
    },
    {
      id: 'domain-2' as any,
      name: 'Order Processing',
      description: 'Manages orders',
      capabilityCount: 3,
      createdAt: '2024-01-02',
      _links: { self: { href: '/api/v1/business-domains/domain-2' } },
    },
  ];

  it('should render all domains', () => {
    const onSelect = vi.fn();
    render(<DomainFilter domains={mockDomains} selected={null} onSelect={onSelect} />);

    expect(screen.getByText('Customer Management')).toBeInTheDocument();
    expect(screen.getByText('Order Processing')).toBeInTheDocument();
  });

  it('should render orphaned capabilities option', () => {
    const onSelect = vi.fn();
    render(<DomainFilter domains={mockDomains} selected={null} onSelect={onSelect} />);

    expect(screen.getByText('Orphaned Capabilities')).toBeInTheDocument();
  });

  it('should show capability counts', () => {
    const onSelect = vi.fn();
    render(<DomainFilter domains={mockDomains} selected={null} onSelect={onSelect} />);

    expect(screen.getByText('5 capabilities')).toBeInTheDocument();
    expect(screen.getByText('3 capabilities')).toBeInTheDocument();
  });

  it('should highlight selected domain', () => {
    const onSelect = vi.fn();
    render(<DomainFilter domains={mockDomains} selected={'domain-1' as any} onSelect={onSelect} />);

    const selectedButton = screen.getByText('Customer Management').closest('button');
    expect(selectedButton).toHaveClass('bg-blue-50', 'border-l-4', 'border-l-blue-600');
  });

  it('should highlight orphaned option when selected', () => {
    const onSelect = vi.fn();
    render(<DomainFilter domains={mockDomains} selected={null} onSelect={onSelect} />);

    const orphanedButton = screen.getByText('Orphaned Capabilities').closest('button');
    expect(orphanedButton).toHaveClass('bg-blue-50', 'border-l-4', 'border-l-blue-600');
  });

  it('should call onSelect when domain clicked', () => {
    const onSelect = vi.fn();
    render(<DomainFilter domains={mockDomains} selected={null} onSelect={onSelect} />);

    const domainButton = screen.getByText('Customer Management').closest('button');
    domainButton?.click();

    expect(onSelect).toHaveBeenCalledWith('domain-1');
  });

  it('should call onSelect with null when orphaned option clicked', () => {
    const onSelect = vi.fn();
    render(<DomainFilter domains={mockDomains} selected={'domain-1' as any} onSelect={onSelect} />);

    const orphanedButton = screen.getByText('Orphaned Capabilities').closest('button');
    orphanedButton?.click();

    expect(onSelect).toHaveBeenCalledWith(null);
  });

  it('should filter domains by search term', async () => {
    const onSelect = vi.fn();
    render(<DomainFilter domains={mockDomains} selected={null} onSelect={onSelect} />);

    const searchInput = screen.getByPlaceholderText('Search domains...');
    fireEvent.change(searchInput, { target: { value: 'Customer' } });

    await waitFor(() => {
      expect(screen.getByText('Customer Management')).toBeInTheDocument();
    });
  });
});
