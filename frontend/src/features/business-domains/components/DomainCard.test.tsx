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
    const onVisualize = vi.fn();
    const onContextMenu = vi.fn();

    render(<DomainCard domain={mockDomain} onVisualize={onVisualize} onContextMenu={onContextMenu} />);

    expect(screen.getByText('Customer Experience')).toBeInTheDocument();
    expect(screen.getByText('All customer-facing capabilities')).toBeInTheDocument();
    expect(screen.getByText('3 capabilities')).toBeInTheDocument();
  });

  it('shows singular capability text when count is 1', () => {
    const domain: BusinessDomain = { ...mockDomain, capabilityCount: 1 };
    const onVisualize = vi.fn();
    const onContextMenu = vi.fn();

    render(<DomainCard domain={domain} onVisualize={onVisualize} onContextMenu={onContextMenu} />);

    expect(screen.getByText('1 capability')).toBeInTheDocument();
  });

  it('calls onVisualize when card is clicked', () => {
    const onVisualize = vi.fn();
    const onContextMenu = vi.fn();

    render(<DomainCard domain={mockDomain} onVisualize={onVisualize} onContextMenu={onContextMenu} />);

    const card = screen.getByTestId('domain-card-domain-1');
    fireEvent.click(card);

    expect(onVisualize).toHaveBeenCalledWith(mockDomain);
  });

  it('calls onContextMenu when right-clicked', () => {
    const onVisualize = vi.fn();
    const onContextMenu = vi.fn();

    render(<DomainCard domain={mockDomain} onVisualize={onVisualize} onContextMenu={onContextMenu} />);

    const card = screen.getByTestId('domain-card-domain-1');
    fireEvent.contextMenu(card);

    expect(onContextMenu).toHaveBeenCalled();
    expect(onVisualize).not.toHaveBeenCalled();
  });

  it('displays "No description" when description is empty', () => {
    const domain: BusinessDomain = { ...mockDomain, description: '' };
    const onVisualize = vi.fn();
    const onContextMenu = vi.fn();

    render(<DomainCard domain={domain} onVisualize={onVisualize} onContextMenu={onContextMenu} />);

    expect(screen.getByText('No description')).toBeInTheDocument();
  });

  it('shows selected state when isSelected is true', () => {
    const onVisualize = vi.fn();
    const onContextMenu = vi.fn();

    render(<DomainCard domain={mockDomain} onVisualize={onVisualize} onContextMenu={onContextMenu} isSelected={true} />);

    const card = screen.getByTestId('domain-card-domain-1');
    expect(card).toHaveStyle({ backgroundColor: '#eff6ff' });
  });
});
