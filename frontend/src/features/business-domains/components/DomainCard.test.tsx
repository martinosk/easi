import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
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

  describe('Capability Count Pluralization', () => {
    it('shows plural "capabilities" when count is greater than 1', () => {
      const onVisualize = vi.fn();
      const onContextMenu = vi.fn();

      render(<DomainCard domain={mockDomain} onVisualize={onVisualize} onContextMenu={onContextMenu} />);

      expect(screen.getByText('3 capabilities')).toBeInTheDocument();
    });

    it('shows singular "capability" when count is 1', () => {
      const domain: BusinessDomain = { ...mockDomain, capabilityCount: 1 };
      const onVisualize = vi.fn();
      const onContextMenu = vi.fn();

      render(<DomainCard domain={domain} onVisualize={onVisualize} onContextMenu={onContextMenu} />);

      expect(screen.getByText('1 capability')).toBeInTheDocument();
    });

    it('shows plural "capabilities" when count is 0', () => {
      const domain: BusinessDomain = { ...mockDomain, capabilityCount: 0 };
      const onVisualize = vi.fn();
      const onContextMenu = vi.fn();

      render(<DomainCard domain={domain} onVisualize={onVisualize} onContextMenu={onContextMenu} />);

      expect(screen.getByText('0 capabilities')).toBeInTheDocument();
    });
  });
});
