import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EnterpriseCapabilityCard } from './EnterpriseCapabilityCard';
import type { EnterpriseCapability } from '../types';
import type { Capability } from '../../../api/types';

describe('EnterpriseCapabilityCard', () => {
  const mockCapability: EnterpriseCapability = {
    id: 'ec-1',
    name: 'Customer Management',
    description: 'Managing customer relationships',
    category: 'Business',
    linkCount: 5,
    domainCount: 3,
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
    _links: { self: { href: '/api/v1/enterprise-capabilities/ec-1' } },
  };

  const mockDomainCapability: Capability = {
    id: 'cap-1',
    name: 'Payment Processing',
    level: 'L1',
    status: 'active',
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
    _links: { self: { href: '/api/v1/capabilities/cap-1' } },
  };

  describe('Visual Rendering', () => {
    it('renders capability name', () => {
      const onDrop = vi.fn();

      render(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Customer Management')).toBeInTheDocument();
    });

    it('renders category badge', () => {
      const onDrop = vi.fn();

      render(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Business')).toBeInTheDocument();
    });

    it('renders description when provided', () => {
      const onDrop = vi.fn();

      render(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Managing customer relationships')).toBeInTheDocument();
    });

    it('does not render description when not provided', () => {
      const capabilityWithoutDescription: EnterpriseCapability = {
        ...mockCapability,
        description: undefined,
      };
      const onDrop = vi.fn();

      render(<EnterpriseCapabilityCard capability={capabilityWithoutDescription} onDrop={onDrop} />);

      expect(screen.queryByText('Managing customer relationships')).not.toBeInTheDocument();
    });

    it('renders link count', () => {
      const onDrop = vi.fn();

      render(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Links:')).toBeInTheDocument();
      expect(screen.getByText('5')).toBeInTheDocument();
    });

    it('renders domain count', () => {
      const onDrop = vi.fn();

      render(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Domains:')).toBeInTheDocument();
      expect(screen.getByText('3')).toBeInTheDocument();
    });
  });

  describe('Drag and Drop Interactions', () => {
    it('calls onDrop with capability data when dropped', () => {
      const onDrop = vi.fn();
      const { container } = render(
        <EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />
      );

      const card = container.firstChild as HTMLElement;
      fireEvent.drop(card, {
        dataTransfer: {
          getData: () => JSON.stringify(mockDomainCapability),
        },
      });

      expect(onDrop).toHaveBeenCalledWith(mockDomainCapability);
    });

    it('does not call onDrop when invalid JSON is dropped', () => {
      const onDrop = vi.fn();
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      const { container } = render(
        <EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />
      );

      const card = container.firstChild as HTMLElement;
      fireEvent.drop(card, {
        dataTransfer: {
          getData: () => 'invalid json',
        },
      });

      expect(onDrop).not.toHaveBeenCalled();

      consoleErrorSpy.mockRestore();
    });

    it('does not call onDrop when empty data is dropped', () => {
      const onDrop = vi.fn();
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      const { container } = render(
        <EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />
      );

      const card = container.firstChild as HTMLElement;
      fireEvent.drop(card, {
        dataTransfer: {
          getData: () => '',
        },
      });

      expect(onDrop).not.toHaveBeenCalled();

      consoleErrorSpy.mockRestore();
    });
  });
});
