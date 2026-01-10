import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { EnterpriseCapabilityCard } from './EnterpriseCapabilityCard';
import type { EnterpriseCapability } from '../types';
import type { Capability } from '../../../api/types';

function renderWithMantine(ui: React.ReactElement) {
  return render(<MantineProvider>{ui}</MantineProvider>);
}

describe('EnterpriseCapabilityCard', () => {
  const mockCapability: EnterpriseCapability = {
    id: 'ec-1',
    name: 'Customer Management',
    description: 'Managing customer relationships',
    category: 'Business',
    linkCount: 5,
    domainCount: 3,
    active: true,
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
    _links: {
      self: { href: '/api/v1/enterprise-capabilities/ec-1', method: 'GET' },
      'x-links': { href: '/api/v1/enterprise-capabilities/ec-1/links', method: 'GET' },
      'x-create-link': { href: '/api/v1/enterprise-capabilities/ec-1/links', method: 'POST' },
      'x-strategic-importance': { href: '/api/v1/enterprise-capabilities/ec-1/strategic-importance', method: 'GET' },
    },
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

      renderWithMantine(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Customer Management')).toBeInTheDocument();
    });

    it('renders category badge', () => {
      const onDrop = vi.fn();

      renderWithMantine(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Business')).toBeInTheDocument();
    });

    it('renders description when provided', () => {
      const onDrop = vi.fn();

      renderWithMantine(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Managing customer relationships')).toBeInTheDocument();
    });

    it('does not render description when not provided', () => {
      const capabilityWithoutDescription: EnterpriseCapability = {
        ...mockCapability,
        description: '',
      };
      const onDrop = vi.fn();

      renderWithMantine(<EnterpriseCapabilityCard capability={capabilityWithoutDescription} onDrop={onDrop} />);

      expect(screen.queryByText('Managing customer relationships')).not.toBeInTheDocument();
    });

    it('renders link count', () => {
      const onDrop = vi.fn();

      renderWithMantine(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Links:')).toBeInTheDocument();
      expect(screen.getByText('5')).toBeInTheDocument();
    });

    it('renders domain count', () => {
      const onDrop = vi.fn();

      renderWithMantine(<EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />);

      expect(screen.getByText('Domains:')).toBeInTheDocument();
      expect(screen.getByText('3')).toBeInTheDocument();
    });
  });

  describe('Drag and Drop Interactions', () => {
    it('calls onDrop with capability data when dropped', () => {
      const onDrop = vi.fn();
      renderWithMantine(
        <EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />
      );

      const card = screen.getByText('Customer Management').closest('div')?.parentElement as HTMLElement;
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

      renderWithMantine(
        <EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />
      );

      const card = screen.getByText('Customer Management').closest('div')?.parentElement as HTMLElement;
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

      renderWithMantine(
        <EnterpriseCapabilityCard capability={mockCapability} onDrop={onDrop} />
      );

      const card = screen.getByText('Customer Management').closest('div')?.parentElement as HTMLElement;
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
