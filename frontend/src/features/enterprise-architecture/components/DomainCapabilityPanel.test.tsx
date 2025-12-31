import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DomainCapabilityPanel } from './DomainCapabilityPanel';
import type { Capability } from '../../../api/types';
import type { CapabilityLinkStatusResponse } from '../types';

describe('DomainCapabilityPanel', () => {
  const mockCapabilities: Capability[] = [
    {
      id: 'cap-1',
      name: 'Payment Processing',
      level: 'L1',
      status: 'active',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
      _links: { self: { href: '/api/v1/capabilities/cap-1' } },
    },
    {
      id: 'cap-2',
      name: 'Order Management',
      level: 'L1',
      status: 'active',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
      _links: { self: { href: '/api/v1/capabilities/cap-2' } },
    },
  ];

  describe('Visual State Rendering', () => {
    it('renders available capabilities without status indicators', () => {
      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        ['cap-1', { capabilityId: 'cap-1', status: 'available' }],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
        />
      );

      expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      expect(screen.queryByText(/Linked/)).not.toBeInTheDocument();
      expect(screen.queryByText(/blocked/i)).not.toBeInTheDocument();
    });

    it('renders linked capabilities with arrow indicator', () => {
      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        [
          'cap-1',
          {
            capabilityId: 'cap-1',
            status: 'linked',
            linkedTo: { id: 'ec-1', name: 'Customer Management' },
          },
        ],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
        />
      );

      expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      expect(screen.getByText('──► Customer Management')).toBeInTheDocument();
    });

    it('renders blocked_by_parent capabilities with appropriate styling', () => {
      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        [
          'cap-1',
          {
            capabilityId: 'cap-1',
            status: 'blocked_by_parent',
            blockingCapability: { id: 'cap-parent', name: 'Parent Capability' },
          },
        ],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
        />
      );

      expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      expect(screen.getByText(/Parent linked to Parent Capability/i)).toBeInTheDocument();
    });

    it('renders blocked_by_child capabilities with appropriate styling', () => {
      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        [
          'cap-1',
          {
            capabilityId: 'cap-1',
            status: 'blocked_by_child',
            blockingCapability: { id: 'cap-child', name: 'Child Capability' },
          },
        ],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
        />
      );

      expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      expect(screen.getByText(/Child linked to Child Capability/i)).toBeInTheDocument();
    });
  });

  describe('Drag and Drop', () => {
    it('makes available capabilities draggable', () => {
      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        ['cap-1', { capabilityId: 'cap-1', status: 'available' }],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
        />
      );

      const draggableElement = screen.getByText('Payment Processing').closest('div[draggable="true"]');
      expect(draggableElement).toBeInTheDocument();
      expect(draggableElement).toHaveAttribute('draggable', 'true');
    });

    it('prevents dragging for linked capabilities', () => {
      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        [
          'cap-1',
          {
            capabilityId: 'cap-1',
            status: 'linked',
            linkedTo: { id: 'ec-1', name: 'Customer Management' },
          },
        ],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
        />
      );

      const nonDraggableElement = screen.getByText('Payment Processing').closest('div[draggable="false"]');
      expect(nonDraggableElement).toBeInTheDocument();
      expect(nonDraggableElement).toHaveAttribute('draggable', 'false');
    });

    it('prevents dragging for blocked_by_parent capabilities', () => {
      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        [
          'cap-1',
          {
            capabilityId: 'cap-1',
            status: 'blocked_by_parent',
            blockingCapability: { id: 'cap-parent', name: 'Parent Capability' },
          },
        ],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
        />
      );

      const nonDraggableElement = screen.getByText('Payment Processing').closest('div[draggable="false"]');
      expect(nonDraggableElement).toBeInTheDocument();
    });

    it('calls onDragStart when dragging begins', () => {
      const onDragStart = vi.fn();
      const onDragEnd = vi.fn();
      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        ['cap-1', { capabilityId: 'cap-1', status: 'available' }],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
          onDragStart={onDragStart}
          onDragEnd={onDragEnd}
        />
      );

      const draggableElement = screen.getByText('Payment Processing').closest('div[draggable="true"]')!;
      fireEvent.dragStart(draggableElement, {
        dataTransfer: {
          setData: vi.fn(),
          effectAllowed: 'move',
        },
      });

      expect(onDragStart).toHaveBeenCalledWith(mockCapabilities[0]);
    });
  });

  describe('Data Loading', () => {
    it('shows loading message when isLoading is true', () => {
      render(
        <DomainCapabilityPanel
          capabilities={[]}
          linkStatuses={new Map()}
          isLoading={true}
        />
      );

      expect(screen.getByText('Loading domain capabilities...')).toBeInTheDocument();
    });

    it('shows empty message when no capabilities are available', () => {
      render(
        <DomainCapabilityPanel
          capabilities={[]}
          linkStatuses={new Map()}
          isLoading={false}
        />
      );

      expect(screen.getByText('No domain capabilities available')).toBeInTheDocument();
    });

    it('renders capability list when data is loaded', () => {
      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        ['cap-1', { capabilityId: 'cap-1', status: 'available' }],
        ['cap-2', { capabilityId: 'cap-2', status: 'available' }],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
        />
      );

      expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      expect(screen.getByText('Order Management')).toBeInTheDocument();
    });
  });

  describe('Hierarchy Tree Structure', () => {
    it('displays nested capabilities in tree structure', () => {
      const hierarchicalCapabilities: Capability[] = [
        {
          id: 'parent-1',
          name: 'Parent Capability',
          level: 'L1',
          status: 'active',
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/parent-1' } },
        },
        {
          id: 'child-1',
          name: 'Child Capability',
          level: 'L2',
          parentId: 'parent-1',
          status: 'active',
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/child-1' } },
        },
      ];

      const linkStatuses = new Map<string, CapabilityLinkStatusResponse>([
        ['parent-1', { capabilityId: 'parent-1', status: 'available' }],
        ['child-1', { capabilityId: 'child-1', status: 'available' }],
      ]);

      render(
        <DomainCapabilityPanel
          capabilities={hierarchicalCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
        />
      );

      expect(screen.getByText('Parent Capability')).toBeInTheDocument();
      expect(screen.getByText('Child Capability')).toBeInTheDocument();
    });
  });
});
